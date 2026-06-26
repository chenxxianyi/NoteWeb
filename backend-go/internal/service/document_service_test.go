package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Pure function tests for document_service.go

func TestTextFileContent(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"plain text", []byte("hello world"), "hello world"},
		{"with BOM", []byte{0xEF, 0xBB, 0xBF, 'h', 'e', 'l', 'l', 'o'}, "hello"},
		{"empty", []byte{}, ""},
		{"only BOM", []byte{0xEF, 0xBB, 0xBF}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := textFileContent(tt.data)
			if got != tt.want {
				t.Fatalf("textFileContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTextStats(t *testing.T) {
	lines, words := textStats("hello world")
	if lines != 1 {
		t.Fatalf("expected 1 line, got %d", lines)
	}
	if words != 2 {
		t.Fatalf("expected 2 words, got %d", words)
	}

	lines, words = textStats("")
	if lines != 0 || words != 0 {
		t.Fatalf("expected 0/0 for empty, got %d/%d", lines, words)
	}

	lines, words = textStats("line1\nline2\nline3")
	if lines != 3 {
		t.Fatalf("expected 3 lines, got %d", lines)
	}
	if words != 3 {
		t.Fatalf("expected 3 words, got %d", words)
	}
}

func TestMarkdownStats(t *testing.T) {
	lines, words := markdownStats("# Hello\n\nWorld")
	if lines != 3 {
		t.Fatalf("expected 3 lines (# Hello\\n\\nWorld = 2 newlines + 1), got %d", lines)
	}
	if words != 14 {
		t.Fatalf("expected 14 chars for '# Hello\\n\\nWorld', got %d", words)
	}

	lines, words = markdownStats("")
	if lines != 0 || words != 0 {
		t.Fatalf("expected 0/0 for empty, got %d/%d", lines, words)
	}
}

func TestShouldParseContent(t *testing.T) {
	cases := []struct {
		ext  string
		want bool
	}{
		{"md", true},
		{"txt", true},
		{"docx", true},
		{"pdf", true},
		{".md", true},
		{".PDF", true},
		{"jpg", false},
		{"png", false},
		{"", false},
	}
	for _, c := range cases {
		t.Run(c.ext, func(t *testing.T) {
			got := shouldParseContent(c.ext)
			if got != c.want {
				t.Fatalf("shouldParseContent(%q) = %v, want %v", c.ext, got, c.want)
			}
		})
	}
}

func TestShouldParseAsText(t *testing.T) {
	cases := []struct {
		ext  string
		want bool
	}{
		{"md", true},
		{"txt", true},
		{"docx", false},
		{"pdf", false},
		{"", false},
	}
	for _, c := range cases {
		t.Run(c.ext, func(t *testing.T) {
			got := shouldParseAsText(c.ext)
			if got != c.want {
				t.Fatalf("shouldParseAsText(%q) = %v, want %v", c.ext, got, c.want)
			}
		})
	}
}

func TestImageMimeFromExt(t *testing.T) {
	cases := []struct {
		ext  string
		want string
	}{
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".png", "image/png"},
		{".gif", "image/gif"},
		{".webp", "image/webp"},
		{".pdf", ""},
		{"", ""},
	}
	for _, c := range cases {
		t.Run(c.ext, func(t *testing.T) {
			got := imageMimeFromExt(c.ext)
			if got != c.want {
				t.Fatalf("imageMimeFromExt(%q) = %q, want %q", c.ext, got, c.want)
			}
		})
	}
}

func TestParseStoredContent_TxtFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("Hello World\nLine 2"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	content, pages, words, err := parseStoredContent("txt", path)
	if err != nil {
		t.Fatalf("parseStoredContent() error = %v", err)
	}
	if content != "Hello World\nLine 2" {
		t.Fatalf("unexpected content: %q", content)
	}
	if pages < 1 {
		t.Fatalf("expected >0 pages, got %d", pages)
	}
	if words < 3 {
		t.Fatalf("expected >=3 words, got %d", words)
	}
}

func TestParseStoredContent_TxtFileBOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bom.txt")
	bom := []byte{0xEF, 0xBB, 0xBF}
	if err := os.WriteFile(path, append(bom, []byte("BOM content")...), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	content, _, _, err := parseStoredContent("txt", path)
	if err != nil {
		t.Fatalf("parseStoredContent() error = %v", err)
	}
	if content != "BOM content" {
		t.Fatalf("expected 'BOM content' without BOM, got %q", content)
	}
}

func TestParseStoredContent_MdFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte("# Title\n\nParagraph."), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	content, pages, words, err := parseStoredContent("md", path)
	if err != nil {
		t.Fatalf("parseStoredContent() error = %v", err)
	}
	if !strings.Contains(content, "# Title") {
		t.Fatalf("expected '# Title' in content, got %q", content)
	}
	if pages < 1 || words < 1 {
		t.Fatalf("expected positive stats, got pages=%d words=%d", pages, words)
	}
}

func TestParseStoredContent_Unsupported(t *testing.T) {
	_, _, _, err := parseStoredContent("jpg", "/nonexistent.jpg")
	if err == nil {
		t.Fatal("expected error for unsupported file type")
	}
}

func TestParseStoredContent_FileNotFound(t *testing.T) {
	_, _, _, err := parseStoredContent("txt", "/nonexistent/path.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestExtractPDFText_Simple(t *testing.T) {
	// A minimal PDF-like content with parentheses text
	pdfData := []byte("%PDF-1.4\nstream\nBT /F1 24 Tf 100 700 Td(Hello PDF World)Tj ET\nendstream\nxref\n0 1\ntrailer<</Size 1>>\nstartxref\n0\n%%EOF")

	text := extractPDFText(pdfData)
	if !strings.Contains(text, "Hello PDF World") {
		t.Fatalf("expected extracted text to contain 'Hello PDF World', got %q", text)
	}
}

func TestExtractPDFText_MultipleStrings(t *testing.T) {
	pdfData := []byte("stream\nBT(Foo Bar)Tj ET\nBT(Baz Qux)Tj ET\nendstream")
	text := extractPDFText(pdfData)
	if !strings.Contains(text, "Foo Bar") || !strings.Contains(text, "Baz Qux") {
		t.Fatalf("expected multiple strings, got %q", text)
	}
}

func TestExtractPDFText_Empty(t *testing.T) {
	text := extractPDFText([]byte("no parenthesized content"))
	if text != "" {
		t.Fatalf("expected empty for no parenthesized content, got %q", text)
	}
}

func TestExtractPDFText_NonStream(t *testing.T) {
	// Text outside stream should not be extracted
	pdfData := []byte("obj<<>>endobj\nstream\n(Stream Text)Tj\nendstream\n(Outside Text)")
	text := extractPDFText(pdfData)
	if strings.Contains(text, "Outside") {
		t.Fatalf("should not extract text outside stream, got %q", text)
	}
	if !strings.Contains(text, "Stream Text") {
		t.Fatalf("should extract text inside stream, got %q", text)
	}
}

func TestExtractPDFText_EscapedParentheses(t *testing.T) {
	pdfData := []byte("stream\n(Escaped \\\\(parentheses\\\\) here)Tj\nendstream")
	text := extractPDFText(pdfData)
	if !strings.Contains(text, "Escaped (parentheses) here") {
		t.Fatalf("expected properly unescaped parentheses, got %q", text)
	}
}
