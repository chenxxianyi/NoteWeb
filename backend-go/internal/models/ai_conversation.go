package models

import "time"

// Message 对话消息结构
type Message struct {
	Role    string `json:"role"`    // user | assistant | system
	Content string `json:"content"` // 消息内容
}

// AIConversation 存储用户的AI对话历史
type AIConversation struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint      `gorm:"index;not null" json:"user_id"`
	DocumentID       *uint     `gorm:"index" json:"document_id"`                       // 关联文档(可为空,表示独立对话)
	Title            string    `gorm:"size:256" json:"title"`                          // 对话标题
	ConversationType string    `gorm:"size:16;default:'chat'" json:"conversation_type"` // chat | search | summary
	Messages         string    `gorm:"type:text" json:"messages"`                      // JSON格式存储完整对话历史
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AIConversation) TableName() string {
	return "ai_conversations"
}