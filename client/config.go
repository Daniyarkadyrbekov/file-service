package main

import "time"

type Config struct {
	ServerUrl    string        `mapstructure:"server-url"`
	updatePeriod time.Duration `mapstructure:"update-period"`
}
