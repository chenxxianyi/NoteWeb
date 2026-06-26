package models

import "time"

// AIProviderConfig 存储用户的AI服务商配置
type AIProviderConfig struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	Provider  string    `gorm:"size:32;default:'mock'" json:"provider"` // mock | deepseek | openai
	APIKey    string    `gorm:"size:256" json:"-"`                       // 不返回给前端,加密存储
	BaseURL   string    `gorm:"size:256" json:"base_url"`
	Model     string    `gorm:"size:64" json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AIProviderConfig) TableName() string {
	return "ai_provider_configs"
}