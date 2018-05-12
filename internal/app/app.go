package app

import (
	"github.com/wybiral/pub/internal/model"
	"github.com/wybiral/pub/pkg/tor"
)

type App struct {
	Config *Config
	Model  *model.Model
	Self   *model.Self
	Tor    *tor.Tor
}

type Config struct {
	TorConfig    *tor.Config
	DatabasePath string
}

func NewDefaultConfig() *Config {
	return &Config{
		TorConfig:    tor.NewDefaultConfig(),
		DatabasePath: "database.sqlite",
	}
}

func NewApp(config *Config) (*App, error) {
	if config == nil {
		config = NewDefaultConfig()
	}
	// Create model
	model, err := model.NewModel(config.DatabasePath)
	if err != nil {
		return nil, err
	}
	// Get self
	self, err := model.GetSelf()
	if err != nil {
		return nil, err
	}
	// Create Tor instance
	tor, err := tor.NewTor(config.TorConfig)
	if err != nil {
		return nil, err
	}
	app := &App{
		Config: config,
		Model:  model,
		Self:   self,
		Tor:    tor,
	}
	return app, nil
}
