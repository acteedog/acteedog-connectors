//go:build wasip1

package auth

import (
	"github.com/extism/go-pdk"
)

// storeOAuthTokens is a host function provided by acteedog.
// It persists updated OAuth tokens to the OS keychain on the host side.
//
// Input: JSON-encoded string of the form:
//
//	{ "access_token": "...", "refresh_token": "..." (optional) }
//
// Returns 0 on success, non-zero on failure (treated as a non-fatal warning).
//
//go:wasmimport acteedog store_oauth_tokens
func hostStoreOAuthTokens(ptr uint64)

// storeOAuthTokens calls the host function to persist tokens.
// json is expected to be a JSON-encoded string matching the input schema above.
func storeOAuthTokens(json string) {
	mem := pdk.AllocateString(json)
	defer mem.Free()
	hostStoreOAuthTokens(mem.Offset())
}
