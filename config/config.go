package config

import (
	"time"

	errors "golang.org/x/xerrors"
)

// Config is used to configure vaingogh.
type Config struct {
	CheckInterval time.Duration `yaml:"checkInterval" split_words:"true"`

	Github struct {
		Username    string `yaml:"username"`
		IsOrg       bool   `yaml:"isOrg"`
		Token       string
		Concurrency int `yaml:"concurrency"`
	} `yaml:"github"`
}

func defaultConfig() *Config {
	cfg := new(Config)
	cfg.CheckInterval = time.Hour
	cfg.Github.Concurrency = 5
	return cfg
}

// Validate returns an error if the Config is not valid.
func (cfg *Config) Validate() error {
	if cfg.CheckInterval <= 0 {
		return errors.New("check interval must be positive (checkInterval)")
	}

	gh := &cfg.Github
	if gh.Username == "" {
		return errors.New("GitHub username is required (github.username)")
	}
	if gh.Concurrency <= 0 {
		return errors.New("GitHub lister concurrency must be positive " +
			"(github.concurrency)")
	}
	return nil
}
