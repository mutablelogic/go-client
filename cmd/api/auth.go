package main

import (
	"context"
	"time"

	// Packages
	tablewriter "github.com/djthorpe/go-tablewriter"
	client "github.com/mutablelogic/go-client"
	auth "github.com/mutablelogic/go-server/pkg/handler/auth/client"
)

var (
	authClient   *auth.Client
	authName     = "tokenauth"
	authDuration time.Duration
)

func authRegister(flags *Flags) {
	// Register flags
	flags.String(authName, "tokenauth-endpoint", "${TOKENAUTH_ENDPOINT}", "tokenauth endpoint (ie, http://host/api/auth/)")
	flags.String(authName, "tokenauth-token", "${TOKENAUTH_TOKEN}", "tokenauth token")
	flags.Duration(authName, "expiry", 0, "token expiry duration")

	// Register commands
	flags.Register(Cmd{
		Name:        authName,
		Description: "Manage token authentication",
		Parse:       authParse,
		Fn: []Fn{
			// Default caller
			{Call: authList, Description: "List authentication tokens"},
			{Name: "list", Call: authList, Description: "List authentication tokens"},
			{Name: "create", Call: authCreate, Description: "Create a token", MinArgs: 1},
			{Name: "delete", Call: authDelete, Description: "Delete a token", MinArgs: 1, MaxArgs: 1},
		},
	})
}

func authParse(flags *Flags, opts ...client.ClientOpt) error {
	endpoint := flags.GetString("tokenauth-endpoint")
	if token := flags.GetString("tokenauth-token"); token != "" {
		opts = append(opts, client.OptReqToken(client.Token{
			Scheme: "Bearer",
			Value:  token,
		}))
	}

	if duration := flags.GetString("expiry"); duration != "" {
		if d, err := time.ParseDuration(duration); err != nil {
			return err
		} else {
			authDuration = d
		}
	}

	if client, err := auth.New(endpoint, opts...); err != nil {
		return err
	} else {
		authClient = client
	}
	return nil
}

func authList(_ context.Context, w *tablewriter.Writer, _ []string) error {
	tokens, err := authClient.List()
	if err != nil {
		return err
	}
	return w.Write(tokens)
}

func authCreate(_ context.Context, w *tablewriter.Writer, params []string) error {
	name := params[0]
	scopes := params[1:]
	token, err := authClient.Create(name, authDuration, scopes...)
	if err != nil {
		return err
	}
	return w.Write(token)
}

func authDelete(ctx context.Context, w *tablewriter.Writer, params []string) error {
	name := params[0]
	err := authClient.Delete(name)
	if err != nil {
		return err
	}
	return authList(ctx, w, nil)
}
