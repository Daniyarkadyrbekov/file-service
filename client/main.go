package main

import (
	"log"
	"net/http"

	"github.com/file-service/client/service"
	"github.com/ory/viper"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{
		"./client.log",
	}
	return cfg.Build()
}

func main() {

	l, err := NewLogger()
	if err != nil {
		log.Printf("create logger err = %s\n", err.Error())
		return
	}

	defer func() {
		if err := recover(); err != nil {
			l.Error("panic occurred:", zap.Any("err", err))
		}
	}()
	defer func() {
		l.Info("client closed")
	}()

	l.Info("client started")

	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		l.Error("read config", zap.Error(err))
		return
	}

	c := &service.Config{}
	if err := viper.GetViper().Unmarshal(c); err != nil {
		l.Error("unmarshal config", zap.Error(err))
		return
	}
	if err := c.Check(); err != nil {
		l.Error("cfg check err", zap.Error(err))
		return
	}

	svc, err := service.New(c, l)
	if err != nil {
		l.Error("service creation err", zap.Error(err))
		return
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
	}()

	svc.Run()
}
