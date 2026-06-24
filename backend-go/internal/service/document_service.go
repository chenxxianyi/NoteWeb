package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/repository"
	"github.com/google/uuid"
)

type DocumentService struct {
	repo      *repository.DocumentRepo
	userRepo  *repository.UserRepo
	uploadDir string
}

func NewDocumentService(repo *repository.DocumentRepo, userRepo *repository.UserRepo, uploadDir string) *DocumentService {
	return &DocumentService{repo: repo, userRepo: userRepo, uploadDir: uploadDir}
}

type DocumentResponse struct {
	ID           uint    `json:"id"`
	Title        string  `json:"title"`
	FileType     string  `json:"file_type"`
	FileSize     int64   `json:"file_size"`
	ReadProgress float64 `json:"read_progress"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type DocumentDetailResponse struct {
	ID            uint    `json:"id"`
	Title         string  `json:"title"`
	FileName      string  `json:"file_name"`
	FileType      string  `json:"file_type"`
	MimeType      string  `json:"mime_type"`
	FileSize      int64   `json:"file_size"`
	ParsedContent string  `json:"parsed_content"`
	PageCount     int     `json:"page_count"`
	WordCount     int     `json:"word_count"`
	ReadProgress  float64 `json:"read_progress"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func toDocResponse(d *models.Document) DocumentResponse {
	return DocumentResponse{
		ID: d.ID, Title: d.Title, FileType: d.FileType,
		FileSize: d.FileSize, ReadProgress: d.ReadProgress,
		CreatedAt: d.CreatedAt.Format(time.RFC3339),
		UpdatedAt: d.UpdatedAt.Format(time.RFC3339),
	}
}

func shouldParseAsText(extOrType string) bool {
	switch strings.TrimPrefix(strings.ToLower(extOrType), ".") {
	case "md", "txt":
		return true
	default:
		return false
	}
}

func textFileContent(data []byte) string {
	return string(bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}))
}

func textStats(content string) (int, int) {
	if content == "" {
		return 0, 0
	}
	return strings.Count(content, "\n") + 1, len(strings.Fields(content))
}

func (s *DocumentService) List(userID uint, search, fileType string, page, pageSize int) ([]DocumentResponse, error) {
	docs, err := s.repo.ListByUser(userID, search, fileType, page, pageSize)
	if err != nil {
		return nil, err
	}
	result := make([]DocumentResponse, len(docs))
	for i, d := range docs {
		result[i] = toDocResponse(&d)
	}
	return result, nil
}

