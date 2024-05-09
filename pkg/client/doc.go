/*
Implements a generic REST API client which can be used for creating
gateway-specific clients. Basic usage:

	package main

	import (
		client "github.com/mutablelogic/go-client/pkg/client"
	)

	func main() {
		// Create a new client
		c := client.New(client.OptEndpoint("https://api.example.com/api/v1"))

		// Send a GET request, populating a struct with the response
		var response struct {
			Message string `json:"message"`
		}
		if err := c.Do(nil, &response, OptPath("test")); err != nil {
			// Handle error
		}

		// Print the response
		fmt.Println(response.Message)
	}
*/
package client
