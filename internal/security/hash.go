package security

import (
	"crypto/sha256"
	"encoding/hex"
)

const HashHeader = "HashSHA256"

func BuildHash(body []byte, key string) string {
	sum := sha256.Sum256(append(body, []byte(key)...))
	return hex.EncodeToString(sum[:])
}

