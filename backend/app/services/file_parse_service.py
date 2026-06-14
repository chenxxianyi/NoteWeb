"""File parsing service — extracts text content from uploaded files."""

import io


class FileParseService:
    def parse(self, content: bytes, file_type: str) -> dict:
        """Parse file content and return text + metadata."""
        if file_type == "txt":
            return self._parse_txt(content)
        elif file_type == "md":
            return self._parse_md(content)
        elif file_type == "pdf":
            return self._parse_pdf(content)
        elif file_type == "docx":
            return self._parse_docx(content)
        else:
            return {"content": "", "page_count": 0, "word_count": 0}

    def _parse_txt(self, content: bytes) -> dict:
        text = content.decode("utf-8", errors="replace")
        words = len(text.split())
        return {"content": text, "page_count": 1, "word_count": words}

    def _parse_md(self, content: bytes) -> dict:
        text = content.decode("utf-8", errors="replace")
        words = len(text.split())
        return {"content": text, "page_count": 1, "word_count": words}

    def _parse_pdf(self, content: bytes) -> dict:
        # Basic: just note that it's a PDF — full parsing requires PDF.js or PyMuPDF
        return {
            "content": "[PDF content — use frontend PDF.js for rendering]",
            "page_count": 1,
            "word_count": 0,
        }

    def _parse_docx(self, content: bytes) -> dict:
        try:
            from docx import Document as DocxDoc
            doc = DocxDoc(io.BytesIO(content))
            text = "\n".join(p.text for p in doc.paragraphs)
            words = len(text.split())
            return {"content": text, "page_count": len(doc.sections), "word_count": words}
        except ImportError:
            return {
                "content": "[DOCX parsing requires python-docx — install with: pip install python-docx]",
                "page_count": 0,
                "word_count": 0,
            }
