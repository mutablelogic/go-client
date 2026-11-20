// Command ipify is an example of go-client: it prints the public IP address
// of the host, from api.ipify.org.
//
//	go run ./examples/ipify
//	go run ./examples/ipify -6
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	// Packages
	client "github.com/mutablelogic/go-client"
	transport "github.com/mutablelogic/go-client/pkg/transport"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Client is a client of the ipify API
type Client struct {
	*client.Client
}

// Response is the public IP address
type Response struct {
	IP string `json:"ip"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	endpoint   = "https://api.ipify.org/"   // IPv4
	endpoint64 = "https://api64.ipify.org/" // IPv6, when there's one, or IPv4
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New returns a client of the endpoint
func New(endpoint string, opts ...client.ClientOpt) (*Client, error) {
	c, err := client.New(append([]client.ClientOpt{client.OptEndpoint(endpoint)}, opts...)...)
	if err != nil {
		return nil, err
	}
	return &Client{c}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get returns the public IP address
func (c *Client) Get(ctx context.Context) (string, error) {
	var response Response
	if err := c.DoWithContext(ctx, nil, &response, client.OptQuery(url.Values{"format": {"json"}})); err != nil {
		return "", err
	}
	return response.IP, nil
}

///////////////////////////////////////////////////////////////////////////////
// MAIN

func main() {
	ipv6 := flag.Bool("6", false, "Return an IPv6 address, when the host has one")
	debug := flag.Bool("debug", false, "Trace requests and responses, with their bodies")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, *ipv6, *debug); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, ipv6, debug bool) error {
	var opts []client.ClientOpt
	if debug {
		opts = append(opts, client.OptTransport(func(parent http.RoundTripper) http.RoundTripper {
			return transport.NewLogging(os.Stderr, parent, true)
		}))
	}
	e := endpoint
	if ipv6 {
		e = endpoint64
	}
	c, err := New(e, opts...)
	if err != nil {
		return err
	}
	ip, err := c.Get(ctx)
	if err != nil {
		return err
	}
	fmt.Println(ip)
	return nil
}
