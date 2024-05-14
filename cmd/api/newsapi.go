package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/newsapi"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	newsapiName     = "newsapi"
	newsapiClient   *newsapi.Client
	newsapiCategory string
	newsapiLanguage string
	newsapiCountry  string
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newsapiRegister(flags *Flags) {
	// Register flags required
	flags.String(newsapiName, "newsapi-key", "${NEWSAPI_KEY}", "API Key")
	flags.String(newsapiName, "category", "", "News category: business, entertainment, general, health, science, sports, technology")
	flags.String(newsapiName, "language", "", "ISO 639 language code")
	flags.String(newsapiName, "country", "", "ISO 3166 country code")

	flags.Register(Cmd{
		Name:        newsapiName,
		Description: "Obtain news headlines from https://newsapi.org/",
		Parse:       inewsapiParse,
		Fn: []Fn{
			{Name: "sources", Call: newsapiSources, Description: "Return sources of news"},
			{Name: "headlines", Call: newsapiHeadlines, Description: "Return top headlines from news sources"},
			{Name: "search", Call: newsapiArticles, Description: "Return articles from news sources with search term", MaxArgs: 1},
		},
	})
}

func inewsapiParse(flags *Flags, opts ...client.ClientOpt) error {
	apiKey := flags.GetString("newsapi-key")
	if apiKey == "" {
		return fmt.Errorf("missing -newsapi-key flag")
	}
	if client, err := newsapi.New(flags.GetString("newsapi-key"), opts...); err != nil {
		return err
	} else {
		newsapiClient = client
	}

	// Set category
	newsapiCategory = strings.ToLower(flags.GetString("category"))
	newsapiLanguage = strings.ToLower(flags.GetString("language"))
	newsapiCountry = strings.ToLower(flags.GetString("country"))

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func newsapiSources(_ context.Context, w *tablewriter.Writer, _ []string) error {
	// Set options
	opts := []newsapi.Opt{}
	if newsapiCategory != "" {
		opts = append(opts, newsapi.OptCategory(newsapiCategory))
	}
	if newsapiLanguage != "" {
		opts = append(opts, newsapi.OptLanguage(newsapiLanguage))
	}
	if newsapiCountry != "" {
		opts = append(opts, newsapi.OptCountry(newsapiCountry))
	}

	// Request -> Response
	sources, err := newsapiClient.Sources(opts...)
	if err != nil {
		return err
	}

	// Write table
	return w.Write(sources)
}

func newsapiHeadlines(_ context.Context, w *tablewriter.Writer, _ []string) error {
	// Set options
	opts := []newsapi.Opt{}
	if newsapiCategory != "" {
		opts = append(opts, newsapi.OptCategory(newsapiCategory))
	}
	if newsapiLanguage != "" {
		opts = append(opts, newsapi.OptLanguage(newsapiLanguage))
	}
	if newsapiCountry != "" {
		opts = append(opts, newsapi.OptCountry(newsapiCountry))
	}

	// Request -> Response
	articles, err := newsapiClient.Headlines(opts...)
	if err != nil {
		return err
	}

	// Write table
	return w.Write(articles)
}

func newsapiArticles(_ context.Context, w *tablewriter.Writer, args []string) error {
	// Set options
	opts := []newsapi.Opt{}
	if newsapiCategory != "" {
		opts = append(opts, newsapi.OptCategory(newsapiCategory))
	}
	if newsapiLanguage != "" {
		opts = append(opts, newsapi.OptLanguage(newsapiLanguage))
	}
	if newsapiCountry != "" {
		opts = append(opts, newsapi.OptCountry(newsapiCountry))
	}
	// Set query
	if len(args) > 0 {
		q := strconv.Quote(args[0])
		opts = append(opts, newsapi.OptQuery(q))
	}

	// Request -> Response
	articles, err := newsapiClient.Articles(opts...)
	if err != nil {
		return err
	}

	// Write table
	return w.Write(articles)
}
