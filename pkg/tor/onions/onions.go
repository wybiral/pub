// Package onions provides methods to generate onion keys and format them for
// use with Tor.
package onions

import (
	"errors"
)

type Onion struct {
	// Onion service ID (same as address without the .onion ending)
	Onion string
	// Onion key type (RSA1024 or ED25519-V3)
	KeyType string
	// Onion key content
	KeyContent []byte
}

// Generate a new onion using the default method.
func Generate() (*Onion, error) {
	// Current default is RSA1024
	return GenerateRSA1024()
}

// Convert a keyType, keyContent pair into a Tor-friendly base64 string.
func TorFormat(keyType string, keyContent []byte) (string, error) {
	if keyType == "RSA1024" {
		return TorFormatRSA1024(keyContent), nil
	} else if keyType == "ED25519-V3" {
		return TorFormatED25519(keyContent), nil
	} else {
		return "", errors.New("bad KeyType")
	}
}
