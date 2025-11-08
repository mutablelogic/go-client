package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	// Packages
	kong "github.com/alecthomas/kong"
	tablewriter "github.com/djthorpe/go-tablewriter"
	client "github.com/mutablelogic/go-client"
	attribute "go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlptracehttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	// Version info
	"github.com/mutablelogic/go-client/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Kong CLI root struct
type CLI struct {
	Globals
	IPify         `embed:"" prefix:"ipify."`
	HomeAssistant `embed:"" prefix:"ha."`
}

type Globals struct {
	ctx         context.Context
	tablewriter *tablewriter.Writer
	opts        []client.ClientOpt

	// Global options
	Path         string `help:"Output file path, defaults to stdout" short:"o"`
	OtelEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" help:"OpenTelemetry collector endpoint"`
	OtelHeader   string `env:"OTEL_EXPORTER_OTLP_HEADERS" help:"OpenTelemetry collector headers"`
	OtelName     string `env:"OTEL_SERVICE_NAME" help:"OpenTelemetry service name" default:"go-client"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	var err error

	// Parse command-line arguments
	name, _ := os.Executable()
	cli := CLI{}
	cmd := kong.Parse(&cli,
		kong.Name(path.Base(name)),
		kong.Description("API client"),
		kong.UsageOnError(),
	)

	// Create a context
	var cancel context.CancelFunc
	cli.Globals.ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGQUIT)
	defer cancel()

	// Tablewriter
	cli.Globals.tablewriter, err = NewTableWriter(&cli)
	cmd.FatalIfErrorf(err)

	// Open Telemetry
	var rootSpanEnd func()
	if cli.Globals.OtelEndpoint != "" {
		if provider, err := NewTracerProvider(cli.Globals.OtelEndpoint, cli.Globals.OtelHeader, cli.Globals.OtelName); err != nil {
			cmd.Fatalf("Failed to create tracer: %v", err)
		} else {
			tracer := provider.Tracer(cli.Globals.OtelName)
			cli.Globals.opts = append(cli.Globals.opts, client.OptTracer(tracer, "api"))
			// Start root span
			ctx, span := tracer.Start(cli.Globals.ctx, "cli."+cmd.Command())
			cli.Globals.ctx = ctx
			rootSpanEnd = func() { span.End() }
			defer func() {
				if rootSpanEnd != nil {
					rootSpanEnd()
				}
				// Give the provider time to flush spans
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := provider.Shutdown(shutdownCtx); err != nil {
					fmt.Fprintf(os.Stderr, "Error shutting down tracer provider: %v\n", err)
				}
			}()
		}
	}

	// Run the command
	if err := cmd.Run(&cli.Globals); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func NewTableWriter(ctx *CLI) (*tablewriter.Writer, error) {
	// By default, the writer is stdout
	w := os.Stdout
	if filename := ctx.Path; filename != "" {
		if file, err := os.Create(filename); err != nil {
			return nil, err
		} else {
			w = file
		}
	}

	// Set table output options
	opts := []tablewriter.TableOpt{
		tablewriter.OptHeader(),
		tablewriter.OptTerminalWidth(w),
	}
	ext := strings.ToLower(ctx.Path)
	switch ext {
	case "csv":
		opts = append(opts, tablewriter.OptOutputCSV())
	case "tsv":
		opts = append(opts, tablewriter.OptOutputTSV())
	default:
		opts = append(opts, tablewriter.OptOutputText())
	}

	// Return success
	return tablewriter.New(w, opts...), nil
}

func NewTracerProvider(endpoint, header, name string) (*sdktrace.TracerProvider, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // Use HTTP instead of HTTPS
	}
	if header != "" {
		headers := make(map[string]string)
		for _, pair := range strings.Split(header, ",") {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(opts...),
	)
	if err != nil {
		return nil, err
	}

	// Build resource with service attributes
	attrs := []attribute.KeyValue{}
	if name != "" {
		attrs = append(attrs, semconv.ServiceName(name))
	}
	if version.GitTag != "" {
		attrs = append(attrs, semconv.ServiceVersion(version.GitTag))
	}
	if version.GitBranch != "" {
		attrs = append(attrs, semconv.DeploymentEnvironment(version.GitBranch))
	}

	// Add additional process metadata
	attrs = append(attrs,
		semconv.TelemetrySDKLanguageGo,
		semconv.TelemetrySDKName("opentelemetry"),
		attribute.String("process.runtime.name", "go"),
	)

	res, err := sdkresource.New(
		context.Background(),
		sdkresource.WithAttributes(attrs...),
		sdkresource.WithHost(),    // Adds hostname
		sdkresource.WithProcess(), // Adds process info (PID, executable, etc.)
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
