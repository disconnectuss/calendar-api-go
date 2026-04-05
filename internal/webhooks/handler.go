package webhooks

import (
	"log/slog"
	"net/http"

	"api-go/internal/common"
	"api-go/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

type Handler struct {
	service  *Service
	cfg      *config.Config
	validate *validator.Validate
}

func NewHandler(cfg *config.Config, service *Service) *Handler {
	return &Handler{
		service:  service,
		cfg:      cfg,
		validate: validator.New(),
	}
}

func (h *Handler) getClientOption(c *gin.Context) option.ClientOption {
	if h.cfg.Google.AuthType == "service-account" {
		return option.WithCredentialsFile(h.cfg.Google.ServiceAccountPath)
	}
	accessToken, _ := c.Get("accessToken")
	token := &oauth2.Token{AccessToken: accessToken.(string)}
	return option.WithTokenSource(oauth2.StaticTokenSource(token))
}

// POST /v1/webhooks/subscribe
func (h *Handler) Subscribe(c *gin.Context) {
	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}

	opt := h.getClientOption(c)
	channel, err := h.service.Subscribe(c.Request.Context(), &req, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, channel)
}

// POST /v1/webhooks/unsubscribe
func (h *Handler) Unsubscribe(c *gin.Context) {
	var req StopWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}

	opt := h.getClientOption(c)
	if err := h.service.Unsubscribe(c.Request.Context(), &req, opt); err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Webhook unsubscribed successfully"})
}

// POST /v1/webhooks/notifications (no auth)
func (h *Handler) Notifications(c *gin.Context) {
	channelID := c.GetHeader("X-Goog-Channel-Id")
	resourceState := c.GetHeader("X-Goog-Resource-State")
	resourceID := c.GetHeader("X-Goog-Resource-Id")

	slog.Info("webhook notification received",
		"channelId", channelID,
		"resourceState", resourceState,
		"resourceId", resourceID,
	)

	switch resourceState {
	case "sync":
		slog.Info("webhook sync confirmation", "channelId", channelID)
	case "exists":
		slog.Info("resource updated", "channelId", channelID, "resourceId", resourceID)
	case "not_exists":
		slog.Info("resource deleted", "channelId", channelID, "resourceId", resourceID)
	}

	c.Status(http.StatusOK)
}

// GET /v1/webhooks/channels
func (h *Handler) ListChannels(c *gin.Context) {
	channels := h.service.ListChannels()
	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	webhooks := rg.Group("/webhooks")

	// No auth for notifications
	webhooks.POST("/notifications", h.Notifications)

	// Auth required
	protected := webhooks.Group("")
	protected.Use(authMiddleware)
	protected.POST("/subscribe", h.Subscribe)
	protected.POST("/unsubscribe", h.Unsubscribe)
	protected.GET("/channels", h.ListChannels)
}
