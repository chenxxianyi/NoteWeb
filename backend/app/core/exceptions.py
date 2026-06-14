"""Custom exception classes and global exception handler."""

from fastapi import Request
from fastapi.responses import JSONResponse


class AppException(Exception):
    """Base application exception with a code and message."""

    def __init__(self, code: int = 40000, message: str = "Internal error"):
        self.code = code
        self.message = message


class NotFoundException(AppException):
    def __init__(self, detail: str = "Resource not found"):
        super().__init__(code=40400, message=detail)


class UnauthorizedException(AppException):
    def __init__(self, detail: str = "Unauthorized"):
        super().__init__(code=40100, message=detail)


class ForbiddenException(AppException):
    def __init__(self, detail: str = "Forbidden"):
        super().__init__(code=40300, message=detail)


class ValidationException(AppException):
    def __init__(self, detail: str = "Validation error"):
        super().__init__(code=42200, message=detail)


async def app_exception_handler(request: Request, exc: AppException):
    return JSONResponse(
        status_code=exc.code // 100,
        content={"detail": exc.message},
    )
