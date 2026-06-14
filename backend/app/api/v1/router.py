from fastapi import APIRouter

from app.api.v1.endpoints import auth, users, documents, annotations, ai
from app.api.v1.endpoints import note_ as notes

router = APIRouter(prefix="/api/v1")

router.include_router(auth.router)
router.include_router(users.router)
router.include_router(documents.router)
router.include_router(annotations.router)
router.include_router(notes.router)
router.include_router(ai.router)
