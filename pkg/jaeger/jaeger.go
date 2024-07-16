package jaeger

import (
	"Service/constant"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"net"
	"strconv"
)

type Config struct {
	LogSpans bool   `yaml:"log_span"`
	Host     string `env:"JAEGER_AGENT_HOST" env-required:"true"`
	Port     int    `env:"JAEGER_AGENT_PORT" env-required:"true"`
}

func InitJaeger(cfg Config) (*sdktrace.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://" + net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)) + "/api/traces")))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(constant.ServiceName),
			semconv.DeploymentEnvironmentKey.String("production"),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}
