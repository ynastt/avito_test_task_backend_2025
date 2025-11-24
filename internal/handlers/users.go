package handlers

import (
	"net/http"

	"ynastt/avito_test_task_backend_2025/internal/domain"

	"github.com/gin-gonic/gin"
)

func (h *Handler) SetIsActive(c *gin.Context) {
	var req domain.SetActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	user, err := h.services.UserService.SetIsActive(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			h.errorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			h.errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	h.successResponse(c, http.StatusOK, gin.H{"user": user})
}

func (h *Handler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		h.errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "user_id is required")
		return
	}

	response, err := h.services.UserService.GetUserReviewerPRs(c.Request.Context(), userID)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			h.errorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			h.errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	h.successResponse(c, http.StatusOK, response)
}
