package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrustedSubnetMiddleware_EmptyAllowsAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw, err := TrustedSubnetMiddleware("")
	require.NoError(t, err)

	router := gin.New()
	router.Use(mw)
	router.POST("/update/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/update/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTrustedSubnetMiddleware_RejectsOutsideSubnet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw, err := TrustedSubnetMiddleware("192.168.1.0/24")
	require.NoError(t, err)

	router := gin.New()
	router.Use(mw)
	router.POST("/update/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/update/", nil)
	req.Header.Set(security.RealIPHeader, "10.0.0.1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTrustedSubnetMiddleware_AllowsInsideSubnet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw, err := TrustedSubnetMiddleware("192.168.1.0/24")
	require.NoError(t, err)

	router := gin.New()
	router.Use(mw)
	router.POST("/update/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/update/", nil)
	req.Header.Set(security.RealIPHeader, "192.168.1.42")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTrustedSubnetMiddleware_RejectsMissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw, err := TrustedSubnetMiddleware("192.168.1.0/24")
	require.NoError(t, err)

	router := gin.New()
	router.Use(mw)
	router.POST("/update/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/update/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTrustedSubnetMiddleware_InvalidCIDR(t *testing.T) {
	_, err := TrustedSubnetMiddleware("not-a-cidr")
	assert.Error(t, err)
}
