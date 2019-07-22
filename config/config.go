package config

import (
	"time"

	"github.com/cockroachdb/errors"
)

// Config is used to configure vaingogh.
type Config struct {
	Server struct {
		BaseURL         string         `yaml:"baseURL"`
		ShutdownTimeout *time.Duration `yaml:"shutdownTimeout"`
	} `yaml:"server"`

	Watcher struct {
		CheckInterval time.Duration `yaml:"checkInterval"`
	} `yaml:"watcher"`

	Lister struct {
		Concurrency int `yaml:"concurrency"`
		GitHub      struct {
			Username string `yaml:"username"`
		} `yaml:"github"`
	} `yaml:"lister"`
}

func defaultConfig() *Config {
	cfg := new(Config)
	cfg.Watcher.CheckInterval = time.Hour
	cfg.Lister.Concurrency = 5
	return cfg
}

// Validate returns an error if the config is not valid.
func (cfg *Config) Validate() error {
	if cfg.Watcher.CheckInterval <= 0 {
		return errors.New("watcher check interval must be positive " +
			"(watcher.checkInterval)")
	}

	// Validate lister config.
	{
		gh := &cfg.Lister
		if gh.GitHub.Username == "" {
			return errors.New("GitHub username is required (lister.github.username)")
		}
		if gh.Concurrency <= 0 {
			return errors.New("lister concurrency must be positive " +
				"(lister.concurrency)")
		}
	}

	// Validate server config.
	if cfg.Server.BaseURL == "" {
		return errors.New("server base URL must not be empty (server.baseURL)")
	}
	return nil
}
