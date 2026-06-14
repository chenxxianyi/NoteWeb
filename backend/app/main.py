"""NoteWeb Backend - FastAPI application entry point."""

from contextlib import asynccontextmanager
import os
import time

from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware

from app.core.config import settings
from app.core.exceptions import AppException, app_exception_handler
from app.api.v1.router import router


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Auto-create tables for SQLite dev mode
    if settings.DATABASE_URL and "sqlite" in settings.DATABASE_URL:
        from app.db.base import Base
        from app.db.session import engine
        Base.metadata.create_all(bind=engine)
    yield


app = FastAPI(
    title=settings.APP_NAME,
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
    lifespan=lifespan,
)

# Request logging middleware
@app.middleware("http")
async def log_requests(request: Request, call_next):
    start_time = time.time()
    print(f"[REQUEST] {request.method} {request.url.path} from {request.client.host if request.client else 'unknown'}")
    response = await call_next(request)
    duration = time.time() - start_time
    print(f"[RESPONSE] {request.method} {request.url.path} -> {response.status_code} ({duration:.3f}s)")
    return response

# CORS - allow frontend dev server
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Exception handlers
app.add_exception_handler(AppException, app_exception_handler)

# API routes
app.include_router(router)


@app.get("/health")
def health_check():
    return {"status": "ok", "app": settings.APP_NAME}


@app.get("/__debug/ping")
def debug_ping():
    return {"status": "ok", "pid": os.getpid(), "app": settings.APP_NAME}
