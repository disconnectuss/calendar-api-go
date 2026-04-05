package calendar

import (
	"net/http"
	"strconv"

	"api-go/internal/auth"
	"api-go/internal/common"
	"api-go/internal/config"
	"api-go/internal/middleware"

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

func (h *Handler) ListEvents(c *gin.Context) {
	maxResults, _ := strconv.ParseInt(c.DefaultQuery("maxResults", "20"), 10, 64)
	pageToken := c.Query("pageToken")

	opt := h.getClientOption(c)
	result, err := h.service.ListEvents(c.Request.Context(), maxResults, pageToken, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetEvent(c *gin.Context) {
	id := c.Param("id")
	opt := h.getClientOption(c)

	event, err := h.service.GetEvent(c.Request.Context(), id, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
}

func (h *Handler) CreateGeneralEvent(c *gin.Context) {
	var req CreateGeneralEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}

	opt := h.getClientOption(c)
	event, err := h.service.CreateGeneralEvent(c.Request.Context(), &req, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, event)
}

func (h *Handler) CreatePrivateEvent(c *gin.Context) {
	var req CreatePrivateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}

	opt := h.getClientOption(c)
	event, err := h.service.CreatePrivateEvent(c.Request.Context(), &req, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, event)
}

func (h *Handler) UpdateEvent(c *gin.Context) {
	id := c.Param("id")
	var req UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}

	user := middleware.GetAuthUser(c)
	opt := h.getClientOption(c)

	event, err := h.service.UpdateEvent(c.Request.Context(), id, &req, user.Email, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
}

func (h *Handler) DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	user := middleware.GetAuthUser(c)
	opt := h.getClientOption(c)

	if err := h.service.DeleteEvent(c.Request.Context(), id, user.Email, opt); err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	cal := rg.Group("/calendar/events")
	cal.GET("", h.ListEvents)
	cal.GET("/:id", h.GetEvent)
	cal.POST("/general", h.CreateGeneralEvent)
	cal.POST("/private", h.CreatePrivateEvent)
	cal.PATCH("/:id", h.UpdateEvent)
	cal.DELETE("/:id", h.DeleteEvent)
}
