from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from app.core.dependencies import get_db, get_current_user
from app.models.user import User
from app.services.ai_service import AiService
from app.schemas.ai import (
    ExplainRequest, TranslateRequest, ChatRequest,
    SummaryResponse, ExplainResponse, TranslateResponse, ChatResponse,
)

router = APIRouter(prefix="/ai", tags=["AI"])


@router.get("/documents/{document_id}/summary", response_model=SummaryResponse)
async def get_summary(document_id: int):
    svc = AiService()
    summary = await svc.get_summary(document_id)
    return {"summary": summary}


@router.post("/explain", response_model=ExplainResponse)
async def explain(body: ExplainRequest):
    svc = AiService()
    explanation = await svc.explain(body.text)
    return {"explanation": explanation}


@router.post("/translate", response_model=TranslateResponse)
async def translate(body: TranslateRequest):
    svc = AiService()
    translation = await svc.translate(body.text, body.target_lang)
    return {"translation": translation}


@router.post("/chat", response_model=ChatResponse)
async def chat(body: ChatRequest):
    svc = AiService()
    answer = await svc.chat(body.document_id, body.question)
    return {"answer": answer}
