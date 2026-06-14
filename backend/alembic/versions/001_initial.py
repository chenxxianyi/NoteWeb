"""Initial migration: create all tables.

Revision ID: 001
Create Date: 2025-01-01
"""
from alembic import op
import sqlalchemy as sa
from datetime import datetime

revision = "001"
down_revision = None
branch_labels = None
depends_on = None


def upgrade() -> None:
    # users
    op.create_table(
        "users",
        sa.Column("id", sa.Integer(), autoincrement=True, nullable=False),
        sa.Column("username", sa.String(64), nullable=False, unique=True),
        sa.Column("email", sa.String(128), nullable=False, unique=True),
        sa.Column("password_hash", sa.String(256), nullable=False),
        sa.Column("avatar_url", sa.String(512), default=""),
        sa.Column("storage_used", sa.BigInteger(), default=0),
        sa.Column("storage_limit", sa.BigInteger(), default=1073741824),
        sa.Column("created_at", sa.DateTime(), server_default=sa.func.now()),
        sa.Column("updated_at", sa.DateTime(), server_default=sa.func.now()),
        sa.PrimaryKeyConstraint("id"),
    )
    op.create_index("ix_users_email", "users", ["email"])
    op.create_index("ix_users_username", "users", ["username"])

    # documents
    op.create_table(
        "documents",
        sa.Column("id", sa.Integer(), autoincrement=True, nullable=False),
        sa.Column("user_id", sa.Integer(), sa.ForeignKey("users.id"), nullable=False),
        sa.Column("title", sa.String(256), nullable=False),
        sa.Column("file_name", sa.String(256), nullable=False),
        sa.Column("file_type", sa.String(16), nullable=False),
        sa.Column("mime_type", sa.String(64), default=""),
        sa.Column("file_size", sa.BigInteger(), default=0),
        sa.Column("storage_path", sa.String(512), default=""),
        sa.Column("preview_url", sa.String(512), default=""),
        sa.Column("parsed_status", sa.String(16), default="pending"),
        sa.Column("parsed_content", sa.Text(), default=""),
        sa.Column("page_count", sa.Integer(), default=0),
        sa.Column("word_count", sa.Integer(), default=0),
        sa.Column("last_read_position", sa.String(256), default=""),
        sa.Column("last_read_at", sa.DateTime(), nullable=True),
        sa.Column("deleted_at", sa.DateTime(), nullable=True),
        sa.Column("created_at", sa.DateTime(), server_default=sa.func.now()),
        sa.Column("updated_at", sa.DateTime(), server_default=sa.func.now()),
        sa.PrimaryKeyConstraint("id"),
    )
    op.create_index("ix_documents_user_id", "documents", ["user_id"])

    # annotations
    op.create_table(
        "annotations",
        sa.Column("id", sa.Integer(), autoincrement=True, nullable=False),
        sa.Column("user_id", sa.Integer(), sa.ForeignKey("users.id"), nullable=False),
        sa.Column("document_id", sa.Integer(), sa.ForeignKey("documents.id"), nullable=False),
        sa.Column("page_number", sa.Integer(), default=1),
        sa.Column("annotation_type", sa.String(16), default="highlight"),
        sa.Column("color", sa.String(32), default="#FFD700"),
        sa.Column("selected_text", sa.Text(), default=""),
        sa.Column("note", sa.Text(), default=""),
        sa.Column("position_data", sa.Text(), default="{}"),
        sa.Column("range_data", sa.Text(), default="{}"),
        sa.Column("tags", sa.String(256), default=""),
        sa.Column("deleted_at", sa.DateTime(), nullable=True),
        sa.Column("created_at", sa.DateTime(), server_default=sa.func.now()),
        sa.PrimaryKeyConstraint("id"),
    )
    op.create_index("ix_annotations_user_id", "annotations", ["user_id"])
    op.create_index("ix_annotations_document_id", "annotations", ["document_id"])

    # notes
    op.create_table(
        "notes",
        sa.Column("id", sa.Integer(), autoincrement=True, nullable=False),
        sa.Column("user_id", sa.Integer(), sa.ForeignKey("users.id"), nullable=False),
        sa.Column("document_id", sa.Integer(), sa.ForeignKey("documents.id"), nullable=True),
        sa.Column("source_annotation_id", sa.Integer(), nullable=True),
        sa.Column("title", sa.String(256), default=""),
        sa.Column("content", sa.Text(), default=""),
        sa.Column("content_type", sa.String(16), default="markdown"),
        sa.Column("tags", sa.String(256), default=""),
        sa.Column("deleted_at", sa.DateTime(), nullable=True),
        sa.Column("created_at", sa.DateTime(), server_default=sa.func.now()),
        sa.Column("updated_at", sa.DateTime(), server_default=sa.func.now()),
        sa.PrimaryKeyConstraint("id"),
    )
    op.create_index("ix_notes_user_id", "notes", ["user_id"])


def downgrade() -> None:
    op.drop_table("notes")
    op.drop_table("annotations")
    op.drop_table("documents")
    op.drop_table("users")
