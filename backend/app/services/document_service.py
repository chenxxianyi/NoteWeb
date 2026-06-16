import os
import uuid
from datetime import datetime
from sqlalchemy.orm import Session
from fastapi import HTTPException, UploadFile

from app.repositories.document_repository import DocumentRepository
from app.repositories.user_repository import UserRepository
from app.services.storage_service import StorageService
from app.services.file_parse_service import FileParseService
from app.utils.mime import guess_mime, is_allowed, file_type_from_ext
from app.models.user import User


class DocumentService:
    def __init__(self, db: Session):
        self.repo = DocumentRepository(db)
        self.user_repo = UserRepository(db)
        self._storage: StorageService | None = None
        self._parser: FileParseService | None = None

    @property
    def storage(self) -> StorageService:
        if self._storage is None:
            self._storage = StorageService()
        return self._storage

    @property
    def parser(self) -> FileParseService:
        if self._parser is None:
            self._parser = FileParseService()
        return self._parser

    def list_documents(self, user: User, search: str = "", file_type: str = "",
                       page: int = 1, page_size: int = 50) -> list[dict]:
        skip = (page - 1) * page_size
        docs = self.repo.list_by_user(user.id, search, file_type, skip, page_size)
        return [
            {
                "id": d.id,
                "title": d.title,
                "file_type": d.file_type,
                "file_size": d.file_size or 0,
                "read_progress": 0.0,  # simplified
                "created_at": d.created_at.isoformat() if d.created_at else "",
                "updated_at": d.updated_at.isoformat() if d.updated_at else "",
            }
            for d in docs
        ]

    def get_document(self, doc_id: int, user: User) -> dict:
        doc = self.repo.get_by_id(doc_id)
        if not doc or doc.user_id != user.id:
            raise HTTPException(status_code=404, detail="Document not found")
        return {
            "id": doc.id,
            "title": doc.title,
            "file_name": doc.file_name,
            "file_type": doc.file_type,
            "mime_type": doc.mime_type,
            "file_size": doc.file_size or 0,
            "storage_path": doc.storage_path,
            "parsed_content": doc.parsed_content or "",
            "page_count": doc.page_count or 0,
            "word_count": doc.word_count or 0,
            "read_progress": 0.0,
            "created_at": doc.created_at.isoformat() if doc.created_at else "",
            "updated_at": doc.updated_at.isoformat() if doc.updated_at else "",
        }

    async def upload(self, file: UploadFile, user: User) -> dict:
        if not is_allowed(file.filename or ""):
            raise HTTPException(status_code=400, detail="File type not allowed")

        content = await file.read()
        if len(content) > 50 * 1024 * 1024:
            raise HTTPException(status_code=400, detail="File too large (max 50 MB)")

        # Upload to MinIO
        ext = os.path.splitext(file.filename or "file")[1]
        object_name = f"users/{user.id}/{uuid.uuid4().hex}{ext}"
        self.storage.upload_bytes(object_name, content, file.content_type or "")

        # Parse content
        ft = file_type_from_ext(file.filename or "")
        parse_result = self.parser.parse(content, ft)

        # Create DB record
        doc = self.repo.create(
            user_id=user.id,
            title=os.path.splitext(file.filename or "untitled")[0],
            file_name=file.filename or "untitled",
            file_type=ft,
            mime_type=guess_mime(file.filename or ""),
            file_size=len(content),
            storage_path=object_name,
        )

        # Save parsed content
        self.repo.update_parsed_content(
            doc.id,
            parse_result.get("content", ""),
            parse_result.get("page_count", 0),
            parse_result.get("word_count", 0),
        )

        # Update user storage
        self.user_repo.update_storage(user.id, len(content))

        return {
            "id": doc.id,
            "title": doc.title,
            "file_type": doc.file_type,
            "file_size": doc.file_size or 0,
            "read_progress": 0.0,
            "created_at": doc.created_at.isoformat() if doc.created_at else "",
            "updated_at": doc.updated_at.isoformat() if doc.updated_at else "",
        }

    def rename(self, doc_id: int, title: str, user: User) -> dict | None:
        doc = self.repo.get_by_id(doc_id)
        if not doc or doc.user_id != user.id:
            raise HTTPException(status_code=404, detail="Document not found")
        updated = self.repo.update_title(doc_id, title)
        if updated:
            return {
                "id": updated.id,
                "title": updated.title,
            }
        return None

    def delete(self, doc_id: int, user: User):
        doc = self.repo.get_by_id(doc_id)
        if not doc or doc.user_id != user.id:
            raise HTTPException(status_code=404, detail="Document not found")
        self.repo.soft_delete(doc_id)

    def update_read_position(self, doc_id: int, position: str, user: User):
        doc = self.repo.get_by_id(doc_id)
        if not doc or doc.user_id != user.id:
            raise HTTPException(status_code=404, detail="Document not found")
        self.repo.update_read_position(doc_id, position)
