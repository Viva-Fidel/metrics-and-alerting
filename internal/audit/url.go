package audit

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
)

// URLObserver отправляет события аудита на внешний HTTP-эндпоинт
type URLObserver struct {
	url    string
	logger *slog.Logger
}

// NewURLObserver создаёт новый URLObserver
func NewURLObserver(logger *slog.Logger, url string) *URLObserver {
	return &URLObserver{url: url, logger: logger}
}

// ID возвращает идентификатор наблюдателя
func (u *URLObserver) ID() string {
	return "url:" + u.url
}

// Update отправляет событие на URL
func (u *URLObserver) Update(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		u.logger.Error("failed to marshal audit event", slog.Any("error", err))
		return
	}

	resp, err := http.Post(u.url, "application/json", bytes.NewReader(data))
	if err != nil {
		u.logger.Error("failed to send audit event", slog.String("url", u.url), slog.Any("error", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		u.logger.Error(
			"audit endpoint returned error status",
			slog.String("url", u.url),
			slog.Int("status", resp.StatusCode),
		)
	}
}
