package handler

import (
	"net/http"
	"strconv"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/repository"
	"github.com/gin-gonic/gin"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)


type MetricsHandlerDeps struct {
	MemStorage repository.MetricsStorage
}

type MetricsHandler struct {
	MemStorage repository.MetricsStorage
}

func NewMetricsHandler(r *gin.Engine, deps MetricsHandlerDeps) {
	handler := &MetricsHandler{
		MemStorage: deps.MemStorage,
	}

	r.POST("/update/:metrics_type/:metrics_name/:metrics_value", handler.Create()) // Создание
	r.GET("/value/:metrics_type/:metrics_name", handler.GetMetric()) // Получение метрики
	r.GET("/", handler.GetAllMetrics()) // Получение всех метрик и вывод в html
}


func (handler *MetricsHandler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Получение данных из строки
		metricsType := c.Param("metrics_type")
		metricsName := c.Param("metrics_name")
		metricsValue := c.Param("metrics_value")

		// Сохранение метрик
		switch metricsType {
			case Gauge:
				value, err := strconv.ParseFloat(metricsValue, 64)
			    if err != nil {
			    	c.String(http.StatusBadRequest, "Invalid gauge value")
			    	return
			    }
			    handler.MemStorage.SetGauge(metricsName, value)
			case Counter:
				value, err := strconv.ParseInt(metricsValue, 10, 64)
			    if err != nil {
			    	c.String(http.StatusBadRequest, "Invalid counter value")
			    	return
			    }
			    handler.MemStorage.AddCounter(metricsName, value)

			default:
				c.String(http.StatusBadRequest, "Unknown metric type")
				return
		}
		
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Status(http.StatusOK)

	}
}

func (handler *MetricsHandler) GetMetric() gin.HandlerFunc {
	return func(c *gin.Context) {

		metricsType := c.Param("metrics_type")
		metricsName := c.Param("metrics_name")

		// Получение метрик
		switch metricsType {
		case Gauge:
			val, ok := handler.MemStorage.GetGauge(metricsName)
			if !ok {
				c.String(http.StatusNotFound, "Metric not found")
				return
			}
			c.String(http.StatusOK, strconv.FormatFloat(val, 'f', -1, 64))

		case Counter:
			val, ok := handler.MemStorage.GetCounter(metricsName)
			if !ok {
				c.String(http.StatusNotFound, "Metric not found")
				return
			}
			c.String(http.StatusOK, strconv.FormatInt(val, 10))

		default:
			c.String(http.StatusNotFound, "Unknown metric type")
			return
		}
	}
}


func (handler *MetricsHandler) GetAllMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем все метрики
		gauges, counters := handler.MemStorage.GetAll()

		// Начало HTML
		html := `<html>
<head><title>Metrics</title></head>
<body>
<h1>All Metrics</h1>
<h2>Gauges</h2>
<ul>`

		// Список gauge
		for name, value := range gauges {
			html += "<li>" + name + ": " + strconv.FormatFloat(value, 'f', -1, 64) + "</li>"
		}

		html += `</ul>
<h2>Counters</h2>
<ul>`

		// Список counter
		for name, value := range counters {
			html += "<li>" + name + ": " + strconv.FormatInt(value, 10) + "</li>"
		}

		html += `</ul>
</body>
</html>`

		// Отправляем HTML
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	}
}