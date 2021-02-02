package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ory/viper"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	errAlreadyExists = errors.New("file already exists on server")
)

type Service struct {
	cfg        *Config
	fileHashes map[string]struct{}
	logsPath   string
	logger     *zap.Logger
}

func New(cfg *Config, logger *zap.Logger) *Service {
	return &Service{
		cfg:        cfg,
		logger:     logger,
		fileHashes: make(map[string]struct{}),
		logsPath:   cfg.LogsPath,
	}
}

func (c *Service) Run() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	t := time.NewTicker(c.cfg.UpdatePeriod)
	l := c.logger
	for {
		select {
		case <-t.C:

			l.Debug("ticker loop")

			err := filepath.Walk("./"+c.logsPath,
				func(path string, info os.FileInfo, err error) error {
					if info.IsDir() {
						return nil
					}
					l.Debug("fileWalk for file", zap.String("fileName", path))

					request, err := c.newFileUploadRequest(c.cfg.ServerUrl, "myFile", path)
					if err == errAlreadyExists {
						return nil
					}
					if err != nil {
						l.Error("getting request err", zap.Error(err))
						return err
					}
					resp, err := client.Do(request)
					if err != nil {
						l.Error("making request err", zap.Error(err))
						return err
					}

					body := &bytes.Buffer{}
					_, err = body.ReadFrom(resp.Body)
					if err != nil {
						l.Error("reading form err", zap.Error(err))
						return err
					}
					resp.Body.Close()
					l.Debug("loading file success",
						zap.Int("code", resp.StatusCode),
						zap.Any("header", resp.Header),
						zap.Any("body", body),
					)

					return nil
				})
			if err != nil {
				l.Error("file path walk err", zap.Error(err))
			}
		}
	}
}

// Creates a new file upload http request with optional extra params
func (c *Service) newFileUploadRequest(uri string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	//defer writer.Close()
	part, err := writer.CreateFormFile(paramName, path)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	tee := io.TeeReader(file, &buf)

	h := sha256.New()
	h.Write([]byte(path))
	if _, err := io.Copy(h, tee); err != nil {
		return nil, errors.Wrap(err, "copy to hash err")
	}

	if _, exists := c.fileHashes[fmt.Sprintf("%x", h.Sum(nil))]; exists {
		return nil, errAlreadyExists
	}
	c.fileHashes[fmt.Sprintf("%x", h.Sum(nil))] = struct{}{}

	_, err = io.Copy(part, &buf)

	//params := map[string]string{
	//	"title":       "My Document",
	//	"author":      "Matt Aimonetti",
	//	"description": "A document with all the Go programming language secrets",
	//}
	//for key, val := range params {
	//	_ = writer.WriteField(key, val)
	//}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Clinic-ID", c.cfg.ClinicID)
	req.Header.Set("Device-ID", c.cfg.DeviceID)
	return req, err
}

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

	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		l.Error("read config", zap.Error(err))
		return
	}

	c := &Config{}
	if err := viper.GetViper().Unmarshal(c); err != nil {
		l.Error("unmarshal config", zap.Error(err))
		return
	}
	if err := c.Check(); err != nil {
		l.Error("cfg check err", zap.Error(err))
		return
	}

	svc := New(c, l)
	svc.Run()
}
