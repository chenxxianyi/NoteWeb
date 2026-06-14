import datetime

from sqlalchemy import Column, Integer, String, BigInteger, DateTime, func

from app.db.base import Base


class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, autoincrement=True)
    username = Column(String(64), unique=True, nullable=False, index=True)
    email = Column(String(128), unique=True, nullable=False, index=True)
    password_hash = Column(String(256), nullable=False)
    avatar_url = Column(String(512), default="")
    storage_used = Column(BigInteger, default=0)  # bytes
    storage_limit = Column(BigInteger, default=1024 * 1024 * 1024)  # 1 GB
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
