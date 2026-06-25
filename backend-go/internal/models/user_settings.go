package models

import "time"

type UserSettings struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	Theme        string    `gorm:"size:16;default:'warm'" json:"theme"`
	Font         string    `gorm:"size:64;default:'Noto Serif SC'" json:"font"`
	ReadingMode  bool      `gorm:"default:true" json:"reading_mode"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
