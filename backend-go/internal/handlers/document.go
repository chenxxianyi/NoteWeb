package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/service"
)

type DocumentHandler struct {
	svc *service.DocumentService
}

func NewDocumentHandler(svc *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

func (h *DocumentHandler) List(c *gin.Context) {
	userID := c.GetUint("userID")
	search := c.Query("search")
	fileType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	docs, err := h.svc.List(userID, search, fileType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, docs)
}

func (h *DocumentHandler) Get(c *gin.Context) {
	userID := c.GetUint("userID")
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	doc, err := h.svc.GetDetail(uint(docID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	userID := c.GetUint("userID")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "请选择文件"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "文件读取失败"})
		return
	}

	resp, err := h.svc.Upload(userID, header.Filename, int64(len(data)), bytes.NewReader(data))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *DocumentHandler) Rename(c *gin.Context) {
	userID := c.GetUint("userID")
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}

	if err := h.svc.Rename(uint(docID), userID, req.Title); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "ok"})
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	userID := c.GetUint("userID")
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.svc.Delete(uint(docID), userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "ok"})
}

func (h *DocumentHandler) GetContent(c *gin.Context) {
	userID := c.GetUint("userID")
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	doc, err := h.svc.GetContent(uint(docID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": doc.ParsedContent})
}

func (h *DocumentHandler) GetFile(c *gin.Context) {
	userID := c.GetUint("userID")
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	data, mimeType, err := h.svc.GetFileData(uint(docID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}
	c.Data(http.StatusOK, mimeType, data)
}

func (h *DocumentHandler) UpdateReadProgress(c *gin.Context) {
	userID := c.GetUint("userID")
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		Progress float64 `json:"progress"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}

	if err := h.svc.UpdateReadProgress(uint(docID), userID, req.Progress); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "ok"})
}
