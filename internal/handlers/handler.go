package handlers

import (
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/ynastt/avito_test_task_backend_2025/internal/domain"
	"github.com/ynastt/avito_test_task_backend_2025/internal/service"
)

type Handler struct {
	services *service.Services
	logger   *slog.Logger
}

func NewHandler(services *service.Services, logger *slog.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	config := cors.DefaultConfig() // CORS
	config.AllowAllOrigins = true  // разрешить все источники
	config.AllowMethods = []string{"GET", "POST"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}

	router.Use(cors.New(config))

	team := router.Group("/team")
	{
		team.POST("/add", h.CreateTeam)
		team.GET("/get", h.GetTeam)
	}

	users := router.Group("/users")
	{
		users.POST("/setIsActive", h.SetIsActive)
		users.GET("/getReview", h.GetReview)
	}

	pullRequest := router.Group("/pullRequest")
	{
		pullRequest.POST("/create", h.CreatePullRequest)
		pullRequest.POST("/merge", h.MergePullRequest)
		pullRequest.POST("/reassign", h.ReassignPullRequest)
	}

	//endpoint для статистики
	router.GET("/stats", h.GetStatistics)

	return router
}

func (h *Handler) errorResponse(c *gin.Context, status int, code, message string) {
	h.logger.Error("handler error", "code", code, "message", message, "status", status)
	c.JSON(status, domain.ErrorResponse{
		Error: domain.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func (h *Handler) successResponse(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}
