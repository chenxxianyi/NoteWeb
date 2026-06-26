package handlers

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/service"
)

// ---- fake annotation service ----

type fakeAnnSvc struct {
	listResp   []service.AnnotationResponse
	listErr    error
	createResp *service.AnnotationResponse
	createErr  error
	deleteErr  error
	replaceResp []service.AnnotationResponse
	replaceErr  error
}

func (f *fakeAnnSvc) ListByDocument(docID, userID uint) ([]service.AnnotationResponse, error) {
	return f.listResp, f.listErr
}
func (f *fakeAnnSvc) Create(userID uint, req service.AnnotationCreateRequest) (*service.AnnotationResponse, error) {
	return f.createResp, f.createErr
}
func (f *fakeAnnSvc) Delete(annID, userID uint) error {
	return f.deleteErr
}
func (f *fakeAnnSvc) Replace(userID uint, req service.AnnotationReplaceRequest) ([]service.AnnotationResponse, error) {
	return f.replaceResp, f.replaceErr
}

// ---- fake note service ----

type fakeNoteSvc struct {
	listResp  []service.NoteResponse
	listErr   error
	getResp   *service.NoteResponse
	getErr    error
	createResp *service.NoteResponse
	createErr error
	updateResp *service.NoteResponse
	updateErr error
	deleteErr error
}

func (f *fakeNoteSvc) List(userID uint, documentID, tag string) ([]service.NoteResponse, error) {
	return f.listResp, f.listErr
}
func (f *fakeNoteSvc) GetByID(noteID, userID uint) (*service.NoteResponse, error) {
	return f.getResp, f.getErr
}
func (f *fakeNoteSvc) Create(userID uint, req service.NoteCreateRequest) (*service.NoteResponse, error) {
	return f.createResp, f.createErr
}
func (f *fakeNoteSvc) Update(noteID, userID uint, req service.NoteUpdateRequest) (*service.NoteResponse, error) {
	return f.updateResp, f.updateErr
}
func (f *fakeNoteSvc) Delete(noteID, userID uint) error {
	return f.deleteErr
}

// ---- fake dashboard service ----

type fakeDashSvc struct {
	summary *models.DashboardSummary
	sumErr  error
}

func (f *fakeDashSvc) GetSummary(userID uint) (*models.DashboardSummary, error) {
	return f.summary, f.sumErr
}

// ---- fake settings service ----

type fakeSettingsSvc struct {
	settings *models.UserSettings
	getErr   error
	updateErr error
}

func (f *fakeSettingsSvc) GetSettings(userID uint) (*models.UserSettings, error) {
	return f.settings, f.getErr
}
func (f *fakeSettingsSvc) UpdateSettings(userID uint, theme, font string, readingMode bool) error {
	return f.updateErr
}

// ---- router helpers ----

func authMW(c *gin.Context) {
	c.Set("userID", uint(1))
	c.Next()
}

func TestNotesList_Success(t *testing.T) {
	svc := &fakeNoteSvc{listResp: []service.NoteResponse{{ID: 1, Title: "mynote"}}}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewNoteHandler(svc)
	r.Group("/api/v1/notes").Use(authMW).GET("", h.List)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/notes", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []service.NoteResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 1 || resp[0].Title != "mynote" {
		t.Fatalf("unexpected: %#v", resp)
	}
}

func TestNotesCreate_Success(t *testing.T) {
	svc := &fakeNoteSvc{createResp: &service.NoteResponse{ID: 2, Title: "new note"}}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewNoteHandler(svc)
	r.Group("/api/v1/notes").Use(authMW).POST("", h.Create)

	w := httptest.NewRecorder()
	body := `{"title":"new note","content":"hello"}`
	req := httptest.NewRequest("POST", "/api/v1/notes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestNotesGet_NotFound(t *testing.T) {
	svc := &fakeNoteSvc{getErr: errors.New("笔记不存在")}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewNoteHandler(svc)
	r.Group("/api/v1/notes").Use(authMW).GET("/:id", h.Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/notes/999", nil)
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestNotesDelete_Success(t *testing.T) {
	svc := &fakeNoteSvc{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewNoteHandler(svc)
	r.Group("/api/v1/notes").Use(authMW).DELETE("/:id", h.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/v1/notes/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAnnotationsList_Success(t *testing.T) {
	svc := &fakeAnnSvc{listResp: []service.AnnotationResponse{{ID: 1, Type: "highlight"}}}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAnnotationHandler(svc)
	r.Group("/api/v1/documents/:id/annotations").Use(authMW).GET("", h.ListByDocument)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/documents/1/annotations", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAnnotationsCreate_Success(t *testing.T) {
	svc := &fakeAnnSvc{createResp: &service.AnnotationResponse{ID: 1, Type: "highlight"}}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAnnotationHandler(svc)
	r.Group("/api/v1/annotations").Use(authMW).POST("", h.Create)

	w := httptest.NewRecorder()
	body := `{"document_id":1,"page":1,"type":"highlight","color":"#ff0"}`
	req := httptest.NewRequest("POST", "/api/v1/annotations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDashboardSummary_Success(t *testing.T) {
	svc := &fakeDashSvc{summary: &models.DashboardSummary{DocumentCount: 5}}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDashboardHandler(svc)
	r.Group("/api/v1/dashboard").Use(authMW).GET("/summary", h.GetSummary)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/dashboard/summary", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp models.DashboardSummary
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.DocumentCount != 5 {
		t.Fatalf("expected DocumentCount=5, got %d", resp.DocumentCount)
	}
}

func TestSettingsGet_Success(t *testing.T) {
	svc := &fakeSettingsSvc{settings: &models.UserSettings{Theme: "dark"}}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewUserSettingsHandler(svc)
	r.Group("/api/v1/settings").Use(authMW).GET("", h.Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/settings", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSettingsUpdate_Success(t *testing.T) {
	svc := &fakeSettingsSvc{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewUserSettingsHandler(svc)
	r.Group("/api/v1/settings").Use(authMW).PATCH("", h.Update)

	w := httptest.NewRecorder()
	body := `{"theme":"blue","font":"Inter","reading_mode":true}`
	req := httptest.NewRequest("PATCH", "/api/v1/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
