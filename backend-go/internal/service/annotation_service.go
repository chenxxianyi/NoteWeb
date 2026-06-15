package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/repository"
)

type AnnotationService struct {
	repo *repository.AnnotationRepo
}

func NewAnnotationService(repo *repository.AnnotationRepo) *AnnotationService {
	return &AnnotationService{repo: repo}
}

type AnnotationResponse struct {
	ID           uint                   `json:"id"`
	DocumentID   uint                   `json:"document_id"`
	Page         int                    `json:"page"`
	SelectedText string                 `json:"selected_text"`
	Color        string                 `json:"color"`
	Type         string                 `json:"type"`
	Note         string                 `json:"note"`
	PositionData map[string]interface{} `json:"position_data"`
	CreatedAt    string                 `json:"created_at"`
}

func toAnnResponse(a *models.Annotation) AnnotationResponse {
	posData := make(map[string]interface{})
	json.Unmarshal([]byte(a.PositionData), &posData)
	return AnnotationResponse{
		ID: a.ID, DocumentID: a.DocumentID, Page: a.PageNumber,
		SelectedText: a.SelectedText, Color: a.Color, Type: a.AnnotationType,
		Note: a.Note, PositionData: posData,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
	}
}

func (s *AnnotationService) ListByDocument(docID, userID uint) ([]AnnotationResponse, error) {
	anns, err := s.repo.ListByDocument(docID)
	if err != nil {
		return nil, err
	}
	result := make([]AnnotationResponse, 0)
	for _, a := range anns {
		if a.UserID == userID {
			result = append(result, toAnnResponse(&a))
		}
	}
	return result, nil
}

func (s *AnnotationService) Create(userID uint, req AnnotationCreateRequest) (*AnnotationResponse, error) {
	posBytes, _ := json.Marshal(req.PositionData)
	ann := &models.Annotation{
		UserID: userID, DocumentID: req.DocumentID, PageNumber: req.Page,
		SelectedText: req.SelectedText, Color: req.Color,
		AnnotationType: req.AnnType, Note: req.Note,
		PositionData: string(posBytes),
	}
	if err := s.repo.Create(ann); err != nil {
		return nil, err
	}
	resp := toAnnResponse(ann)
	return &resp, nil
}

func (s *AnnotationService) Delete(annID, userID uint) error {
	ann, err := s.repo.GetByID(annID)
	if err != nil || ann.UserID != userID {
		return errors.New("批注不存在")
	}
	return s.repo.SoftDelete(annID)
}

type AnnotationCreateRequest struct {
	DocumentID   uint                   `json:"document_id"`
	Page         int                    `json:"page"`
	SelectedText string                 `json:"selected_text"`
	Color        string                 `json:"color"`
	AnnType      string                 `json:"type"`
	Note         string                 `json:"note"`
	PositionData map[string]interface{} `json:"position_data"`
}
