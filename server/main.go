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

		//if err := r.ParseMultipartForm(1); err != nil {
		//	l.Error("max file size exceeded", zap.Error(nil))
		//	w.WriteHeader(http.StatusBadRequest)
		//	w.Write([]byte("max file size"))
		//	return
		//}

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
			zap.Any("reqHeaders", r.Header),
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

		tempFile, err := create(handler.Filename, r)
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
		w.Write([]byte("File Successfully Uploaded\n"))
		logger.Debug("file successfully uploaded")
	}
}

func check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Successfully checked\n")
}

func create(fileName string, r *http.Request) (*os.File, error) {
	clinicID, deviceID := "Unknown", "Unknown"
	if len(r.Header["Clinic-Id"]) == 1 {
		clinicID = r.Header["Clinic-Id"][0]
	}
	if len(r.Header["Device-Id"]) == 1 {
		deviceID = r.Header["Device-Id"][0]
	}

	path := "./devices/" + clinicID + "/" + deviceID + "/" + strings.ReplaceAll(fileName, "\\", "/")

	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		return nil, err
	}
	return os.Create(path)
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

	defer func() {
		if err := recover(); err != nil {
			l.Error("panic occurred:", zap.Any("err", err))
		}
	}()
	defer func() {
		l.Info("server closed")
	}()

	l.Info("server started")

	http.HandleFunc("/upload", uploadFile(l))
	http.HandleFunc("/check", check)

	err = http.ListenAndServe(":9123", nil)
	l.Info("server finished with err", zap.Error(err))
}
