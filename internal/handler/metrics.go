// Package handler предоставляет HTTP-обработчики сервера сбора метрик.
package handler

import (
	"net/http"
	"strconv"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/audit"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/payload"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/service"
	"github.com/gin-gonic/gin"
)

type metricsPageData struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

// MetricsHandlerDeps содержит зависимости для инициализации MetricsHandler.
type MetricsHandlerDeps struct {
	*service.MetricsService
	AuditPublisher *audit.Publisher
}

// MetricsHandler обрабатывает HTTP-запросы к API метрик.
type MetricsHandler struct {
	*service.MetricsService
	auditPublisher *audit.Publisher
}

// NewMetricsHandler регистрирует маршруты API метрик на переданном роутере.
func NewMetricsHandler(r *gin.Engine, deps MetricsHandlerDeps) {
	handler := &MetricsHandler{
		MetricsService: deps.MetricsService,
		auditPublisher: deps.AuditPublisher,
	}

	r.POST("/update/:metrics_type/:metrics_name/:metrics_value", handler.CreateMetricFromURL()) // Создание, с данными из строки
	r.POST("/update/", handler.CreateMetricFromJSON())                                          // Создание из JSON
	r.POST("/updates/", handler.CreateMetricsFromJSONBatch())
	r.GET("/value/:metrics_type/:metrics_name", handler.GetMetricFromURL()) // Получение метрики из строки
	r.POST("/value/", handler.GetMetricFromJSON())                          // Получение метрики из JSON
	r.GET("/", handler.GetAllMetrics())                                     // Получение всех метрик и вывод в html
	r.GET("/ping", handler.GetPing())                                       // Проверка соединения к бд

}

// GetPing возвращает обработчик проверки доступности хранилища метрик (GET /ping).
func (h *MetricsHandler) GetPing() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.MetricsService.Ping(c.Request.Context()); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	}
}

// CreateMetricFromURL возвращает обработчик создания метрики из параметров URL (POST /update/:metrics_type/:metrics_name/:metrics_value).
func (h *MetricsHandler) CreateMetricFromURL() gin.HandlerFunc {
	return func(c *gin.Context) {
		metricType := c.Param("metrics_type")
		name := c.Param("metrics_name")
		value := c.Param("metrics_value")

		if err := h.MetricsService.SetMetricURL(metricType, name, value); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		h.notifyAudit(c, []string{name})
		c.Status(http.StatusOK)
	}
}

// CreateMetricFromJSON возвращает обработчик создания метрики из JSON-тела (POST /update/).
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

		h.notifyAudit(c, []string{body.ID})
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	}
}

// CreateMetricsFromJSONBatch возвращает обработчик пакетного обновления метрик (POST /updates/).
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

		metricNames := make([]string, len(body))
		for i := range body {
			metricNames[i] = body[i].ID
		}
		h.notifyAudit(c, metricNames)
		c.JSON(http.StatusOK, body)
	}
}

// GetMetricFromJSON возвращает обработчик получения метрики по JSON-запросу (POST /value/).
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

// GetMetricFromURL возвращает обработчик получения метрики из параметров URL (GET /value/:metrics_type/:metrics_name).
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

// notifyAudit отправляет событие всем наблюдателям
func (h *MetricsHandler) notifyAudit(c *gin.Context, metrics []string) {
	if h.auditPublisher == nil {
		return
	}
	h.auditPublisher.Notify(metrics, c.ClientIP())
}

// GetAllMetrics возвращает обработчик HTML-страницы со всеми метриками (GET /).
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
