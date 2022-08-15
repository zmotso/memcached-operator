package tracing

import (
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
)

// Bootstrap prepares a new tracer to be used by the operator
func Bootstrap(namespace, instanceID string) error {
	jaegerURL, enableTracing := os.LookupEnv("JAEGER_URL")
	if !enableTracing {
		return nil
	}

	hostPort := strings.Split(jaegerURL, ":")

	var endpoint jaeger.EndpointOption
	if len(hostPort) >= 2 {
		endpoint = jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(hostPort[0]),
			jaeger.WithAgentPort(hostPort[1]),
		)
	} else {
		endpoint = jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(hostPort[0]),
		)
	}

	jexporter, err := jaeger.NewRawExporter(endpoint)
	if err != nil {
		return err
	}

	processor := tracesdk.NewBatchSpanProcessor(jexporter)

	attr := []attribute.KeyValue{
		semconv.ServiceNameKey.String("memcached-operator"),
		semconv.ServiceVersionKey.String("0.0.1"),
		semconv.ServiceNamespaceKey.String(namespace),
	}

	if instanceID != "" {
		attr = append(attr, semconv.ServiceInstanceIDKey.String(instanceID))
	}

	traceProvider := tracesdk.NewTracerProvider(
		tracesdk.WithSpanProcessor(processor),
		tracesdk.WithResource(resource.NewWithAttributes(attr...)),
	)
	otel.SetTracerProvider(traceProvider)

	return nil
}
