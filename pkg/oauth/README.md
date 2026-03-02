# oauth

OAuth 2.0 helper package for Go. Provides discovery, dynamic client registration, and three authorization flows — all composable and using the standard `context` pattern for custom HTTP client injection.

## Package overview

| Function / Type | Inputs | Outputs | Description |
|---|---|---|---|
| `Discover` | `URL` | `*OAuthMetadata` | Fetch server metadata via RFC 8414 / OIDC discovery |
| `Register` | `*OAuthMetadata`, client name, redirect URIs | `*OAuthCredentials` | RFC 7591 dynamic client registration; returned credentials carry `Metadata` |
| `AuthorizeWithBrowser` | `*OAuthCredentials`, `net.Listener`, `OpenFunc`, scopes | `*OAuthCredentials` | Authorization Code + PKCE via loopback redirect |
| `AuthorizeWithCode` | `*OAuthCredentials`, `PromptFunc`, scopes | `*OAuthCredentials` | Authorization Code + PKCE via manual code paste |
| `AuthorizeWithDevice` | `*OAuthCredentials`, `DevicePromptFunc`, scopes | `*OAuthCredentials` | Device Authorization Grant — CLI / headless devices |
| `AuthorizeWithCredentials` | `*OAuthCredentials`, scopes | `*OAuthCredentials` | Client Credentials grant — machine-to-machine, no user interaction |
| `(*OAuthCredentials).Refresh` | `*OAuthCredentials`, `ctx` | `error` | Refresh an access token using a stored refresh token |
| `(*OAuthCredentials).Revoke` | `*OAuthCredentials`, `ctx` | `error` | Revoke access and refresh tokens (RFC 7009); clears token on success |
| `(*OAuthCredentials).Introspect` | `*OAuthCredentials`, `ctx` | `*IntrospectionResponse`, `error` | Query active status and metadata for the current token (RFC 7662) |
| `(*OAuthCredentials).Valid` | `*OAuthCredentials` | `bool` | Returns true if a non-nil, non-expired token is set |

## Custom HTTP client

All functions honour a custom HTTP client injected via context — useful for testing, proxies, or mutual TLS:

```go
ctx = context.WithValue(ctx, oauth2.HTTPClient, myClient)
```

---

## Flow 1 — Browser (Authorization Code + PKCE via loopback)

The most common interactive flow. Opens a browser, starts a local HTTP server to receive the redirect, and exchanges the code for a token automatically.

```go
import (
    "context"
    "net"
    "os/exec"

    "github.com/mutablelogic/go-client/pkg/oauth"
)

func main() {
    ctx := context.Background()

    // 1. Discover server metadata
    metadata, err := oauth.Discover(ctx, "https://mcp.asana.com/v2/mcp")
    if err != nil {
        log.Fatal(err)
    }

    // 2. Allocate listener first — the port determines the redirect URI
    ln, err := net.Listen("tcp", "localhost:0")
    if err != nil {
        log.Fatal(err)
    }
    redirectURI := "http://" + ln.Addr().String() + "/callback"

    // 3. Register the client (if the server supports it)
    //    Pass the redirect URI so the server accepts the callback
    var creds *oauth.OAuthCredentials
    if metadata.SupportsRegistration() {
        creds, err = oauth.Register(ctx, metadata, "my-app", redirectURI)
        if err != nil {
            log.Fatal(err)
        }
    } else {
        // Use a pre-registered client ID from the provider's developer portal
        creds = &oauth.OAuthCredentials{
            Metadata: metadata,
            ClientID: os.Getenv("CLIENT_ID"),
        }
    }

    // 4. Open the browser and wait for the callback
    open := func(url string) error {
        return exec.Command("open", url).Start() // macOS; use "xdg-open" on Linux
    }
    creds, err = oauth.AuthorizeWithBrowser(ctx, creds, ln, open, "openid", "email")
    if err != nil {
        log.Fatal(err)
    }

    // creds.AccessToken, creds.RefreshToken are now set
    fmt.Println("access token:", creds.AccessToken)

    // 5. Later: refresh the token when it expires
    if err := creds.Refresh(ctx); err != nil {
        log.Fatal(err)
    }
}
```

