"""Celery application for async tasks."""

from celery import Celery
from app.core.config import settings

celery_app = Celery(
    "noteweb",
    broker=settings.redis_url,
    backend=settings.redis_url,
)

celery_app.conf.update(
    task_serializer="json",
    accept_content=["json"],
    result_serializer="json",
    timezone="Asia/Shanghai",
    enable_utc=True,
)
