package service

import (
	"errors"
	"strings"
	"time"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/repository"
)

type NoteService struct {
	repo     *repository.NoteRepo
	docRepo  *repository.DocumentRepo
}

func NewNoteService(repo *repository.NoteRepo, docRepo *repository.DocumentRepo) *NoteService {
	return &NoteService{repo: repo, docRepo: docRepo}
}

type NoteResponse struct {
	ID            uint     `json:"id"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	DocumentID    *uint    `json:"document_id,omitempty"`
	DocumentTitle string   `json:"document_title,omitempty"`
	Tags          []string `json:"tags"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

func toNoteResponse(n *models.Note) NoteResponse {
	tags := strings.Split(n.Tags, ",")
	if len(tags) == 1 && tags[0] == "" {
		tags = []string{}
	}
	return NoteResponse{
		ID: n.ID, Title: n.Title, Content: n.Content,
		DocumentID: n.DocumentID,
		Tags:       tags,
		CreatedAt:  n.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  n.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *NoteService) List(userID uint, documentID, tag string) ([]NoteResponse, error) {
	notes, err := s.repo.ListByUser(userID, documentID, tag)
	if err != nil {
		return nil, err
	}
	result := make([]NoteResponse, len(notes))
	for i, n := range notes {
		resp := toNoteResponse(&n)
		if n.DocumentID != nil {
			if doc, err := s.docRepo.GetByID(*n.DocumentID); err == nil {
				resp.DocumentTitle = doc.Title
			}
		}
		result[i] = resp
	}
	return result, nil
}

func (s *NoteService) GetByID(noteID, userID uint) (*NoteResponse, error) {
	note, err := s.repo.GetByID(noteID)
	if err != nil || note.UserID != userID {
		return nil, errors.New("笔记不存在")
	}
	resp := toNoteResponse(note)
	if note.DocumentID != nil {
		if doc, err := s.docRepo.GetByID(*note.DocumentID); err == nil {
			resp.DocumentTitle = doc.Title
		}
	}
	return &resp, nil
}

func (s *NoteService) Create(userID uint, req NoteCreateRequest) (*NoteResponse, error) {
	tags := strings.Join(req.Tags, ",")
	note := &models.Note{
		UserID: userID, Title: req.Title, Content: req.Content,
		DocumentID: req.DocumentID, Tags: tags,
	}
	if err := s.repo.Create(note); err != nil {
		return nil, err
	}
	resp := toNoteResponse(note)
	if note.DocumentID != nil {
		if doc, err := s.docRepo.GetByID(*note.DocumentID); err == nil {
			resp.DocumentTitle = doc.Title
		}
	}
	return &resp, nil
}

func (s *NoteService) Update(noteID, userID uint, req NoteUpdateRequest) (*NoteResponse, error) {
	note, err := s.repo.GetByID(noteID)
	if err != nil || note.UserID != userID {
		return nil, errors.New("笔记不存在")
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Tags != nil {
		updates["tags"] = strings.Join(req.Tags, ",")
	}
	if err := s.repo.UpdateFields(noteID, updates); err != nil {
		return nil, err
	}
	return s.GetByID(noteID, userID)
}

func (s *NoteService) Delete(noteID, userID uint) error {
	note, err := s.repo.GetByID(noteID)
	if err != nil || note.UserID != userID {
		return errors.New("笔记不存在")
	}
	return s.repo.SoftDelete(noteID)
}

type NoteCreateRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	DocumentID *uint    `json:"document_id,omitempty"`
	Tags       []string `json:"tags"`
}

type NoteUpdateRequest struct {
	Title   *string  `json:"title,omitempty"`
	Content *string  `json:"content,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}
