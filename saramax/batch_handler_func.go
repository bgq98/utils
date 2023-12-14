package saramax

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"

	"utils/logger"
)

/**
   @author：biguanqun
   @since： 2023/11/5
   @desc：
**/

type BatchHandler[T any] struct {
	l  logger.Logger
	fn func(msgs []*sarama.ConsumerMessage, t []T) error
	// 用 option 模式来设置这个 batchSize 和 duration
	batchSize     int
	batchDuration time.Duration
}

func NewBatchHandler[T any](l logger.Logger,
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
