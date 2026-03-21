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