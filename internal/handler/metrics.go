package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/domain"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/payload"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/gin-gonic/gin"
)



type MetricsHandlerDeps struct {
	*service.MetricsService
}

type MetricsHandler struct {
	*service.MetricsService
}

func NewMetricsHandler(r *gin.Engine, deps MetricsHandlerDeps) {
	handler := &MetricsHandler{
		MetricsService: deps.MetricsService,
	}

	r.POST("/update/:metrics_type/:metrics_name/:metrics_value", handler.CreateMetricFromURL()) // Создание, с данными из строки
	r.POST("/update", handler.CreateMetricFromJSON()) // Создание из JSON
	r.GET("/value/:metrics_type/:metrics_name", handler.GetMetricFromURL()) // Получение метрики из строки
	r.POST("/value", handler.GetMetricFromJSON()) // Получение метрики из JSON
	r.GET("/", handler.GetAllMetrics()) // Получение всех метрик и вывод в html
}


func (h *MetricsHandler) CreateMetricFromURL() gin.HandlerFunc {
	return func(c *gin.Context) {
		metricType := c.Param("metrics_type")
		name := c.Param("metrics_name")
		value := c.Param("metrics_value")

		if err := h.MetricsService.SetMetric(metricType, name, value); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		c.Status(http.StatusOK)
	}
}

func (h *MetricsHandler) CreateMetricFromJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body payload.MetricCreateRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrInvalidJSONBody.Error()})
			return
		}

		if err := h.MetricsService.SetMetric(body.MetricsType, body.MetricsName, body.MetricsValue); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	}
}

func (h *MetricsHandler) GetMetricFromJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body payload.MetricGetRequest

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrInvalidJSONBody.Error()})
			return
		}

		metric, err := h.MetricsService.GetMetricJson(body.Type, body.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, metric)
	}
}


func (h *MetricsHandler) GetMetricFromURL() gin.HandlerFunc {
	return func(c *gin.Context) {
		metricsType := c.Param("metrics_type")
		metricsName := c.Param("metrics_name")

		metric, err := h.MetricsService.GetMetricUrl(metricsType, metricsName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		switch metric.Type {
		case service.Gauge:
			c.String(http.StatusOK, strconv.FormatFloat(*metric.Gauge, 'f', -1, 64))
		case service.Counter:
			c.String(http.StatusOK, strconv.FormatInt(*metric.Count, 10))
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrUnknownMetricType.Error()})
		}
	}
}

func (h *MetricsHandler) GetAllMetrics() gin.HandlerFunc {
    return func(c *gin.Context) {
        gauges, counters := h.MetricsService.MetricsRepository.GetAll()

        var b strings.Builder
        b.WriteString("<html><head><title>Metrics</title></head><body>")
        b.WriteString("<h1>All Metrics</h1><h2>Gauges</h2><ul>")
        for name, val := range gauges {
            b.WriteString(fmt.Sprintf("<li>%s: %f</li>", name, val))
        }
        b.WriteString("</ul><h2>Counters</h2><ul>")
        for name, val := range counters {
            b.WriteString(fmt.Sprintf("<li>%s: %d</li>", name, val))
        }
        b.WriteString("</ul></body></html>")

        c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(b.String()))
    }
}

