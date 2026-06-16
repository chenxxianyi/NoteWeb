from sqlalchemy.orm import Session
from fastapi import HTTPException

from app.repositories.note_repository import NoteRepository
from app.repositories.document_repository import DocumentRepository
from app.models.user import User


class NoteService:
    def __init__(self, db: Session):
        self.repo = NoteRepository(db)
        self.doc_repo = DocumentRepository(db)

    def list_notes(self, user: User, document_id: int | None = None,
                   tag: str | None = None) -> list[dict]:
        notes = self.repo.list_by_user(user.id, document_id, tag)
        doc_ids = {note.document_id for note in notes if note.document_id}
        docs_by_id = {}
        if doc_ids:
            docs = self.doc_repo.list_by_ids(list(doc_ids), user.id)
            docs_by_id = {doc.id: doc for doc in docs}
        return [self._to_dict(n, docs_by_id) for n in notes]

    def get_note(self, note_id: int, user: User) -> dict:
        note = self.repo.get_by_id(note_id)
        if not note or note.user_id != user.id:
            raise HTTPException(status_code=404, detail="Note not found")
        return self._to_dict(note)

    def create(self, data: dict, user: User) -> dict:
        note = self.repo.create(user.id, data)
        return self._to_dict(note)

    def update(self, note_id: int, data: dict, user: User) -> dict:
        note = self.repo.get_by_id(note_id)
        if not note or note.user_id != user.id:
            raise HTTPException(status_code=404, detail="Note not found")
        updated = self.repo.update(note_id, data)
        return self._to_dict(updated)

    def delete(self, note_id: int, user: User):
        note = self.repo.get_by_id(note_id)
        if not note or note.user_id != user.id:
            raise HTTPException(status_code=404, detail="Note not found")
        self.repo.soft_delete(note_id)

    def _to_dict(self, note, docs_by_id: dict[int, object] | None = None) -> dict:
        doc_title = None
        if note.document_id and docs_by_id:
            doc = docs_by_id.get(note.document_id)
            if doc:
                doc_title = doc.title

        tags = note.tags.split(",") if note.tags else []
        tags = [t.strip() for t in tags if t.strip()]

        return {
            "id": note.id,
            "title": note.title or "",
            "content": note.content or "",
            "document_id": note.document_id,
            "document_title": doc_title,
            "tags": tags,
            "created_at": note.created_at.isoformat() if note.created_at else "",
            "updated_at": note.updated_at.isoformat() if note.updated_at else "",
        }
