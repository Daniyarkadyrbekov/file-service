package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

func uploadFile(l *zap.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Debug("file Upload Endpoint Hit")

		//if err r.ParseMultipartForm(10 << 20)
		if err := r.ParseMultipartForm(100 << 20); err != nil {
			l.Error("max file size exceeded", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("max file size"))
			return
		}

		file, handler, err := r.FormFile("myFile")
		if err != nil {
			l.Error("error retrieving the File", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("error retrieving the File"))
			return
		}
		defer file.Close()
		logger := l.With(
			zap.String("fileName", handler.Filename),
			zap.Int64("size", handler.Size),
			zap.Any("header", handler.Header),
		)

		// read all of the contents of our uploaded file into a
		// byte array
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			logger.Error("error reading file", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error reading file"))
			return
		}

		tempFile, err := create(getFinalPath(handler.Filename, r))
		if err != nil {
			logger.Error("error creating temp file", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error creating temp file"))
			return
		}
		// write this byte array to our temporary file
		if _, err := tempFile.Write(fileBytes); err != nil {
			logger.Error("error writing to file", zap.Error(err))
			return
		}

		// return that we have successfully uploaded our file!
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Successfully Uploaded File\n"))
	}
}

func getFinalPath(fileName string, r *http.Request) string {
	clinicID, deviceID := "Unknown", "Unknown"
	if len(r.Header["Clinic-ID"]) == 1 {
		clinicID = r.Header["Clinic-ID"][0]
	}
	if len(r.Header["Device-ID"]) == 1 {
		clinicID = r.Header["Device-ID"][0]
	}

	return "./devices/" + clinicID + "/" + deviceID + "/" + strings.ReplaceAll(fileName, "\\", "/")
}

func check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Successfully checked\n")
}

func create(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}

func NewLogger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{
		"./server.log",
	}
	return cfg.Build()
}

func main() {
	l, err := NewLogger()
	if err != nil {
		log.Printf("create logger err = %s\n", err.Error())
		return
	}

	l.Info("server started")

	http.HandleFunc("/upload", uploadFile(l))
	http.HandleFunc("/check", check)

	err = http.ListenAndServe(":8080", nil)
	l.Info("server finished with err", zap.Error(err))
}
