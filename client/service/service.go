package service

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/file-service/client/hashes"
	"github.com/file-service/client/metrics"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Service struct {
	cfg     *Config
	hashes  *hashes.Hashes
	logger  *zap.Logger
	metrics metrics.Metrics
}

func New(cfg *Config, logger *zap.Logger) (*Service, error) {
	m, err := metrics.New("files_client")
	if err != nil {
		return nil, err
	}

	h, err := hashes.New(cfg.HashesCfg.HashesPath, logger)
	if err != nil {
		return nil, err
	}

	return &Service{
		cfg:     cfg,
		logger:  logger,
		hashes:  h,
		metrics: m,
	}, nil
}

func (c *Service) Run() {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	t := time.NewTicker(c.cfg.UpdatePeriod)
	l := c.logger

	err := c.uploadDirFiles(l, "./"+c.cfg.LogsPath, client)
	if err != nil {
		l.Error("file path walk err", zap.Error(err))
	}

	for {
		select {
		case <-t.C:
			c.metrics.IncTicker()

			err := c.uploadDirFiles(l, "./"+c.cfg.LogsPath, client)
			if err != nil {
				l.Error("file path walk err", zap.Error(err))
			}
		}
	}
}

func (c *Service) uploadDirFiles(l *zap.Logger, path string, client *http.Client) error {
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {

			if info.IsDir() {
				return nil
			}

			l = l.With(zap.String("fileName", path))
			request, fileHash, err := c.newFileUploadRequest(c.cfg.ServerUrl, "myFile", path)
			if err != nil {
				l.Error("getting request err", zap.Error(err))
				return nil
			}

			if c.hashes.Exists(fileHash) {
				return nil
			}

			resp, err := client.Do(request)
			if err != nil {
				l.Error("making request err", zap.Error(err))
				return nil
			}

			l := l.With(zap.Int("status", resp.StatusCode), zap.Any("header", resp.Header))
			if resp.StatusCode != http.StatusOK {
				l.Debug("loading file err", zap.Int("status", resp.StatusCode))
			}

			body := &bytes.Buffer{}
			_, err = body.ReadFrom(resp.Body)
			if err != nil {
				l.Error("reading form err", zap.Error(err))
				return nil
			}
			resp.Body.Close()

			l.Debug("loading file success")

			c.hashes.Put(fileHash)

			return nil
		})

	return err
}

// Creates a new file upload http request with optional extra params
func (c *Service) newFileUploadRequest(uri string, paramName, path string) (*http.Request, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	//defer writer.Close()
	part, err := writer.CreateFormFile(paramName, path)
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	tee := io.TeeReader(file, &buf)

	hash := sha256.New()
	hash.Write([]byte(path))
	if _, err := io.Copy(hash, tee); err != nil {
		return nil, "", errors.Wrap(err, "copy to hash err")
	}

	//fileHash := fmt.Sprintf("%x", hash.Sum(nil))
	//if _, exists := c.fileHashes[fmt.Sprintf("%x", c.hash.Sum(nil))]; exists {
	//	return nil, "", errAlreadyExists
	//}

	_, err = io.Copy(part, &buf)

	err = writer.Close()
	if err != nil {
		return nil, "", err
	}
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Clinic-ID", c.cfg.ClinicID)
	req.Header.Set("Device-ID", c.cfg.DeviceID)
	return req, fmt.Sprintf("%x", hash.Sum(nil)), err
}

//// Creates a new file upload http request with optional extra params
//func (c *Service) newFileUploadRequest(uri string, paramName, path string) (*http.Request, error) {
//	file, err := os.Open(path)
//	if err != nil {
//		return nil, err
//	}
//	defer file.Close()
//
//	body := &bytes.Buffer{}
//	writer := multipart.NewWriter(body)
//	//defer writer.Close()
//	part, err := writer.CreateFormFile(paramName, path)
//	if err != nil {
//		return nil, err
//	}
//
//	var buf bytes.Buffer
//	tee := io.TeeReader(file, &buf)
//
//	c.hash = sha256.New()
//	c.hash.Write([]byte(path))
//	if _, err := io.Copy(c.hash, tee); err != nil {
//		return nil, errors.Wrap(err, "copy to hash err")
//	}
//
//	if _, exists := c.fileHashes[fmt.Sprintf("%x", c.hash.Sum(nil))]; exists {
//		return nil, errAlreadyExists
//	}
//
//	_, err = io.Copy(part, &buf)
//
//	err = writer.Close()
//	if err != nil {
//		return nil, err
//	}
//	req, err := http.NewRequest("POST", uri, body)
//	if err != nil {
//		return nil, err
//	}
//	req.Header.Set("Content-Type", writer.FormDataContentType())
//	req.Header.Set("Clinic-ID", c.cfg.ClinicID)
//	req.Header.Set("Device-ID", c.cfg.DeviceID)
//	return req, err
//}
