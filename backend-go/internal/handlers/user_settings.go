package handlers

import (
	"net/http"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/service"
	"github.com/gin-gonic/gin"
)

type UserSettingsHandler struct {
	svc *service.UserSettingsService
}

func NewUserSettingsHandler(svc *service.UserSettingsService) *UserSettingsHandler {
	return &UserSettingsHandler{svc: svc}
}

func (h *UserSettingsHandler) Get(c *gin.Context) {
	userID := c.GetUint("userID")
	settings, err := h.svc.GetSettings(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *UserSettingsHandler) Update(c *gin.Context) {
	userID := c.GetUint("userID")

	var req struct {
		Theme       string `json:"theme"`
		Font        string `json:"font"`
		ReadingMode bool   `json:"reading_mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}

	if req.Theme == "" {
		req.Theme = "warm"
	}
	if req.Font == "" {
		req.Font = "Noto Serif SC"
	}

	if err := h.svc.UpdateSettings(userID, req.Theme, req.Font, req.ReadingMode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "ok"})
}