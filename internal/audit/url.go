package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// URLObserver отправляет события аудита на внешний HTTP-эндпоинт.
type URLObserver struct {
	url string
}

// NewURLObserver создает новый URLObserver
func NewURLObserver(url string) *URLObserver {
	return &URLObserver{url: url}
}

// ID возвращает идентификатор наблюдателя
func (u *URLObserver) ID() string {
	return "url:" + u.url
}

// Update отправляет событие на URL
func (u *URLObserver) Update(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	resp, err := http.Post(u.url, "application/json", bytes.NewReader(data))
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}
