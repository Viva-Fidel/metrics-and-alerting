package agent_test

import (
	"testing"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/agent"
	"github.com/stretchr/testify/assert"
)

func TestPollMetrics(t *testing.T) {
	metrics := agent.NewMetrics()

	initialPoll := metrics.Counter["PollCount"]

	agent.PollMetrics(metrics)

	assert.Equal(t, initialPoll+1, metrics.Counter["PollCount"])

	assert.GreaterOrEqual(t, metrics.Gauge["RandomValue"], 0.0)
	assert.Less(t, metrics.Gauge["RandomValue"], 1.0)

	assert.Contains(t, metrics.Gauge, "Alloc")
	assert.Contains(t, metrics.Gauge, "HeapAlloc")
	assert.Contains(t, metrics.Gauge, "TotalAlloc")
}
