package onions

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base32"
	"encoding/base64"
	"strings"
)

func GenerateRSA1024() (*Onion, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}
	// Encode public key as DER
	pubder, err := asn1.Marshal(key.PublicKey)
	if err != nil {
		return nil, err
	}
	// Onion id is base32(firstHalf(sha1(publicKeyDER)))
	hash := sha1.Sum(pubder)
	half := hash[:len(hash)/2]
	onion := base32.StdEncoding.EncodeToString(half)
	onion = strings.ToLower(onion)
	prider := x509.MarshalPKCS1PrivateKey(key)
	return &Onion{
		Onion:      onion,
		KeyType:    "RSA1024",
		KeyContent: prider,
	}, nil
}

func TorFormatRSA1024(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}
