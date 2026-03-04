package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"sync"

	// Packages
	oauth2 "golang.org/x/oauth2"
)

/////////////////////////////////////////////////////////////////////////////
// TYPES

// OpenFunc is called with the authorization URL and should open it in a
// browser. On macOS the caller can use:
//
//	func(u string) error { return exec.Command("open", u).Start() }
type OpenFunc func(url string) error

// callbackResult carries the authorization code (or an error) from the
// loopback HTTP handler back to AuthorizeWithBrowser.
type callbackResult struct {
	code string
	err  error
}

/////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const callbackPath = "/callback"

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AuthorizeWithBrowser performs an OAuth 2.0 Authorization Code flow with
// PKCE using a loopback redirect. It:
//
//  1. Derives redirect_uri from the listener address (http://localhost:PORT/callback)
//  2. Starts a temporary HTTP server on the listener
//  3. Calls open(authURL) — typically to launch a browser
//  4. Waits for the provider to redirect back with the authorization code
//  5. Exchanges the code for tokens and shuts the server down
//
// The creds parameter must have Metadata and ClientID set (e.g. obtained from Register
// or constructed manually). If no scopes are provided, the scope parameter is omitted
// and the server applies its own defaults (RFC 6749 §3.3).
// The caller is responsible for creating the listener, e.g.:
//
//	ln, _ := net.Listen("tcp", "localhost:0") // random port
//
// To use a custom HTTP client for the token exchange, inject it into the
// context:
//
//	ctx = context.WithValue(ctx, oauth2.HTTPClient, myClient)
func AuthorizeWithBrowser(ctx context.Context, creds *OAuthCredentials, listener net.Listener, open OpenFunc, scopes ...string) (*OAuthCredentials, error) {
	switch {
	case creds == nil:
		return nil, fmt.Errorf("credentials are required")
	case creds.Metadata == nil:
		return nil, fmt.Errorf("credentials missing metadata")
	case creds.ClientID == "":
		return nil, fmt.Errorf("client ID is required")
	case listener == nil:
		return nil, fmt.Errorf("listener is required")
	case open == nil:
		return nil, fmt.Errorf("open function is required")
	case len(scopes) == 0:
		// No scopes requested — omit the scope parameter so the server applies its own defaults (RFC 6749 §3.3).
	}

	if err := creds.Metadata.SupportsFlow(OAuthFlowAuthorizationCode); err != nil {
		return nil, err
	}
	if err := creds.Metadata.ValidateScopes(scopes...); err != nil {
		return nil, err
	}

	// Derive redirect URI from the listener address.
	redirectURI := "http://" + listener.Addr().String() + callbackPath

	cfg := &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Scopes:       scopes,
		RedirectURL:  redirectURI,
		Endpoint:     creds.Metadata.OAuthEndpoint(),
	}

	// Generate PKCE verifier if the server supports it (RFC 7636).
	// Servers that predate PKCE (e.g. GitHub) reject requests that include
	// code_challenge, so we only enable it when advertised.
	var verifier string
	if creds.Metadata.SupportsPKCE() {
		verifier = oauth2.GenerateVerifier()
	}
	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("generate state: %w", err)
	}

	// Build the authorization URL, adding PKCE challenge only when supported.
	authOpts := []oauth2.AuthCodeOption{oauth2.AccessTypeOffline}
	if verifier != "" {
		authOpts = append(authOpts, oauth2.S256ChallengeOption(verifier))
	}
	authURL := cfg.AuthCodeURL(state, authOpts...)

	// Channel that receives exactly one result from the callback handler.
	resultCh := make(chan callbackResult, 1)

	// Start the loopback server.
	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux}
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("state"); got != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			select {
			case resultCh <- callbackResult{err: fmt.Errorf("state mismatch: got %q", got)}:
			default:
			}
			return
		}
		if errVal := q.Get("error"); errVal != "" {
			msg := errVal
			if desc := q.Get("error_description"); desc != "" {
				msg += ": " + desc
			}
			http.Error(w, msg, http.StatusBadRequest)
			select {
			case resultCh <- callbackResult{err: fmt.Errorf("authorization error: %s", msg)}:
			default:
			}
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			select {
			case resultCh <- callbackResult{err: fmt.Errorf("no authorization code in callback")}:
			default:
			}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `<!DOCTYPE html><html><body><p>Authorization successful. You may close this window.</p></body></html>`)
		select {
		case resultCh <- callbackResult{code: code}:
		default:
		}
	})

	// Serve in the background. Serve always returns ErrServerClosed after
	// Close is called, so we ignore that error. The WaitGroup ensures we
	// don't return until the goroutine has fully exited.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		srv.Serve(listener) //nolint:errcheck
	}()
	defer func() {
		srv.Close()
		wg.Wait()
	}()

	// Open the browser.
	if err := open(authURL); err != nil {
		return nil, fmt.Errorf("open browser: %w", err)
	}

	// Wait for the callback or context cancellation.
	var result callbackResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if result.err != nil {
		return nil, result.err
	}

	// Exchange the code for a token, including the PKCE verifier when used.
	exchangeOpts := []oauth2.AuthCodeOption{}
	if verifier != "" {
		exchangeOpts = append(exchangeOpts, oauth2.VerifierOption(verifier))
	}
	tok, err := cfg.Exchange(ctx, result.code, exchangeOpts...)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	return creds.withToken(tok), nil
}

/////////////////////////////////////////////////////////////////////////////
// PRIVATE HELPERS

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
