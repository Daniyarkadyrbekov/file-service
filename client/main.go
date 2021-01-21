package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, basePath, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fmt.Printf("path = %s\n", path)
	//fileName, err := filepath.Rel(basePath, path)
	if err != nil {
		return nil, err
	}
	part, err := writer.CreateFormFile(paramName, path)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

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
	//path, _ := os.Getwd()
	//path += "/example2"
	//extraParams := map[string]string{
	//	"title":       "My Document",
	//	"author":      "Matt Aimonetti",
	//	"description": "A document with all the Go programming language secrets",
	//}
	//request, err := newfileUploadRequest("http://localhost:8080/upload", extraParams, "myFile", path)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//client := &http.Client{}
	//resp, err := client.Do(request)
	//if err != nil {
	//	log.Fatal(err)
	//} else {
	//	body := &bytes.Buffer{}
	//	_, err := body.ReadFrom(resp.Body)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	resp.Body.Close()
	//	fmt.Println(resp.StatusCode)
	//	fmt.Println(resp.Header)
	//	fmt.Println(body)
	//}

	extraParams := map[string]string{
		"title":       "My Document",
		"author":      "Matt Aimonetti",
		"description": "A document with all the Go programming language secrets",
	}
	client := &http.Client{}
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk("./logs",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			request, err := newfileUploadRequest("http://localhost:8080/upload", extraParams, "myFile", workDir, path)
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
}
