package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// responseError reads an RFC 6749 / 7009 / 7662 / 7591 error response body and
// returns a descriptive error. Falls back to the HTTP status line if the body
// cannot be decoded or contains no "error" field.
func responseError(prefix string, resp *http.Response) error {
	var body struct {
		Err  string `json:"error"`
		Desc string `json:"error_description"`
	}
	if json.NewDecoder(resp.Body).Decode(&body) == nil && body.Err != "" {
		if body.Desc != "" {
			return fmt.Errorf("%s: %s: %s", prefix, body.Err, body.Desc)
		}
		return fmt.Errorf("%s: %s", prefix, body.Err)
	}
	return fmt.Errorf("%s: server returned %s", prefix, resp.Status)
}
