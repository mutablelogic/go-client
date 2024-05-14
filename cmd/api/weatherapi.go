package main

import (
	"context"
	"fmt"

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/weatherapi"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	weatherapiName   = "weatherapi"
	weatherapiClient *weatherapi.Client
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func weatherapiRegister(flags *Flags) {
	// Register flags required
	flags.String(newsapiName, "weatherapi-key", "${WEATHERAPI_KEY}", "API Key")

	flags.Register(Cmd{
		Name:        weatherapiName,
		Description: "Obtain weather information from https://www.weatherapi.com/",
		Parse:       weatherapiParse,
		Fn: []Fn{
			{Call: weatherapiCurrent, Description: "Return current weather, given city, zip code, IP address or lat,long", MaxArgs: 1},
		},
	})
}

func weatherapiParse(flags *Flags, opts ...client.ClientOpt) error {
	apiKey := flags.GetString("weatherapi-key")
	if apiKey == "" {
		return fmt.Errorf("missing -weatherapi-key flag")
	}
	if client, err := weatherapi.New(apiKey, opts...); err != nil {
		return err
	} else {
		weatherapiClient = client
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func weatherapiCurrent(_ context.Context, w *tablewriter.Writer, args []string) error {
	var q string
	if len(args) == 1 {
		q = args[0]
	} else {
		q = "auto:ip"
	}

	// Request -> Response
	weather, err := weatherapiClient.Current(q)
	if err != nil {
		return err
	}

	// Write table
	w.Write(weather.Location)
	w.Write(weather.Current)

	// Return success
	return nil
}
