package trace

import (
	"context"
	"net"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

/**
   @author：biguanqun
   @since： 2023/12/17
   @desc：
**/

type TracingHook struct {
	tracer    trace.Tracer
	isCluster bool
}

func (t *TracingHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (t *TracingHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		ctx, span := t.tracer.Start(ctx, "redisx"+cmd.Name(), trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()

		var cmdName string
		if t.isCluster {
			cmdName = cmd.FullName()
		} else {
			cmdName = cmd.Name()
		}

		span.SetAttributes(attribute.String("cmd.Name", cmdName))
		span.SetAttributes(attribute.String("cmd.String", cmd.String()))
		err := next(ctx, cmd)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}
}

func (t *TracingHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}
