package service

import (
	"strconv"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/repository"
)

func formatPageNumber(page int) string {
	return strconv.Itoa(page)
}

type DashboardService struct {
	docRepo   *repository.DocumentRepo
	annRepo   *repository.AnnotationRepo
	noteRepo  *repository.NoteRepo
}

func NewDashboardService(docRepo *repository.DocumentRepo, annRepo *repository.AnnotationRepo, noteRepo *repository.NoteRepo) *DashboardService {
	return &DashboardService{
		docRepo:  docRepo,
		annRepo:  annRepo,
		noteRepo: noteRepo,
	}
}

func (s *DashboardService) GetSummary(userID uint) (*models.DashboardSummary, error) {
	summary := &models.DashboardSummary{}

	// Get document count
	docs, err := s.docRepo.ListByUser(userID, "", "", 1, 1000)
	if err != nil {
		return nil, err
	}
	summary.DocumentCount = len(docs)

	// Get annotations for all user's documents
	var totalAnnotations int
	docIDs := make([]uint, 0, len(docs))
	for _, doc := range docs {
		docIDs = append(docIDs, doc.ID)
	}
	
	// Count annotations
	for _, docID := range docIDs {
		anns, err := s.annRepo.ListByDocument(docID)
		if err == nil {
			totalAnnotations += len(anns)
		}
	}
	summary.AnnotationCount = totalAnnotations

	// Get notes count
	notes, err := s.noteRepo.ListByUser(userID, "", "")
	if err == nil {
		summary.NoteCount = len(notes)
	}

	// Get continue reading (documents with progress > 0)
	continueReading := make([]models.ContinueReading, 0)
	for _, doc := range docs {
		if doc.ReadProgress > 0 && doc.ReadProgress < 100 {
			continueReading = append(continueReading, models.ContinueReading{
				ID:           doc.ID,
				Title:        doc.Title,
				FileType:     doc.FileType,
				ReadProgress: doc.ReadProgress,
				UpdatedAt:    doc.UpdatedAt.Format("2006-01-02"),
			})
		}
	}
	// Limit to 5 recent documents
	if len(continueReading) > 5 {
		continueReading = continueReading[:5]
	}
	summary.ContinueReading = continueReading

	// Build recent activities
	activities := make([]models.RecentActivity, 0)

	// Add recent annotations
	for _, docID := range docIDs {
		anns, err := s.annRepo.ListByDocument(docID)
		if err != nil {
			continue
		}
		for i, ann := range anns {
			if i >= 3 {
				break
			}
			docTitle := ""
			for _, d := range docs {
				if d.ID == docID {
					docTitle = d.Title
					break
				}
			}
			text := ann.SelectedText
			if len(text) > 40 {
				text = text[:40] + "..."
			}
			activities = append(activities, models.RecentActivity{
				Type:       "highlight",
				ID:         ann.ID,
				DocumentID: docID,
				Text:       "\"" + text + "\"",
				Document:   docTitle,
				Page:       "第" + formatPageNumber(ann.PageNumber) + "页",
				CreatedAt:  ann.CreatedAt.Format("2006-01-02 15:04"),
			})
		}
	}

	// Add recent notes
	for i, note := range notes {
		if i >= 3 {
			break
		}
		text := note.Content
		if len(text) > 40 {
			text = text[:40] + "..."
		}
		docTitle := ""
		if note.DocumentID != nil {
			for _, d := range docs {
				if d.ID == *note.DocumentID {
					docTitle = d.Title
					break
				}
			}
		}
		activities = append(activities, models.RecentActivity{
			Type:       "note",
			ID:         note.ID,
			DocumentID: 0,
			Title:      note.Title,
			Text:       text,
			Document:   docTitle,
			Page:       "—",
			CreatedAt:  note.UpdatedAt.Format("2006-01-02 15:04"),
		})
	}

	// Sort by created_at descending and limit to 6
	// For simplicity, we just limit to 6 items
	if len(activities) > 6 {
		activities = activities[:6]
	}
	summary.RecentActivities = activities

	return summary, nil
}
