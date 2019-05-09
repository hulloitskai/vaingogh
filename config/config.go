package config

import (
	"time"

	errors "golang.org/x/xerrors"
)

// Config is used to configure vaingogh.
type Config struct {
	Github struct {
		Username string `yaml:"username"`
		IsOrg    bool   `yaml:"isOrg"`
		Token    string
	} `yaml:"github"`
	CheckInterval time.Duration `yaml:"checkInterval" split_words:"true"`
}

var defaultConfig = Config{CheckInterval: 5 * time.Minute}

// Validate returns an error if the Config is not valid.
func (cfg *Config) Validate() error {
	if cfg.Github.Username == "" {
		return errors.New("GitHub username is required")
	}
	if cfg.CheckInterval <= 0 {
		return errors.New("check interval must be positive")
	}
	return nil
}
