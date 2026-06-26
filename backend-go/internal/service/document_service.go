package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
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

func shouldParseContent(extOrType string) bool {
	switch strings.TrimPrefix(strings.ToLower(extOrType), ".") {
	case "md", "txt", "docx", "pdf":
		return true
	default:
		return false
	}
}

func textFileContent(data []byte) string {
	return string(bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}))
}

func parseStoredContent(fileType, path string) (string, int, int, error) {
	switch strings.TrimPrefix(strings.ToLower(fileType), ".") {
	case "md", "txt":
		data, err := os.ReadFile(path)
		if err != nil {
			return "", 0, 0, fmt.Errorf("文件解析失败: %w", err)
		}
		content := textFileContent(data)
		pageCount, wordCount := textStats(content)
		return content, pageCount, wordCount, nil
	case "pdf":
		data, err := os.ReadFile(path)
		if err != nil {
			return "", 0, 0, fmt.Errorf("PDF 读取失败: %w", err)
		}
		content := extractPDFText(data)
		pageCount := strings.Count(content, "\f") + 1
		wordCount := len(strings.Fields(content))
		if wordCount == 0 {
			return "", 0, 0, errors.New("当前 PDF 不支持文本提取")
		}
		return content, pageCount, wordCount, nil
	case "docx":
		content, err := parseDocxContent(path)
		if err != nil {
			return "", 0, 0, err
		}
		pageCount, wordCount := markdownStats(content)
		return content, pageCount, wordCount, nil
	default:
		return "", 0, 0, errors.New("当前文档不支持内容解析")
	}
}

// extractPDFText extracts readable text from a PDF by finding content
// between parentheses in PDF stream objects — a simple heuristic that works
// for most text-based PDFs.
func extractPDFText(data []byte) string {
	var buf strings.Builder
	inStream := false

	// Scan through the raw PDF content
	i := 0
	for i < len(data) {
		// Track stream objects
		if bytes.HasPrefix(data[i:], []byte("stream\n")) || bytes.HasPrefix(data[i:], []byte("stream\r")) {
			inStream = true
			i += 7
			continue
		}
		if inStream && bytes.HasPrefix(data[i:], []byte("endstream")) {
			inStream = false
			i += 9
			continue
		}
		if !inStream {
			i++
			continue
		}

		// Inside a stream — look for parenthesized text: (Hello World)
		if data[i] == '(' {
			depth := 1
			j := i + 1
			for j < len(data) && depth > 0 {
				if data[j] == '\\' {
					j += 2 // skip escaped char
					continue
				}
				if data[j] == '(' {
					depth++
				} else if data[j] == ')' {
					depth--
				}
				j++
			}
			if depth == 0 {
				// Extract the text between the parentheses, handling escapes
				text := string(data[i+1 : j-1])
				text = strings.ReplaceAll(text, "\\n", "\n")
				text = strings.ReplaceAll(text, "\\r", "\r")
				text = strings.ReplaceAll(text, "\\t", "\t")
				text = strings.ReplaceAll(text, "\\(", "(")
				text = strings.ReplaceAll(text, "\\)", ")")
				text = strings.ReplaceAll(text, "\\\\", "\\")
				if strings.TrimSpace(text) != "" {
					if buf.Len() > 0 {
						buf.WriteByte(' ')
					}
					buf.WriteString(text)
				}
				i = j
				continue
			}
		}
		i++
	}

	return buf.String()
}

func (s *DocumentService) ensureParsedContent(doc *models.Document) {
	if doc == nil || doc.ParsedStatus == "done" || !shouldParseContent(doc.FileType) || doc.StoragePath == "" {
		return
	}
	savePath := filepath.Join(s.uploadDir, doc.StoragePath)
	content, pageCount, wordCount, err := parseStoredContent(doc.FileType, savePath)
	if err != nil {
		_ = s.repo.UpdateParsedStatus(doc.ID, "failed")
		return
	}
	if err := s.repo.UpdateParsedContent(doc.ID, content, pageCount, wordCount); err == nil {
		doc.ParsedContent = content
		doc.ParsedStatus = "done"
		doc.PageCount = pageCount
		doc.WordCount = wordCount
	}
}

func textStats(content string) (int, int) {
	if content == "" {
		return 0, 0
	}
	return strings.Count(content, "\n") + 1, len(strings.Fields(content))
}

func markdownStats(content string) (int, int) {
	if content == "" {
		return 0, 0
	}
	return strings.Count(content, "\n") + 1, len([]rune(strings.TrimSpace(content)))
}

