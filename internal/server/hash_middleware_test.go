package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/audit"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHashMiddleware_BadHashRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	repo := repository.NewMemRepository(logger, "", 0, false)
	router := NewRouter(repo, func(c *gin.Context) { c.Next() }, "secret", nil, audit.NewPublisher(logger, "", ""))

	body := []byte(`{"id":"GaugeMetric","type":"gauge","value":100.5}`)
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(security.HashHeader, "wrong-hash")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHashMiddleware_ResponseHashAdded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	repo := repository.NewMemRepository(logger, "", 0, false)
	router := NewRouter(repo, func(c *gin.Context) { c.Next() }, "secret", nil, audit.NewPublisher(logger, "", ""))

	updateBody := []byte(`{"id":"GaugeMetric","type":"gauge","value":100.5}`)
	updateHash := sha256.Sum256(append(updateBody, []byte("secret")...))

	updateReq := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set(security.HashHeader, hex.EncodeToString(updateHash[:]))
	updateW := httptest.NewRecorder()
	router.ServeHTTP(updateW, updateReq)
	assert.Equal(t, http.StatusOK, updateW.Code)

	readBody := []byte(`{"id":"GaugeMetric","type":"gauge"}`)
	readHash := sha256.Sum256(append(readBody, []byte("secret")...))
	readReq := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(readBody))
	readReq.Header.Set("Content-Type", "application/json")
	readReq.Header.Set(security.HashHeader, hex.EncodeToString(readHash[:]))
	readW := httptest.NewRecorder()

	router.ServeHTTP(readW, readReq)

	assert.Equal(t, http.StatusOK, readW.Code)
	assert.NotEmpty(t, readW.Header().Get(security.HashHeader))
}
