package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"api-go/internal/auth"
	"api-go/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddleware_ServiceAccount(t *testing.T) {
	cfg := &config.Config{
		Google: config.GoogleConfig{AuthType: "service-account"},
	}
	ts := auth.NewTokenStorage()

	r := gin.New()
	r.Use(AuthMiddleware(cfg, ts))
	r.GET("/test", func(c *gin.Context) {
		user := GetAuthUser(c)
		c.JSON(http.StatusOK, gin.H{"type": user.Type, "email": user.Email})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "service-account")
}

func TestAuthMiddleware_OAuth2_MissingHeader(t *testing.T) {
	cfg := &config.Config{
		Google: config.GoogleConfig{AuthType: "oauth2"},
	}
	ts := auth.NewTokenStorage()

	r := gin.New()
	r.Use(AuthMiddleware(cfg, ts))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_OAuth2_InvalidFormat(t *testing.T) {
	cfg := &config.Config{
		Google: config.GoogleConfig{AuthType: "oauth2"},
	}
	ts := auth.NewTokenStorage()

	r := gin.New()
	r.Use(AuthMiddleware(cfg, ts))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
