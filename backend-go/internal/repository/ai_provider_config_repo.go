package repository

import (
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"gorm.io/gorm"
)

type AIProviderConfigRepo struct {
	db *gorm.DB
}

// NewAIProviderConfigRepo 创建新的AIProviderConfig仓库
func NewAIProviderConfigRepo(db *gorm.DB) *AIProviderConfigRepo {
	return &AIProviderConfigRepo{db: db}
}

// GetByUserID 根据用户ID获取AI配置
func (r *AIProviderConfigRepo) GetByUserID(userID uint) (*models.AIProviderConfig, error) {
	var config models.AIProviderConfig
	err := r.db.Where("user_id = ?", userID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Upsert 创建或更新AI配置
func (r *AIProviderConfigRepo) Upsert(config *models.AIProviderConfig) error {
	// 检查是否已存在
	existing, err := r.GetByUserID(config.UserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if existing != nil {
		// 更新
		config.ID = existing.ID
		config.CreatedAt = existing.CreatedAt
		return r.db.Save(config).Error
	}

	// 创建新记录
	return r.db.Create(config).Error
}
