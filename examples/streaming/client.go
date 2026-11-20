package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"
)

func runClient(ctx context.Context, listenAddr string) error {
	// Create a client
	c, err := client.New(client.OptEndpoint("http://" + listenAddr))
	if err != nil {
		return err
	}

	// Create a bi-directional stream
	if err := c.Stream(ctx, func(ctx context.Context, stream client.JSONStream) error {
		ticker := time.NewTimer(time.Second)
		defer ticker.Stop()
		var seq int
		for {
			select {
			case <-ticker.C:
				if err := stream.Send(Event{Message: fmt.Sprintf("client seq %d", seq)}.JSON()); err != nil {
					return err
				}
				seq++
				ticker.Reset(time.Second * time.Duration(rand.Int31n(8)))
			case evt, ok := <-stream.Recv():
				if !ok {
					return nil
				}
				if e, err := NewEvent(evt); err != nil {
					return err
				} else {
					fmt.Println("received event:", e)
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}); err != nil {
		return err
	}

	// Create a ticker
	return nil
}
