package agent_test

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/agent"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/model"
	"github.com/stretchr/testify/assert"
	"resty.dev/v3"
)

func gzipReaderFromRequest(r *http.Request) (*gzip.Reader, error) {
	return gzip.NewReader(r.Body)
}

func TestReportMetrics_SendsBatch(t *testing.T) {
	receivedPath := ""
	receivedEncoding := ""
	var receivedBatch []models.Metrics

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		receivedEncoding = r.Header.Get("Content-Encoding")
		assert.Equal(t, "/updates/", r.URL.Path)

		reader, err := gzipReaderFromRequest(r)
		assert.NoError(t, err)
		defer reader.Close()

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
	defer client.Close()

	agent.ReportMetrics(context.Background(), client, metrics)

	assert.Equal(t, "/updates/", receivedPath)
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
	defer client.Close()

	agent.ReportMetrics(context.Background(), client, agent.NewMetrics())
	assert.Equal(t, 0, requestsCount)
}

func TestReportMetrics_FallbackToLegacy(t *testing.T) {
	received := make(map[string]int)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received[r.URL.Path]++

		if r.URL.Path == "/updates/" {
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
	defer client.Close()

	agent.ReportMetrics(context.Background(), client, metrics)

	assert.Equal(t, 1, received["/updates/"])
	assert.Equal(t, 1, received["/update/gauge/TestGauge/12.34"])
	assert.Equal(t, 1, received["/update/counter/TestCounter/42"])
}
