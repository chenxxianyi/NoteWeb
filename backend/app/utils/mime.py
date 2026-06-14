"""MIME type helpers."""

MIME_MAP = {
    ".pdf": "application/pdf",
    ".md": "text/markdown",
    ".txt": "text/plain",
    ".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    ".jpg": "image/jpeg",
    ".jpeg": "image/jpeg",
    ".png": "image/png",
    ".gif": "image/gif",
    ".webp": "image/webp",
}

ALLOWED_EXTENSIONS = {".pdf", ".md", ".txt", ".docx", ".jpg", ".jpeg", ".png", ".gif", ".webp"}


def guess_mime(filename: str) -> str:
    import os
    ext = os.path.splitext(filename)[1].lower()
    return MIME_MAP.get(ext, "application/octet-stream")


def is_allowed(filename: str) -> bool:
    import os
    ext = os.path.splitext(filename)[1].lower()
    return ext in ALLOWED_EXTENSIONS


def file_type_from_ext(filename: str) -> str:
    import os
    ext = os.path.splitext(filename)[1].lower()
    mapping = {
        ".pdf": "pdf",
        ".md": "md",
        ".txt": "txt",
        ".docx": "docx",
    }
    return mapping.get(ext, "other")
