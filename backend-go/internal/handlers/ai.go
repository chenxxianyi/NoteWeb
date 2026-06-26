package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/repository"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/service"
	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	svc        *service.AIService
	configRepo *repository.AIProviderConfigRepo
}

func NewAIHandler(svc *service.AIService, configRepo *repository.AIProviderConfigRepo) *AIHandler {
	return &AIHandler{svc: svc, configRepo: configRepo}
}

// Chat 聊天接口
func (h *AIHandler) Chat(c *gin.Context) {
	var req struct {
		DocumentID       *uint  `json:"document_id"`
		Question         string `json:"question" binding:"required"`
		ConversationType string `json:"conversation_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}

	// 默认对话类型为chat
	if req.ConversationType == "" {
		req.ConversationType = "chat"
	}

	// 获取当前用户
	userID := c.MustGet("userID").(uint)

	// 处理documentID
	var docID uint
	if req.DocumentID != nil {
		docID = *req.DocumentID
	} else {
		docID = 0
	}

	// 调用服务
	resp, err := h.svc.Chat(c.Request.Context(), userID, docID, req.Question, req.ConversationType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ChatStream 聊天流式接口
func (h *AIHandler) ChatStream(c *gin.Context) {
	var req struct {
		DocumentID       *uint  `json:"document_id"`
		Question         string `json:"question" binding:"required"`
		ConversationType string `json:"conversation_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}

	if req.ConversationType == "" {
		req.ConversationType = "chat"
	}

	userID := c.MustGet("userID").(uint)

	var docID uint
	if req.DocumentID != nil {
		docID = *req.DocumentID
	}

	prepareSSE(c)

	resp, err := h.svc.ChatStream(c.Request.Context(), userID, docID, req.Question, req.ConversationType, func(delta string) error {
		return writeSSE(c, "delta", gin.H{"content": delta})
	})
	if err != nil {
		_ = writeSSE(c, "error", gin.H{"detail": err.Error()})
		return
	}

	_ = writeSSE(c, "done", gin.H{"conversation_id": resp.ConversationID})
}

// GetSummary 获取文档总结
func (h *AIHandler) GetSummary(c *gin.Context) {
	// 获取文档ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "无效的文档ID"})
		return
	}

	// 获取当前用户
	userID := c.MustGet("userID").(uint)

	// 调用服务
	resp, err := h.svc.GetSummary(c.Request.Context(), userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetSummaryStream 获取文档流式总结
func (h *AIHandler) GetSummaryStream(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "无效的文档ID"})
		return
	}

	userID := c.MustGet("userID").(uint)

	prepareSSE(c)
	resp, err := h.svc.GetSummaryStream(c.Request.Context(), userID, uint(id), func(delta string) error {
		return writeSSE(c, "delta", gin.H{"content": delta})
	})
	if err != nil {
		_ = writeSSE(c, "error", gin.H{"detail": err.Error()})
		return
	}

	_ = writeSSE(c, "done", gin.H{"conversation_id": resp.ConversationID})
}

// Search 联网搜索
func (h *AIHandler) Search(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}

	// 获取当前用户
	userID := c.MustGet("userID").(uint)

	// 调用服务
	resp, err := h.svc.Search(c.Request.Context(), userID, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SearchStream 流式联网搜索
func (h *AIHandler) SearchStream(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}

	userID := c.MustGet("userID").(uint)

	prepareSSE(c)
	resp, err := h.svc.SearchStream(c.Request.Context(), userID, req.Query, func(delta string) error {
		return writeSSE(c, "delta", gin.H{"content": delta})
	})
	if err != nil {
		_ = writeSSE(c, "error", gin.H{"detail": err.Error()})
		return
	}

	_ = writeSSE(c, "done", gin.H{"conversation_id": resp.ConversationID})
}

// GetConversations 获取对话历史
func (h *AIHandler) GetConversations(c *gin.Context) {
	// 获取当前用户
	userID := c.MustGet("userID").(uint)

	// 获取文档ID(可选)
	var docID uint
	if docIDStr := c.Query("document_id"); docIDStr != "" {
		id, err := strconv.ParseUint(docIDStr, 10, 32)
		if err == nil {
			docID = uint(id)
		}
	}

	// 调用服务
	convs, err := h.svc.GetConversations(userID, docID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, convs)
}

// DeleteConversation 删除对话
func (h *AIHandler) DeleteConversation(c *gin.Context) {
	// 获取对话ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "无效的对话ID"})
		return
	}

	// 获取当前用户
	userID := c.MustGet("userID").(uint)

	// 调用服务
	if err := h.svc.DeleteConversation(userID, uint(id)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetProviderConfig 获取AI配置
func (h *AIHandler) GetProviderConfig(c *gin.Context) {
	// 获取当前用户
	userID := c.MustGet("userID").(uint)

	// 从数据库获取配置
	config, err := h.configRepo.GetByUserID(userID)
	if err != nil {
		// 没有配置,返回默认值
		c.JSON(http.StatusOK, gin.H{
			"provider":    "deepseek",
			"base_url":    "",
			"model":       "deepseek-chat",
			"has_api_key": false,
		})
		return
	}

	// 返回配置(不包含API Key)
	c.JSON(http.StatusOK, gin.H{
		"provider":    config.Provider,
		"base_url":    config.BaseURL,
		"model":       config.Model,
		"has_api_key": strings.TrimSpace(config.APIKey) != "",
	})
}

// UpdateProviderConfig 更新AI配置
func (h *AIHandler) UpdateProviderConfig(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required"`
		APIKey   string `json:"api_key"`
		BaseURL  string `json:"base_url"`
		Model    string `json:"model"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误"})
		return
	}

	// 获取当前用户
	userID := c.MustGet("userID").(uint)

	provider := normalizeAIProvider(req.Provider)
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "不支持的 AI 提供商"})
		return
	}
	if provider == "mock" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Mock AI 服务已停用，请配置 DeepSeek API Key"})
		return
	}

	model := strings.TrimSpace(req.Model)
	if model == "" && provider == "deepseek" {
		model = "deepseek-chat"
	}

	apiKey := strings.TrimSpace(req.APIKey)
	if apiKey == "" {
		if existing, err := h.configRepo.GetByUserID(userID); err == nil {
			apiKey = existing.APIKey
		}
	}

	// 构建配置对象
	config := &models.AIProviderConfig{
		UserID:   userID,
		Provider: provider,
		APIKey:   apiKey,
		BaseURL:  strings.TrimSpace(req.BaseURL),
		Model:    model,
	}

	// 保存配置
	if err := h.configRepo.Upsert(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "保存配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider":    config.Provider,
		"base_url":    config.BaseURL,
		"model":       config.Model,
		"has_api_key": strings.TrimSpace(config.APIKey) != "",
	})
}

func normalizeAIProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "deepseek":
		return "deepseek"
	case "openai":
		return "openai"
	case "custom", "自定义":
		return "custom"
	case "mock", "mock api":
		return "mock"
	default:
		return ""
	}
}

func prepareSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	c.Writer.Flush()
}

func writeSSE(c *gin.Context, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if event != "" {
		if _, err := fmt.Fprintf(c.Writer, "event: %s\n", event); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", data); err != nil {
		return err
	}
	c.Writer.Flush()
	return nil
}
