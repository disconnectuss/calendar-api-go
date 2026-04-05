package webhooks

import (
	"log/slog"
	"net/http"

	"api-go/internal/auth"
	"api-go/internal/common"
	"api-go/internal/config"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

type Handler struct {
	service  *Service
	cfg      *config.Config
	validate *common.Validator
}

func NewHandler(cfg *config.Config, service *Service) *Handler {
	return &Handler{
		service:  service,
		cfg:      cfg,
		validate: common.NewValidator(),
	}
}

func (h *Handler) getClientOption(c *gin.Context) option.ClientOption {
	if h.cfg.Google.AuthType == "service-account" {
		opt, err := auth.NewServiceAccountOption(h.cfg.Google.ServiceAccountPath, h.cfg.Google.Scopes)
		if err != nil {
			return nil
		}
		return opt
	}
	accessToken, _ := c.Get("accessToken")
	token := &oauth2.Token{AccessToken: accessToken.(string)}
	return option.WithTokenSource(oauth2.StaticTokenSource(token))
}

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

func (h *Handler) Notifications(c *gin.Context) {
	channelID := c.GetHeader("X-Goog-Channel-Id")
	resourceState := c.GetHeader("X-Goog-Resource-State")
	resourceID := c.GetHeader("X-Goog-Resource-Id")

	if channelID == "" {
		c.JSON(http.StatusBadRequest, common.BadRequestError("missing X-Goog-Channel-Id header"))
		return
	}

	// Reject notifications for channels we didn't create
	if !h.service.HasChannel(channelID) {
		slog.Warn("webhook notification from unknown channel", "channelId", channelID)
		c.Status(http.StatusForbidden)
		return
	}

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

func (h *Handler) ListChannels(c *gin.Context) {
	channels := h.service.ListChannels()
	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	webhooks := rg.Group("/webhooks")

	webhooks.POST("/notifications", h.Notifications)

	protected := webhooks.Group("")
	protected.Use(authMiddleware)
	protected.POST("/subscribe", h.Subscribe)
	protected.POST("/unsubscribe", h.Unsubscribe)
	protected.GET("/channels", h.ListChannels)
}
