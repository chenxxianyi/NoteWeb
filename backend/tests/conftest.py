"""Test configuration and fixtures for NoteWeb backend tests."""

import os
import pytest
from sqlalchemy import create_engine
from sqlalchemy.orm import Session, sessionmaker
from fastapi.testclient import TestClient
from unittest.mock import patch

from app.db.base import Base
from app.main import app

TEST_DB_PATH = os.path.join(os.path.dirname(__file__), "test.db")


@pytest.fixture(scope="session")
def engine():
    """Create SQLite engine for testing — session-scoped, created once."""
    url = f"sqlite:///{TEST_DB_PATH}"
    eng = create_engine(url, connect_args={"check_same_thread": False})
    Base.metadata.create_all(bind=eng)
    yield eng
    Base.metadata.drop_all(bind=eng)
    try:
        if os.path.exists(TEST_DB_PATH):
            os.remove(TEST_DB_PATH)
    except PermissionError:
        pass


@pytest.fixture()
def db(engine):
    """Per-test database session — rolled back after each test."""
    connection = engine.connect()
    transaction = connection.begin()
    session = Session(bind=connection)

    yield session

    session.close()
    transaction.rollback()
    connection.close()


# Patch StorageService to avoid MinIO connection
@pytest.fixture(autouse=True)
def mock_storage():
    with patch("app.services.document_service.StorageService", autospec=True) as mock:
        instance = mock.return_value
        instance.upload_bytes.return_value = None
        instance.get_url.return_value = ""
        yield mock


@pytest.fixture()
def client(db):
    """TestClient with test database session."""
    from app.core.dependencies import get_db

    def override_get_db():
        yield db

    app.dependency_overrides[get_db] = override_get_db
    yield TestClient(app)
    app.dependency_overrides.clear()


@pytest.fixture()
def auth_headers(client, db):
    """Register a test user and return Authorization header."""
    resp = client.post("/api/v1/auth/register", json={
        "username": "tester",
        "email": "tester@test.com",
        "password": "Test1234",
        "confirm_password": "Test1234",
    })
    if resp.status_code != 200:
        resp = client.post("/api/v1/auth/login", json={
            "email": "tester@test.com",
            "password": "Test1234",
        })
    token = resp.json()["token"]
    return {"Authorization": f"Bearer {token}"}
