package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetStatistics(c *gin.Context) {
    // Параметр для включения детальной статистики
    includeDetails := false
    if detailsParam := c.Query("details"); detailsParam != "" {
        if parsed, err := strconv.ParseBool(detailsParam); err == nil {
            includeDetails = parsed
        }
    }
    
    stats, err := h.services.StatsService.GetStats(c.Request.Context(), includeDetails)
    if err != nil {
        h.errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get statistics")
        return
    }
    
    h.successResponse(c, http.StatusOK, gin.H{
        "stats": stats,
    })
}