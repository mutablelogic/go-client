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
	Name     string
	Username string
	URI      string
	Folder   string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	bwName    = "bitwarden"
	bwDirPerm = 0700
)

var (
	bwClient                   *bitwarden.Client
	bwClientId, bwClientSecret string
	bwPassword                 string
	bwConfigDir                string
	bwForce                    bool
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
			{Name: "sync", Description: "Sync items from Bitwarden", Call: bwSync},
			{Name: "folders", Description: "Retrieve folders", Call: bwFolders},
			{Name: "logins", Description: "Retrieve login items", Call: bwLogins},
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

	// Set defaults
	bwClientId = flags.GetString("bitwarden-client-id")
	bwClientSecret = flags.GetString("bitwarden-client-secret")
	if bwClientId == "" || bwClientSecret == "" {
		return ErrBadParameter.With("Missing -bitwarden-client-id or -bitwarden-client-secret argument")
	}
	bwForce = flags.GetBool("force")
	bwPassword = flags.GetString("bitwarden-password")

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// API METHODS

func bwAuth(w *tablewriter.TableWriter) error {
	// Load session or create a new one
	session, err := bwReadSession()
	if err != nil {
		return err
	}

	// Login options
	opts := []bitwarden.LoginOpt{
		bitwarden.OptCredentials(bwClientId, bwClientSecret),
	}
	if session.Device == nil {
		opts = append(opts, bitwarden.OptDevice(schema.Device{
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
	// Load the profile
	profile, err := bwReadProfile()
	if err != nil {
		return err
	}

	// Load session or create a new one
	session, err := bwReadSession()
	if err != nil {
		return err
	}

	// Make an encryption key
	if bwPassword == "" {
		if v, err := bwReadPasswordFromTerminal(); err != nil {
			return err
		} else {
			bwPassword = v
		}
	}
	if err := session.CacheKey(profile.Key, profile.Email, bwPassword); err != nil {
		return err
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
			folders[i] = decrypted.(*schema.Folder)
		}
	}

	// Output the folders
	w.Write(folders)

	// Return success
	return nil
}

func bwLogins(w *tablewriter.TableWriter) error {
	// Load the profile
	profile, err := bwReadProfile()
	if err != nil {
		return err
	}

	// Load session or create a new one
	session, err := bwReadSession()
	if err != nil {
		return err
	}

	// Make an encryption key
	if bwPassword == "" {
		if v, err := bwReadPasswordFromTerminal(); err != nil {
			return err
		} else {
			bwPassword = v
		}
	}
	if err := session.CacheKey(profile.Key, profile.Email, bwPassword); err != nil {
		return err
	}

	// Read the ciphers
	ciphers := schema.Ciphers{}
	result := []bwCipher{}
	if err := bwRead("ciphers.json", &ciphers); err != nil {
		return err
	}
	// Decrypt the ciphers from the session
	for _, cipher := range ciphers {
		if cipher.Type != schema.CipherTypeLogin {
			continue
		}
		if decrypted, err := cipher.Decrypt(session); err != nil {
			return err
		} else {
			result = append(result, bwCipher{
				Name:     decrypted.(*schema.Cipher).Name,
				Username: decrypted.(*schema.Cipher).Login.Username,
				URI:      decrypted.(*schema.Cipher).Login.URI,
				Folder:   decrypted.(*schema.Cipher).FolderId,
			})
		}
	}

	// Output the ciphers
	w.Write(result)

	// Return success
	return nil
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

func bwReadProfile() (*schema.Profile, error) {
	result := schema.NewProfile()
	filename := filepath.Join(bwConfigDir, "profile.json")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Return an error
		return nil, ErrNotFound.With("Profile not found")
	} else if err != nil {
		return nil, err
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

func bwReadSession() (*schema.Session, error) {
	result := schema.NewSession()
	filename := filepath.Join(bwConfigDir, "session.json")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Return a new, empty session
		return result, nil
	} else if err != nil {
		return nil, err
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

func bwWriteSession(session *schema.Session) error {
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
