package otel

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"

	// Packages
	gootel "go.opentelemetry.io/otel"
	attribute "go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlptracegrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otlptracehttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	propagation "go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	noop "go.opentelemetry.io/otel/trace/noop"
)

////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	providerMu     sync.Mutex
	globalProvider *sdktrace.TracerProvider
	propagatorOnce sync.Once
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type Attr struct {
	Key   string
	Value string
}

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewProvider creates a new OpenTelemetry tracer provider. It expects a
// endpoint formatted as host:port for HTTPS endpoints, or a URL with a
// http, https, grpc or grpcs scheme, host, port and optional path.
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
		exporter, err = toHTTP(parsed, toHeaders(header))
	case "grpc", "grpcs":
		exporter, err = toGRPC(parsed, toHeaders(header))
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

	// Set global propagator for trace context propagation (W3C Trace Context).
	// Uses sync.Once to ensure the propagator is only set on the first call.
	// For custom propagator configurations, callers should set the propagator
	// themselves using otel.SetTextMapPropagator() before calling NewProvider.
	propagatorOnce.Do(func() {
		gootel.SetTextMapPropagator(propagation.TraceContext{})
	})

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Register as the global provider so instrumentation libraries
	// (e.g. otelaws) pick it up via gootel.GetTracerProvider().
	// Return an error if a provider is already registered to avoid span
	// discontinuity. Call ShutdownProvider first to replace it intentionally.
	providerMu.Lock()
	defer providerMu.Unlock()
	if globalProvider != nil {
		return nil, fmt.Errorf("global OTel provider already set; call ShutdownProvider first")
	}
	gootel.SetTracerProvider(provider)
	globalProvider = provider
	return provider, nil
}

// ShutdownProvider shuts down the global tracer provider, flushing and
// exporting any remaining spans. After this call, NewProvider can be used
// again. It is a no-op if no provider has been registered.
//
// The OpenTelemetry global provider is reset to a no-op provider so that any
// instrumentation that calls otel.GetTracerProvider() after shutdown receives
// a safe, inert implementation rather than a shut-down SDK provider.
func ShutdownProvider(ctx context.Context) error {
	providerMu.Lock()
	p := globalProvider
	globalProvider = nil
	if p != nil {
		gootel.SetTracerProvider(noop.NewTracerProvider())
	}
	providerMu.Unlock()
	if p == nil {
		return nil
	}
	return p.Shutdown(ctx)
}

func toHTTP(endpoint *url.URL, headers map[string]string) (sdktrace.SpanExporter, error) {
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
		pathComponent = pathComponent + "/v1/traces"
	}
	clientOpts = append(clientOpts, otlptracehttp.WithURLPath("/"+pathComponent))

	if len(headers) > 0 {
		clientOpts = append(clientOpts, otlptracehttp.WithHeaders(headers))
	}

	exporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient(clientOpts...))
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func toGRPC(endpoint *url.URL, headers map[string]string) (sdktrace.SpanExporter, error) {
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
