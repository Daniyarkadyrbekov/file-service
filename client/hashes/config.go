package hashes

import (
	"github.com/pkg/errors"
)

type Config struct {
	HashesPath string `mapstructure:"hashes-path"`
}

func (c *Config) Check() error {
	if c.HashesPath == "" {
		return errors.New("hashes path is required")
	}

	return nil
}
