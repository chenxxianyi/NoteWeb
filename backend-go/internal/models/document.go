package models

import "time"

type Document struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint      `gorm:"index;not null" json:"user_id"`
	Title          string    `gorm:"size:256;not null" json:"title"`
	FileName       string    `gorm:"size:256;not null" json:"file_name"`
	FileType       string    `gorm:"size:16;not null" json:"file_type"`
	MimeType       string    `gorm:"size:64;default:''" json:"mime_type"`
	FileSize       int64     `gorm:"default:0" json:"file_size"`
	StoragePath    string    `gorm:"size:512;default:''" json:"storage_path"`
	ParsedStatus   string    `gorm:"size:16;default:'pending'" json:"parsed_status"`
	ParsedContent  string    `gorm:"type:text" json:"parsed_content,omitempty"`
	PageCount      int       `gorm:"default:0" json:"page_count,omitempty"`
	WordCount      int       `gorm:"default:0" json:"word_count,omitempty"`
	ReadProgress   float64   `gorm:"default:0" json:"read_progress"`
	DeletedAt      *time.Time `gorm:"index" json:"-"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
