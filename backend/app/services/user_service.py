from sqlalchemy.orm import Session
from app.repositories.user_repository import UserRepository
from app.models.user import User


class UserService:
    def __init__(self, db: Session):
        self.repo = UserRepository(db)

    def get_user(self, user: User) -> dict:
        return {
            "id": user.id,
            "username": user.username,
            "email": user.email,
            "avatar": user.avatar_url or "",
            "storage_used": user.storage_used or 0,
            "storage_limit": user.storage_limit,
        }
