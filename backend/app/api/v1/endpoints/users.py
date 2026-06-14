from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from app.core.dependencies import get_db, get_current_user
from app.models.user import User
from app.services.user_service import UserService

router = APIRouter(prefix="/users", tags=["Users"])


@router.get("/me")
def get_current_user_info(current_user: User = Depends(get_current_user)):
    svc = UserService(Depends(get_db))  # not used — direct response
    return current_user
