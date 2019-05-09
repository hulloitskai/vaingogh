package config

import (
	"io"

	"github.com/kelseyhightower/envconfig"
	errors "golang.org/x/xerrors"
	"gopkg.in/yaml.v2"
)

// ReadFrom reads a Config from an io.Reader.
//
// It also reads in values from the environment.
func ReadFrom(r io.Reader) (*Config, error) {
	var (
		dec      = yaml.NewDecoder(r)
		cfg, err = readFromEnv()
	)
	if err != nil {
		return nil, errors.Errorf("config: reading values from env: %w", err)
	}
	if err = dec.Decode(cfg); err != nil {
		return nil, errors.Errorf("config: decoding YAML: %w", err)
	}
	return cfg, nil
}

func readFromEnv() (*Config, error) {
	cfg := defaultConfig
	if err := envconfig.Process("github", &cfg.Github); err != nil {
		return nil, err
	}
	return &cfg, nil
}
