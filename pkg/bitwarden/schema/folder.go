package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Folders []*Folder

type Folder struct {
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	RevisionDate time.Time `json:"revisionDate"`
	Object       string    `json:"object"`
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
	fmt.Println("TODO: Decrypt")
	return f, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f Folder) String() string {
	data, _ := json.MarshalIndent(f, "", "  ")
	return string(data)
}
