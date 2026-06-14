"""Application configuration via environment variables."""

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    APP_NAME: str = "NoteWeb API"
    ENV: str = "development"
    DEBUG: bool = True
    API_PREFIX: str = "/api/v1"

    # JWT
    SECRET_KEY: str = "change-me-in-production-noteweb-secret"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 60 * 24  # 24 hours

    # MySQL (used when DATABASE_URL is not set)
    MYSQL_HOST: str = "localhost"
    MYSQL_PORT: int = 3306
    MYSQL_USER: str = "noteweb"
    MYSQL_PASSWORD: str = "noteweb123"
    MYSQL_DATABASE: str = "noteweb"

    # Override both with DATABASE_URL for SQLite local dev
    # e.g. DATABASE_URL=sqlite:///./noteweb.db
    DATABASE_URL: str | None = None

    # Redis (reserved for Celery / caching)
    REDIS_HOST: str = "localhost"
    REDIS_PORT: int = 6379

    # MinIO
    MINIO_ENDPOINT: str = "localhost:9000"
    MINIO_ACCESS_KEY: str = "noteweb"
    MINIO_SECRET_KEY: str = "noteweb123"
    MINIO_BUCKET: str = "noteweb-files"
    MINIO_SECURE: bool = False

    # Upload
    MAX_UPLOAD_SIZE: int = 50 * 1024 * 1024  # 50 MB

    @property
    def database_url(self) -> str:
        if self.DATABASE_URL:
            return self.DATABASE_URL
        return (
            f"mysql+pymysql://{self.MYSQL_USER}:{self.MYSQL_PASSWORD}"
            f"@{self.MYSQL_HOST}:{self.MYSQL_PORT}/{self.MYSQL_DATABASE}?charset=utf8mb4"
        )

    @property
    def redis_url(self) -> str:
        return f"redis://{self.REDIS_HOST}:{self.REDIS_PORT}/0"

    model_config = {"env_file": ".env", "env_file_encoding": "utf-8"}


settings = Settings()
