import datetime

from sqlalchemy import (
    Column, Integer, String, Text, DateTime,
    ForeignKey, func,
)

from app.db.base import Base


class Annotation(Base):
    __tablename__ = "annotations"

    id = Column(Integer, primary_key=True, autoincrement=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False, index=True)
    document_id = Column(Integer, ForeignKey("documents.id"), nullable=False, index=True)
    page_number = Column(Integer, default=1)
    annotation_type = Column(String(16), default="highlight")  # highlight | underline | comment
    color = Column(String(32), default="#FFD700")
    selected_text = Column(Text, default="")
    note = Column(Text, default="")
    position_data = Column(Text, default="{}")  # JSON string
    range_data = Column(Text, default="{}")  # JSON string
    tags = Column(String(256), default="")
    deleted_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, server_default=func.now())
