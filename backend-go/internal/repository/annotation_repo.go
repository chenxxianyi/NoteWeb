package repository

import (
	"time"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"gorm.io/gorm"
)

type AnnotationRepo struct {
	db *gorm.DB
}

func NewAnnotationRepo(db *gorm.DB) *AnnotationRepo {
	return &AnnotationRepo{db: db}
}

func (r *AnnotationRepo) GetByID(id uint) (*models.Annotation, error) {
	var ann models.Annotation
	err := r.db.Where("deleted_at IS NULL").First(&ann, id).Error
	return &ann, err
}

func (r *AnnotationRepo) ListByDocument(docID uint) ([]models.Annotation, error) {
	var anns []models.Annotation
	err := r.db.Where("document_id = ? AND deleted_at IS NULL", docID).
		Order("created_at ASC").Find(&anns).Error
	return anns, err
}

func (r *AnnotationRepo) Create(ann *models.Annotation) error {
	return r.db.Create(ann).Error
}

func (r *AnnotationRepo) SoftDelete(id uint) error {
	now := time.Now()
	return r.db.Model(&models.Annotation{}).Where("id = ?", id).Update("deleted_at", &now).Error
}
