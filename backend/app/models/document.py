import datetime

from sqlalchemy import (
    Column, Integer, String, BigInteger, Text, DateTime,
    ForeignKey, Float, Index, func,
)
from sqlalchemy.orm import relationship

from app.db.base import Base


class Document(Base):
    __tablename__ = "documents"

    id = Column(Integer, primary_key=True, autoincrement=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False, index=True)
    title = Column(String(256), nullable=False)
    file_name = Column(String(256), nullable=False)
    file_type = Column(String(16), nullable=False)  # pdf, md, docx, txt
    mime_type = Column(String(128), default="")
    file_size = Column(BigInteger, default=0)
    storage_path = Column(String(512), default="")
    preview_url = Column(String(512), default="")
    parsed_status = Column(String(16), default="pending")  # pending | done | failed
    parsed_content = Column(Text, default="")
    page_count = Column(Integer, default=0)
    word_count = Column(Integer, default=0)
    last_read_position = Column(String(256), default="")
    last_read_at = Column(DateTime, nullable=True)
    deleted_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())

    owner = relationship("User", backref="documents")
