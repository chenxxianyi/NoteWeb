package models

import "time"

type Note struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID             uint      `gorm:"index;not null" json:"user_id"`
	DocumentID         *uint     `gorm:"index" json:"document_id,omitempty"`
	SourceAnnotationID *uint     `json:"source_annotation_id,omitempty"`
	Title              string    `gorm:"size:256;default:''" json:"title"`
	Content            string    `gorm:"type:text" json:"content"`
	ContentType        string    `gorm:"size:16;default:'markdown'" json:"content_type"`
	Tags               string    `gorm:"size:256;default:''" json:"-"`
	DeletedAt          *time.Time `gorm:"index" json:"-"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	// Computed fields
	DocumentTitle *string `gorm:"-" json:"document_title,omitempty"`
	TagList       []string `gorm:"-" json:"tags"`
}
