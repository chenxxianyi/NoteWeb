from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from app.core.dependencies import get_db, get_current_user
from app.models.user import User
from app.services.annotation_service import AnnotationService
from app.schemas.annotation import AnnotationCreate, AnnotationResponse

router = APIRouter(tags=["Annotations"])


@router.get("/documents/{document_id}/annotations", response_model=list[AnnotationResponse])
def list_annotations(
    document_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = AnnotationService(db)
    return svc.list_by_document(document_id, current_user)


@router.post("/annotations", response_model=AnnotationResponse)
def create_annotation(
    body: AnnotationCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = AnnotationService(db)
    return svc.create(body.model_dump(), current_user)


@router.delete("/annotations/{annotation_id}")
def delete_annotation(
    annotation_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = AnnotationService(db)
    svc.delete(annotation_id, current_user)
    return {"detail": "ok"}
