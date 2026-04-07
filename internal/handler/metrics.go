package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/payload"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/gin-gonic/gin"
)

type metricsPageData struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

type MetricsHandlerDeps struct {
	*service.MetricsService
	DB *sql.DB
}

type MetricsHandler struct {
	*service.MetricsService
	DB *sql.DB
}

func NewMetricsHandler(r *gin.Engine, deps MetricsHandlerDeps) {
	handler := &MetricsHandler{
		MetricsService: deps.MetricsService,
		DB:             deps.DB,
	}

	r.POST("/update/:metrics_type/:metrics_name/:metrics_value", handler.CreateMetricFromURL()) // Создание, с данными из строки
	r.POST("/update/", handler.CreateMetricFromJSON())                                          // Создание из JSON
	r.POST("/updates/", handler.CreateMetricsFromJSONBatch())
	r.GET("/value/:metrics_type/:metrics_name", handler.GetMetricFromURL()) // Получение метрики из строки
	r.POST("/value/", handler.GetMetricFromJSON())                          // Получение метрики из JSON
	r.GET("/", handler.GetAllMetrics())                                     // Получение всех метрик и вывод в html
	r.GET("/ping", handler.GetPing())                                       // Проверка соединения к бд

}

func (h *MetricsHandler) GetPing() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.DB == nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		if err := h.DB.PingContext(c.Request.Context()); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)
	}
}

func (h *MetricsHandler) CreateMetricFromURL() gin.HandlerFunc {
	return func(c *gin.Context) {
		metricType := c.Param("metrics_type")
		name := c.Param("metrics_name")
		value := c.Param("metrics_value")

		if err := h.MetricsService.SetMetricURL(metricType, name, value); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		c.Status(http.StatusOK)
	}
}

func (h *MetricsHandler) CreateMetricFromJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body payload.MetricsJSON
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
			return
		}

		if err := h.MetricsService.SetMetricJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	}
}

func (h *MetricsHandler) CreateMetricsFromJSONBatch() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body []payload.MetricsJSON
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
			return
		}

		if err := h.MetricsService.SetMetricsJSONBatch(body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, body)
	}
}

func (h *MetricsHandler) GetMetricFromJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body payload.MetricsJSON

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
			return
		}

		metric, err := h.MetricsService.GetMetricJSON(body.MType, body.ID)
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

		metric, err := h.MetricsService.GetMetricURL(metricsType, metricsName)
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown metrics type"})
		}
	}
}

func (h *MetricsHandler) GetAllMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		gauges, counters := h.MetricsService.MetricsRepository.GetAll()

		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := metricsPageTmpl.Execute(c.Writer, metricsPageData{
			Gauges:   gauges,
			Counters: counters,
		}); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
}
