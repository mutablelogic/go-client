// Package oauth provides helpers for OAuth 2.0 authorization flows, token
// lifecycle management, and server metadata discovery.
//
// # Discovery and registration
//
// [Discover] fetches OAuth 2.0 Authorization Server Metadata (RFC 8414) or
// OpenID Connect Discovery metadata (OIDC Discovery 1.0). Pass any URL on the
// server — the function walks up the path until it finds the well-known
// document.
//
// [Register] performs RFC 7591 Dynamic Client Registration and returns an
// [OAuthCredentials] with the assigned ClientID and Secret embedded, ready to
// pass straight into an authorization function.
//
// # Authorization flows
//
//   - [AuthorizeWithBrowser]     — Authorization Code + PKCE via loopback redirect (RFC 6749)
//   - [AuthorizeWithCode]        — Authorization Code + PKCE via manual code paste (RFC 6749)
//   - [AuthorizeWithDevice]      — Device Authorization Grant (RFC 8628)
//   - [AuthorizeWithCredentials] — Client Credentials grant (RFC 6749 §4.4)
//
// # Token lifecycle
//
//   - [OAuthCredentials.Refresh]    — exchange a refresh token for a new access token
//   - [OAuthCredentials.Revoke]     — revoke access and refresh tokens (RFC 7009)
//   - [OAuthCredentials.Introspect] — query token metadata from the server (RFC 7662)
//   - [OAuthCredentials.Valid]      — local check that the token is set and non-expired
//
// # Custom HTTP client
//
// Every function that makes a network call honours a custom *http.Client
// injected via context:
//
//	ctx = context.WithValue(ctx, oauth2.HTTPClient, myClient)
package oauth
