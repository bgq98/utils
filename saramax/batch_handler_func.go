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

package saramax

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"

	"github.com/bgq98/utils/logger"
)

/**
   @author：biguanqun
   @since： 2023/11/5
   @desc：
**/

type BatchHandler[T interface{}] struct {
	l  logger.Logger
	fn func(msgs []*sarama.ConsumerMessage, t []T) error
	// 用 option 模式来设置这个 batchSize 和 duration
	batchSize     int
	batchDuration time.Duration
}

func NewBatchHandler[T interface{}](l logger.Logger,
	fn func(msgs []*sarama.ConsumerMessage, t []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{
		l:             l,
		fn:            fn,
		batchDuration: time.Second,
		batchSize:     10,
	}
}

func (h *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	msgCh := claim.Messages()
	batchSize := h.batchSize
	for {
		msgs := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		ctx, cancelFunc := context.WithTimeout(context.Background(), h.batchDuration)
		done := false
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				// 这一批次已经超时了或者整个 consume 被关闭了
				done = true
			case msg, ok := <-msgCh:
				if !ok {
					// chan 被关闭了
					cancelFunc()
					return nil
				}
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					h.l.Error("反序列化消息体失败",
						logger.Int64("offset", msg.Offset),
						logger.Int32("partition", msg.Partition),
						logger.String("topic", msg.Topic),
						logger.Error(err))
					continue
				}
				msgs = append(msgs, msg)
				ts = append(ts, t)
			}
		}
		err := h.fn(msgs, ts)
		if err == nil {
			for _, msg := range msgs {
				session.MarkMessage(msg, "")
			}
		} else {
			// 考虑重试
		}
		cancelFunc()
	}
}
