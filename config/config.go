package config

import (
	"time"

	"github.com/cockroachdb/errors"
)

// Config is used to configure vaingogh.
type Config struct {
	Watcher struct {
		CheckInterval time.Duration `yaml:"checkInterval"`
	} `yaml:"watcher"`

	GitHub struct {
		Username string `yaml:"username"`

		Lister struct {
			Concurrency int `yaml:"concurrency"`
		} `yaml:"lister"`
	} `yaml:"github"`
}

func defaultConfig() *Config {
	cfg := new(Config)
	cfg.Watcher.CheckInterval = time.Hour
	cfg.GitHub.Lister.Concurrency = 5
	return cfg
}

// Validate returns an error if the Config is not valid.
func (cfg *Config) Validate() error {
	if cfg.Watcher.CheckInterval <= 0 {
		return errors.New("Watcher check interval must be positive " +
			"(watcher.checkInterval)")
	}

	// Validate Github credentials.
	{
		gh := &cfg.GitHub
		if gh.Username == "" {
			return errors.New("GitHub username is required (github.username)")
		}
		if gh.Lister.Concurrency <= 0 {
			return errors.New("Lister concurrency must be positive " +
				"(github.lister.concurrency)")
		}
	}
	return nil
}
