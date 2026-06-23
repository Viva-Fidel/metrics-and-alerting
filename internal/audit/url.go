package audit

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// URLObserver отправляет события аудита на внешний HTTP-эндпоинт
type URLObserver struct {
	logger *slog.Logger
	client *http.Client
	url    string
}

const (
	urlObserverTimeout    = 5 * time.Second
	urlObserverMaxRetries = 3
)

// NewURLObserver создаёт новый URLObserver
func NewURLObserver(logger *slog.Logger, url string) *URLObserver {
	return &URLObserver{
		url:    url,
		logger: logger,
		client: &http.Client{Timeout: urlObserverTimeout},
	}
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

	var resp *http.Response
	for attempt := 1; attempt <= urlObserverMaxRetries; attempt++ {
		req, reqErr := http.NewRequest(http.MethodPost, u.url, bytes.NewReader(data))
		if reqErr != nil {
			u.logger.Error("failed to build audit request", slog.String("url", u.url), slog.Any("error", reqErr))
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = u.client.Do(req)
		if err == nil && resp.StatusCode < http.StatusInternalServerError {
			break
		}
		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				u.logger.Error("failed to close audit response body", slog.Any("error", closeErr))
			}
		}

		if attempt == urlObserverMaxRetries {
			if err != nil {
				u.logger.Error("failed to send audit event", slog.String("url", u.url), slog.Any("error", err))
			} else {
				u.logger.Error(
					"audit endpoint returned retryable status",
					slog.String("url", u.url),
					slog.Int("status", resp.StatusCode),
					slog.Int("attempts", urlObserverMaxRetries),
				)
			}
			return
		}
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			u.logger.Error("failed to close audit response body", slog.Any("error", closeErr))
		}
	}()

	if resp.StatusCode >= http.StatusMultipleChoices {
		u.logger.Error(
			"audit endpoint returned error status",
			slog.String("url", u.url),
			slog.Int("status", resp.StatusCode),
		)
	}
}
