package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/service"
)

type AnnotationHandler struct {
	svc *service.AnnotationService
}

func NewAnnotationHandler(svc *service.AnnotationService) *AnnotationHandler {
	return &AnnotationHandler{svc: svc}
}

func (h *AnnotationHandler) ListByDocument(c *gin.Context) {
	userID := c.GetUint("userID")
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	anns, err := h.svc.ListByDocument(uint(docID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, anns)
}

func (h *AnnotationHandler) Create(c *gin.Context) {
	userID := c.GetUint("userID")
	var req service.AnnotationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}

	ann, err := h.svc.Create(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ann)
}

func (h *AnnotationHandler) Delete(c *gin.Context) {
	userID := c.GetUint("userID")
	annID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.svc.Delete(uint(annID), userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "ok"})
}
