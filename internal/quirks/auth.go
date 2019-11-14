// Package quirks implements utility functions for dealing with the
// quirks of HG659 API.
package quirks

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// A LoginRequest contains a username and a hashed password.
type LoginRequest struct {
	Username string `json:"UserName"`
	Password string `json:"Password"`
}

func sha256hex(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func base64encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// NewLoginRequest generates a LoginRequest with given credentials.
func NewLoginRequest(username, password string, csrf *CSRF) *LoginRequest {
	return &LoginRequest{
		Username: username,
		Password: sha256hex(username + base64encode(sha256hex(password)) + csrf.Param + csrf.Token),
	}
}
