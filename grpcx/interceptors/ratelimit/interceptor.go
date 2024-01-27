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

package ratelimit

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bgq98/utils/ginx/middlewares/ratelimit"
	"github.com/bgq98/utils/logger"
)

type InterceptorBuilder struct {
	limiter ratelimit.Limiter // 滑动窗口算法限流器
	l       logger.Logger
	key     string // 限流器key
	name    string // 服务名
}

// BuildServerInterceptor 整个应用,集群的限流
// key limiter:service:user
func (s *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := s.limiter.Limit(ctx, fmt.Sprintf(s.key+":"+s.name))
		if err != nil {
			s.l.Error("判定限流出了问题", logger.Error(err))
			return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
		}
		if limited {
			return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
		}
		return handler(ctx, req)
	}
}

// BuildServerInterceptorIO 配合后续业务做限流处理
// key limiter:service:user
func (s *InterceptorBuilder) BuildServerInterceptorIO() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := s.limiter.Limit(ctx, fmt.Sprintf(s.key+":"+s.name))
		if err != nil || limited {
			ctx = context.WithValue(ctx, "limited", "true")
		}
		return handler(ctx, req)
	}
}

// BuildClientInterceptor 客户端限流
// key limiter:service:user
func (s *InterceptorBuilder) BuildClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context,
		method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {
		limited, err := s.limiter.Limit(ctx, fmt.Sprintf(s.key+":"+s.name))
		if err != nil {
			s.l.Error("判定限流出了问题", logger.Error(err))
			return status.Errorf(codes.ResourceExhausted, "触发限流")
		}
		if limited {
			return status.Errorf(codes.ResourceExhausted, "触发限流")
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// BuildServerInterceptorService 服务级别限流
// key limiter:service:user:UserService user 里面的 UserService 限流
func (s *InterceptorBuilder) BuildServerInterceptorService() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		if strings.HasPrefix(info.FullMethod, "/UserService") {
			limited, err := s.limiter.Limit(ctx, fmt.Sprintf(s.key+":"+s.name+":"+"UserService"))
			if err != nil {
				s.l.Error("判定限流出了问题", logger.Error(err))
				return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
			}
			if limited {
				return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
			}
		}
		return handler(ctx, req)
	}
}
