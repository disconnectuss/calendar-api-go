package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"api-go/internal/config"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type Handler struct {
	oauthConfig  *oauth2.Config
	tokenStorage *TokenStorage
	cfg          *config.Config
}

func NewHandler(cfg *config.Config, tokenStorage *TokenStorage) *Handler {
	h := &Handler{
		tokenStorage: tokenStorage,
		cfg:          cfg,
	}
	if cfg.Google.AuthType == "oauth2" {
		h.oauthConfig = NewOAuth2Config(cfg)
	}
	return h
}

// GET /v1/auth/google
func (h *Handler) GoogleAuth(c *gin.Context) {
	if h.oauthConfig == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth2 is not configured"})
		return
	}
	url := h.oauthConfig.AuthCodeURL("state",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GET /v1/auth/google/callback
func (h *Handler) GoogleCallback(c *gin.Context) {
	if h.oauthConfig == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth2 is not configured"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
		return
	}

	token, err := h.oauthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange token: " + err.Error()})
		return
	}

	// Validate token to get user info
	userInfo, err := ValidateToken(c.Request.Context(), token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate token: " + err.Error()})
		return
	}

	// Generate session ID
	sessionBytes := make([]byte, 32)
	if _, err := rand.Read(sessionBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate session ID"})
		return
	}
	sessionID := hex.EncodeToString(sessionBytes)

	// Store tokens
	h.tokenStorage.Store(sessionID, &StoredTokens{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiryDate:   token.Expiry,
		Email:        userInfo.Email,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":   "Authentication successful",
		"sessionId": sessionID,
		"expiresIn": token.Expiry,
	})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	auth.GET("/google", h.GoogleAuth)
	auth.GET("/google/callback", h.GoogleCallback)
}
