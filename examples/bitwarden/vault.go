package main

import (
	"errors"
	"fmt"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Vault is an account's vault, as it's synced, which is encrypted
type Vault struct {
	Profile struct {
		Email         string `json:"email"`
		Key           string `json:"key"`
		PrivateKey    string `json:"privateKey"`
		Organizations []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Key  string `json:"key"`
		} `json:"organizations"`
	} `json:"profile"`
	Folders []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"folders"`
	Ciphers []encItem `json:"ciphers"`
	kdf     kdf
}

// encItem is an item of the vault, which is encrypted
type encItem struct {
	ID             string  `json:"id"`
	Type           int     `json:"type"`
	OrganizationID string  `json:"organizationId"`
	FolderID       string  `json:"folderId"`
	Key            string  `json:"key"` // the item's own key, when it has one
	Name           string  `json:"name"`
	Notes          string  `json:"notes"`
	DeletedDate    *string `json:"deletedDate"`
	Login          *struct {
		Username string `json:"username"`
		Password string `json:"password"`
		URIs     []struct {
			URI string `json:"uri"`
		} `json:"uris"`
	} `json:"login"`
}

// Folder is a folder, which is decrypted
type Folder struct {
	ID, Name string
}

// Item is an item, which is decrypted
type Item struct {
	ID, Type, Folder, Organization string
	Name, Username, Password, URI  string
	Notes                          string
}

// unlocked is a vault's keys, once it's unlocked
type unlocked struct {
	user          key
	organizations map[string]key
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var itemTypes = map[int]string{1: "login", 2: "note", 3: "card", 4: "identity", 5: "ssh-key"}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// DecryptFolders returns the folders, decrypted with the master password
func (v *Vault) DecryptFolders(password string) ([]Folder, error) {
	keys, err := v.unlock(password)
	if err != nil {
		return nil, err
	}
	result := make([]Folder, 0, len(v.Folders))
	for _, f := range v.Folders {
		name, err := keys.user.decryptString(f.Name)
		if err != nil {
			return nil, fmt.Errorf("folder %s: %w", f.ID, err)
		}
		result = append(result, Folder{ID: f.ID, Name: name})
	}
	return result, nil
}

// DecryptItems returns the items which aren't deleted, decrypted with the master
// password, with their folders' and organizations' names
func (v *Vault) DecryptItems(password string) ([]Item, error) {
	keys, err := v.unlock(password)
	if err != nil {
		return nil, err
	}
	folders := make(map[string]string, len(v.Folders))
	for _, f := range v.Folders {
		if folders[f.ID], err = keys.user.decryptString(f.Name); err != nil {
			return nil, fmt.Errorf("folder %s: %w", f.ID, err)
		}
	}
	organizations := make(map[string]string, len(v.Profile.Organizations))
	for _, o := range v.Profile.Organizations {
		organizations[o.ID] = o.Name
	}
	result := make([]Item, 0, len(v.Ciphers))
	for _, c := range v.Ciphers {
		if c.DeletedDate != nil {
			continue
		}
		item, err := keys.item(c)
		if err != nil {
			return nil, fmt.Errorf("item %s: %w", c.ID, err)
		}
		item.Folder = folders[c.FolderID]
		item.Organization = organizations[c.OrganizationID]
		result = append(result, item)
	}
	return result, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// unlock returns the vault's keys, from the master password: the user key,
// and the keys of the organizations
func (v *Vault) unlock(password string) (*unlocked, error) {
	master, err := masterKey(v.kdf, v.Profile.Email, password)
	if err != nil {
		return nil, err
	}
	stretched, err := stretch(master)
	if err != nil {
		return nil, err
	}
	data, err := stretched.decrypt(v.Profile.Key)
	if errors.Is(err, errMAC) {
		return nil, errors.New("the master password is wrong")
	} else if err != nil {
		return nil, err
	}
	user, err := newKey(data)
	if err != nil {
		return nil, err
	}
	keys := &unlocked{user: user, organizations: make(map[string]key)}
	if len(v.Profile.Organizations) == 0 {
		return keys, nil
	}
	private, err := user.privateKey(v.Profile.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("private key: %w", err)
	}
	for _, o := range v.Profile.Organizations {
		data, err := decryptRSA(private, o.Key)
		if err != nil {
			return nil, fmt.Errorf("organization %s: %w", o.ID, err)
		}
		if keys.organizations[o.ID], err = newKey(data); err != nil {
			return nil, fmt.Errorf("organization %s: %w", o.ID, err)
		}
	}
	return keys, nil
}

// item decrypts an item, with its own key, when it has one, which the
// user's, or its organization's, key decrypts
func (u *unlocked) item(c encItem) (Item, error) {
	k := u.user
	if c.OrganizationID != "" {
		var ok bool
		if k, ok = u.organizations[c.OrganizationID]; !ok {
			return Item{}, fmt.Errorf("no key of organization %s", c.OrganizationID)
		}
	}
	if c.Key != "" {
		data, err := k.decrypt(c.Key)
		if err != nil {
			return Item{}, err
		}
		if k, err = newKey(data); err != nil {
			return Item{}, err
		}
	}
	item := Item{ID: c.ID, Type: itemTypes[c.Type]}
	var err error
	decrypt := func(s string) string {
		value, e := k.decryptString(s)
		err = errors.Join(err, e)
		return value
	}
	item.Name = decrypt(c.Name)
	item.Notes = decrypt(c.Notes)
	if c.Login != nil {
		item.Username = decrypt(c.Login.Username)
		item.Password = decrypt(c.Login.Password)
		if len(c.Login.URIs) > 0 {
			item.URI = decrypt(c.Login.URIs[0].URI)
		}
	}
	return item, err
}

// match returns true when an item's name, username or URI contains the
// filter, ignoring case, or the filter is empty
func (i Item) match(filter string) bool {
	filter = strings.ToLower(filter)
	for _, s := range []string{i.Name, i.Username, i.URI} {
		if strings.Contains(strings.ToLower(s), filter) {
			return true
		}
	}
	return false
}
