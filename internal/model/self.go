package model

import (
	"bytes"
	"crypto/rand"
	"errors"
	"github.com/wybiral/pub/pkg/tor/onions"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/sign"
	"net/http"
	"strconv"
	"time"
)

const selfSchema = `
create table Self (
	onion string not null,
	name string not null,
	about string not null,
	onion_key_type string not null,
	private_onion_key blob not null,
	public_box_key blob not null,
	private_box_key blob not null,
	public_sign_key blob not null,
	private_sign_key blob not null
);
`

type Self struct {
	model *Model
	Peer
	OnionKeyType    string `json:"-"`
	PrivateOnionKey []byte `json:"-"`
	PrivateBoxKey   []byte `json:"-"`
	PrivateSignKey  []byte `json:"-"`
}

// Get self identity from DB.
func (m *Model) GetSelf() (*Self, error) {
	s := &Self{model: m}
	row := m.db.QueryRow(`
		select
			onion,
			name,
			about,
			onion_key_type,
			private_onion_key,
			public_box_key,
			private_box_key,
			public_sign_key,
			private_sign_key
		from Self
	`)
	err := row.Scan(
		&s.Onion,
		&s.Name,
		&s.About,
		&s.OnionKeyType,
		&s.PrivateOnionKey,
		&s.PublicBoxKey,
		&s.PrivateBoxKey,
		&s.PublicSignKey,
		&s.PrivateSignKey,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Create new instance of self identity.
func (m *Model) CreateSelf(name, about string) (*Self, error) {
	s := &Self{model: m}
	s.Name = name
	s.About = about
	onion, err := onions.Generate()
	if err != nil {
		return nil, err
	}
	s.Onion = onion.Onion
	s.OnionKeyType = onion.KeyType
	s.PrivateOnionKey = onion.KeyContent
	// Generate box keys
	publicBoxKey, privateBoxKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	s.PublicBoxKey = publicBoxKey[:]
	s.PrivateBoxKey = privateBoxKey[:]
	// Generate signing keys
	publicSignKey, privateSignKey, err := sign.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	s.PublicSignKey = publicSignKey[:]
	s.PrivateSignKey = privateSignKey[:]
	_, err = m.db.Exec(
		`insert into Self (
			onion,
			name,
			about,
			onion_key_type,
			private_onion_key,
			public_box_key,
			private_box_key,
			public_sign_key,
			private_sign_key
		) values (
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?
		)`,
		s.Onion,
		s.Name,
		s.About,
		s.OnionKeyType,
		s.PrivateOnionKey,
		s.PublicBoxKey,
		s.PrivateBoxKey,
		s.PublicSignKey,
		s.PrivateSignKey,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Seal data using peer public key.
func (s *Self) Seal(data []byte, peerPublicBoxKey []byte) []byte {
	var publicKey [32]byte
	var privateKey [32]byte
	var nonce [24]byte
	copy(publicKey[:], peerPublicBoxKey)
	copy(privateKey[:], s.PrivateBoxKey)
	rand.Read(nonce[:])
	return box.Seal(nonce[:], data, &nonce, &publicKey, &privateKey)
}

// Open data sealed by peer with public key.
func (s *Self) Open(data []byte, peerPublicBoxKey []byte) ([]byte, bool) {
	var publicKey [32]byte
	var privateKey [32]byte
	var nonce [24]byte
	copy(publicKey[:], peerPublicBoxKey)
	copy(privateKey[:], s.PrivateBoxKey)
	copy(nonce[:], data[:24])
	data = data[24:]
	return box.Open(nil, data, &nonce, &publicKey, &privateKey)
}

// Make subscribe request to peer at onion.
func (s *Self) SubscribeRequest(c *http.Client, onion string) (*Peer, error) {
	peer, err := s.model.GetPeerByOnion(c, onion)
	if err != nil {
		return nil, err
	}
	// Create random secret
	secret := make([]byte, 32)
	rand.Read(secret)
	now := time.Now().Unix()
	// Construct auth payload
	auth := []byte("subscribe:")
	auth = append(auth, []byte(strconv.FormatInt(now, 10))...)
	auth = append(auth, []byte(":")...)
	auth = append(auth, secret...)
	auth = s.Seal(auth, peer.PublicBoxKey)
	// Make request
	addr := "http://" + onion + ".onion/subscribe"
	req, err := http.NewRequest("POST", addr, bytes.NewBuffer(auth))
	req.Header.Set("Peer", s.Onion)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("unable to subscribe")
	}
	peer.SecretAuthKey = secret
	err = peer.Insert(s.model)
	if err != nil {
		return nil, err
	}
	return peer, nil
}

// Accept a subscribe request by onion with auth payload.
func (s *Self) SubscribeAccept(c *http.Client, onion string, auth []byte) (*Peer, error) {
	// Get info for peer at onion
	peer, err := s.model.GetPeerByOnion(c, onion)
	if err != nil {
		return nil, err
	}
	// Open sealed auth payload
	opened, ok := s.Open(auth, peer.PublicBoxKey)
	if !ok {
		return nil, errors.New("box not opened")
	}
	parts := bytes.SplitN(opened, []byte(":"), 3)
	if len(parts) != 3 {
		return nil, errors.New("bad auth")
	}
	// Verify payload prefix
	if string(parts[0]) != "subscribe" {
		return nil, errors.New("no subcribe tag")
	}
	timestamp, err := strconv.ParseInt(string(parts[1]), 10, 64)
	// Verify integer timestamp
	if err != nil {
		return nil, errors.New("bad timestamp")
	}
	// Verify timestamp TTL
	now := time.Now().Unix()
	difference := now - timestamp
	ttl := int64(60 * 15)
	if difference < -ttl || difference > ttl {
		return nil, errors.New("timestamp out of range")
	}
	// Verify length of session secret
	if len(parts[2]) != 32 {
		return nil, errors.New("invalid secret length")
	}
	peer.SecretAuthKey = parts[2]
	err = peer.Insert(s.model)
	if err != nil {
		return nil, err
	}
	return peer, nil
}
