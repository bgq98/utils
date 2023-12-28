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

package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/bgq98/utils/ginx"
	"github.com/bgq98/utils/gormx/connpool"
	"github.com/bgq98/utils/logger"
	"github.com/bgq98/utils/migrator"
	"github.com/bgq98/utils/migrator/events"
	"github.com/bgq98/utils/migrator/validator"
)

type Scheduler[T migrator.Entity] struct {
	lock       sync.Mutex
	src        *gorm.DB
	dst        *gorm.DB
	pool       *connpool.DoubleWritePool
	l          logger.Logger
	pattern    string
	cancelFull func()
	cancelIntr func()
	producer   events.Producer
}

func NewScheduler[T migrator.Entity](src *gorm.DB, dst *gorm.DB,
	pool *connpool.DoubleWritePool, l logger.Logger, producer events.Producer) *Scheduler[T] {
	return &Scheduler[T]{
		src:     src,
		dst:     dst,
		pool:    pool,
		l:       l,
		pattern: connpool.PatternSrcOnly,
		cancelFull: func() {
			// 初始化,什么都不用做
		},
		cancelIntr: func() {
			// 初始化,什么都不用做
		},
		producer: producer,
	}
}

func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	server.POST("/src_only", ginx.Wrap(s.SrcOnly))
	server.POST("/src_first", ginx.Wrap(s.SrcFirst))
	server.POST("/dst_only", ginx.Wrap(s.DstOnly))
	server.POST("/dst_first", ginx.Wrap(s.DstFirst))
	server.POST("/full/start", ginx.Wrap(s.StartFullvalidation))
	server.POST("/full/stop", ginx.Wrap(s.StopFullValidation))
	server.POST("/incr/stop", ginx.Wrap(s.StopIncrementValidation))
	server.POST("/incr/start", ginx.WrapReq[StartIncrRequest](s.StartIncrementValidation))
}

func (s *Scheduler[T]) SrcOnly(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcOnly
	s.pool.ChangePattern(connpool.PatternSrcOnly)
	return ginx.Result{
		Msg: "ok",
	}, nil
}

func (s *Scheduler[T]) SrcFirst(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcFirst
	s.pool.ChangePattern(connpool.PatternSrcFirst)
	return ginx.Result{
		Msg: "ok",
	}, nil
}

func (s *Scheduler[T]) DstOnly(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstOnly
	s.pool.ChangePattern(connpool.PatternDstOnly)
	return ginx.Result{
		Msg: "ok",
	}, nil
}

func (s *Scheduler[T]) DstFirst(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstFirst
	s.pool.ChangePattern(connpool.PatternDstFirst)
	return ginx.Result{
		Msg: "ok",
	}, nil
}

// StartIncrementValidation 开启增量校验
func (s *Scheduler[T]) StartIncrementValidation(c *gin.Context,
	req StartIncrRequest) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	cancel := s.cancelIntr
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		}, nil
	}
	v.SleepInterval(time.Duration(req.Interval) * time.Millisecond).Utime(req.Utime)
	var ctx context.Context
	ctx, s.cancelIntr = context.WithCancel(context.Background())

	go func() {
		cancel()
		err := v.Validate(ctx)
		s.l.Warn("退出增量校验", logger.Error(err))
	}()
	return ginx.Result{
		Msg: "启动增量校验成功",
	}, nil
}

// StopIncrementValidation 停止增量校验
func (s *Scheduler[T]) StopIncrementValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIntr()
	return ginx.Result{
		Msg: "ok",
	}, nil
}

// StartFullvalidation 开启全量校验
func (s *Scheduler[T]) StartFullvalidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	cancel := s.cancelFull // 上一次的取消 func
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{}, err
	}
	var ctx context.Context
	ctx, s.cancelFull = context.WithCancel(context.Background())

	go func() {
		cancel() // 先取消上一次的
		err := v.Validate(ctx)
		if err != nil {
			s.l.Warn("退出全量校验", logger.Error(err))
		}
	}()
	return ginx.Result{
		Msg: "ok",
	}, nil
}

// StopFullValidation 停止全量校验
func (s *Scheduler[T]) StopFullValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelFull()
	return ginx.Result{
		Msg: "ok",
	}, nil
}

func (s *Scheduler[T]) newValidator() (*validator.Validator[T], error) {
	switch s.pattern {
	case connpool.PatternSrcOnly, connpool.PatternSrcFirst:
		return validator.NewValidator[T](s.src, s.dst, s.l, s.producer, "Src"), nil
	case connpool.PatternDstOnly, connpool.PatternDstFirst:
		return validator.NewValidator[T](s.dst, s.src, s.l, s.producer, "Dst"), nil
	default:
		return nil, fmt.Errorf("未知的 pattern %s", s.pattern)
	}
}

type StartIncrRequest struct {
	Utime    int64 `json:"utime"`
	Interval int64 `json:"interval"` // [毫秒数] json 不能正确处理 time.Duration 类型
}
