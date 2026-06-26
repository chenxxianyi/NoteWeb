package repository

import (
	"encoding/json"
	"time"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"gorm.io/gorm"
)

type AIConversationRepo struct {
	db *gorm.DB
}

// NewAIConversationRepo 创建新的AIConversation仓库
func NewAIConversationRepo(db *gorm.DB) *AIConversationRepo {
	return &AIConversationRepo{db: db}
}

// Create 创建新的AI对话
func (r *AIConversationRepo) Create(conv *models.AIConversation) error {
	now := time.Now()
	conv.CreatedAt = now
	conv.UpdatedAt = now
	return r.db.Create(conv).Error
}

// GetByID 根据ID获取AI对话
func (r *AIConversationRepo) GetByID(id uint) (*models.AIConversation, error) {
	var conv models.AIConversation
	err := r.db.First(&conv, id).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// ListByUser 获取用户的所有AI对话
func (r *AIConversationRepo) ListByUser(userID uint, limit int) ([]models.AIConversation, error) {
	if limit == 0 {
		limit = 50
	}
	var convs []models.AIConversation
	err := r.db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Limit(limit).
		Find(&convs).Error
	return convs, err
}

// ListByDocument 获取用户在指定文档上的所有AI对话
func (r *AIConversationRepo) ListByDocument(userID, documentID uint) ([]models.AIConversation, error) {
	var convs []models.AIConversation
	err := r.db.Where("user_id = ? AND document_id = ?", userID, documentID).
		Order("updated_at DESC").
		Find(&convs).Error
	return convs, err
}

// Update 更新AI对话
func (r *AIConversationRepo) Update(conv *models.AIConversation) error {
	now := time.Now()
	conv.UpdatedAt = now
	return r.db.Save(conv).Error
}

// Delete 删除AI对话
func (r *AIConversationRepo) Delete(id uint) error {
	return r.db.Delete(&models.AIConversation{}, id).Error
}

// GetMessages 获取对话的消息列表(反序列化JSON)
func (r *AIConversationRepo) GetMessages(conv *models.AIConversation) ([]models.Message, error) {
	if conv.Messages == "" {
		return []models.Message{}, nil
	}

	var msgs []models.Message
	err := json.Unmarshal([]byte(conv.Messages), &msgs)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

// SaveMessages 保存对话的消息列表(序列化为JSON)
func (r *AIConversationRepo) SaveMessages(conv *models.AIConversation, messages []models.Message) error {
	data, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	conv.Messages = string(data)
	return r.db.Model(conv).UpdateColumn("messages", string(data)).Error
}
