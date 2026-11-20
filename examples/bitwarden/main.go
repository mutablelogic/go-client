// Command bitwarden is an example of go-client: a client of the Bitwarden
// API, which signs in with an API key, syncs the vault, and decrypts it
// with the master password, to list folders and items, and print a
// password.
//
//	export BW_CLIENTID=... BW_CLIENTSECRET=...
//	go run ./examples/bitwarden items github
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/tabwriter"

	// Packages
	kong "github.com/alecthomas/kong"
	client "github.com/mutablelogic/go-client"
	transport "github.com/mutablelogic/go-client/pkg/transport"
	term "golang.org/x/term"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	Server         string `env:"BW_SERVER" default:"us" help:"Server: us, eu, or the URL of a self-hosted server"`
	ClientID       string `env:"BW_CLIENTID" required:"" help:"API key client id"`
	ClientSecret   string `env:"BW_CLIENTSECRET" required:"" help:"API key client secret"`
	MasterPassword string `env:"BW_PASSWORD" name:"password" help:"Master password, which is prompted for when it isn't set"`
	Cache          string `type:"path" help:"File of the session, by default in the user's cache directory"`
	Debug          bool   `help:"Trace requests and responses, without their bodies"`
	Verbose        bool   `help:"Trace the bodies of requests and responses too, which include the client secret and access token"`

	Login    LoginCmd    `cmd:"" help:"Sign in, even when the session is still valid"`
	Folders  FoldersCmd  `cmd:"" help:"List folders"`
	Items    ItemsCmd    `cmd:"" help:"List items, whose name, username or URI contains the filter"`
	Password PasswordCmd `cmd:"" help:"Print an item's password"`
}

type LoginCmd struct{}

type FoldersCmd struct{}

type ItemsCmd struct {
	Filter string `arg:"" optional:""`
}

type PasswordCmd struct {
	Name string `arg:"" help:"Item name, or ID"`
}

///////////////////////////////////////////////////////////////////////////////
// MAIN

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	var cli CLI
	kctx := kong.Parse(&cli, kong.Name("bitwarden"), kong.Description("Bitwarden API client example"), kong.UsageOnError(), kong.BindTo(ctx, (*context.Context)(nil)))
	if err := kctx.Run(&cli); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (cmd *LoginCmd) Run(ctx context.Context, cli *CLI) error {
	c, err := cli.client()
	if err != nil {
		return err
	}
	return c.Login(ctx, true)
}

func (cmd *FoldersCmd) Run(ctx context.Context, cli *CLI) error {
	vault, password, err := cli.sync(ctx)
	if err != nil {
		return err
	}
	folders, err := vault.DecryptFolders(password)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME")
	for _, f := range folders {
		fmt.Fprintf(w, "%s\t%s\n", f.ID, f.Name)
	}
	return w.Flush()
}

func (cmd *ItemsCmd) Run(ctx context.Context, cli *CLI) error {
	vault, password, err := cli.sync(ctx)
	if err != nil {
		return err
	}
	items, err := vault.DecryptItems(password)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "TYPE\tNAME\tUSERNAME\tURI\tFOLDER")
	for _, item := range items {
		if item.match(cmd.Filter) {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", item.Type, item.Name, item.Username, item.URI, item.Folder)
		}
	}
	return w.Flush()
}

func (cmd *PasswordCmd) Run(ctx context.Context, cli *CLI) error {
	vault, password, err := cli.sync(ctx)
	if err != nil {
		return err
	}
	items, err := vault.DecryptItems(password)
	if err != nil {
		return err
	}
	var found []Item
	for _, item := range items {
		if item.ID == cmd.Name || item.Name == cmd.Name {
			found = append(found, item)
		}
	}
	switch len(found) {
	case 0:
		return fmt.Errorf("no item %q", cmd.Name)
	case 1:
		// The password is the command's output, to stdout, such as for a
		// pipe, not a log
		_, err := os.Stdout.WriteString(found[0].Password + "\n")
		return err
	default:
		return fmt.Errorf("%d items are named %q, use an ID", len(found), cmd.Name)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// client returns the client, with its session in the cache
func (cli *CLI) client() (*Client, error) {
	path := cli.Cache
	if path == "" {
		dir, err := os.UserCacheDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(dir, "go-client", "bitwarden.json")
	}
	var opts []client.ClientOpt
	if cli.Debug || cli.Verbose {
		opts = append(opts, client.OptTransport(func(parent http.RoundTripper) http.RoundTripper {
			return transport.NewLogging(os.Stderr, parent, cli.Verbose)
		}))
	}
	return New(cli.Server, cli.ClientID, cli.ClientSecret, path, opts...)
}

// sync returns the vault, and the master password, which is prompted for
// when it isn't set
func (cli *CLI) sync(ctx context.Context) (*Vault, string, error) {
	c, err := cli.client()
	if err != nil {
		return nil, "", err
	}
	vault, err := c.Sync(ctx)
	if err != nil {
		return nil, "", err
	}
	password := cli.MasterPassword
	if password == "" {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return nil, "", errors.New("the master password is required: set BW_PASSWORD")
		}
		fmt.Fprint(os.Stderr, "Master password: ")
		data, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return nil, "", err
		}
		password = string(data)
	}
	return vault, password, nil
}
