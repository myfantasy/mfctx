package trace

import (
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// NewJaegerOpenTracerFromEnv create jaeger tracer and sets it as global
func NewJaegerOpenTracerFromEnv() (tracer opentracing.Tracer, closer io.Closer, err error) {
	cfg, err := config.FromEnv()
	if err != nil {
		return nil, nil, errors.Wrap(err, "jaeger env parse")
	}

	return NewJaegerOpenTracer(cfg)
}

// NewJaegerOpenTracer create jaeger tracer and sets it as global
func NewJaegerOpenTracer(cfg *config.Configuration) (tracer opentracing.Tracer, closer io.Closer, err error) {

	tracer, closer, err = cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		return nil, nil, errors.Wrap(err, "jaeger init")
	}

	opentracing.SetGlobalTracer(tracer)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	return tracer, closer, nil
}

func NewConsoleTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	return tracerProvider, err
}
