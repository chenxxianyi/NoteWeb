package repository

import (
	"errors"
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

func (r *AnnotationRepo) Replace(
	userID uint,
	documentID uint,
	deleteIDs []uint,
	creates []models.Annotation,
) ([]models.Annotation, error) {
	uniqueDeleteIDs := make([]uint, 0, len(deleteIDs))
	seen := make(map[uint]struct{}, len(deleteIDs))
	for _, id := range deleteIDs {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		uniqueDeleteIDs = append(uniqueDeleteIDs, id)
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var documentCount int64
		if err := tx.Model(&models.Document{}).
			Where("id = ? AND user_id = ? AND deleted_at IS NULL", documentID, userID).
			Count(&documentCount).Error; err != nil {
			return err
		}
		if documentCount != 1 {
			return errors.New("annotation replacement document is unavailable")
		}

		if len(uniqueDeleteIDs) > 0 {
			var count int64
			if err := tx.Model(&models.Annotation{}).
				Where(
					"id IN ? AND user_id = ? AND document_id = ? AND deleted_at IS NULL",
					uniqueDeleteIDs,
					userID,
					documentID,
				).
				Count(&count).Error; err != nil {
				return err
			}
			if count != int64(len(uniqueDeleteIDs)) {
				return errors.New("annotation replacement contains unavailable annotations")
			}
		}

		for index := range creates {
			creates[index].ID = 0
			creates[index].UserID = userID
			creates[index].DocumentID = documentID
			creates[index].DeletedAt = nil
		}
		if len(creates) > 0 {
			if err := tx.Create(&creates).Error; err != nil {
				return err
			}
		}

		if len(uniqueDeleteIDs) > 0 {
			now := time.Now()
			result := tx.Model(&models.Annotation{}).
				Where(
					"id IN ? AND user_id = ? AND document_id = ? AND deleted_at IS NULL",
					uniqueDeleteIDs,
					userID,
					documentID,
				).
				Update("deleted_at", &now)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != int64(len(uniqueDeleteIDs)) {
				return errors.New("annotation replacement changed concurrently")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return creates, nil
}
