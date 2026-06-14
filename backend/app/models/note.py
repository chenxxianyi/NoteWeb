import datetime

from sqlalchemy import (
    Column, Integer, String, Text, DateTime,
    ForeignKey, func,
)

from app.db.base import Base


class Note(Base):
    __tablename__ = "notes"

    id = Column(Integer, primary_key=True, autoincrement=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False, index=True)
    document_id = Column(Integer, ForeignKey("documents.id"), nullable=True)
    source_annotation_id = Column(Integer, nullable=True)
    title = Column(String(256), default="")
    content = Column(Text, default="")
    content_type = Column(String(16), default="markdown")  # markdown | rich_text
    tags = Column(String(256), default="")
    deleted_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
