from sqlalchemy.orm import Session
from sqlalchemy import desc
from app.models.document import Document


class DocumentRepository:
    def __init__(self, db: Session):
        self.db = db

    def get_by_id(self, document_id: int) -> Document | None:
        return (
            self.db.query(Document)
            .filter(Document.id == document_id, Document.deleted_at.is_(None))
            .first()
        )

    def list_by_user(
        self, user_id: int, search: str = "", file_type: str = "", skip: int = 0, limit: int = 50
    ) -> list[Document]:
        query = self.db.query(Document).filter(
            Document.user_id == user_id, Document.deleted_at.is_(None)
        )
        if search:
            query = query.filter(Document.title.ilike(f"%{search}%"))
        if file_type:
            query = query.filter(Document.file_type == file_type)
        return query.order_by(desc(Document.created_at)).offset(skip).limit(limit).all()

    def create(self, user_id: int, title: str, file_name: str, file_type: str,
               mime_type: str, file_size: int, storage_path: str) -> Document:
        doc = Document(
            user_id=user_id,
            title=title,
            file_name=file_name,
            file_type=file_type,
            mime_type=mime_type,
            file_size=file_size,
            storage_path=storage_path,
        )
        self.db.add(doc)
        self.db.commit()
        self.db.refresh(doc)
        return doc

    def update_title(self, document_id: int, title: str) -> Document | None:
        doc = self.get_by_id(document_id)
        if doc:
            doc.title = title
            self.db.commit()
            self.db.refresh(doc)
        return doc

    def soft_delete(self, document_id: int) -> bool:
        from datetime import datetime
        doc = self.get_by_id(document_id)
        if doc:
            doc.deleted_at = datetime.utcnow()
            self.db.commit()
            return True
        return False

    def update_read_position(self, document_id: int, position: str):
        from datetime import datetime
        doc = self.get_by_id(document_id)
        if doc:
            doc.last_read_position = position
            doc.last_read_at = datetime.utcnow()
            self.db.commit()

    def update_parsed_content(self, document_id: int, content: str, page_count: int = 0, word_count: int = 0):
        doc = self.get_by_id(document_id)
        if doc:
            doc.parsed_content = content
            doc.parsed_status = "done"
            doc.page_count = page_count
            doc.word_count = word_count
            self.db.commit()
