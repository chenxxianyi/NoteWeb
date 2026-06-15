package repository

import (
	"time"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"gorm.io/gorm"
)

type NoteRepo struct {
	db *gorm.DB
}

func NewNoteRepo(db *gorm.DB) *NoteRepo {
	return &NoteRepo{db: db}
}

func (r *NoteRepo) GetByID(id uint) (*models.Note, error) {
	var note models.Note
	err := r.db.Where("deleted_at IS NULL").First(&note, id).Error
	return &note, err
}

func (r *NoteRepo) ListByUser(userID uint, documentID, tag string) ([]models.Note, error) {
	query := r.db.Where("user_id = ? AND deleted_at IS NULL", userID)
	if documentID != "" {
		query = query.Where("document_id = ?", documentID)
	}
	if tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}
	var notes []models.Note
	err := query.Order("updated_at DESC").Find(&notes).Error
	return notes, err
}

func (r *NoteRepo) Create(note *models.Note) error {
	return r.db.Create(note).Error
}

func (r *NoteRepo) UpdateFields(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Note{}).Where("id = ?", id).Updates(updates).Error
}

func (r *NoteRepo) SoftDelete(id uint) error {
	now := time.Now()
	return r.db.Model(&models.Note{}).Where("id = ?", id).Update("deleted_at", &now).Error
}
