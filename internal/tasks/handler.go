package tasks

import (
	"net/http"
	"strconv"

	"api-go/internal/auth"
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

func (h *Handler) ListTaskLists(c *gin.Context) {
	maxResults, _ := strconv.ParseInt(c.DefaultQuery("maxResults", "20"), 10, 64)
	pageToken := c.Query("pageToken")
	opt := h.getClientOption(c)

	result, err := h.service.ListTaskLists(c.Request.Context(), maxResults, pageToken, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) CreateTaskList(c *gin.Context) {
	var req CreateTaskListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}

	opt := h.getClientOption(c)
	taskList, err := h.service.CreateTaskList(c.Request.Context(), &req, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, taskList)
}

func (h *Handler) ListDefaultTasks(c *gin.Context) {
	maxResults, _ := strconv.ParseInt(c.DefaultQuery("maxResults", "20"), 10, 64)
	pageToken := c.Query("pageToken")
	opt := h.getClientOption(c)

	result, err := h.service.ListTasks(c.Request.Context(), "", maxResults, pageToken, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) CreateDefaultTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}

	opt := h.getClientOption(c)
	task, err := h.service.CreateTask(c.Request.Context(), "", &req, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *Handler) ListTasks(c *gin.Context) {
	taskListID := c.Param("taskListId")
	maxResults, _ := strconv.ParseInt(c.DefaultQuery("maxResults", "20"), 10, 64)
	pageToken := c.Query("pageToken")
	opt := h.getClientOption(c)

	result, err := h.service.ListTasks(c.Request.Context(), taskListID, maxResults, pageToken, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) CreateTask(c *gin.Context) {
	taskListID := c.Param("taskListId")
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}

	opt := h.getClientOption(c)
	task, err := h.service.CreateTask(c.Request.Context(), taskListID, &req, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *Handler) GetTask(c *gin.Context) {
	taskListID := c.Param("taskListId")
	taskID := c.Param("taskId")
	opt := h.getClientOption(c)

	task, err := h.service.GetTask(c.Request.Context(), taskListID, taskID, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *Handler) UpdateTask(c *gin.Context) {
	taskListID := c.Param("taskListId")
	taskID := c.Param("taskId")
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.BadRequestError(err.Error()))
		return
	}

	opt := h.getClientOption(c)
	task, err := h.service.UpdateTask(c.Request.Context(), taskListID, taskID, &req, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *Handler) DeleteTask(c *gin.Context) {
	taskListID := c.Param("taskListId")
	taskID := c.Param("taskId")
	opt := h.getClientOption(c)

	if err := h.service.DeleteTask(c.Request.Context(), taskListID, taskID, opt); err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func (h *Handler) CompleteTask(c *gin.Context) {
	taskListID := c.Param("taskListId")
	taskID := c.Param("taskId")
	opt := h.getClientOption(c)

	task, err := h.service.CompleteTask(c.Request.Context(), taskListID, taskID, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *Handler) UncompleteTask(c *gin.Context) {
	taskListID := c.Param("taskListId")
	taskID := c.Param("taskId")
	opt := h.getClientOption(c)

	task, err := h.service.UncompleteTask(c.Request.Context(), taskListID, taskID, opt)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/tasks/lists", h.ListTaskLists)
	rg.POST("/tasks/lists", h.CreateTaskList)

	rg.GET("/tasks", h.ListDefaultTasks)
	rg.POST("/tasks", h.CreateDefaultTask)

	rg.GET("/tasks/lists/:taskListId/tasks", h.ListTasks)
	rg.POST("/tasks/lists/:taskListId/tasks", h.CreateTask)
	rg.GET("/tasks/lists/:taskListId/tasks/:taskId", h.GetTask)
	rg.PATCH("/tasks/lists/:taskListId/tasks/:taskId", h.UpdateTask)
	rg.DELETE("/tasks/lists/:taskListId/tasks/:taskId", h.DeleteTask)
	rg.POST("/tasks/lists/:taskListId/tasks/:taskId/complete", h.CompleteTask)
	rg.POST("/tasks/lists/:taskListId/tasks/:taskId/uncomplete", h.UncompleteTask)
}
