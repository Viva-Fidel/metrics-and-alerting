package audit

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
)

// FileObserver записывает события аудита в файл
type FileObserver struct {
	path   string
	file   *os.File
	mu     sync.Mutex
	logger *slog.Logger
}

// NewFileObserver создает новый FileObserver и открывает файл для записи
func NewFileObserver(logger *slog.Logger, path string) *FileObserver {
	f := &FileObserver{path: path, logger: logger}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		logger.Error("failed to open audit file", slog.String("path", path), slog.Any("error", err))
		return f
	}
	f.file = file
	return f
}

// ID возвращает идентификатор наблюдателя
func (f *FileObserver) ID() string {
	return "file:" + f.path
}

// Close закрывает файл аудита
func (f *FileObserver) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.file == nil {
		return nil
	}
	err := f.file.Close()
	f.file = nil
	return err
}

// Update записывает событие в файл
func (f *FileObserver) Update(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		f.logger.Error("failed to marshal audit event", slog.Any("error", err))
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.file == nil {
		return
	}

	if _, err := f.file.Write(append(data, '\n')); err != nil {
		f.logger.Error("failed to write audit event", slog.String("path", f.path), slog.Any("error", err))
	}
}
