"""MinIO storage service."""

import io
import urllib3
from minio import Minio
from minio.error import S3Error

from app.core.config import settings

# Patch urllib3 timeout for MinIO health checks
_http = urllib3.PoolManager(timeout=urllib3.Timeout(connect=2.0, read=5.0))


class StorageService:
    def __init__(self):
        try:
            self.client = Minio(
                settings.MINIO_ENDPOINT,
                access_key=settings.MINIO_ACCESS_KEY,
                secret_key=settings.MINIO_SECRET_KEY,
                secure=settings.MINIO_SECURE,
                http_client=_http,
            )
            self._ensure_bucket()
        except Exception:
            self.client = None

    def _ensure_bucket(self):
        if self.client is None:
            return
        try:
            if not self.client.bucket_exists(settings.MINIO_BUCKET):
                self.client.make_bucket(settings.MINIO_BUCKET)
        except Exception:
            pass  # MinIO not available — graceful fallback

    def upload_bytes(self, object_name: str, data: bytes, content_type: str = "application/octet-stream"):
        if self.client is None:
            return
        try:
            self.client.put_object(
                settings.MINIO_BUCKET,
                object_name,
                io.BytesIO(data),
                length=len(data),
                content_type=content_type,
            )
        except Exception:
            pass

    def get_url(self, object_name: str) -> str:
        if self.client is None:
            return ""
        try:
            return self.client.presigned_get_object(settings.MINIO_BUCKET, object_name)
        except Exception:
            return ""
