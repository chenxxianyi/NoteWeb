package repository

import (
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"gorm.io/gorm"
)

type UserSettingsRepo struct {
	db *gorm.DB
}

func NewUserSettingsRepo(db *gorm.DB) *UserSettingsRepo {
	return &UserSettingsRepo{db: db}
}

func (r *UserSettingsRepo) GetByUserID(userID uint) (*models.UserSettings, error) {
	var settings models.UserSettings
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default settings if not exists
			settings = models.UserSettings{
				UserID:      userID,
				Theme:       "warm",
				Font:        "Noto Serif SC",
				ReadingMode: true,
			}
			if createErr := r.db.Create(&settings).Error; createErr != nil {
				return nil, createErr
			}
			return &settings, nil
		}
		return nil, err
	}
	return &settings, nil
}

func (r *UserSettingsRepo) Update(userID uint, theme string, font string, readingMode bool) error {
	var settings models.UserSettings
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			settings = models.UserSettings{
				UserID:      userID,
				Theme:       theme,
				Font:        font,
				ReadingMode: readingMode,
			}
			return r.db.Create(&settings).Error
		}
		return err
	}
	return r.db.Model(&settings).Updates(map[string]interface{}{
		"theme":        theme,
		"font":         font,
		"reading_mode": readingMode,
	}).Error
}