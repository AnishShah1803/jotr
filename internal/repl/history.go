package repl

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	maxHistorySize  = 1000
	historyFileName = "repl_history"
)

type History struct {
	mu       sync.RWMutex
	entries  []string
	position int
	filePath string
}

func NewHistory() *History {
	h := &History{
		entries:  make([]string, 0),
		position: -1,
	}
	h.filePath = h.getHistoryPath()
	h.load()
	return h
}

func (h *History) getHistoryPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".jotr", historyFileName)
	}
	return filepath.Join(configDir, "jotr", historyFileName)
}

func (h *History) load() {
	file, err := os.Open(h.filePath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			h.entries = append(h.entries, line)
		}
	}
}

func (h *History) save() {
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	file, err := os.Create(h.filePath)
	if err != nil {
		return
	}
	defer file.Close()

	for _, entry := range h.entries {
		if _, err := file.WriteString(entry + "\n"); err != nil {
			return
		}
	}
}

func (h *History) Add(entry string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	entry = strings.TrimSpace(entry)
	if entry == "" {
		return
	}

	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == entry {
		h.position = len(h.entries)
		return
	}

	h.entries = append(h.entries, entry)

	if len(h.entries) > maxHistorySize {
		h.entries = h.entries[len(h.entries)-maxHistorySize:]
	}

	h.position = len(h.entries)
	h.save()
}

func (h *History) Previous() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.entries) == 0 {
		return ""
	}

	if h.position > 0 {
		h.position--
	}

	if h.position < 0 || h.position >= len(h.entries) {
		return ""
	}

	return h.entries[h.position]
}

func (h *History) Next() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.entries) == 0 {
		return ""
	}

	if h.position < len(h.entries)-1 {
		h.position++
	}

	if h.position >= len(h.entries) {
		h.position = len(h.entries) - 1
		return ""
	}

	return h.entries[h.position]
}

func (h *History) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.position = len(h.entries)
}

func (h *History) All() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]string, len(h.entries))
	copy(result, h.entries)
	return result
}
