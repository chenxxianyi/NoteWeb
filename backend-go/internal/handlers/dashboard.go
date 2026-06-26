package handlers

import (
	"net/http"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"github.com/gin-gonic/gin"
)

type DashboardHandlerService interface {
	GetSummary(userID uint) (*models.DashboardSummary, error)
}

type DashboardHandler struct {
	svc DashboardHandlerService
}

func NewDashboardHandler(svc DashboardHandlerService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) GetSummary(c *gin.Context) {
	userID := c.GetUint("userID")
	summary, err := h.svc.GetSummary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}