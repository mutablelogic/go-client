package otel_test

import (
	"context"
	"testing"

	// Packages
	"github.com/mutablelogic/go-client/pkg/otel"
	"github.com/stretchr/testify/assert"
)

func TestNewProvider_EmptyEndpoint(t *testing.T) {
	assert := assert.New(t)

	provider, err := otel.NewProvider("", "", "test-service")
	assert.Nil(provider)
	assert.Error(err)
	assert.Contains(err.Error(), "missing OTLP endpoint")
}

func TestNewProvider_InvalidEndpoint(t *testing.T) {
	assert := assert.New(t)

	provider, err := otel.NewProvider("://invalid", "", "test-service")
	assert.Nil(provider)
	assert.Error(err)
}

func TestNewProvider_MissingScheme(t *testing.T) {
	assert := assert.New(t)

	// Just a hostname without scheme or port should fail
	provider, err := otel.NewProvider("example.com", "", "test-service")
	assert.Nil(provider)
	assert.Error(err)
	assert.Contains(err.Error(), "missing a scheme")
}

func TestNewProvider_UnsupportedScheme(t *testing.T) {
	assert := assert.New(t)

	provider, err := otel.NewProvider("ftp://example.com:4318", "", "test-service")
	assert.Nil(provider)
	assert.Error(err)
	assert.Contains(err.Error(), "unsupported OTLP scheme")
}

func TestNewProvider_HostPortDefaultsToHTTPS(t *testing.T) {
	assert := assert.New(t)

	// Note: This will fail to connect but should parse correctly
	// We're testing that host:port format is accepted and defaults to https
	provider, err := otel.NewProvider("localhost:4318", "", "test-service")

	// The provider creation should succeed even if we can't connect
	assert.NotNil(provider)
	assert.NoError(err)

	if provider != nil {
		provider.Shutdown(context.Background())
	}
}

func TestNewProvider_HTTPEndpoint(t *testing.T) {
	assert := assert.New(t)

	provider, err := otel.NewProvider("http://localhost:4318", "", "test-service")
	assert.NotNil(provider)
	assert.NoError(err)

	if provider != nil {
		provider.Shutdown(context.Background())
	}
}

func TestNewProvider_HTTPSEndpoint(t *testing.T) {
	assert := assert.New(t)

	provider, err := otel.NewProvider("https://localhost:4318", "", "test-service")
	assert.NotNil(provider)
	assert.NoError(err)

	if provider != nil {
		provider.Shutdown(context.Background())
	}
}

func TestNewProvider_GRPCEndpoint(t *testing.T) {
	assert := assert.New(t)

	provider, err := otel.NewProvider("grpc://localhost:4317", "", "test-service")
	assert.NotNil(provider)
	assert.NoError(err)

	if provider != nil {
		provider.Shutdown(context.Background())
	}
}

func TestNewProvider_GRPCSEndpoint(t *testing.T) {
	assert := assert.New(t)

	provider, err := otel.NewProvider("grpcs://localhost:4317", "", "test-service")
	assert.NotNil(provider)
	assert.NoError(err)

	if provider != nil {
		provider.Shutdown(context.Background())
	}
}

func TestNewProvider_WithPath(t *testing.T) {
	assert := assert.New(t)

	provider, err := otel.NewProvider("https://localhost:4318/custom/path", "", "test-service")
	assert.NotNil(provider)
	assert.NoError(err)

	if provider != nil {
		provider.Shutdown(context.Background())
	}
}

func TestNewProvider_WithHeaders(t *testing.T) {
	assert := assert.New(t)

	provider, err := otel.NewProvider(
		"https://localhost:4318",
		"api-key=test123,other-header=value",
		"test-service",
	)
	assert.NotNil(provider)
	assert.NoError(err)

	if provider != nil {
		provider.Shutdown(context.Background())
	}
}

func TestNewProvider_WithAttributes(t *testing.T) {
	assert := assert.New(t)

	provider, err := otel.NewProvider(
		"https://localhost:4318",
		"",
		"test-service",
		otel.Attr{Key: "environment", Value: "test"},
		otel.Attr{Key: "version", Value: "1.0.0"},
	)
	assert.NotNil(provider)
	assert.NoError(err)

	if provider != nil {
		provider.Shutdown(context.Background())
	}
}

func TestNewProvider_EmptyServiceName(t *testing.T) {
	assert := assert.New(t)

	// Empty service name should still work
	provider, err := otel.NewProvider("https://localhost:4318", "", "")
	assert.NotNil(provider)
	assert.NoError(err)

	if provider != nil {
		provider.Shutdown(context.Background())
	}
}

func TestNewProvider_GRPCWithPath(t *testing.T) {
	assert := assert.New(t)

	// gRPC endpoints should not have a path
	provider, err := otel.NewProvider("grpc://localhost:4317/some/path", "", "test-service")
	assert.Nil(provider)
	assert.Error(err)
	assert.Contains(err.Error(), "should not include a path")
}
