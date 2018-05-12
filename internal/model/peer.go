package model

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const peerSchema = `
create table Peer (
	onion string primary key,
	name string not null,
	about string not null,
	public_sign_key blob not null,
	public_box_key blob not null,
	secret_auth_key blob not null
);
`

type Peer struct {
	Onion         string `json:"onion"`
	Name          string `json:"name"`
	About         string `json:"about"`
	PublicBoxKey  []byte `json:"box_key"`
	PublicSignKey []byte `json:"sign_key"`
	SecretAuthKey []byte `json:"-"`
}

// Return array of all peers.
func (m *Model) GetPeers() ([]*Peer, error) {
	rows, err := m.db.Query(`
		select
			onion,
			name,
			about,
			public_sign_key,
			public_box_key,
			secret_auth_key
		from Peer
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	peers := make([]*Peer, 0)
	for rows.Next() {
		p := &Peer{}
		err = rows.Scan(
			&p.Onion,
			&p.Name,
			&p.About,
			&p.PublicSignKey,
			&p.PublicBoxKey,
			&p.SecretAuthKey,
		)
		if err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return peers, nil
}

// Return Peer instance from onion id (and tor http client).
func (m *Model) GetPeerByOnion(c *http.Client, onion string) (*Peer, error) {
	addr := "http://" + onion + ".onion/info"
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	res.Body.Close()
	p := &Peer{}
	err = json.Unmarshal(data, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Insert model into DB.
func (p *Peer) Insert(m *Model) error {
	_, err := m.db.Exec(
		`insert into Peer (
			onion,
			name,
			about,
			public_box_key,
			public_sign_key,
			secret_auth_key
		) values (
			?,
			?,
			?,
			?,
			?,
			?
		)`,
		p.Onion,
		p.Name,
		p.About,
		p.PublicBoxKey,
		p.PublicSignKey,
		p.SecretAuthKey,
	)
	if err != nil {
		return err
	}
	return nil
}
