package audit

import (
	"encoding/json"
	"os"
	"sync"
)

// FileObserver записывает события аудита в файл.
type FileObserver struct {
	path string
	mu   sync.Mutex
}

// NewFileObserver создает новый FileObserver
func NewFileObserver(path string) *FileObserver {
	return &FileObserver{path: path}
}

// ID возвращает идентификатор наблюдателя
func (f *FileObserver) ID() string {
	return "file:" + f.path
}

// Update записывает событие в файл
func (f *FileObserver) Update(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return
	}
	defer file.Close()

	_, _ = file.Write(append(data, '\n'))
}
