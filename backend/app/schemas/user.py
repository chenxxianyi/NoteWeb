from pydantic import BaseModel, Field
from datetime import datetime


class UserResponse(BaseModel):
    id: int
    username: str
    email: str
    avatar: str | None = Field(default="", validation_alias="avatar_url")
    storage_used: int
    storage_limit: int

    model_config = {"from_attributes": True, "populate_by_name": True}