---

## Flow 2 — Manual code paste (Authorization Code + PKCE)

For environments where opening a browser is possible but automating the redirect isn't — e.g. a remote SSH session. The user visits the URL manually and pastes the code back.

```go
import (
    "bufio"
    "fmt"
    "os"

    "github.com/mutablelogic/go-client/pkg/oauth"
)

func main() {
    ctx := context.Background()

    metadata, err := oauth.Discover(ctx, "https://accounts.google.com")
    if err != nil {
        log.Fatal(err)
    }

    // PromptFunc: display the URL and read the pasted code from stdin
    prompt := func(authURL string) (string, error) {
        fmt.Println("Open this URL in your browser:")
        fmt.Println(authURL)
        fmt.Print("Paste the authorization code: ")
        scanner := bufio.NewScanner(os.Stdin)
        scanner.Scan()
        return strings.TrimSpace(scanner.Text()), nil
    }

    creds, err := oauth.AuthorizeWithCode(ctx, &oauth.OAuthCredentials{
        Metadata: metadata,
        ClientID: "my-client-id",
    }, prompt, "openid", "email")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("access token:", creds.AccessToken)
}
```

---

## Flow 3 — Refresh

Once you have credentials with a refresh token (from either flow above), refresh them on demand. The `oauth2` library handles the 10-second expiry buffer automatically — if the token is still valid, no network call is made.

```go
// Persist creds to disk / keychain after authorization, then on each run:
// Wire up the callback before refreshing so rotated tokens are saved.
creds.OnRefresh = func(c *oauth.OAuthCredentials) error {
    return saveCredentials(c)
}
if err := creds.Refresh(ctx); err != nil {
    log.Fatal(err)
}
// creds.AccessToken is now fresh
```

To use a custom HTTP client for the refresh request:

```go
ctx = context.WithValue(ctx, oauth2.HTTPClient, myClient)
if err := creds.Refresh(ctx); err != nil {
    log.Fatal(err)
}
```

---

## Flow 4 — Device Code (CLI / headless)

For devices or CLI tools that can display a URL and code but cannot receive a browser redirect. The user visits the URL on any device, enters the code, and the CLI polls until authorization completes.

```go
prompt := func(userCode, verificationURI string) error {
    fmt.Printf("Visit %s and enter code: %s\n", verificationURI, userCode)
    return nil
}

creds, err := oauth.AuthorizeWithDevice(ctx, &oauth.OAuthCredentials{
    Metadata: metadata,
    ClientID: clientID,
}, prompt, "openid", "email")
if err != nil {
    log.Fatal(err)
}

fmt.Println("access token:", creds.AccessToken)
```

---

## Flow 5 — Client Credentials (machine-to-machine)

No user interaction. The client authenticates directly with its own credentials. Used for daemons, background services, and server-to-server calls. Note: servers do not issue refresh tokens for this grant — call `AuthorizeWithCredentials` again when the token expires.

```go
metadata, err := oauth.Discover(ctx, "https://auth.example.com")
if err != nil {
    log.Fatal(err)
}

creds, err := oauth.AuthorizeWithCredentials(ctx, &oauth.OAuthCredentials{
    Metadata:     metadata,
    ClientID:     os.Getenv("CLIENT_ID"),
    ClientSecret: os.Getenv("CLIENT_SECRET"),
}, "api:read", "api:write") // scopes optional
if err != nil {
    log.Fatal(err)
}

fmt.Println("access token:", creds.AccessToken)
```

---

## Discovery

`Discover` tries RFC 8414 root paths first (`/.well-known/oauth-authorization-server`, `/.well-known/openid-configuration`), then walks up the URL path for servers like Keycloak that publish metadata under realm-specific paths.

```go
// Root discovery (Google, Asana, Facebook, etc.)
metadata, err := oauth.Discover(ctx, "https://accounts.google.com")

// Path-relative discovery (Keycloak realm)
metadata, err := oauth.Discover(ctx, "https://keycloak.example.com/realms/myrealm")

// Any path under the server — walks up to find the metadata
metadata, err := oauth.Discover(ctx, "https://mcp.asana.com/v2/mcp")
```

