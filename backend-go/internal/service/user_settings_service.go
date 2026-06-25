package service

import (
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/repository"
)

type UserSettingsService struct {
	repo *repository.UserSettingsRepo
}

func NewUserSettingsService(repo *repository.UserSettingsRepo) *UserSettingsService {
	return &UserSettingsService{repo: repo}
}

func (s *UserSettingsService) GetSettings(userID uint) (*models.UserSettings, error) {
	return s.repo.GetByUserID(userID)
}

func (s *UserSettingsService) UpdateSettings(userID uint, theme string, font string, readingMode bool) error {
	return s.repo.Update(userID, theme, font, readingMode)
}