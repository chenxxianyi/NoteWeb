from pydantic import BaseModel
from datetime import datetime
from typing import Optional


class DocumentResponse(BaseModel):
    id: int
    title: str
    file_type: str  # pdf | md | docx | txt
    file_size: int
    read_progress: float = 0.0
    created_at: datetime
    updated_at: datetime

    model_config = {"from_attributes": True}


class DocumentDetailResponse(BaseModel):
    id: int
    title: str
    file_name: str
    file_type: str
    mime_type: str
    file_size: int
    storage_path: str
    parsed_content: str
    page_count: int
    word_count: int
    read_progress: float = 0.0
    created_at: datetime
    updated_at: datetime

    model_config = {"from_attributes": True}


class DocumentRenameRequest(BaseModel):
    title: str


class ReadPositionRequest(BaseModel):
    position: str
