package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;size:128;not null" json:"email"`
	PasswordHash string    `gorm:"size:256;not null" json:"-"`
	AvatarURL    string    `gorm:"size:512;default:''" json:"avatar,omitempty"`
	StorageUsed  int64     `gorm:"default:0" json:"storage_used"`
	StorageLimit int64     `gorm:"default:1073741824" json:"storage_limit"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
