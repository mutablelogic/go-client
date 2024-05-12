package schema

import (
	"encoding/json"
	"io"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Folders []*Folder

type Folder struct {
	Id           string    `json:"id,width:36"`
	Name         string    `json:"name"` // Encrypted
	RevisionDate time.Time `json:"revisionDate,width:29"`
	Object       string    `json:"object,width:6"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Read a list of folders
func (f *Folders) Read(r io.Reader) error {
	return json.NewDecoder(r).Decode(f)
}

// Write a list of folders
func (f *Folders) Write(w io.Writer) error {
	return json.NewEncoder(w).Encode(f)
}

// Decrypt a folder
func (f Folder) Decrypt(s *Session) (Crypter, error) {
	result := &f
	if value, err := s.DecryptStr(result.Name); err != nil {
		return nil, err
	} else {
		result.Name = value
	}
	return result, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f Folder) String() string {
	data, _ := json.MarshalIndent(f, "", "  ")
	return string(data)
}
