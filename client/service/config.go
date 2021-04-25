package service

import (
	"errors"
	"time"
)

type Config struct {
	ServerUrl    string        `mapstructure:"server-url"`
	UpdatePeriod time.Duration `mapstructure:"update-period"`
	LogsPath     string        `mapstructure:"logs-path"`
	ClinicID     string        `mapstructure:"clinic-id"`
	DeviceID     string        `mapstructure:"device-id"`
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

	return nil
}
