import request from '../utils/request'

export function getSummary(documentId: number) {
  return request.get<{ summary: string }>(`/ai/documents/${documentId}/summary`)
}

export function explain(text: string) {
  return request.post<{ explanation: string }>('/ai/explain', { text })
}

export function translate(text: string, targetLang = 'zh') {
  return request.post<{ translation: string }>('/ai/translate', { text, target_lang: targetLang })
}

export function chat(documentId: number, question: string) {
  return request.post<{ answer: string }>('/ai/chat', { document_id: documentId, question })
}
