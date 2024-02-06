/*
   Copyright 2023 bgq98

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package validator

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"github.com/bgq98/utils/logger"
	"github.com/bgq98/utils/migrator"
	"github.com/bgq98/utils/migrator/events"
	"github.com/bgq98/utils/slice"
)

type Validator[T migrator.Entity] struct {
	// 校验 以 XXX 为准
	base          *gorm.DB
	target        *gorm.DB
	l             logger.Logger
	producer      events.Producer
	direction     string
	batchSize     int
	utime         int64
	sleepInterval time.Duration
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, l logger.Logger,
	producer events.Producer, direction string) *Validator[T] {
	return &Validator[T]{
		base:          base,
		target:        target,
		l:             l,
		producer:      producer,
		direction:     direction,
		batchSize:     100,
		sleepInterval: 0,
	}
}

func (v *Validator[T]) Utime(utime int64) *Validator[T] {
	v.utime = utime
	return v
}

func (v *Validator[T]) SleepInterval(i time.Duration) *Validator[T] {
	v.sleepInterval = i
	return v
}

// Validate 调用者可以通过 ctx 来控制校验程序退出
func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		return v.baseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.targetToBase(ctx)
	})
	return eg.Wait()
}

func (v *Validator[T]) baseToTarget(ctx context.Context) error {
	offset := 0
	for {
		var ts []T
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		err := v.base.WithContext(dbCtx).
			Where("utime >= ?", v.utime).
			Order("utime asc,id asc").
			Offset(offset).Limit(v.batchSize).
			Find(&ts).Error
		cancel()
		switch err {
		case gorm.ErrRecordNotFound:
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		case context.DeadlineExceeded, context.Canceled:
			return nil
		case nil:
			v.targetMissingRecords(ctx, ts)
		default:
			v.l.Error("base => target 查询源表失败", logger.Error(err))
			time.Sleep(time.Second)
		}
		if len(ts) < v.batchSize {
			// 没数据了 退出循环
			return nil
		}
		offset += v.batchSize
	}
}

func (v *Validator[T]) targetToBase(ctx context.Context) error {
	// 先找 target 再找 base 中已经删除的
	offset := 0
	for {
		var ts []T
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		err := v.target.WithContext(dbCtx).Model(new(T)).
			Select("id").Offset(offset).
			Limit(v.batchSize).Find(&ts).Error
		cancel()
		if len(ts) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		switch err {
		case gorm.ErrRecordNotFound:
			if v.sleepInterval > 0 {
				time.Sleep(v.sleepInterval)
				continue
			}
		case context.DeadlineExceeded, context.Canceled:
			return nil
		case nil:
			v.baseMissingRecords(ctx, ts)
		default:
			// 记录日志
			v.l.Error("target => base 查询目标表失败", logger.Error(err))
		}
		if len(ts) < v.batchSize {
			// 没数据了 退出循环
			return nil
		}
		offset += v.batchSize
	}
}

func (v *Validator[T]) baseMissingRecords(ctx context.Context, ts []T) {
	ids := slice.Map[T, int64](ts, func(idx int, src T) int64 {
		return src.ID()
	})
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	base := v.base.WithContext(dbCtx)
	var srcTs []T
	err := base.Select("id").Where("id in ?", ids).Find(&srcTs).Error
	switch err {
	case gorm.ErrRecordNotFound:
		// 说明 ids 都没有
		v.notifyBaseMissing(ts, events.InconsistentEventTypeBaseMissing)
	case nil:
		// 计算差集
		missing := slice.DiffSetFunc[T](ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyBaseMissing(missing, events.InconsistentEventTypeBaseMissing)
	default:
		v.l.Error("targe => base 查询源表失败", logger.Error(err))
	}
}

func (v *Validator[T]) targetMissingRecords(ctx context.Context, ts []T) {
	ids := slice.Map[T, int64](ts, func(idx int, src T) int64 {
		return src.ID()
	})
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var srcTs []T
	target := v.target.WithContext(dbCtx)
	err := target.Select("id").Where("id in ?", ids).Find(&srcTs).Error
	switch err {
	case gorm.ErrRecordNotFound:
		v.notifyTargetMissing(ts, events.InconsistentEventTypeTargetMissing)
	case nil:
		missing := slice.DiffSetFunc[T](ts, srcTs, func(src, dst T) bool {
			if !src.CompareTo(dst) {
				return src.ID() == dst.ID()
			}
			return false
		})
		v.notifyTargetNotEqual(missing, events.InconsistentEventTypeNotEqual)
	default:
		v.l.Error("base => target 查询目标表失败", logger.Error(err))
	}
}

func (v *Validator[T]) notifyBaseMissing(ts []T, typ string) {
	for _, t := range ts {
		v.notify(t.ID(), typ)
	}
}

func (v *Validator[T]) notifyTargetMissing(ts []T, typ string) {
	for _, t := range ts {
		v.notify(t.ID(), typ)
	}
}

func (v *Validator[T]) notifyTargetNotEqual(ts []T, typ string) {
	for _, t := range ts {
		v.notify(t.ID(), typ)
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	evt := events.InconsistentEvent{
		Id:        id,
		Direction: v.direction,
		Type:      typ,
	}

	err := v.producer.ProduceInsistentEvent(ctx, evt)
	if err != nil {
		// 记录日志,告警手动去修
		// 或者下一轮修复和校验还会找出来
		v.l.Error("发送 kafka 失败", logger.Error(err))
	}
}
