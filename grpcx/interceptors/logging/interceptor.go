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

package logging

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bgq98/utils/grpcx/interceptors"
	"github.com/bgq98/utils/logger"
)

type InterceptorBuilder struct {
	l        logger.Logger
	reqBody  bool
	respBody bool
	interceptors.Builder
}

func (s *InterceptorBuilder) BuildCliet() grpc.UnaryClientInterceptor {
	return func(ctx context.Context,
		method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		start := time.Now()
		var event = "normal"
		defer func() {
			duration := time.Since(start)
			if rec := recover(); rec != nil {
				switch recType := rec.(type) {
				case error:
					err = recType
				default:
					err = fmt.Errorf("%v", rec)
				}
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				event = "recover"
				err = status.New(codes.Internal, "panic,err"+err.Error()).Err()
			}
			fields := []logger.Field{
				logger.Int64("cost", duration.Milliseconds()),
				logger.String("type", "unary"),
				logger.String("event", event),
				logger.String("peer", s.PeerName(ctx)),
				logger.String("peer_ip", s.PeerIP(ctx)),
			}
			if err != nil {
				st, _ := status.FromError(err)
				fields = append(fields,
					logger.String("code", st.Code().String()),
					logger.String("code_msg", st.Message()))
			}
			s.l.Info("RPC请求", fields...)
		}()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (s *InterceptorBuilder) BuildServer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		var event = "normal"
		defer func() {
			duration := time.Since(start)
			if rec := recover(); rec != nil {
				switch recType := rec.(type) {
				case error:
					err = recType
				default:
					err = fmt.Errorf("%v", rec)
				}
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				event = "recover"
				err = status.New(codes.Internal, "panic,err"+err.Error()).Err()
			}
			fields := []logger.Field{
				logger.Int64("cost", duration.Milliseconds()),
				logger.String("type", "unary"),
				logger.String("method", info.FullMethod),
				logger.String("event", event),
				logger.String("peer", s.PeerName(ctx)),
				logger.String("peer_ip", s.PeerIP(ctx)),
			}
			if err != nil {
				st, _ := status.FromError(err)
				fields = append(fields,
					logger.String("code", st.Code().String()),
					logger.String("code_msg", st.Message()))
			}
			s.l.Info("RPC请求", fields...)
		}()
		resp, err = handler(ctx, req)
		return
	}
}
