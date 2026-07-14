package agent_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/agent"
	models "github.com/Viva-Fidel/metrics-and-alerting/internal/model"
	"github.com/stretchr/testify/assert"
	"resty.dev/v3"
)

func TestReportMetrics_SendsBatch(t *testing.T) {
	receivedPath := ""
	receivedEncoding := ""
	receivedHash := ""
	var receivedBatch []models.Metrics

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		receivedEncoding = r.Header.Get("Content-Encoding")
		receivedHash = r.Header.Get("HashSHA256")
		assert.Equal(t, "/updates", r.URL.Path)

		compressedBody, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		expectedSum := sha256.Sum256(append(compressedBody, []byte("secret")...))
		assert.Equal(t, hex.EncodeToString(expectedSum[:]), receivedHash)

		reader, err := gzip.NewReader(bytes.NewReader(compressedBody))
		assert.NoError(t, err)
		defer func() { _ = reader.Close() }()

		raw, err := io.ReadAll(reader)
		assert.NoError(t, err)
		err = json.Unmarshal(raw, &receivedBatch)
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	metrics := &agent.Metrics{
		Gauge:   map[string]float64{"TestGauge": 12.34},
		Counter: map[string]int64{"TestCounter": 42},
	}

	client := resty.New().SetBaseURL(ts.URL)
	defer func() { _ = client.Close() }()

	agent.ReportMetrics(context.Background(), client, metrics, "secret", nil)

	assert.Equal(t, "/updates", receivedPath)
	assert.Equal(t, "gzip", receivedEncoding)
	assert.Len(t, receivedBatch, 2)
}

func TestReportMetrics_SkipsEmptyBatch(t *testing.T) {
	requestsCount := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestsCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	defer func() { _ = client.Close() }()

	agent.ReportMetrics(context.Background(), client, agent.NewMetrics(), "", nil)
	assert.Equal(t, 0, requestsCount)
}

func TestReportMetrics_FallbackToLegacy(t *testing.T) {
	received := make(map[string]int)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received[r.URL.Path]++

		if r.URL.Path == "/updates" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	metrics := &agent.Metrics{
		Gauge:   map[string]float64{"TestGauge": 12.34},
		Counter: map[string]int64{"TestCounter": 42},
	}

	client := resty.New().SetBaseURL(ts.URL)
	defer func() { _ = client.Close() }()

	agent.ReportMetrics(context.Background(), client, metrics, "", nil)

	assert.Equal(t, 1, received["/updates"])
	assert.Equal(t, 1, received["/update/gauge/TestGauge/12.34"])
	assert.Equal(t, 1, received["/update/counter/TestCounter/42"])
}
