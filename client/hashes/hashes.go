package hashes

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

type Hashes struct {
	filePath   string
	hashes     map[string]bool
	mu         *sync.RWMutex
	logger     *zap.Logger
	unloadChan chan string
}

func New(filePath string, l *zap.Logger) (*Hashes, error) {
	h := &Hashes{
		filePath:   filePath,
		hashes:     make(map[string]bool),
		logger:     l,
		mu:         &sync.RWMutex{},
		unloadChan: make(chan string, 10),
	}

	if err := h.loadHashes(); err != nil {
		return nil, err
	}

	go h.runHashUnloader()

	return h, nil
}

func (h *Hashes) Put(hash string) {
	if !h.hashes[hash] {
		h.unloadChan <- hash
	}

	h.hashes[hash] = true
}

func (h *Hashes) Exists(hash string) bool {
	return h.hashes[hash]
}

func (h *Hashes) loadHashes() error {
	file, err := os.OpenFile(h.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		h.hashes[scanner.Text()] = true
	}

	return scanner.Err()
}

func (h *Hashes) runHashUnloader() {

	file, err := os.OpenFile(h.filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		h.logger.Error("error opening hashes file", zap.Error(err), zap.String("path", h.filePath))
		return
	}
	defer file.Close()

	for {
		hash := <-h.unloadChan

		w := bufio.NewWriter(file)
		if _, err := fmt.Fprintln(w, hash); err != nil {
			h.logger.Error("error unloading hash", zap.Error(err))
		}

		if err := w.Flush(); err != nil {
			h.logger.Debug("unload err", zap.Error(err))
		}
	}
}

//func (h *Hashes) unloadHashes() error {
//	h.mu.RLock()
//	defer h.mu.RUnlock()
//
//	file, err := os.OpenFile(h.filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
//	if err != nil {
//		return err
//	}
//	defer file.Close()
//
//	w := bufio.NewWriter(file)
//	for hash, _ := range h.hashes {
//		if _, err := fmt.Fprintln(w, hash); err != nil {
//			h.logger.Error("error unloading hashes", zap.Error(err))
//		}
//	}
//
//	return w.Flush()
//}
