package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
)

func runServer(ctx context.Context, listenAddr string) error {
	// Create a new server
	server, err := httpserver.New(listenAddr, nil)
	if err != nil {
		return err
	}

	// Bind to the listen address
	if err := server.Listen(); err != nil {
		return err
	}

	// Create a handler for the streaming endpoint
	handler := httpresponse.NewJSONStreamHandler(func(r <-chan json.RawMessage, w chan<- json.RawMessage) error {
		fmt.Println("stream opened")

		// Create a ticker
		ticker := time.NewTimer(time.Second)
		defer ticker.Stop()

		var seq int
	FOR_LOOP:
		for {
			select {
			case evt, ok := <-r:
				if !ok {
					fmt.Println("receiving event failed")
					break FOR_LOOP
				}
				if e, err := NewEvent(evt); err == nil {
					fmt.Println("received event:", e)
				} else {
					return err
				}
			case <-ticker.C:
				fmt.Println("sending event")
				w <- Event{Message: fmt.Sprintf("server seq %d", seq)}.JSON()
				fmt.Println("sent event")
				seq++
				ticker.Reset(time.Second * time.Duration(rand.Int31n(8)))
			case <-ctx.Done():
				break FOR_LOOP
			}
		}
		fmt.Println("stream closed")
		return nil
	})

	// Register the streaming handler
	server.Router().Handle("/", handler)

	// Wait for the context to be done
	return server.Run(ctx)
}
