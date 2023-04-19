// Copyright 2020 Layer5, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type KeyValue struct {
	Key   string
	Value string
}

type Handler interface {
	Tracer(name string) interface{}
	Span(ctx context.Context)
	AddEvent(name string, attrs ...*KeyValue)
}

type handler struct {
	provider trace.TracerProvider
	context  context.Context
	span     trace.Span
}

func New(service string, endpoint string) (Handler, error) {
	if len(endpoint) < 2 {
		return nil, nil
	}

	provider, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(provider),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(service),
		)),
	)
	return &handler{
		provider: tp,
	}, err
}

func (h *handler) Tracer(name string) interface{} {
	return otel.GetTracerProvider().Tracer(name)
}

func (h *handler) Span(ctx context.Context) {
	h.span = trace.SpanFromContext(ctx)
	h.context = ctx
}

func (h *handler) AddEvent(name string, attrs ...*KeyValue) {
	kvstore := make([]trace.EventOption, 0)
	// @TODO still need to fix this portion
	// for _, attr := range attrs {
	// kvstore = append(kvstore, trace.WithAttributes(attribute.String(attr.Key, attr.Value)))
	// }

	h.span.AddEvent(name, kvstore...)
}
