package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ynastt/avito_test_task_backend_2025/internal/domain"
)

func (h *Handler) CreatePullRequest(c *gin.Context) {
	var req domain.CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	pr, err := h.services.PullRequestService.CreatePullRequest(c.Request.Context(), req)
	if err != nil {
		switch err {
		case domain.ErrPRExists:
			h.errorResponse(c, http.StatusConflict, "PR_EXISTS", err.Error())
		case domain.ErrUserNotFound:
			h.errorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			h.errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	h.successResponse(c, http.StatusCreated, gin.H{"pr": pr})
}

func (h *Handler) MergePullRequest(c *gin.Context) {
	var req domain.MergePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	pr, err := h.services.PullRequestService.MergePullRequest(c.Request.Context(), req.ID)
	if err != nil {
		switch err {
		case domain.ErrPRNotFound:
			h.errorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			h.errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	h.successResponse(c, http.StatusOK, gin.H{"pr": pr})
}

func (h *Handler) ReassignPullRequest(c *gin.Context) {
	var req domain.ReassignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	pr, replacedByID, err := h.services.PullRequestService.ReassignReviewer(c.Request.Context(), req.ID, req.OldReviewerID)
	if err != nil {
		switch err {
		case domain.ErrPRNotFound, domain.ErrUserNotFound:
			h.errorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		case domain.ErrPRMerged:
			h.errorResponse(c, http.StatusConflict, "PR_MERGED", err.Error())
		case domain.ErrNotAssigned:
			h.errorResponse(c, http.StatusConflict, "NOT_ASSIGNED", err.Error())
		case domain.ErrNoCandidate:
			h.errorResponse(c, http.StatusConflict, "NO_CANDIDATE", err.Error())
		default:
			h.errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	response := domain.ReassignResponse{
		PR:         pr,
		ReplacedBy: replacedByID,
	}

	h.successResponse(c, http.StatusOK, response)
}
