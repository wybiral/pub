package tor

import (
	"errors"
	"fmt"
	"github.com/wybiral/pub/pkg/tor/onions"
	"github.com/wybiral/torgo"
	"net/http"
)

type Tor struct {
	Config     *Config
	Client     *http.Client
	Controller *torgo.Controller
}

type Config struct {
	SocksHost       string
	SocksPort       int
	ControlHost     string
	ControlPort     int
	ControlPassword string
}

func NewDefaultConfig() *Config {
	return &Config{
		SocksHost:       "127.0.0.1",
		SocksPort:       9050,
		ControlHost:     "127.0.0.1",
		ControlPort:     9051,
		ControlPassword: "",
	}
}

func NewTor(config *Config) (*Tor, error) {
	if config == nil {
		config = NewDefaultConfig()
	}
	// Get client
	client, err := NewClient(config.SocksHost, config.SocksPort)
	if err != nil {
		return nil, err
	}
	// Get controller
	controller, err := NewController(config.ControlHost, config.ControlPort)
	if err != nil {
		return nil, err
	}
	// Authenticate controller
	if len(config.ControlPassword) > 0 {
		err = controller.AuthenticatePassword(config.ControlPassword)
	} else {
		err = controller.AuthenticateCookie()
		if err != nil {
			err = controller.AuthenticateNone()
		}
	}
	if err != nil {
		return nil, errors.New("unable to authenticate tor controller")
	}
	tor := &Tor{
		Config:     config,
		Client:     client,
		Controller: controller,
	}
	return tor, nil
}

// Start hidden service to serve local port using onion keyType, key pair.
func (tor *Tor) StartOnion(port int, keyType string, key []byte) error {
	keyContent, err := onions.TorFormat(keyType, key)
	if err != nil {
		return err
	}
	onion := &torgo.Onion{
		Ports: map[int]string{
			80: fmt.Sprintf("127.0.0.1:%d", port),
		},
		PrivateKeyType: keyType,
		PrivateKey:     keyContent,
	}
	return tor.Controller.AddOnion(onion)
}
