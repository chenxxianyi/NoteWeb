package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/service"
)

type NoteHandler struct {
	svc *service.NoteService
}

func NewNoteHandler(svc *service.NoteService) *NoteHandler {
	return &NoteHandler{svc: svc}
}

func (h *NoteHandler) List(c *gin.Context) {
	userID := c.GetUint("userID")
	docID := c.Query("document_id")
	tag := c.Query("tag")

	notes, err := h.svc.List(userID, docID, tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notes)
}

func (h *NoteHandler) Get(c *gin.Context) {
	userID := c.GetUint("userID")
	noteID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	note, err := h.svc.GetByID(uint(noteID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, note)
}

func (h *NoteHandler) Create(c *gin.Context) {
	userID := c.GetUint("userID")
	var req service.NoteCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}

	note, err := h.svc.Create(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, note)
}

func (h *NoteHandler) Update(c *gin.Context) {
	userID := c.GetUint("userID")
	noteID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req service.NoteUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}

	note, err := h.svc.Update(uint(noteID), userID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, note)
}

func (h *NoteHandler) Delete(c *gin.Context) {
	userID := c.GetUint("userID")
	noteID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.svc.Delete(uint(noteID), userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "ok"})
}
