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

package prometheus

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/event"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/bgq98/utils/logger"
)

/**
   @author：biguanqun
   @since： 2023/12/17
   @desc：
**/

type (
	Started   func(ctx context.Context, startedEvent *event.CommandStartedEvent)
	Succeeded func(ctx context.Context, succeededEvent *event.CommandSucceededEvent)
	Failed    func(ctx context.Context, failedEvent *event.CommandFailedEvent)
)

type MongoPluginMonitor struct {
	vector *prometheus.SummaryVec
	tracer trace.Tracer
	l      logger.Logger
	ctx    context.Context
}

func NewMongoPluginMonitor(opt prometheus.SummaryOpts, l logger.Logger) *MongoPluginMonitor {
	vector := prometheus.NewSummaryVec(opt, []string{"cmd"})
	prometheus.MustRegister(vector)
	return &MongoPluginMonitor{
		vector: vector,
		tracer: otel.GetTracerProvider().Tracer("github.com/bgq98/utils/mongox/prometheus"),
		l:      l,
	}
}

func (m *MongoPluginMonitor) StartedPrometheus() Started {
	return func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
		var span trace.Span
		m.ctx, span = m.tracer.Start(ctx, "mongodbx"+startedEvent.CommandName,
			trace.WithSpanKind(trace.SpanKindClient))
		span.SetAttributes(attribute.String("mongo.database", startedEvent.DatabaseName))
		span.SetAttributes(attribute.String("mongo.command", startedEvent.Command.String()))
		span.SetAttributes(attribute.String("mongo.command.name", startedEvent.CommandName))
		m.l.Debug("mongo", logger.Any("mongoCommand", startedEvent.Command))
	}
}

func (m *MongoPluginMonitor) SucceedPrometheus() Succeeded {
	return func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {
		duration := time.Duration(succeededEvent.DurationNanos)
		cmd := succeededEvent.CommandName
		m.vector.WithLabelValues(cmd).Observe(float64(duration.Milliseconds()))

		span := trace.SpanFromContext(m.ctx)
		if !span.IsRecording() {
			// 判断 span 是否处于活跃状态
			return
		}
		defer span.End()
	}
}

func (m *MongoPluginMonitor) FailedPrometheus() Failed {
	return func(ctx context.Context, failedEvent *event.CommandFailedEvent) {
		duration := time.Duration(failedEvent.DurationNanos)
		cmd := failedEvent.CommandName
		m.vector.WithLabelValues(cmd).Observe(float64(duration.Milliseconds()))

		span := trace.SpanFromContext(m.ctx)
		if !span.IsRecording() {
			return
		}
		defer span.End()
		span.SetStatus(codes.Error, failedEvent.Failure)
	}
}
