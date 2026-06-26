package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/config"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/repository"
)

type AIService struct {
	convRepo   *repository.AIConversationRepo
	configRepo *repository.AIProviderConfigRepo
	docRepo    *repository.DocumentRepo
	cfg        *config.Config
	httpClient *http.Client
}

func NewAIService(
	convRepo *repository.AIConversationRepo,
	configRepo *repository.AIProviderConfigRepo,
	docRepo *repository.DocumentRepo,
	cfg *config.Config,
) *AIService {
	return &AIService{
		convRepo:   convRepo,
		configRepo: configRepo,
		docRepo:    docRepo,
		cfg:        cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Answer         string `json:"answer"`
	ConversationID uint   `json:"conversation_id"`
}

// SummaryResponse 总结响应
type SummaryResponse struct {
	Summary        string `json:"summary"`
	ConversationID uint   `json:"conversation_id"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Answer         string         `json:"answer"`
	Sources        []SearchSource `json:"sources"`
	ConversationID uint           `json:"conversation_id"`
}

// SearchSource 搜索来源
type SearchSource struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// Chat 处理聊天请求
func (s *AIService) Chat(ctx context.Context, userID, documentID uint, question string, conversationType string) (*ChatResponse, error) {
	userConfig, err := s.getActiveAIConfig(userID)
	if err != nil {
		return nil, err
	}

	// 获取或创建对话
	var conv *models.AIConversation
	var messages []models.Message

	// 构建消息
	if documentID > 0 {
		// 有文档上下文
		docContext, err := s.getDocumentContext(documentID)
		if err == nil && docContext != "" {
			// 添加文档上下文作为系统消息
			messages = append(messages, models.Message{
				Role:    "system",
				Content: fmt.Sprintf("你是一个智能文档助手。以下是文档的内容:\n\n%s\n\n请基于文档内容回答用户问题。", truncateText(docContext, 4000)),
			})
		}
	}

	messages = append(messages, models.Message{
		Role:    "user",
		Content: question,
	})

	answer, err := s.callDeepSeekAPI(ctx, userConfig, messages, false)
	if err != nil {
		return nil, err
	}

	// 保存对话
	messages = append(messages, models.Message{
		Role:    "assistant",
		Content: answer,
	})

	messagesJSON, _ := json.Marshal(messages)
	conv = &models.AIConversation{
		UserID:           userID,
		DocumentID:       &documentID,
		Title:            truncateText(question, 50),
		ConversationType: conversationType,
		Messages:         string(messagesJSON),
	}

	if err := s.convRepo.Create(conv); err != nil {
		return nil, err
	}

	return &ChatResponse{
		Answer:         answer,
		ConversationID: conv.ID,
	}, nil
}

// ChatStream 处理流式聊天请求
func (s *AIService) ChatStream(ctx context.Context, userID, documentID uint, question string, conversationType string, onDelta func(string) error) (*ChatResponse, error) {
	userConfig, err := s.getActiveAIConfig(userID)
	if err != nil {
		return nil, err
	}

	var conv *models.AIConversation
	var messages []models.Message

	if documentID > 0 {
		docContext, err := s.getDocumentContext(documentID)
		if err == nil && docContext != "" {
			messages = append(messages, models.Message{
				Role:    "system",
				Content: fmt.Sprintf("你是一个智能文档助手。以下是文档的内容:\n\n%s\n\n请基于文档内容回答用户问题。", truncateText(docContext, 4000)),
			})
		}
	}

	messages = append(messages, models.Message{
		Role:    "user",
		Content: question,
	})

	answer, err := s.callDeepSeekAPIStream(ctx, userConfig, messages, false, onDelta)
	if err != nil {
		return nil, err
	}

	messages = append(messages, models.Message{
		Role:    "assistant",
		Content: answer,
	})

	messagesJSON, _ := json.Marshal(messages)
	conv = &models.AIConversation{
		UserID:           userID,
		DocumentID:       &documentID,
		Title:            truncateText(question, 50),
		ConversationType: conversationType,
		Messages:         string(messagesJSON),
	}

	if err := s.convRepo.Create(conv); err != nil {
		return nil, err
	}

	return &ChatResponse{
		Answer:         answer,
		ConversationID: conv.ID,
	}, nil
}

// GetSummary 获取文档总结
func (s *AIService) GetSummary(ctx context.Context, userID, documentID uint) (*SummaryResponse, error) {
	userConfig, err := s.getActiveAIConfig(userID)
	if err != nil {
		return nil, err
	}

	// 获取文档内容
	docContext, err := s.getDocumentContext(documentID)
	if err != nil {
		return nil, err
	}

	// 构建总结请求
	messages := []models.Message{
		{
			Role:    "system",
			Content: "你是一个专业的文档总结助手。请为用户提供的文档生成一个简洁、准确的总结。总结应包括文档的主要内容和核心观点。",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("请总结以下文档:\n\n%s", truncateText(docContext, 6000)),
		},
	}

	summary, err := s.callDeepSeekAPI(ctx, userConfig, messages, false)
	if err != nil {
		return nil, err
	}

	// 保存对话
	messagesJSON, _ := json.Marshal(messages)
	conv := &models.AIConversation{
		UserID:           userID,
		DocumentID:       &documentID,
		Title:            "文档总结",
		ConversationType: "summary",
		Messages:         string(messagesJSON),
	}

	if err := s.convRepo.Create(conv); err != nil {
		return nil, err
	}

	return &SummaryResponse{
		Summary:        summary,
		ConversationID: conv.ID,
	}, nil
}

// GetSummaryStream 获取文档流式总结
func (s *AIService) GetSummaryStream(ctx context.Context, userID, documentID uint, onDelta func(string) error) (*SummaryResponse, error) {
	userConfig, err := s.getActiveAIConfig(userID)
	if err != nil {
		return nil, err
	}

	docContext, err := s.getDocumentContext(documentID)
	if err != nil {
		return nil, err
	}

	messages := []models.Message{
		{
			Role:    "system",
			Content: "你是一个专业的文档总结助手。请为用户提供的文档生成一个简洁、准确的总结。总结应包括文档的主要内容和核心观点。",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("请总结以下文档:\n\n%s", truncateText(docContext, 6000)),
		},
	}

	summary, err := s.callDeepSeekAPIStream(ctx, userConfig, messages, false, onDelta)
	if err != nil {
		return nil, err
	}

	messages = append(messages, models.Message{
		Role:    "assistant",
		Content: summary,
	})
	messagesJSON, _ := json.Marshal(messages)
	conv := &models.AIConversation{
		UserID:           userID,
		DocumentID:       &documentID,
		Title:            "文档总结",
		ConversationType: "summary",
		Messages:         string(messagesJSON),
	}

	if err := s.convRepo.Create(conv); err != nil {
		return nil, err
	}

	return &SummaryResponse{
		Summary:        summary,
		ConversationID: conv.ID,
	}, nil
}

// Search 执行联网搜索
func (s *AIService) Search(ctx context.Context, userID uint, query string) (*SearchResponse, error) {
	userConfig, err := s.getActiveAIConfig(userID)
	if err != nil {
		return nil, err
	}

	// 构建搜索请求
	messages := []models.Message{
		{
			Role:    "system",
			Content: "你是一个智能搜索助手。用户会给你一个搜索查询,你需要提供准确、有用的信息。如果可能,请提供相关的来源链接。",
		},
		{
			Role:    "user",
			Content: query,
		},
	}

	answer, err := s.callDeepSeekAPI(ctx, userConfig, messages, true)
	if err != nil {
		return nil, err
	}

	// 保存对话
	messagesJSON, _ := json.Marshal(messages)
	conv := &models.AIConversation{
		UserID:           userID,
		DocumentID:       nil,
		Title:            truncateText(query, 50),
		ConversationType: "search",
		Messages:         string(messagesJSON),
	}

	if err := s.convRepo.Create(conv); err != nil {
		return nil, err
	}

	// 暂时不返回来源(DeepSeek会在回答中包含来源信息)
	return &SearchResponse{
		Answer:         answer,
		Sources:        []SearchSource{},
		ConversationID: conv.ID,
	}, nil
}

// SearchStream 执行流式联网搜索
func (s *AIService) SearchStream(ctx context.Context, userID uint, query string, onDelta func(string) error) (*SearchResponse, error) {
	userConfig, err := s.getActiveAIConfig(userID)
	if err != nil {
		return nil, err
	}

	messages := []models.Message{
		{
			Role:    "system",
			Content: "你是一个智能搜索助手。用户会给你一个搜索查询,你需要提供准确、有用的信息。如果可能,请提供相关的来源链接。",
		},
		{
			Role:    "user",
			Content: query,
		},
	}

	answer, err := s.callDeepSeekAPIStream(ctx, userConfig, messages, true, onDelta)
	if err != nil {
		return nil, err
	}

	messages = append(messages, models.Message{
		Role:    "assistant",
		Content: answer,
	})
	messagesJSON, _ := json.Marshal(messages)
	conv := &models.AIConversation{
		UserID:           userID,
		DocumentID:       nil,
		Title:            truncateText(query, 50),
		ConversationType: "search",
		Messages:         string(messagesJSON),
	}

	if err := s.convRepo.Create(conv); err != nil {
		return nil, err
	}

	return &SearchResponse{
		Answer:         answer,
		Sources:        []SearchSource{},
		ConversationID: conv.ID,
	}, nil
}

// GetConversations 获取用户的对话历史
func (s *AIService) GetConversations(userID, documentID uint) ([]models.AIConversation, error) {
	if documentID > 0 {
		return s.convRepo.ListByDocument(userID, documentID)
	}
	return s.convRepo.ListByUser(userID, 50)
}

// DeleteConversation 删除对话
func (s *AIService) DeleteConversation(userID, conversationID uint) error {
	conv, err := s.convRepo.GetByID(conversationID)
	if err != nil {
		return err
	}

	// 验证所有权
	if conv.UserID != userID {
		return errors.New("无权删除此对话")
	}

	return s.convRepo.Delete(conversationID)
}

func (s *AIService) getActiveAIConfig(userID uint) (*models.AIProviderConfig, error) {
	userConfig, err := s.configRepo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("未配置 AI 服务，请先在设置页保存 DeepSeek API Key")
	}

	userConfig.Provider = strings.ToLower(strings.TrimSpace(userConfig.Provider))
	userConfig.Model = strings.TrimSpace(userConfig.Model)
	userConfig.BaseURL = strings.TrimSpace(userConfig.BaseURL)
	userConfig.APIKey = strings.TrimSpace(userConfig.APIKey)

	if userConfig.Provider == "" || userConfig.Provider == "mock" {
		return nil, errors.New("Mock AI 服务已停用，请在设置页配置 DeepSeek API Key")
	}
	if userConfig.Model == "" {
		if userConfig.Provider == "deepseek" {
			userConfig.Model = "deepseek-chat"
		} else {
			return nil, errors.New("未配置 AI 模型，请在设置页选择模型")
		}
	}
	if userConfig.Provider != "deepseek" && userConfig.BaseURL == "" {
		return nil, errors.New("未配置 AI Base URL，请在设置页填写 API 端点地址")
	}
	if userConfig.APIKey == "" && !(userConfig.Provider == "deepseek" && strings.TrimSpace(s.cfg.DeepSeekAPIKey) != "") {
		return nil, errors.New("未配置 AI API Key，请在设置页保存 API Key")
	}

	return userConfig, nil
}

// callDeepSeekAPI 调用DeepSeek API
func (s *AIService) callDeepSeekAPI(ctx context.Context, config *models.AIProviderConfig, messages []models.Message, searchMode bool) (string, error) {
	// 确定API地址
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = s.cfg.DeepSeekBaseURL
	}

	// 构建请求体
	reqBody := map[string]interface{}{
		"model": config.Model,
		"messages": func() []map[string]string {
			result := make([]map[string]string, len(messages))
			for i, msg := range messages {
				result[i] = map[string]string{
					"role":    msg.Role,
					"content": msg.Content,
				}
			}
			return result
		}(),
		"temperature": 0.7,
	}

	// 如果是搜索模式,添加搜索参数
	if searchMode {
		reqBody["search_mode"] = true
	}

	// 序列化请求
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// 创建请求
	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(baseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	apiKey := strings.TrimSpace(config.APIKey)
	if apiKey == "" && config.Provider == "deepseek" {
		apiKey = strings.TrimSpace(s.cfg.DeepSeekAPIKey)
	}
	if apiKey == "" {
		return "", errors.New("未配置 AI API Key，请在设置页保存 API Key")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API请求失败: %s", string(body))
	}

	// 解析响应
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", errors.New("API返回空响应")
	}

	return result.Choices[0].Message.Content, nil
}

// callDeepSeekAPIStream 调用 DeepSeek 流式 API, 并把增量内容交给调用方
func (s *AIService) callDeepSeekAPIStream(ctx context.Context, config *models.AIProviderConfig, messages []models.Message, searchMode bool, onDelta func(string) error) (string, error) {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = s.cfg.DeepSeekBaseURL
	}

	reqBody := map[string]interface{}{
		"model": config.Model,
		"messages": func() []map[string]string {
			result := make([]map[string]string, len(messages))
			for i, msg := range messages {
				result[i] = map[string]string{
					"role":    msg.Role,
					"content": msg.Content,
				}
			}
			return result
		}(),
		"temperature": 0.7,
		"stream":      true,
	}

	if searchMode {
		reqBody["search_mode"] = true
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(baseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	apiKey := strings.TrimSpace(config.APIKey)
	if apiKey == "" && config.Provider == "deepseek" {
		apiKey = strings.TrimSpace(s.cfg.DeepSeekAPIKey)
	}
	if apiKey == "" {
		return "", errors.New("未配置 AI API Key，请在设置页保存 API Key")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API请求失败: %s", string(body))
	}

	var answer strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") || !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}

		var result struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return "", err
		}
		if len(result.Choices) == 0 {
			continue
		}

		delta := result.Choices[0].Delta.Content
		if delta == "" {
			continue
		}

		answer.WriteString(delta)
		if onDelta != nil {
			if err := onDelta(delta); err != nil {
				return "", err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	if answer.Len() == 0 {
		return "", errors.New("API返回空响应")
	}

	return answer.String(), nil
}

// getDocumentContext 获取文档内容上下文
func (s *AIService) getDocumentContext(documentID uint) (string, error) {
	doc, err := s.docRepo.GetByID(documentID)
	if err != nil {
		return "", err
	}

	if doc.ParsedContent == "" {
		return "", errors.New("文档内容为空")
	}

	return doc.ParsedContent, nil
}

// truncateText 截断文本
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
