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

package trace

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/bgq98/utils/grpcx/interceptors"
)

type InterceptorBuilder struct {
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
	interceptors.Builder
}

func NewInterceptorBuilder(tracer trace.Tracer, propagator propagation.TextMapPropagator) *InterceptorBuilder {
	return &InterceptorBuilder{
		tracer:     tracer,
		propagator: propagator,
	}
}

func (s *InterceptorBuilder) BuildClient() grpc.UnaryClientInterceptor {
	propagator := s.propagator
	if propagator == nil {
		propagator = otel.GetTextMapPropagator()
	}
	tracer := s.tracer
	if tracer == nil {
		tracer = otel.Tracer("github.com/bgq98/utils/grpcx/interceptors/trace")
	}
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"),
		attribute.Key("rpc.component").String("client"),
	}
	return func(ctx context.Context,
		method string, req, reply any, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		ctx, span := tracer.Start(ctx, method,
			trace.WithAttributes(attrs...),
			trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()

		defer func() {
			if err != nil {
				span.RecordError(err)
				if e := errors.FromError(err); e != nil {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(e.Code)))
				}
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
			span.End()
		}()

		// inject 过程
		ctx = inject(ctx, propagator)
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}
}

type GrpHeaderCarrier metadata.MD

func (s GrpHeaderCarrier) Get(key string) string {
	vals := metadata.MD(s).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (s GrpHeaderCarrier) Set(key string, value string) {
	metadata.MD(s).Set(key, value)
}

func (s GrpHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(s))
	for k := range metadata.MD(s) {
		keys = append(keys, k)
	}
	return keys
}

func inject(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	propagators.Inject(ctx, GrpHeaderCarrier(md))
	return metadata.NewOutgoingContext(ctx, md)
}

func (s *InterceptorBuilder) BuildServer() grpc.UnaryServerInterceptor {
	propagator := s.propagator
	if propagator == nil {
		propagator = otel.GetTextMapPropagator()
	}
	tracer := s.tracer
	if tracer == nil {
		tracer = otel.Tracer("github.com/bgq98/utils/grpcx/interceptors/trace")
	}
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"),
		attribute.Key("rpc.component").String("server"),
	}
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		ctx = extract(ctx, propagator)
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithAttributes(attrs...),
			trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
		span.SetAttributes(
			semconv.RPCMethodKey.String(info.FullMethod),
			semconv.NetPeerNameKey.String(s.PeerName(ctx)),
			attribute.Key("net.peer.ip").String(s.PeerIP(ctx)),
		)
		defer func() {
			if err != nil {
				span.RecordError(err)
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
		}()
		resp, err = handler(ctx, req)
		return
	}
}

func extract(ctx context.Context, p propagation.TextMapPropagator) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	return p.Extract(ctx, GrpHeaderCarrier(md))
}
