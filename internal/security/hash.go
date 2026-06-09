// Package security содержит утилиты для подписи HTTP-запросов.
package security

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashHeader — имя HTTP-заголовка с SHA-256 хешем тела запроса или ответа.
const HashHeader = "HashSHA256"

// BuildHash вычисляет SHA-256 хеш от тела и ключа.
func BuildHash(body []byte, key string) string {
	sum := sha256.Sum256(append(body, []byte(key)...))
	return hex.EncodeToString(sum[:])
}
