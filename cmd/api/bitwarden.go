package main

import (
	"os"
	"path/filepath"

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/bitwarden"
	"github.com/mutablelogic/go-client/pkg/bitwarden/schema"

	// Namespace import
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	bwName    = "bitwarden"
	bwDirPerm = 0700
)

var (
	bwClient                   *bitwarden.Client
	bwClientId, bwClientSecret string
	bwConfigDir, bwCacheDir    string
	bwForce                    bool
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func bwRegister(flags *Flags) {
	// Register flags required
	flags.String(bwName, "bitwarden-client-id", "${BW_CLIENTID}", "Client ID")
	flags.String(bwName, "bitwarden-client-secret", "${BW_CLIENTSECRET}", "Client Secret")
	flags.Bool(bwName, "force", false, "Force login or sync to Bitwarden, even if existing token or data is valid")
	//	flags.String(bwName, "bitwarden-device-id", "${BW_DEVICEID}", "Device Identifier")
	//	flags.String(bwName, "bitwarden-password", "${BW_PASSWORD}", "Master Password")

	// Register commands
	flags.Register(Cmd{
		Name:        bwName,
		Description: "Interact with the Bitwarden API",
		Parse:       bwParse,
		Fn: []Fn{
			{Name: "login", Description: "Login to Bitwarden", Call: bwLogin},
			{Name: "sync", Description: "Sync items from Bitwarden", Call: bwSync},
			{Name: "folders", Description: "Retrieve folders", Call: bwFolders},
		},
	})
}

func bwParse(flags *Flags, opts ...client.ClientOpt) error {
	if client, err := bitwarden.New(opts...); err != nil {
		return err
	} else {
		bwClient = client
	}

	// Get config directory
	if config, err := os.UserConfigDir(); err != nil {
		return err
	} else {
		bwConfigDir = filepath.Join(config, bwName)
		if err := os.MkdirAll(bwConfigDir, bwDirPerm); err != nil {
			return err
		}
	}
	// Get cache directory
	if cache, err := os.UserCacheDir(); err != nil {
		return err
	} else {
		bwCacheDir = filepath.Join(cache, bwName)
		if err := os.MkdirAll(bwCacheDir, bwDirPerm); err != nil {
			return err
		}
	}

	// Set client ID and secret
	bwClientId = flags.GetString("bitwarden-client-id")
	bwClientSecret = flags.GetString("bitwarden-client-secret")
	if bwClientId == "" || bwClientSecret == "" {
		return ErrBadParameter.With("Missing -bitwarden-client-id or -bitwarden-client-secret argument")
	}

	// Get the force flag
	bwForce = flags.GetBool("force")

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// API METHODS

func bwLogin(w *tablewriter.TableWriter) error {
	// Load session or create a new one
	session, err := bwReadSession()
	if err != nil {
		return err
	}

	// Login options
	opts := []bitwarden.SessionOpt{
		bitwarden.OptCredentials(bwClientId, bwClientSecret),
	}
	if session.Device == nil {
		opts = append(opts, bitwarden.OptDevice(bitwarden.Device{
			Name: bwName,
		}))
	}
	if bwForce {
		opts = append(opts, bitwarden.OptForce())
	}

	// Perform the login
	if err := bwClient.Login(session, opts...); err != nil {
		return err
	}

	// Save session
	if err := bwWriteSession(session); err != nil {
		return err
	}

	// Print out session
	w.Write(session)

	// Return success
	return nil
}

func bwSync(w *tablewriter.TableWriter) error {
	// Load session or create a new one
	session, err := bwReadSession()
	if err != nil {
		return err
	}
	// If the session is not valid, then return an error
	if !session.IsValid() {
		return ErrOutOfOrder.With("Session is not valid, login first")
	}
	// Perform the sync
	sync, err := bwClient.Sync(session)
	if err != nil {
		return err
	} else if err := bwWrite("profile.json", sync.Profile); err != nil {
		return err
	} else if err := bwWrite("folders.json", sync.Folders); err != nil {
		return err
	} else if err := bwWrite("ciphers.json", sync.Ciphers); err != nil {
		return err
	} else if err := bwWrite("domains.json", sync.Domains); err != nil {
		return err
	}

	// Output the profile
	w.Write(sync.Profile)

	// Return success
	return nil
}

func bwFolders(w *tablewriter.TableWriter) error {
	// Load session or create a new one
	session, err := bwReadSession()
	if err != nil {
		return err
	}

	// If the session is not valid, then return an error
	if !session.IsValid() {
		return ErrOutOfOrder.With("Session is not valid, login first")
	}

	// Read the folders
	folders := schema.Folders{}
	if err := bwRead("folders.json", &folders); err != nil {
		return err
	}

	// Decrypt the folders from the session
	for i, folder := range folders {
		if decrypted, err := folder.Decrypt(session); err != nil {
			return err
		} else {
			folders[i] = decrypted
		}
	}

	// Output the folders
	w.Write(folders)

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// OTHER

func bwReadSession() (*bitwarden.Session, error) {
	result := new(bitwarden.Session)
	filename := filepath.Join(bwConfigDir, "session.json")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Return a new, empty session
		return result, nil
	}

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read and return the session
	return result, result.Read(file)
}

func bwWriteSession(session *bitwarden.Session) error {
	return bwWrite("session.json", session)
}

func bwWrite(filename string, obj schema.ReaderWriter) error {
	path := filepath.Join(bwConfigDir, filename)
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	// Write the object and return any errors
	return obj.Write(w)
}

func bwRead(filename string, obj schema.ReaderWriter) error {
	path := filepath.Join(bwConfigDir, filename)
	r, err := os.Open(path)
	if err != nil {
		return err
	}
	defer r.Close()

	// Write the object and return any errors
	return obj.Read(r)
}
