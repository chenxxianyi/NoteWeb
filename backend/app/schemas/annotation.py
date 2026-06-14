from pydantic import BaseModel
from datetime import datetime
from typing import Optional


class AnnotationCreate(BaseModel):
    document_id: int
    page: int = 1
    selected_text: str = ""
    color: str = "#FFD700"
    type: str = "highlight"  # highlight | underline | comment
    note: str = ""
    position_data: dict = {}


class AnnotationResponse(BaseModel):
    id: int
    document_id: int
    page: int
    selected_text: str
    color: str
    type: str
    note: str = ""
    position_data: dict = {}
    created_at: datetime

    model_config = {"from_attributes": True}
