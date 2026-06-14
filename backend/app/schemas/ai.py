from pydantic import BaseModel


class ExplainRequest(BaseModel):
    text: str


class TranslateRequest(BaseModel):
    text: str
    target_lang: str = "zh"


class ChatRequest(BaseModel):
    document_id: int
    question: str


class SummaryResponse(BaseModel):
    summary: str


class ExplainResponse(BaseModel):
    explanation: str


class TranslateResponse(BaseModel):
    translation: str


class ChatResponse(BaseModel):
    answer: str
