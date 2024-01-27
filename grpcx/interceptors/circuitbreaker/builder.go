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

package circuitbreaker

import (
	"context"
	rand2 "math/rand"

	"github.com/go-kratos/aegis/circuitbreaker"
	"google.golang.org/grpc"
)

type InterceptorBuilder struct {
	breaker   circuitbreaker.CircuitBreaker
	threshold int // 标记位 触发熔断就置为 0
}

// BuildServerInterceptor 默认 kratos 熔断
func (s *InterceptorBuilder) BuildServerDefaultInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		// 判断请求是否允许发送
		if s.breaker.Allow() == nil {
			resp, err = handler(ctx, req)
			if err != nil {
				s.breaker.MarkFailed()
			} else {
				s.breaker.MarkSuccess()
			}
		}
		s.breaker.MarkFailed()
		// 触发了熔断器
		return nil, err
	}
}

func (s *InterceptorBuilder) BuildServerThresholdInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		if !s.allow() {
			s.threshold = s.threshold / 2
		}

		rand := rand2.Intn(100)
		if rand <= s.threshold {
			resp, err = handler(ctx, req)
			if err == nil && s.threshold != 0 {
				// 要考虑调大 threshold 说明业务正常
				s.threshold = s.threshold * 2
			} else if s.threshold != 0 {
				// 要考虑调大 threshold
			}
		}
		return
	}
}

func (s *InterceptorBuilder) allow() bool {
	return false
}
