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

	label "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	apitrace "go.opentelemetry.io/otel/trace"
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
	provider apitrace.TracerProvider
	context  context.Context
	span     apitrace.Span
}

func New(service string, endpoint string) (Handler, error) {
	if len(endpoint) < 2 {
		return nil, nil
	}

	exporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)),
	)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			label.String("name", service),
			label.String("exporter", "jaeger"),
		),
		),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	return &handler{
		provider: provider,
	}, nil
}

func (h *handler) Tracer(name string) interface{} {
	return h.provider
}

func (h *handler) Span(ctx context.Context) {
	h.span = apitrace.SpanFromContext(ctx)
	h.context = ctx
}

func (h *handler) AddEvent(name string, attrs ...*KeyValue) {
	kvstore := make([]label.KeyValue, 0)
	for _, attr := range attrs {
		kvstore = append(kvstore, label.String(attr.Key, attr.Value))
	}

	h.span.AddEvent(name, apitrace.WithAttributes(kvstore...))
}
