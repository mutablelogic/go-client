package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/bitwarden"
	"github.com/mutablelogic/go-client/pkg/bitwarden/schema"
	"golang.org/x/term"

	// Namespace import
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type bwCipher struct {
	Name     string `json:"name,wrap"`
	Username string `json:"username,width:30"`
	URI      string `json:"uri,width:40"`
	Folder   string `json:"folder,width:36"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	bwName    = "bitwarden"
	bwDirPerm = 0700
)

var (
	bwClient    *bitwarden.Client
	bwPassword  string
	bwConfigDir string
	bwForce     bool
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func bwRegister(flags *Flags) {
	// Register flags required
	flags.String(bwName, "bitwarden-client-id", "${BW_CLIENTID}", "Client ID")
	flags.String(bwName, "bitwarden-client-secret", "${BW_CLIENTSECRET}", "Client Secret")
	flags.String(bwName, "bitwarden-password", "${BW_PASSWORD}", "Master password")
	flags.Bool(bwName, "force", false, "Force login or sync to Bitwarden, even if existing token or data is valid")

	// Register commands
	flags.Register(Cmd{
		Name:        bwName,
		Description: "Interact with the Bitwarden API",
		Parse:       bwParse,
		Fn: []Fn{
			{Name: "auth", Description: "Authenticate with Bitwarden", Call: bwAuth},
			{Name: "folders", Description: "Retrieve folders", Call: bwFolders},
			{Name: "logins", Description: "Retrieve login items", Call: bwLogins, MaxArgs: 1},
			{Name: "password", Description: "Print a password to stdout", Call: bwGetPassword, MinArgs: 1, MaxArgs: 1},
		},
	})
}

func bwParse(flags *Flags, opts ...client.ClientOpt) error {
	// Get config directory
	if config, err := os.UserConfigDir(); err != nil {
		return err
	} else {
		bwConfigDir = filepath.Join(config, bwName)
		if err := os.MkdirAll(bwConfigDir, bwDirPerm); err != nil {
			return err
		}
	}

	// Set defaults
	clientId := flags.GetString("bitwarden-client-id")
	clientSecret := flags.GetString("bitwarden-client-secret")
	if clientId == "" || clientSecret == "" {
		return ErrBadParameter.With("Missing -bitwarden-client-id or -bitwarden-client-secret argument")
	}
	bwForce = flags.GetBool("force")
	bwPassword = flags.GetString("bitwarden-password")

	// Create the client
	opts = append(opts, bitwarden.OptCredentials(clientId, clientSecret))
	opts = append(opts, bitwarden.OptFileStorage(bwConfigDir))
	if client, err := bitwarden.New(opts...); err != nil {
		return err
	} else {
		bwClient = client
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// API METHODS

func bwAuth(w *tablewriter.Writer, _ []string) error {
	opts := []bitwarden.RequestOpt{}
	if bwForce {
		opts = append(opts, bitwarden.OptForce())
	}

	// Login
	if err := bwClient.Login(opts...); err != nil {
		return err
	}

	// Sync
	if profile, err := bwClient.Sync(opts...); err != nil {
		return err
	} else {
		return w.Write(profile)
	}
}

func bwFolders(w *tablewriter.Writer, _ []string) error {
	opts := []bitwarden.RequestOpt{}
	if bwForce {
		opts = append(opts, bitwarden.OptForce())
	}
	folders, err := bwClient.Folders(opts...)
	if err != nil {
		return err
	}

	// Decrypt the folders from the session
	var result []*schema.Folder
	for folder := folders.Next(); folder != nil; folder = folders.Next() {
		result = append(result, folders.Decrypt(folder))
	}
	return w.Write(result)
}

func bwLogins(w *tablewriter.Writer, _ []string) error {
	opts := []bitwarden.RequestOpt{}
	if bwForce {
		opts = append(opts, bitwarden.OptForce())
	}
	ciphers, err := bwClient.Ciphers(opts...)
	if err != nil {
		return err
	}

	// Decrypt the ciphers from the session
	var result []*schema.Cipher
	for cipher := ciphers.Next(); cipher != nil; cipher = ciphers.Next() {
		result = append(result, ciphers.Decrypt(cipher))
	}
	return w.Write(result)
}

func bwGetPassword(w *tablewriter.Writer, _ []string) error {
	return ErrNotImplemented
}

///////////////////////////////////////////////////////////////////////////////
// OTHER

func bwReadPasswordFromTerminal() (string, error) {
	stdin := int(os.Stdin.Fd())
	if !term.IsTerminal(stdin) {
		return "", ErrBadParameter.With("No password set and not running in terminal")
	}
	fmt.Fprintf(os.Stdout, "Enter password: ")
	defer func() {
		fmt.Fprintf(os.Stdout, "\n")
	}()
	if value, err := term.ReadPassword(stdin); err != nil {
		return "", err
	} else {
		return string(value), nil
	}
}
