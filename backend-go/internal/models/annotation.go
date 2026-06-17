package models

import "time"

type Annotation struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint      `gorm:"index;not null" json:"user_id"`
	DocumentID     uint      `gorm:"index;not null" json:"document_id"`
	PageNumber     int       `gorm:"default:1" json:"page"`
	AnnotationType string    `gorm:"size:16;default:'highlight'" json:"type"`
	Color          string    `gorm:"size:32;default:'#FFD700'" json:"color"`
	SelectedText   string    `gorm:"type:text" json:"selected_text"`
	Note           string    `gorm:"type:text" json:"note"`
	PositionData   string    `gorm:"type:text" json:"position_data"`
	DeletedAt      *time.Time `gorm:"index" json:"-"`
	CreatedAt      time.Time  `json:"created_at"`
}
