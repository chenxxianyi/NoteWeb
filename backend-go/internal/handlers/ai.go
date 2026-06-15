package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/service"
)

type AIHandler struct {
	svc *service.AIService
}

func NewAIHandler(svc *service.AIService) *AIHandler {
	return &AIHandler{svc: svc}
}

func (h *AIHandler) Summary(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.GetSummary(0))
}

func (h *AIHandler) Explain(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}
	c.JSON(http.StatusOK, h.svc.Explain(req.Text))
}

func (h *AIHandler) Translate(c *gin.Context) {
	var req struct {
		Text       string `json:"text" binding:"required"`
		TargetLang string `json:"target_lang"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}
	if req.TargetLang == "" {
		req.TargetLang = "zh"
	}
	c.JSON(http.StatusOK, h.svc.Translate(req.Text, req.TargetLang))
}

func (h *AIHandler) Chat(c *gin.Context) {
	var req struct {
		DocumentID uint   `json:"document_id"`
		Question   string `json:"question" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}
	c.JSON(http.StatusOK, h.svc.Chat(req.DocumentID, req.Question))
}
