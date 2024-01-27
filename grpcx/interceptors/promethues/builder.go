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

package promethues

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/bgq98/utils/grpcx/interceptors"
)

type InterceptorBuilder struct {
	Namespace string
	Subsystem string
	interceptors.Builder
}

func (s *InterceptorBuilder) BuildServer() grpc.UnaryServerInterceptor {
	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: s.Namespace,
			Subsystem: s.Subsystem,
			Name:      "server_handle_seconds",
			Objectives: map[float64]float64{
				0.5:   0.01,
				0.9:   0.01,
				0.95:  0.01,
				0.99:  0.001,
				0.999: 0.0001,
			},
		}, []string{"type", "service", "method", "peer", "code"})
	prometheus.MustRegister(summary)
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		defer func() {
			c, m := s.splitMethodName(info.FullMethod)
			duration := float64(time.Since(start).Milliseconds())
			if err == nil {
				summary.WithLabelValues("unary", c, m, s.PeerName(ctx), "OK").Observe(duration)
			} else {
				st, _ := status.FromError(err)
				summary.WithLabelValues("unary", c, m, s.PeerName(ctx), st.Code().String()).Observe(duration)
			}
		}()
		resp, err = handler(ctx, req)
		return
	}
}

func (s *InterceptorBuilder) splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/")
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}
