from sqlalchemy.orm import Session
from fastapi import HTTPException

from app.repositories.annotation_repository import AnnotationRepository
from app.models.user import User
import json


class AnnotationService:
    def __init__(self, db: Session):
        self.repo = AnnotationRepository(db)

    def list_by_document(self, document_id: int, user: User) -> list[dict]:
        annotations = self.repo.list_by_document(document_id)
        # Filter by user ownership
        result = []
        for ann in annotations:
            if ann.user_id != user.id:
                continue
            result.append(self._to_dict(ann))
        return result

    def create(self, data: dict, user: User) -> dict:
        data["document_id"] = data.get("document_id")
        ann = self.repo.create(user.id, data)
        return self._to_dict(ann)

    def delete(self, annotation_id: int, user: User):
        ann = self.repo.get_by_id(annotation_id)
        if not ann or ann.user_id != user.id:
            raise HTTPException(status_code=404, detail="Annotation not found")
        self.repo.soft_delete(annotation_id)

    def _to_dict(self, ann) -> dict:
        pos_data = {}
        try:
            pos_data = json.loads(ann.position_data) if ann.position_data else {}
        except (json.JSONDecodeError, TypeError):
            pos_data = {}
        return {
            "id": ann.id,
            "document_id": ann.document_id,
            "page": ann.page_number or 1,
            "selected_text": ann.selected_text or "",
            "color": ann.color or "#FFD700",
            "type": ann.annotation_type or "highlight",
            "note": ann.note or "",
            "position_data": pos_data,
            "created_at": ann.created_at.isoformat() if ann.created_at else "",
        }
