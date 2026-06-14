from sqlalchemy.orm import Session
from sqlalchemy import desc
from app.models.annotation import Annotation
import json


class AnnotationRepository:
    def __init__(self, db: Session):
        self.db = db

    def get_by_id(self, annotation_id: int) -> Annotation | None:
        return (
            self.db.query(Annotation)
            .filter(Annotation.id == annotation_id, Annotation.deleted_at.is_(None))
            .first()
        )

    def list_by_document(self, document_id: int) -> list[Annotation]:
        return (
            self.db.query(Annotation)
            .filter(Annotation.document_id == document_id, Annotation.deleted_at.is_(None))
            .order_by(Annotation.created_at)
            .all()
        )

    def create(self, user_id: int, data: dict) -> Annotation:
        ann = Annotation(
            user_id=user_id,
            document_id=data.get("document_id"),
            page_number=data.get("page", 1),
            selected_text=data.get("selected_text", ""),
            color=data.get("color", "#FFD700"),
            annotation_type=data.get("type", "highlight"),
            note=data.get("note", ""),
            position_data=json.dumps(data.get("position_data", {}), ensure_ascii=False),
        )
        self.db.add(ann)
        self.db.commit()
        self.db.refresh(ann)
        return ann

    def soft_delete(self, annotation_id: int) -> bool:
        from datetime import datetime
        ann = self.get_by_id(annotation_id)
        if ann:
            ann.deleted_at = datetime.utcnow()
            self.db.commit()
            return True
        return False
