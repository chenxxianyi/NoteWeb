package service

import (
	"strings"
	"testing"
)

func TestDocxXMLToMarkdown(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p>
      <w:pPr><w:pStyle w:val="Heading1"/></w:pPr>
      <w:r><w:t>网络安全事件处置情况</w:t></w:r>
    </w:p>
    <w:p>
      <w:r><w:t>第一段内容</w:t></w:r>
      <w:r><w:tab/></w:r>
      <w:r><w:t>后续文本</w:t></w:r>
    </w:p>
  </w:body>
</w:document>`

	got, err := docxXMLToMarkdown(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("docxXMLToMarkdown() error = %v", err)
	}

	want := "# 网络安全事件处置情况\n\n第一段内容\t后续文本"
	if got != want {
		t.Fatalf("docxXMLToMarkdown() = %q, want %q", got, want)
	}
}
