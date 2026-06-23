from fastapi import APIRouter, Depends, UploadFile, File, Form, Query
from fastapi.responses import Response
from sqlalchemy.orm import Session

from app.core.dependencies import get_db, get_current_user
from app.models.user import User
from app.services.document_service import DocumentService
from app.schemas.document import DocumentRenameRequest, ReadPositionRequest

router = APIRouter(prefix="/documents", tags=["Documents"])


@router.get("")
def list_documents(
    search: str = Query(default=""),
    type: str = Query(default="", alias="type"),
    page: int = Query(default=1, ge=1),
    page_size: int = Query(default=50, ge=1, le=100),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = DocumentService(db)
    return svc.list_documents(current_user, search, type, page, page_size)


@router.get("/{document_id}")
def get_document(
    document_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = DocumentService(db)
    return svc.get_document(document_id, current_user)


@router.get("/{document_id}/content")
def get_document_content(
    document_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = DocumentService(db)
    return svc.get_content(document_id, current_user)


@router.get("/{document_id}/file")
def get_document_file(
    document_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = DocumentService(db)
    data, mime_type = svc.get_file(document_id, current_user)
    if data is None:
        from fastapi import HTTPException
        raise HTTPException(status_code=404, detail="File not found")
    return Response(content=data, media_type=mime_type or "application/octet-stream")


@router.post("/upload")
async def upload_document(
    file: UploadFile = File(...),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = DocumentService(db)
    return await svc.upload(file, current_user)


@router.patch("/{document_id}")
def rename_document(
    document_id: int,
    body: DocumentRenameRequest,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = DocumentService(db)
    return svc.rename(document_id, body.title, current_user)


@router.delete("/{document_id}")
def delete_document(
    document_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = DocumentService(db)
    svc.delete(document_id, current_user)
    return {"detail": "ok"}


@router.put("/{document_id}/read-position")
def update_read_position(
    document_id: int,
    body: ReadPositionRequest,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = DocumentService(db)
    svc.update_read_position(document_id, body.position, current_user)
    return {"detail": "ok"}
