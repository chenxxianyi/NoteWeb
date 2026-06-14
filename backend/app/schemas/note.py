from pydantic import BaseModel
from datetime import datetime
from typing import Optional


class NoteCreate(BaseModel):
    title: str = ""
    content: str = ""
    document_id: int | None = None
    tags: list[str] = []


class NoteUpdate(BaseModel):
    title: str | None = None
    content: str | None = None
    document_id: int | None = None
    tags: list[str] | None = None


class NoteResponse(BaseModel):
    id: int
    title: str
    content: str
    document_id: int | None = None
    document_title: str | None = None
    tags: list[str] = []
    created_at: datetime
    updated_at: datetime

    model_config = {"from_attributes": True}
