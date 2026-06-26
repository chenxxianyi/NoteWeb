package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/service"
	"github.com/gin-gonic/gin"
)

type fakeDocSvc struct {
	listResp       []service.DocumentResponse
	listErr        error
	detailResp     *service.DocumentDetailResponse
	detailErr      error
	uploadResp     *service.DocumentResponse
	uploadErr      error
	renameErr      error
	deleteErr      error
	contentResp    *service.DocumentDetailResponse
	contentErr     error
	fileData       []byte
	fileMIME       string
	fileErr        error
	progressErr    error
	textContentErr error
	assetURL       string
	assetErr       error
	assetData      []byte
	assetMIME      string
	assetDataErr   error
}

func (f *fakeDocSvc) List(userID uint, search, fileType string, page, pageSize int) ([]service.DocumentResponse, error) {
	return f.listResp, f.listErr
}
func (f *fakeDocSvc) GetDetail(docID, userID uint) (*service.DocumentDetailResponse, error) {
	return f.detailResp, f.detailErr
}
func (f *fakeDocSvc) Upload(userID uint, fileName string, fileSize int64, reader io.Reader) (*service.DocumentResponse, error) {
	return f.uploadResp, f.uploadErr
}
func (f *fakeDocSvc) Rename(docID, userID uint, title string) error {
	return f.renameErr
}
func (f *fakeDocSvc) Delete(docID, userID uint) error {
	return f.deleteErr
}
func (f *fakeDocSvc) GetContent(docID, userID uint) (*service.DocumentDetailResponse, error) {
	return f.contentResp, f.contentErr
}
func (f *fakeDocSvc) GetFileData(docID, userID uint) ([]byte, string, error) {
	return f.fileData, f.fileMIME, f.fileErr
}
func (f *fakeDocSvc) UpdateReadProgress(docID, userID uint, progress float64) error {
	return f.progressErr
}
func (f *fakeDocSvc) UpdateTextContent(docID, userID uint, content string) error {
	return f.textContentErr
}
func (f *fakeDocSvc) UploadAsset(docID, userID uint, fileName string, reader io.Reader) (string, error) {
	return f.assetURL, f.assetErr
}
func (f *fakeDocSvc) GetAssetData(docID uint, assetName string) ([]byte, string, error) {
	return f.assetData, f.assetMIME, f.assetDataErr
}

func setupDocRouter(svc DocumentHandlerService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDocumentHandler(svc)

	docs := r.Group("/api/v1/documents")
	docs.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Next()
	})
	docs.GET("", h.List)
	docs.GET("/:id", h.Get)
	docs.GET("/:id/content", h.GetContent)
	docs.GET("/:id/file", h.GetFile)
	docs.POST("/upload", h.Upload)
	docs.PATCH("/:id", h.Rename)
	docs.PATCH("/:id/read-progress", h.UpdateReadProgress)
	docs.DELETE("/:id", h.Delete)

	return r
}

func TestDocumentList_Success(t *testing.T) {
	svc := &fakeDocSvc{
		listResp: []service.DocumentResponse{{
			ID: 1, Title: "doc1", FileType: "txt", FileSize: 100,
		}},
	}
	router := setupDocRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/documents", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	var resp []service.DocumentResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(resp) != 1 || resp[0].Title != "doc1" {
		t.Fatalf("unexpected list: %#v", resp)
	}
}

func TestDocumentList_Empty(t *testing.T) {
	svc := &fakeDocSvc{listResp: []service.DocumentResponse{}}
	router := setupDocRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/documents", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "[]\n" && w.Body.String() != "[]" {
		t.Fatalf("expected empty array, got %s", w.Body.String())
	}
}

func TestDocumentGet_Success(t *testing.T) {
	svc := &fakeDocSvc{
		detailResp: &service.DocumentDetailResponse{
			ID: 42, Title: "mydoc", FileType: "md", ParsedContent: "# Hello",
		},
	}
	router := setupDocRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/documents/42", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp service.DocumentDetailResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.ID != 42 || resp.Title != "mydoc" {
		t.Fatalf("unexpected detail: %#v", resp)
	}
}

func TestDocumentGet_NotFound(t *testing.T) {
	svc := &fakeDocSvc{detailErr: errors.New("文档不存在")}
	router := setupDocRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/documents/999", nil)
	router.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDocumentContent_Success(t *testing.T) {
	svc := &fakeDocSvc{
		contentResp: &service.DocumentDetailResponse{
			ID: 1, Title: "doc", ParsedContent: "hello world",
		},
	}
	router := setupDocRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/documents/1/content", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct{ Content string }
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Content != "hello world" {
		t.Fatalf("expected 'hello world', got %q", resp.Content)
	}
}

func TestDocumentFile_Success(t *testing.T) {
	svc := &fakeDocSvc{
		fileData: []byte("PDF content"),
		fileMIME: "application/pdf",
	}
	router := setupDocRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/documents/1/file", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/pdf" {
		t.Fatalf("expected PDF content type, got %s", w.Header().Get("Content-Type"))
	}
	if w.Body.String() != "PDF content" {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestDocumentFile_NotFound(t *testing.T) {
	svc := &fakeDocSvc{fileErr: errors.New("文件读取失败")}
	router := setupDocRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/documents/1/file", nil)
	router.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDocumentUpload(t *testing.T) {
	svc := &fakeDocSvc{
		uploadResp: &service.DocumentResponse{
			ID: 10, Title: "uploaded", FileType: "txt", FileSize: 12,
		},
	}
	router := setupDocRouter(svc)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("hello world"))
	writer.Close()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/documents/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestDocumentRename_Success(t *testing.T) {
	svc := &fakeDocSvc{}
	router := setupDocRouter(svc)

	w := httptest.NewRecorder()
	body := `{"title":"newname"}`
	req := httptest.NewRequest("PATCH", "/api/v1/documents/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDocumentDelete_Success(t *testing.T) {
	svc := &fakeDocSvc{}
	router := setupDocRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/v1/documents/1", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDocumentReadProgress_Success(t *testing.T) {
	svc := &fakeDocSvc{}
	router := setupDocRouter(svc)

	w := httptest.NewRecorder()
	body := `{"progress":50}`
	req := httptest.NewRequest("PATCH", "/api/v1/documents/1/read-progress", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDocumentReadProgress_BadRequest(t *testing.T) {
	svc := &fakeDocSvc{}
	router := setupDocRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/api/v1/documents/1/read-progress", strings.NewReader(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
