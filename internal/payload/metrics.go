package payload

type MetricCreateRequest struct {
	MetricsType  string `json:"metrics_type"`
	MetricsName  string `json:"metrics_name"`
	MetricsValue string `json:"metrics_value"`
}

type MetricGetRequest struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}