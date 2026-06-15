package repository

import (
	"time"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"gorm.io/gorm"
)

type DocumentRepo struct {
	db *gorm.DB
}

func NewDocumentRepo(db *gorm.DB) *DocumentRepo {
	return &DocumentRepo{db: db}
}

func (r *DocumentRepo) GetByID(id uint) (*models.Document, error) {
	var doc models.Document
	err := r.db.Where("deleted_at IS NULL").First(&doc, id).Error
	return &doc, err
}

func (r *DocumentRepo) ListByUser(userID uint, search, fileType string, page, pageSize int) ([]models.Document, error) {
	query := r.db.Where("user_id = ? AND deleted_at IS NULL", userID)
	if search != "" {
		query = query.Where("title LIKE ?", "%"+search+"%")
	}
	if fileType != "" {
		query = query.Where("file_type = ?", fileType)
	}
	var docs []models.Document
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&docs).Error
	return docs, err
}

func (r *DocumentRepo) Create(doc *models.Document) error {
	return r.db.Create(doc).Error
}

func (r *DocumentRepo) UpdateTitle(id uint, title string) error {
	return r.db.Model(&models.Document{}).Where("id = ?", id).Update("title", title).Error
}

func (r *DocumentRepo) SoftDelete(id uint) error {
	now := time.Now()
	return r.db.Model(&models.Document{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

func (r *DocumentRepo) UpdateParsedContent(id uint, content string, pageCount, wordCount int) error {
	return r.db.Model(&models.Document{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"parsed_content": content,
			"parsed_status":  "done",
			"page_count":     pageCount,
			"word_count":     wordCount,
		}).Error
}
