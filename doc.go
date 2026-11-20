/*
Package client implements a generic REST API client which can be used for
creating gateway-specific clients. Basic usage:

	package main

	import (
		"fmt"
		"log"

		client "github.com/mutablelogic/go-client"
	)

	func main() {
		// Create a new client
		c, err := client.New(client.OptEndpoint("https://api.example.com/api/v1"))
		if err != nil {
			log.Fatal(err)
		}

		// Send a GET request, populating a struct with the response
		var response struct {
			Message string `json:"message"`
		}
		if err := c.Do(nil, &response, client.OptPath("test")); err != nil {
			log.Fatal(err)
		}

		// Print the response
		fmt.Println(response.Message)
	}
*/
package client