Useful metadata checks:

```go
metadata.SupportsRegistration()               // RFC 7591 dynamic client registration
metadata.SupportsDeviceFlow()                 // RFC 8628 device authorization
metadata.SupportsPKCE()                       // any PKCE method
metadata.SupportsS256()                       // S256 specifically
metadata.SupportsRevocation()                 // RFC 7009 token revocation
metadata.SupportsIntrospection()              // RFC 7662 token introspection
metadata.SupportsFlow(oauth.OAuthFlowAuthorizationCode) // returns error if unsupported
metadata.ValidateScopes("openid", "email")    // returns error listing any unsupported scopes
```

---

## Dynamic client registration

`Register` performs RFC 7591 Dynamic Client Registration. The returned `*OAuthCredentials` has `ClientID` and `ClientSecret` set but no token — pass it directly into an authorization flow.

```go
// Servers that require redirect_uris (e.g. Asana MCP):
creds, err := oauth.Register(ctx, metadata, "my-app", "http://localhost:8080/callback")

// Servers that don't require redirect_uris (device / client_credentials flows):
creds, err := oauth.Register(ctx, metadata, "my-app")
```

---

## Token lifecycle — Revoke, Introspect, Valid

### Revoke (RFC 7009)

Revoke both the access token and (if present) the refresh token. The credentials' `Token` field is cleared on success so accidental reuse is prevented. If the server does not advertise a `revocation_endpoint`, the call is a no-op.

```go
if metadata.SupportsRevocation() {
    if err := creds.Revoke(ctx); err != nil {
        log.Fatal(err)
    }
    // creds.Token is now nil
}
```

### Introspect (RFC 7662)

Query the server for metadata about the current access token. Useful for debugging, or for verifying that a token received from a third party is still active.

```go
if metadata.SupportsIntrospection() {
    info, err := creds.Introspect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if !info.Active {
        fmt.Println("token has been revoked or expired server-side")
    } else {
        fmt.Println("scopes:", info.Scope)
        fmt.Println("expires:", info.Expiry)
    }
}
```

The `IntrospectionResponse` struct mirrors RFC 7662 §2.2:

```go
type IntrospectionResponse struct {
    Active    bool      // false means token is invalid
    Scope     string    // space-separated list of scopes
    ClientID  string
    Username  string
    TokenType string    // e.g. "Bearer"
    Expiry    time.Time // zero if not returned by server
    Subject   string    // sub
    Issuer    string    // iss
    JWTID     string    // jti
}
```

### Valid

Convenience check that the credentials contain a non-nil, non-expired token without making any network calls:

```go
if !creds.Valid() {
    if err := creds.Refresh(ctx); err != nil {
        log.Fatal(err)
    }
}
```

---

## OAuthFlow constants

```go
oauth.OAuthFlowAuthorizationCode   // "authorization_code"
oauth.OAuthFlowDeviceCode          // "urn:ietf:params:oauth:grant-type:device_code"
oauth.OAuthFlowClientCredentials   // "client_credentials"
oauth.OAuthFlowRefreshToken        // "refresh_token"
```

Used with `metadata.SupportsFlow(flow)` and `metadata.SupportsGrantType(flow)`.

---

## OAuthCredentials

```go
type OAuthCredentials struct {
    *oauth2.Token               // AccessToken, RefreshToken, Expiry, TokenType

    ClientID     string         // OAuth client ID
    ClientSecret string         // OAuth client secret (omitted for public clients)
    TokenURL     string         // token endpoint — stored for refresh without re-discovery
    Metadata     *OAuthMetadata // server metadata — carried from Register into Authorize functions
    OnRefresh    func(*OAuthCredentials) error // called when a new token is fetched; not serialised
}
```

`OnRefresh` is invoked only when the server actually returns a new token (i.e. the old one was expired or about to expire). Set it once after loading credentials to persist rotated refresh tokens:

```go
creds.OnRefresh = func(c *OAuthCredentials) error {
    return saveCredentials(c) // write to keychain / file / database
}
```