func (s *DocumentService) GetDetail(docID, userID uint) (*DocumentDetailResponse, error) {
	doc, err := s.repo.GetByID(docID)
	if err != nil {
		return nil, errors.New("文档不存在")
	}
	if doc.UserID != userID {
		return nil, errors.New("文档不存在")
	}
	if doc.ParsedContent == "" && shouldParseAsText(doc.FileType) && doc.StoragePath != "" {
		savePath := filepath.Join(s.uploadDir, doc.StoragePath)
		if data, err := os.ReadFile(savePath); err == nil {
			content := textFileContent(data)
			pageCount, wordCount := textStats(content)
			if err := s.repo.UpdateParsedContent(doc.ID, content, pageCount, wordCount); err == nil {
				doc.ParsedContent = content
				doc.ParsedStatus = "done"
				doc.PageCount = pageCount
				doc.WordCount = wordCount
			}
		}
	}
	return &DocumentDetailResponse{
		ID: doc.ID, Title: doc.Title, FileName: doc.FileName,
		FileType: doc.FileType, MimeType: doc.MimeType, FileSize: doc.FileSize,
		ParsedContent: doc.ParsedContent, PageCount: doc.PageCount, WordCount: doc.WordCount,
		ReadProgress: doc.ReadProgress,
		CreatedAt:    doc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    doc.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *DocumentService) Upload(userID uint, fileName string, fileSize int64, reader io.Reader) (*DocumentResponse, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	allowed := map[string]bool{".pdf": true, ".md": true, ".txt": true, ".docx": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		return nil, errors.New("不支持的文件类型")
	}

	objectName := fmt.Sprintf("users/%d/%s%s", userID, uuid.New().String(), ext)
	savePath := filepath.Join(s.uploadDir, objectName)
	os.MkdirAll(filepath.Dir(savePath), 0755)

	f, err := os.Create(savePath)
	if err != nil {
		return nil, fmt.Errorf("文件保存失败: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return nil, fmt.Errorf("文件写入失败: %w", err)
	}

	title := strings.TrimSuffix(fileName, ext)
	fileType := strings.TrimPrefix(ext, ".")
	mimeMap := map[string]string{
		".pdf": "application/pdf", ".md": "text/markdown", ".txt": "text/plain",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
	mimeType := mimeMap[ext]
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	doc := &models.Document{
		UserID: userID, Title: title, FileName: fileName,
		FileType: fileType, MimeType: mimeType, FileSize: fileSize,
		StoragePath: objectName,
	}
	if err := s.repo.Create(doc); err != nil {
		return nil, err
	}

	if shouldParseAsText(ext) {
		data, err := os.ReadFile(savePath)
		if err != nil {
			return nil, fmt.Errorf("鏂囦欢瑙ｆ瀽澶辫触: %w", err)
		}
		content := textFileContent(data)
		pageCount, wordCount := textStats(content)
		if err := s.repo.UpdateParsedContent(doc.ID, content, pageCount, wordCount); err != nil {
			return nil, err
		}
		doc.ParsedContent = content
		doc.ParsedStatus = "done"
		doc.PageCount = pageCount
		doc.WordCount = wordCount
	}

	// Update user storage
	s.userRepo.UpdateStorage(userID, fileSize)

	resp := toDocResponse(doc)
	return &resp, nil
}

func (s *DocumentService) Rename(docID, userID uint, title string) error {
	doc, err := s.repo.GetByID(docID)
	if err != nil || doc.UserID != userID {
		return errors.New("文档不存在")
	}
	return s.repo.UpdateTitle(docID, title)
}

func (s *DocumentService) Delete(docID, userID uint) error {
	doc, err := s.repo.GetByID(docID)
	if err != nil || doc.UserID != userID {
		return errors.New("文档不存在")
	}
	return s.repo.SoftDelete(docID)
}

// GetContent returns the parsed text content of a document.
func (s *DocumentService) GetContent(docID, userID uint) (*DocumentDetailResponse, error) {
	doc, err := s.repo.GetByID(docID)
	if err != nil || doc.UserID != userID {
		return nil, errors.New("文档不存在")
	}
	return &DocumentDetailResponse{
		ID: doc.ID, Title: doc.Title, FileName: doc.FileName,
		FileType: doc.FileType, MimeType: doc.MimeType, FileSize: doc.FileSize,
		ParsedContent: doc.ParsedContent, PageCount: doc.PageCount, WordCount: doc.WordCount,
		ReadProgress: doc.ReadProgress,
		CreatedAt:    doc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    doc.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetFileData returns the original file bytes and its MIME type.
func (s *DocumentService) GetFileData(docID, userID uint) ([]byte, string, error) {
	doc, err := s.repo.GetByID(docID)
	if err != nil || doc.UserID != userID {
		return nil, "", errors.New("文档不存在")
	}
	savePath := filepath.Join(s.uploadDir, doc.StoragePath)
	data, err := os.ReadFile(savePath)
	if err != nil {
		return nil, "", fmt.Errorf("文件读取失败: %w", err)
	}
	return data, doc.MimeType, nil
}

func (s *DocumentService) UpdateTextContent(docID, userID uint, content string) error {
	doc, err := s.repo.GetByID(docID)
	if err != nil || doc.UserID != userID {
		return errors.New("文档不存在")
	}
	if !shouldParseAsText(doc.FileType) {
		return errors.New("当前文档不支持文本编辑")
	}

	data := []byte(content)
	if doc.StoragePath != "" {
		savePath := filepath.Join(s.uploadDir, doc.StoragePath)
		if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
			return fmt.Errorf("文件保存失败: %w", err)
		}
		if err := os.WriteFile(savePath, data, 0644); err != nil {
			return fmt.Errorf("文件保存失败: %w", err)
		}
	}

	pageCount, wordCount := textStats(content)
	return s.repo.UpdateTextContent(docID, content, pageCount, wordCount, int64(len(data)))
}

// UpdateReadProgress sets the document's read progress (0–100).
func (s *DocumentService) UpdateReadProgress(docID, userID uint, progress float64) error {
	doc, err := s.repo.GetByID(docID)
	if err != nil || doc.UserID != userID {
		return errors.New("文档不存在")
	}
	return s.repo.UpdateReadProgress(docID, progress)
}
