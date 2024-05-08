package main

import (
	// Package imports
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/newsapi"
)

/////////////////////////////////////////////////////////////////////
// REGISTER FUNCTIONS

func NewsAPIFlags(flags *Flags) {
	flags.String("news-api-key", "${NEWSAPI_KEY}", "NewsAPI key")
	flags.String("q", "", "Search query")
}

func NewsAPIRegister(cmd []Client, opts []client.ClientOpt, flags *Flags) ([]Client, error) {
	newsapi, err := newsapi.New(flags.GetString("news-api-key"), opts...)
	if err != nil {
		return nil, err
	}

	// Register commands
	cmd = append(cmd, Client{
		ns: "newsapi",
		cmd: []Command{
			{Name: "sources", Description: "Return news sources", MinArgs: 2, MaxArgs: 2, Fn: newsAPISources(newsapi, flags)},
			{Name: "headlines", Description: "Return news headlines", MinArgs: 2, MaxArgs: 2, Fn: newsAPIHeadlines(newsapi, flags)},
			{Name: "articles", Description: "Return news articles", MinArgs: 2, MaxArgs: 2, Fn: newsAPIArticles(newsapi, flags)},
		},
	})

	// Return success
	return cmd, nil
}

/////////////////////////////////////////////////////////////////////
// API CALL FUNCTIONS

func newsAPISources(client *newsapi.Client, flags *Flags) CommandFn {
	return func() error {
		if sources, err := client.Sources(); err != nil {
			return err
		} else {
			return flags.Write(sources)
		}
	}
}

func newsAPIHeadlines(client *newsapi.Client, flags *Flags) CommandFn {
	return func() error {
		opts := []newsapi.Opt{}
		if q := flags.GetString("q"); q != "" {
			opts = append(opts, newsapi.OptQuery(q))
		}
		if articles, err := client.Headlines(opts...); err != nil {
			return err
		} else {
			return flags.Write(articles)
		}
	}
}

func newsAPIArticles(client *newsapi.Client, flags *Flags) CommandFn {
	return func() error {
		opts := []newsapi.Opt{}
		if q := flags.GetString("q"); q != "" {
			opts = append(opts, newsapi.OptQuery(q))
		}
		if articles, err := client.Articles(opts...); err != nil {
			return err
		} else {
			return flags.Write(articles)
		}
	}
}
