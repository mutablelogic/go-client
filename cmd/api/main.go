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
	otel "github.com/mutablelogic/go-client/pkg/otel"
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

	// Open Telemetry - update the context to include root span
	if cli.Globals.OtelEndpoint != "" {
		provider, err := otel.NewProvider(cli.Globals.OtelEndpoint, cli.Globals.OtelHeader, cli.Globals.OtelName)
		if err != nil {
			cmd.Fatalf("Failed to create tracer: %v", err)
		}

		tracer := provider.Tracer(cli.Globals.OtelName)
		cli.Globals.opts = append(cli.Globals.opts, client.OptTracer(tracer))

		// Start root span
		ctx, span := tracer.Start(cli.Globals.ctx, "cli."+cmd.Command())
		cli.Globals.ctx = ctx
		defer func() {
			// Give the provider time to flush spans
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Perform shutdown
			if err := provider.Shutdown(shutdownCtx); err != nil {
				fmt.Fprintf(os.Stderr, "Error shutting down tracer provider: %v\n", err)
			}
		}()
		defer span.End()
	}

	// Run the command
	if err := cmd.Run(&cli.Globals); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
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
