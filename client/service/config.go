package service

import (
	"errors"
	"time"

	"github.com/file-service/client/hashes"
)

type Config struct {
	ServerUrl    string        `mapstructure:"server-url"`
	UpdatePeriod time.Duration `mapstructure:"update-period"`
	LogsPath     string        `mapstructure:"logs-path"`
	ClinicID     string        `mapstructure:"clinic-id"`
	DeviceID     string        `mapstructure:"device-id"`
	HashesCfg    hashes.Config `mapstructure:"hashes"`
}

func (c *Config) Check() error {
	if c.ServerUrl == "" {
		return errors.New("empty server url")
	}

	if c.UpdatePeriod < time.Second {
		return errors.New("wrong update period = " + c.UpdatePeriod.String())
	}

	if c.LogsPath == "" {
		return errors.New("empty logs path")
	}

	if err := c.HashesCfg.Check(); err != nil {
		return err
	}

	return nil
}
