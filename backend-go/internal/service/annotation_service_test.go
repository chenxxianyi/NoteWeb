package service

import (
	"errors"
	"testing"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
)

type fakeAnnotationRepo struct {
	replaceUserID     uint
	replaceDocumentID uint
	replaceDeleteIDs  []uint
	replaceCreates    []models.Annotation
	replaceResult     []models.Annotation
	replaceErr        error
}

func (f *fakeAnnotationRepo) GetByID(uint) (*models.Annotation, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeAnnotationRepo) ListByDocument(uint) ([]models.Annotation, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeAnnotationRepo) Create(*models.Annotation) error {
	return errors.New("not implemented")
}

func (f *fakeAnnotationRepo) SoftDelete(uint) error {
	return errors.New("not implemented")
}

func (f *fakeAnnotationRepo) Replace(
	userID uint,
	documentID uint,
	deleteIDs []uint,
	creates []models.Annotation,
) ([]models.Annotation, error) {
	f.replaceUserID = userID
	f.replaceDocumentID = documentID
	f.replaceDeleteIDs = append([]uint(nil), deleteIDs...)
	f.replaceCreates = append([]models.Annotation(nil), creates...)
	return f.replaceResult, f.replaceErr
}

func TestAnnotationServiceReplaceBuildsDrawingAnnotations(t *testing.T) {
	repo := &fakeAnnotationRepo{
		replaceResult: []models.Annotation{{
			ID:             91,
			UserID:         7,
			DocumentID:     12,
			PageNumber:     3,
			AnnotationType: "drawing",
			Color:          "#ff0000",
			PositionData:   `{"tool":"pen","width":3}`,
		}},
	}
	service := NewAnnotationService(repo)

	result, err := service.Replace(7, AnnotationReplaceRequest{
		DocumentID: 12,
		DeleteIDs:  []uint{31, 32},
		Creates: []AnnotationReplacementCreateRequest{{
			Page:    3,
			Color:   "#ff0000",
			AnnType: "drawing",
			PositionData: map[string]interface{}{
				"tool":  "pen",
				"width": float64(3),
			},
		}},
	})

	if err != nil {
		t.Fatalf("Replace returned error: %v", err)
	}
	if repo.replaceUserID != 7 || repo.replaceDocumentID != 12 {
		t.Fatalf("unexpected replace scope: user=%d document=%d", repo.replaceUserID, repo.replaceDocumentID)
	}
	if len(repo.replaceDeleteIDs) != 2 || repo.replaceDeleteIDs[0] != 31 || repo.replaceDeleteIDs[1] != 32 {
		t.Fatalf("unexpected delete IDs: %#v", repo.replaceDeleteIDs)
	}
	if len(repo.replaceCreates) != 1 {
		t.Fatalf("expected one create, got %d", len(repo.replaceCreates))
	}
	created := repo.replaceCreates[0]
	if created.UserID != 7 || created.DocumentID != 12 || created.AnnotationType != "drawing" {
		t.Fatalf("unexpected created annotation: %#v", created)
	}
	if created.PositionData != `{"tool":"pen","width":3}` {
		t.Fatalf("unexpected position data: %s", created.PositionData)
	}
	if len(result) != 1 || result[0].ID != 91 {
		t.Fatalf("unexpected response: %#v", result)
	}
}

func TestAnnotationServiceReplacePropagatesRepositoryError(t *testing.T) {
	repo := &fakeAnnotationRepo{replaceErr: errors.New("transaction failed")}
	service := NewAnnotationService(repo)

	_, err := service.Replace(7, AnnotationReplaceRequest{DocumentID: 12})

	if err == nil || err.Error() != "transaction failed" {
		t.Fatalf("expected repository error, got %v", err)
	}
}
