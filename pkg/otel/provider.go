package otel

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"

	// Packages
	attribute "go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlptracegrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otlptracehttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type Attr struct {
	Key   string
	Value string
}

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewProvider creates a new OpenTelemetry tracer provider
func NewProvider(endpoint, header, name string, attrs ...Attr) (*sdktrace.TracerProvider, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("missing OTLP endpoint")
	}

	// If the endpoint is a simple host:port, add the https scheme
	if host, port, err := net.SplitHostPort(endpoint); err == nil && host != "" && port != "" && !strings.Contains(endpoint, "://") {
		endpoint = "https://" + endpoint
	}

	// Parse the endpoint as a URL
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid OTLP endpoint %q: %w", endpoint, err)
	} else if parsed.Scheme == "" {
		return nil, fmt.Errorf("OTLP endpoint %q is missing a scheme", parsed)
	} else if parsed.Host == "" {
		return nil, fmt.Errorf("OTLP endpoint %q is missing a host", parsed)
	}

	var exporter sdktrace.SpanExporter
	switch parsed.Scheme {
	case "http", "https":
		exporter, err = newHTTPTraceExporter(parsed, toHeaders(header))
	case "grpc", "grpcs":
		exporter, err = newGRPCTraceExporter(parsed, toHeaders(header))
	default:
		return nil, fmt.Errorf("unsupported OTLP scheme %q", parsed.Scheme)
	}
	if err != nil {
		return nil, err
	}

	// Add in the service name attribute
	if name != "" {
		attrs = append(attrs, Attr{
			Key:   "service.name",
			Value: name,
		})
	}

	res, err := sdkresource.New(
		context.Background(),
		sdkresource.WithAttributes(toAttributes(attrs)...),
		sdkresource.WithHost(),         // Adds hostname
		sdkresource.WithProcess(),      // Adds process info (PID, executable, etc.)
		sdkresource.WithTelemetrySDK(), // Adds SDK info
	)
	if err != nil {
		return nil, err
	}

	// Return tracer provider
	return sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	), nil
}

func newHTTPTraceExporter(endpoint *url.URL, headers map[string]string) (sdktrace.SpanExporter, error) {
	clientOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint.Host),
	}

	if endpoint.Scheme == "http" {
		clientOpts = append(clientOpts, otlptracehttp.WithInsecure())
	}

	pathComponent := strings.Trim(endpoint.Path, "/")
	switch {
	case pathComponent == "":
		pathComponent = "v1/traces"
	case strings.HasSuffix(pathComponent, "v1/traces"):
		// leave as-is
	default:
		pathComponent = path.Join(pathComponent, "v1/traces")
	}
	clientOpts = append(clientOpts, otlptracehttp.WithURLPath(pathComponent))

	if len(headers) > 0 {
		clientOpts = append(clientOpts, otlptracehttp.WithHeaders(headers))
	}

	exporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient(clientOpts...))
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func newGRPCTraceExporter(endpoint *url.URL, headers map[string]string) (sdktrace.SpanExporter, error) {
	if endpoint.Path != "" && endpoint.Path != "/" {
		return nil, fmt.Errorf("gRPC OTLP endpoint should not include a path: %q", endpoint.Path)
	}

	clientOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint.Host),
	}

	if endpoint.Scheme == "grpc" {
		clientOpts = append(clientOpts, otlptracegrpc.WithInsecure())
	}

	if len(headers) > 0 {
		clientOpts = append(clientOpts, otlptracegrpc.WithHeaders(headers))
	}

	exporter, err := otlptrace.New(context.Background(), otlptracegrpc.NewClient(clientOpts...))
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func toHeaders(header string) map[string]string {
	headers := make(map[string]string)
	if header == "" {
		return headers
	}
	for _, pair := range strings.Split(header, ",") {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		if key := strings.TrimSpace(kv[0]); key != "" {
			headers[key] = strings.TrimSpace(kv[1])
		}
	}
	return headers
}

func toAttributes(attrs []Attr) []attribute.KeyValue {
	var result []attribute.KeyValue
	for _, attr := range attrs {
		result = append(result, attribute.String(attr.Key, attr.Value))
	}
	return result
}
