package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

var fileHashes = map[string]struct{}{}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, path)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	tee := io.TeeReader(file, &buf)

	h := sha256.New()
	h.Write([]byte(path))
	if _, err := io.Copy(h, tee); err != nil {
		log.Fatal(err)
	}

	if _, exists := fileHashes[fmt.Sprintf("%x", h.Sum(nil))]; exists {
		return nil, errors.New("file already exists on server")
	}
	fileHashes[fmt.Sprintf("%x", h.Sum(nil))] = struct{}{}

	_, err = io.Copy(part, &buf)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func main() {

	extraParams := map[string]string{
		"title":       "My Document",
		"author":      "Matt Aimonetti",
		"description": "A document with all the Go programming language secrets",
	}
	client := &http.Client{}

	err := filepath.Walk("./logs",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			request, err := newfileUploadRequest("http://localhost:8080/upload", extraParams, "myFile", path)
			if err != nil {
				return err
			}
			resp, err := client.Do(request)
			if err != nil {
				log.Fatal(err)
			} else {
				body := &bytes.Buffer{}
				_, err := body.ReadFrom(resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				resp.Body.Close()
				fmt.Println(resp.StatusCode)
				fmt.Println(resp.Header)
				fmt.Println(body)
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("hashes = %v\n", fileHashes)
}
