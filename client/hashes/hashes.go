package hashes

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Hashes struct {
	filePath       string
	hashes         map[string]bool
	mu             *sync.RWMutex
	logger         *zap.Logger
	unloadInterval time.Duration
}

func New(filePath string, unloadInterval time.Duration, l *zap.Logger) (*Hashes, error) {
	h := &Hashes{
		filePath:       filePath,
		unloadInterval: unloadInterval,
		hashes:         make(map[string]bool),
		logger:         l,
		mu:             &sync.RWMutex{},
	}

	if err := h.loadHashes(); err != nil {
		return nil, err
	}

	go h.runHashUnloader()

	return h, nil
}

func (h *Hashes) Put(hash string) {
	h.mu.Lock()
	h.hashes[hash] = true
	h.mu.Unlock()
}

func (h *Hashes) Exists(hash string) bool {
	h.mu.RLock()
	res := h.hashes[hash]
	h.mu.RUnlock()

	return res
}

func (h *Hashes) loadHashes() error {
	h.mu.Lock()
	defer h.mu.Unlock()

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
	t := time.NewTicker(h.unloadInterval)

	for {
		select {
		case <-t.C:
			if err := h.unloadHashes(); err != nil {
				h.logger.Debug("unload err", zap.Error(err))
			}
		}
	}
}

func (h *Hashes) unloadHashes() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	file, err := os.OpenFile(h.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for hash, _ := range h.hashes {
		fmt.Fprintln(w, hash)
	}

	return w.Flush()
}
