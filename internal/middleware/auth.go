package middleware

import (
	"net/http"
	"strings"

	"api-go/internal/auth"
	"api-go/internal/config"

	"github.com/gin-gonic/gin"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthUser struct {
	Type  string `json:"type"`
	Email string `json:"email"`
}

func AuthMiddleware(cfg *config.Config, tokenStorage *auth.TokenStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.Google.AuthType == "service-account" {
			c.Set(string(UserContextKey), &AuthUser{
				Type:  "service-account",
				Email: "service-account",
			})
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"statusCode": http.StatusUnauthorized,
				"message":    "Missing authorization header",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"statusCode": http.StatusUnauthorized,
				"message":    "Invalid authorization header format",
			})
			return
		}

		accessToken := parts[1]

		userInfo, err := auth.ValidateToken(c.Request.Context(), accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"statusCode": http.StatusUnauthorized,
				"message":    "Invalid or expired token",
			})
			return
		}

		c.Set(string(UserContextKey), &AuthUser{
			Type:  "oauth2",
			Email: userInfo.Email,
		})
		c.Set("accessToken", accessToken)

		if stored, _, ok := tokenStorage.GetByAccessToken(accessToken); ok {
			c.Set("storedTokens", stored)
		}

		c.Next()
	}
}

func GetAuthUser(c *gin.Context) *AuthUser {
	val, exists := c.Get(string(UserContextKey))
	if !exists {
		return nil
	}
	user, ok := val.(*AuthUser)
	if !ok {
		return nil
	}
	return user
}
