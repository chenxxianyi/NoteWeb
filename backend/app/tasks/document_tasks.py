"""Async document processing tasks."""

from app.tasks.celery_app import celery_app


@celery_app.task
def parse_document(document_id: int):
    """Background task to parse document content after upload."""
    # TODO: implement async parsing
    pass
