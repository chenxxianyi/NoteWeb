from sqlalchemy.orm import Session
from fastapi import HTTPException, status

from app.repositories.user_repository import UserRepository
from app.utils.password import hash_password, verify_password
from app.utils.jwt import create_token
from app.schemas.auth import AuthResponse
from app.schemas.user import UserResponse


class AuthService:
    def __init__(self, db: Session):
        self.repo = UserRepository(db)

    def register(self, username: str, email: str, password: str, confirm_password: str) -> AuthResponse:
        if password != confirm_password:
            raise HTTPException(status_code=400, detail="Passwords do not match")
        if self.repo.get_by_email(email):
            raise HTTPException(status_code=400, detail="Email already registered")

        pwd_hash = hash_password(password)
        user = self.repo.create(username, email, pwd_hash)
        token = create_token(user.id)
        return AuthResponse(
            token=token,
            user=UserResponse(
                id=user.id,
                username=user.username,
                email=user.email,
                avatar=user.avatar_url or "",
                storage_used=user.storage_used or 0,
                storage_limit=user.storage_limit,
            ),
        )

    def login(self, email: str, password: str) -> AuthResponse:
        user = self.repo.get_by_email(email)
        if not user or not verify_password(password, user.password_hash):
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid email or password",
            )
        token = create_token(user.id)
        return AuthResponse(
            token=token,
            user=UserResponse(
                id=user.id,
                username=user.username,
                email=user.email,
                avatar=user.avatar_url or "",
                storage_used=user.storage_used or 0,
                storage_limit=user.storage_limit,
            ),
        )
