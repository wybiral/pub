package onions

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
	"strings"
)

func GenerateED25519() (*Onion, error) {
	pub, pri, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	onion := ed25519ToOnion(pub)
	onion = strings.ToLower(onion)
	return &Onion{
		Onion:      onion,
		KeyType:    "ED25519-V3",
		KeyContent: pri[:],
	}, nil
}

// Construct onion address base32(publicKey || checkdigits || version).
func ed25519ToOnion(pub ed25519.PublicKey) string {
	checkdigits := ed25519Checkdigits(pub)
	combined := pub[:]
	combined = append(combined, checkdigits...)
	combined = append(combined, 0x03)
	return base32.StdEncoding.EncodeToString(combined)
}

// Calculate checksum sha3(".onion checksum" || publicKey || version).
func ed25519Checkdigits(pub ed25519.PublicKey) []byte {
	checkstr := []byte(".onion checksum")
	checkstr = append(checkstr, pub...)
	checkstr = append(checkstr, 0x03)
	checksum := sha3.Sum256(checkstr)
	return checksum[:2]
}

// Convert key to Tor format.
func TorFormatED25519(key []byte) string {
	h := sha512.Sum512(key[:32])
	// Set bits so that h[:32] is private scalar "a"
	h[0] &= 248
	h[31] &= 127
	h[31] |= 64
	// Since h[32:] is RH, h is now (a || RH)
	return base64.StdEncoding.EncodeToString(h[:])
}
