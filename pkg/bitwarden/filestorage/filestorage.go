package filestorage

import (
	"os"
	"path/filepath"

	// Packages

	schema "github.com/mutablelogic/go-client/pkg/bitwarden/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type fileStorage struct {
	cachePath string
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	fileNameSession = "session.json"
	fileNameProfile = "profile.json"
	fileNameFolders = "folders.json"
	fileNameCiphers = "ciphers.json"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a file storage object, which will store files in the cache path
// which must be a directory and exist
func New(path string) (*fileStorage, error) {
	return &fileStorage{
		cachePath: path,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Read the session from storage for a session, returns nil if there is no session
func (f *fileStorage) ReadSession() (*schema.Session, error) {
	session := schema.NewSession()

	// Read the session file - return nil if it doesn't exist
	fileName := filepath.Join(f.cachePath, fileNameSession)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil, nil
	}
	r, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	if err := session.Read(r); err != nil {
		return nil, err
	}

	// Return the session
	return session, nil
}

// Write the session to storage
func (f *fileStorage) WriteSession(s *schema.Session) error {
	// Create the session file
	fileName := filepath.Join(f.cachePath, fileNameSession)
	w, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer w.Close()

	// Write the session
	return s.Write(w)
}

// Read the profile from storage
func (f *fileStorage) ReadProfile() (*schema.Profile, error) {
	profile := schema.NewProfile()

	// Read the profile file - return nil if it doesn't exist
	fileName := filepath.Join(f.cachePath, fileNameProfile)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil, nil
	}
	r, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	if err := profile.Read(r); err != nil {
		return nil, err
	}

	// Return the session
	return profile, nil
}

// Write the profile to storage
func (f *fileStorage) WriteProfile(p *schema.Profile) error {
	// Create the profile
	fileName := filepath.Join(f.cachePath, fileNameProfile)
	w, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer w.Close()
	return p.Write(w)
}

// Write the folders to storage
func (f *fileStorage) WriteFolders(v schema.Folders) error {
	// Create the folders
	fileName := filepath.Join(f.cachePath, fileNameFolders)
	w, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer w.Close()
	return v.Write(w)
}

// Write the ciphers to storage
func (f *fileStorage) WriteCiphers(v schema.Ciphers) error {
	// Create the ciphers
	fileName := filepath.Join(f.cachePath, fileNameCiphers)
	w, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer w.Close()
	return v.Write(w)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - READ CIPHERS AND FOLDERS

// Read ciphers and return an iterator
func (f *fileStorage) ReadCiphers(profile *schema.Profile) (schema.Iterator[*schema.Cipher], error) {
	// Profile is a required argument
	if profile == nil {
		return nil, ErrBadParameter.With("missing profile")
	}

	// Read the ciphers file
	ciphers := schema.Ciphers{}
	fileName := filepath.Join(f.cachePath, fileNameCiphers)
	r, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	if err := ciphers.Read(r); err != nil {
		return nil, err
	}

	// Return an iterator
	return schema.NewIterator[*schema.Cipher](profile, ciphers), nil
}

// Read folders and return an iterator
func (f *fileStorage) ReadFolders(profile *schema.Profile) (schema.Iterator[*schema.Folder], error) {
	// Profile is a required argument
	if profile == nil {
		return nil, ErrBadParameter.With("missing profile")
	}

	// Read the folders file
	folders := schema.Folders{}
	fileName := filepath.Join(f.cachePath, fileNameFolders)
	r, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	if err := folders.Read(r); err != nil {
		return nil, err
	}

	// Return an iterator
	return schema.NewIterator[*schema.Folder](profile, folders), nil
}
