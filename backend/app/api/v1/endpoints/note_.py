from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session

from app.core.dependencies import get_db, get_current_user
from app.models.user import User
from app.services.note_service import NoteService
from app.schemas.note import NoteCreate, NoteUpdate, NoteResponse

router = APIRouter(tags=["Notes"])


@router.get("/notes", response_model=list[NoteResponse])
def list_notes(
    document_id: int | None = Query(default=None),
    tag: str | None = Query(default=None),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = NoteService(db)
    return svc.list_notes(current_user, document_id, tag)


@router.get("/notes/{note_id}", response_model=NoteResponse)
def get_note(
    note_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = NoteService(db)
    return svc.get_note(note_id, current_user)


@router.post("/notes", response_model=NoteResponse)
def create_note(
    body: NoteCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = NoteService(db)
    return svc.create(body.model_dump(), current_user)


@router.patch("/notes/{note_id}", response_model=NoteResponse)
def update_note(
    note_id: int,
    body: NoteUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = NoteService(db)
    return svc.update(note_id, body.model_dump(exclude_none=True), current_user)


@router.delete("/notes/{note_id}")
def delete_note(
    note_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    svc = NoteService(db)
    svc.delete(note_id, current_user)
    return {"detail": "ok"}
