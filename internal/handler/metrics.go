package handler

import (
	"net/http"
	"strconv"

	"github.com/Viva-Fidel/metrics-and-alerting/pkg/storage"
)


type MetricsHandlerDeps struct {
	Storage storage.MetricsStorage
}

type MetricsHandler struct {
	storage storage.MetricsStorage
}



func NewMetricsHandler(router *http.ServeMux, deps MetricsHandlerDeps) {
	handler := &MetricsHandler{
		storage: deps.Storage,
	}

	router.HandleFunc("POST /update/{metrics_type}/{metrics_name}/{metrics_value}", handler.Create())
}

func (handler *MetricsHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricsType := r.PathValue("metrics_type")
		metricsName := r.PathValue("metrics_name")
		metricsValue := r.PathValue("metrics_value")

		if metricsName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		switch metricsType {
			case "gauge":
				value, err := strconv.ParseFloat(metricsValue, 64)
			    if err != nil {
			    	http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			    	return
			    }
			    handler.storage.SetGauge(metricsName, value)
			case "counter":
				value, err := strconv.ParseInt(metricsValue, 10, 64)
			    if err != nil {
			    	http.Error(w, "Invalid counter value", http.StatusBadRequest)
			    	return
			    }
			    handler.storage.AddCounter(metricsName, value)

			default:
				http.Error(w, "Unknown metric type", http.StatusBadRequest)
				return
		}



		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

	}
}