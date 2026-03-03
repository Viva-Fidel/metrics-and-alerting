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
	Storage storage.MetricsStorage
}

func NewMetricsHandler(router *http.ServeMux, deps MetricsHandlerDeps) {
	handler := &MetricsHandler{
		Storage: deps.Storage,
	}

	router.HandleFunc("POST /update/{metrics_type}/{metrics_name}/{metrics_value}", handler.Create())
}

func (handler *MetricsHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Получение данных из строки
		metricsType := r.PathValue("metrics_type")
		metricsName := r.PathValue("metrics_name")
		metricsValue := r.PathValue("metrics_value")

		// Сохранение метрик
		switch metricsType {
			case "gauge":
				value, err := strconv.ParseFloat(metricsValue, 64)
			    if err != nil {
			    	http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			    	return
			    }
			    handler.Storage.SetGauge(metricsName, value)
			case "counter":
				value, err := strconv.ParseInt(metricsValue, 10, 64)
			    if err != nil {
			    	http.Error(w, "Invalid counter value", http.StatusBadRequest)
			    	return
			    }
			    handler.Storage.AddCounter(metricsName, value)

			default:
				http.Error(w, "Unknown metric type", http.StatusBadRequest)
				return
		}
		
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

	}
}