func parseDocxContent(path string) (string, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("DOCX 读取失败: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if file.Name != "word/document.xml" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("DOCX 正文读取失败: %w", err)
		}
		defer rc.Close()
		return docxXMLToMarkdown(rc)
	}
	return "", errors.New("DOCX 正文不存在")
}

func docxXMLToMarkdown(reader io.Reader) (string, error) {
	decoder := xml.NewDecoder(reader)
	var blocks []string
	var text strings.Builder
	headingLevel := 0
	inText := false
	inParagraph := false
	inStyle := false

	flushParagraph := func() {
		content := strings.TrimSpace(text.String())
		text.Reset()
		if content == "" {
			return
		}
		if headingLevel >= 1 && headingLevel <= 6 {
			blocks = append(blocks, strings.Repeat("#", headingLevel)+" "+content)
		} else {
			blocks = append(blocks, content)
		}
		headingLevel = 0
	}

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("DOCX XML 解析失败: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				inParagraph = true
				headingLevel = 0
			case "pStyle":
				if inParagraph && headingLevel == 0 {
					for _, attr := range t.Attr {
						if attr.Name.Local != "val" {
							continue
						}
						value := strings.ToLower(attr.Value)
						if strings.HasPrefix(value, "heading") {
							if level := strings.TrimPrefix(value, "heading"); len(level) > 0 {
								if level[0] >= '1' && level[0] <= '6' {
									headingLevel = int(level[0] - '0')
								}
							}
						}
					}
				}
			case "t":
				inText = true
			case "tab":
				if text.Len() > 0 {
					text.WriteByte('\t')
				}
			case "br", "cr":
				if text.Len() > 0 {
					text.WriteByte('\n')
				}
			case "instrText":
				inStyle = true
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "p":
				flushParagraph()
				inParagraph = false
			case "t":
				inText = false
			case "instrText":
				inStyle = false
			}
		case xml.CharData:
			if inText && !inStyle {
				text.Write([]byte(t))
			}
		}
	}
	flushParagraph()
	return strings.Join(blocks, "\n\n"), nil
}

func imageMimeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
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
	s.ensureParsedContent(doc)
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

	if shouldParseContent(ext) {
		content, pageCount, wordCount, err := parseStoredContent(ext, savePath)
		if err != nil {
			return nil, fmt.Errorf("文件解析失败: %w", err)
		}
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
	s.ensureParsedContent(doc)
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

func (s *DocumentService) UploadAsset(docID, userID uint, fileName string, reader io.Reader) (string, error) {
	doc, err := s.repo.GetByID(docID)
	if err != nil || doc.UserID != userID {
		return "", errors.New("文档不存在")
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	mimeType := imageMimeFromExt(ext)
	if mimeType == "" {
		return "", errors.New("仅支持上传 JPG、PNG、GIF、WEBP 图片")
	}

	assetName := uuid.New().String() + ext
	assetPath := filepath.Join(s.uploadDir, "assets", "documents", fmt.Sprintf("%d", docID), assetName)
	if err := os.MkdirAll(filepath.Dir(assetPath), 0755); err != nil {
		return "", fmt.Errorf("图片保存失败: %w", err)
	}

	out, err := os.Create(assetPath)
	if err != nil {
		return "", fmt.Errorf("图片保存失败: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, reader); err != nil {
		return "", fmt.Errorf("图片写入失败: %w", err)
	}

	return fmt.Sprintf("/api/v1/documents/%d/assets/%s", docID, assetName), nil
}

func (s *DocumentService) GetAssetData(docID uint, assetName string) ([]byte, string, error) {
	if assetName != filepath.Base(assetName) {
		return nil, "", errors.New("图片不存在")
	}

	ext := strings.ToLower(filepath.Ext(assetName))
	mimeType := imageMimeFromExt(ext)
	if mimeType == "" {
		if detected := mime.TypeByExtension(ext); strings.HasPrefix(detected, "image/") {
			mimeType = detected
		} else {
			return nil, "", errors.New("图片不存在")
		}
	}

	assetPath := filepath.Join(s.uploadDir, "assets", "documents", fmt.Sprintf("%d", docID), assetName)
	data, err := os.ReadFile(assetPath)
	if err != nil {
		return nil, "", errors.New("图片不存在")
	}
	return data, mimeType, nil
}

func (s *DocumentService) UpdateTextContent(docID, userID uint, content string) error {
	doc, err := s.repo.GetByID(docID)
	if err != nil || doc.UserID != userID {
		return errors.New("文档不存在")
	}
	if !shouldParseContent(doc.FileType) {
		return errors.New("当前文档不支持文本编辑")
	}

	data := []byte(content)
	if doc.StoragePath != "" && shouldParseAsText(doc.FileType) {
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
