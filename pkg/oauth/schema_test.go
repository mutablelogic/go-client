package oauth

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
	oauth2 "golang.org/x/oauth2"
)

func TestValidateScopes_NoScopesSupported(t *testing.T) {
	// Server did not advertise scopes_supported — all scopes should be accepted.
	m := &OAuthMetadata{}
	assert.NoError(t, m.ValidateScopes("openid", "email", "anything"))
}

func TestValidateScopes_EmptyRequest(t *testing.T) {
	m := &OAuthMetadata{ScopesSupported: []string{"openid", "email"}}
	assert.NoError(t, m.ValidateScopes())
}

func TestValidateScopes_AllSupported(t *testing.T) {
	m := &OAuthMetadata{ScopesSupported: []string{"openid", "email", "profile"}}
	assert.NoError(t, m.ValidateScopes("openid", "email"))
}

func TestValidateScopes_OneUnsupported(t *testing.T) {
	m := &OAuthMetadata{ScopesSupported: []string{"openid", "email"}}
	err := m.ValidateScopes("openid", "offline_access")
	assert.EqualError(t, err, "unsupported scopes: offline_access")
}

func TestValidateScopes_MultipleUnsupported(t *testing.T) {
	m := &OAuthMetadata{ScopesSupported: []string{"openid"}}
	err := m.ValidateScopes("email", "profile", "openid")
	assert.ErrorContains(t, err, "email")
	assert.ErrorContains(t, err, "profile")
	assert.NotContains(t, err.Error(), "openid")
}

func TestValidateScopes_AllUnsupported(t *testing.T) {
	m := &OAuthMetadata{ScopesSupported: []string{"openid"}}
	err := m.ValidateScopes("email", "profile")
	assert.EqualError(t, err, "unsupported scopes: email, profile")
}

func TestValid_NilToken(t *testing.T) {
	creds := &OAuthCredentials{}
	assert.False(t, creds.Valid(), "nil Token should be invalid")
}

func TestValid_ExpiredToken(t *testing.T) {
	creds := &OAuthCredentials{
		Token: &oauth2.Token{
			AccessToken: "tok",
			Expiry:      time.Now().Add(-time.Minute),
		},
	}
	assert.False(t, creds.Valid(), "expired token should be invalid")
}

func TestValid_ValidToken(t *testing.T) {
	creds := &OAuthCredentials{
		Token: &oauth2.Token{
			AccessToken: "tok",
			Expiry:      time.Now().Add(time.Hour),
		},
	}
	assert.True(t, creds.Valid(), "non-expired token should be valid")
}

func TestValid_ZeroExpiry(t *testing.T) {
	// oauth2.Token.Valid() returns true for a zero expiry (no expiry set).
	creds := &OAuthCredentials{
		Token: &oauth2.Token{AccessToken: "tok"},
	}
	assert.True(t, creds.Valid(), "token with zero expiry should be considered valid")
}
