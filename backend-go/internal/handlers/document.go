package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/service"
	"github.com/gin-gonic/gin"
)

type DocumentHandlerService interface {
	List(userID uint, search, fileType string, page, pageSize int) ([]service.DocumentResponse, error)
	GetDetail(docID, userID uint) (*service.DocumentDetailResponse, error)
	Upload(userID uint, fileName string, fileSize int64, reader io.Reader) (*service.DocumentResponse, error)
	Rename(docID, userID uint, title string) error
	Delete(docID, userID uint) error
	GetContent(docID, userID uint) (*service.DocumentDetailResponse, error)
	GetFileData(docID, userID uint) ([]byte, string, error)
	UpdateReadProgress(docID, userID uint, progress float64) error
	UpdateTextContent(docID, userID uint, content string) error
	UploadAsset(docID, userID uint, fileName string, reader io.Reader) (string, error)
	GetAssetData(docID uint, assetName string) ([]byte, string, error)
}

type DocumentHandler struct {
	svc DocumentHandlerService
}

func NewDocumentHandler(svc DocumentHandlerService) *DocumentHandler {
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

func (h *DocumentHandler) UploadAsset(c *gin.Context) {
	userID := c.GetUint("userID")
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "请选择图片"})
		return
	}
	defer file.Close()

	url, err := h.svc.UploadAsset(uint(docID), userID, header.Filename, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *DocumentHandler) GetAsset(c *gin.Context) {
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	assetName := c.Param("name")

	data, mimeType, err := h.svc.GetAssetData(uint(docID), assetName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		return
	}
	c.Data(http.StatusOK, mimeType, data)
}

func (h *DocumentHandler) UpdateContent(c *gin.Context) {
	userID := c.GetUint("userID")
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}

	if err := h.svc.UpdateTextContent(uint(docID), userID, req.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "ok"})
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
