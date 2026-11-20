package main

import (
	"context"
	"crypto/aes"
	gocipher "crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

const (
	email    = "Nobody@Example.com"
	password = "p4ssw0rd"
)

// encrypt encrypts with AES-256-CBC and an HMAC, as Bitwarden does
func encrypt(t *testing.T, k key, data []byte) string {
	t.Helper()
	n := aes.BlockSize - len(data)%aes.BlockSize
	for range n {
		data = append(data, byte(n))
	}
	iv := make([]byte, aes.BlockSize)
	rand.Read(iv)
	block, err := aes.NewCipher(k.Enc)
	require.NoError(t, err)
	out := make([]byte, len(data))
	gocipher.NewCBCEncrypter(block, iv).CryptBlocks(out, data)
	mac := hmac.New(sha256.New, k.MAC)
	mac.Write(iv)
	mac.Write(out)
	b64 := base64.StdEncoding.EncodeToString
	return fmt.Sprintf("2.%s|%s|%s", b64(iv), b64(out), b64(mac.Sum(nil)))
}

// randomKey returns a symmetric key, and its bytes
func randomKey(t *testing.T) (key, []byte) {
	data := make([]byte, 64)
	rand.Read(data)
	k, err := newKey(data)
	require.NoError(t, err)
	return k, data
}

// The master key of PBKDF2 is the same as it was
func Test_MasterKey(t *testing.T) {
	k, err := masterKey(kdf{Type: kdfPBKDF2, Iterations: 5000}, "nobody@example.com", password)
	require.NoError(t, err)
	assert.Equal(t, `"\x13\x88j`+"`"+`\x99m\xe3FA\x94\xee'\xf0\xb2\x1a!\xb6>\\)\xf4\xd5\xca#\xe5\x1b\xa6f5o{\xaa"`, fmt.Sprintf("%q", k))

	// The email address isn't case sensitive
	k2, err := masterKey(kdf{Type: kdfPBKDF2, Iterations: 5000}, email, password)
	require.NoError(t, err)
	assert.Equal(t, k, k2)

	// Argon2id
	k3, err := masterKey(kdf{Type: kdfArgon2id, Iterations: 3, Memory: 16, Parallelism: 2}, email, password)
	require.NoError(t, err)
	assert.Len(t, k3, 32)
	assert.NotEqual(t, k, k3)
	_, err = masterKey(kdf{Type: kdfArgon2id, Iterations: 3}, email, password)
	assert.Error(t, err)
	_, err = masterKey(kdf{Type: 9, Iterations: 3}, email, password)
	assert.Error(t, err)
}

func Test_Decrypt(t *testing.T) {
	k, _ := randomKey(t)
	s := encrypt(t, k, []byte("hello, world"))
	data, err := k.decrypt(s)
	require.NoError(t, err)
	assert.Equal(t, "hello, world", string(data))

	// Another key, or a value which is changed, isn't authentic
	other, _ := randomKey(t)
	_, err = other.decrypt(s)
	assert.ErrorIs(t, err, errMAC)
	e, err := parseEncString(s)
	require.NoError(t, err)
	e.Data[0] ^= 1
	b64 := base64.StdEncoding.EncodeToString
	_, err = k.decrypt(fmt.Sprintf("2.%s|%s|%s", b64(e.IV), b64(e.Data), b64(e.MAC)))
	assert.ErrorIs(t, err, errMAC)

	// Strings which aren't encrypted, or aren't of a supported type
	for _, s := range []string{"", "x", "0.abc|def", "2.abc", "2.!!|!!|!!"} {
		_, err := k.decrypt(s)
		assert.Error(t, err, s)
	}
}

func Test_Sync(t *testing.T) {
	// The account's keys
	kdf := kdf{Type: kdfPBKDF2, Iterations: 5000}
	master, err := masterKey(kdf, email, password)
	require.NoError(t, err)
	stretched, err := stretch(master)
	require.NoError(t, err)
	user, userBytes := randomKey(t)
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(private)
	require.NoError(t, err)
	org, orgBytes := randomKey(t)
	orgKey, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, &private.PublicKey, orgBytes, nil)
	require.NoError(t, err)
	item, itemBytes := randomKey(t)

	vault := map[string]any{
		"profile": map[string]any{
			"email":         email,
			"key":           encrypt(t, stretched, userBytes),
			"privateKey":    encrypt(t, user, der),
			"organizations": []any{map[string]any{"id": "o1", "name": "Family", "key": "4." + base64.StdEncoding.EncodeToString(orgKey)}},
		},
		"folders": []any{map[string]any{"id": "f1", "name": encrypt(t, user, []byte("Work"))}},
		"ciphers": []any{
			map[string]any{"id": "c1", "type": 1, "folderId": "f1", "name": encrypt(t, user, []byte("GitHub")), "login": map[string]any{
				"username": encrypt(t, user, []byte("octocat")),
				"password": encrypt(t, user, []byte("secret1")),
				"uris":     []any{map[string]any{"uri": encrypt(t, user, []byte("https://github.com"))}},
			}},
			// An item with its own key
			map[string]any{"id": "c2", "type": 1, "key": encrypt(t, user, itemBytes), "name": encrypt(t, item, []byte("Example")), "login": map[string]any{
				"password": encrypt(t, item, []byte("secret2")),
			}},
			// An organization's item
			map[string]any{"id": "c3", "type": 2, "organizationId": "o1", "name": encrypt(t, org, []byte("Wi-Fi")), "notes": encrypt(t, org, []byte("on the fridge"))},
			// A deleted item
			map[string]any{"id": "c4", "type": 1, "name": "not decrypted", "deletedDate": "2026-01-01T00:00:00Z"},
		},
	}

	// The server
	var tokens, syncs atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /identity/connect/token", func(w http.ResponseWriter, r *http.Request) {
		tokens.Add(1)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
		assert.Equal(t, "id", r.FormValue("client_id"))
		assert.Equal(t, "secret", r.FormValue("client_secret"))
		assert.NotEmpty(t, r.FormValue("deviceIdentifier"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"access_token": "token", "expires_in": 3600, "token_type": "Bearer", "Kdf": kdf.Type, "KdfIterations": kdf.Iterations})
	})
	mux.HandleFunc("GET /api/sync", func(w http.ResponseWriter, r *http.Request) {
		syncs.Add(1)
		if r.Header.Get("Authorization") != "Bearer token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(vault)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// Sync, which signs in first, and decrypt
	path := filepath.Join(t.TempDir(), "session.json")
	c, err := New(server.URL, "id", "secret", path)
	require.NoError(t, err)
	v, err := c.Sync(context.Background())
	require.NoError(t, err)
	folders, err := v.DecryptFolders(password)
	require.NoError(t, err)
	assert.Equal(t, []Folder{{ID: "f1", Name: "Work"}}, folders)
	items, err := v.DecryptItems(password)
	require.NoError(t, err)
	assert.Equal(t, []Item{
		{ID: "c1", Type: "login", Folder: "Work", Name: "GitHub", Username: "octocat", Password: "secret1", URI: "https://github.com"},
		{ID: "c2", Type: "login", Name: "Example", Password: "secret2"},
		{ID: "c3", Type: "note", Organization: "Family", Name: "Wi-Fi", Notes: "on the fridge"},
	}, items)
	assert.True(t, items[0].match("GITHUB"))
	assert.False(t, items[0].match("gitlab"))

	// A wrong password
	_, err = v.DecryptItems("wrong")
	assert.ErrorContains(t, err, "master password is wrong")

	// The session is kept, so another client doesn't sign in again, unless
	// it's forced to
	c2, err := New(server.URL, "id", "secret", path)
	require.NoError(t, err)
	assert.Equal(t, c.session.Device, c2.session.Device)
	_, err = c2.Sync(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int32(1), tokens.Load())
	require.NoError(t, c2.Login(context.Background(), true))
	assert.Equal(t, int32(2), tokens.Load())

	// The session file is only the user's, even when it was another's
	require.NoError(t, os.Chmod(path, 0o644))
	require.NoError(t, c2.Login(context.Background(), true))
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	// The token isn't used for another account, or another server, which
	// signs in, with the same device
	c3, err := New(server.URL, "other", "secret", path)
	require.NoError(t, err)
	assert.Empty(t, c3.session.Token)
	assert.Equal(t, c.session.Device, c3.session.Device)
	c4, err := New(server.URL+"/", "id", "secret", path)
	require.NoError(t, err)
	assert.Empty(t, c4.session.Token)

	// A cancelled context is an error
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = c2.Sync(ctx)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, int32(2), syncs.Load())
}
