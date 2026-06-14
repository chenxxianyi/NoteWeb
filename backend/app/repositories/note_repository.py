from sqlalchemy.orm import Session
from sqlalchemy import desc
from app.models.note import Note


class NoteRepository:
    def __init__(self, db: Session):
        self.db = db

    def get_by_id(self, note_id: int) -> Note | None:
        return (
            self.db.query(Note)
            .filter(Note.id == note_id, Note.deleted_at.is_(None))
            .first()
        )

    def list_by_user(self, user_id: int, document_id: int | None = None,
                     tag: str | None = None, skip: int = 0, limit: int = 50) -> list[Note]:
        query = self.db.query(Note).filter(
            Note.user_id == user_id, Note.deleted_at.is_(None)
        )
        if document_id:
            query = query.filter(Note.document_id == document_id)
        if tag:
            query = query.filter(Note.tags.ilike(f"%{tag}%"))
        return query.order_by(desc(Note.updated_at)).offset(skip).limit(limit).all()

    def create(self, user_id: int, data: dict) -> Note:
        tags = data.get("tags", [])
        note = Note(
            user_id=user_id,
            title=data.get("title", ""),
            content=data.get("content", ""),
            document_id=data.get("document_id"),
            tags=",".join(tags) if isinstance(tags, list) else tags,
        )
        self.db.add(note)
        self.db.commit()
        self.db.refresh(note)
        return note

    def update(self, note_id: int, data: dict) -> Note | None:
        note = self.get_by_id(note_id)
        if not note:
            return None
        for key, value in data.items():
            if key == "tags" and isinstance(value, list):
                setattr(note, key, ",".join(value))
            elif key == "tags" and isinstance(value, str):
                setattr(note, key, value)
            elif hasattr(note, key) and value is not None:
                setattr(note, key, value)
        self.db.commit()
        self.db.refresh(note)
        return note

    def soft_delete(self, note_id: int) -> bool:
        from datetime import datetime
        note = self.get_by_id(note_id)
        if note:
            note.deleted_at = datetime.utcnow()
            self.db.commit()
            return True
        return False
