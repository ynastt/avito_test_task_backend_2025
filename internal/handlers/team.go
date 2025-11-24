package handlers

import (
	"net/http"

	"ynastt/avito_test_task_backend_2025/internal/domain"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateTeam(c *gin.Context) {
	var req domain.Team
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	_, err := h.services.TeamService.CreateTeam(c.Request.Context(), req)
	if err != nil {
		switch err {
		case domain.ErrTeamExists:
			h.errorResponse(c, http.StatusBadRequest, "TEAM_EXISTS", err.Error())
		default:
			h.errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	h.successResponse(c, http.StatusCreated, gin.H{"team": req})
}

func (h *Handler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		h.errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "team_name is required")
		return
	}

	team, err := h.services.TeamService.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		switch err {
		case domain.ErrTeamNotFound:
			h.errorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			h.errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	h.successResponse(c, http.StatusOK, team)
}
