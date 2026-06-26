import request from '../utils/request'
import type {
  ChatResponse,
  SummaryResponse,
  SearchResponse,
  AIConversation,
  AIProviderConfig,
} from '../types/ai'

type StreamHandlers = {
  onDelta?: (content: string) => void
  signal?: AbortSignal
}

type StreamDone = {
  conversation_id?: number
}

const API_BASE = (import.meta.env.VITE_API_BASE || '/api/v1').replace(/\/$/, '')

// 获取文档总结
export function getSummary(documentId: number) {
  return request.get<SummaryResponse>(`/ai/documents/${documentId}/summary`)
}

export function getSummaryStream(documentId: number, handlers: StreamHandlers = {}) {
  return postStream<StreamDone>(`/ai/documents/${documentId}/summary/stream`, {}, handlers)
}

// AI聊天
export function chat(data: {
  document_id?: number
  question: string
  conversation_type?: string
}) {
  return request.post<ChatResponse>('/ai/chat', data)
}

export function chatStream(data: {
  document_id?: number
  question: string
  conversation_type?: string
}, handlers: StreamHandlers = {}) {
  return postStream<StreamDone>('/ai/chat/stream', data, handlers)
}

// 联网搜索
export function search(data: { query: string }) {
  return request.post<SearchResponse>('/ai/search', data)
}

export function searchStream(data: { query: string }, handlers: StreamHandlers = {}) {
  return postStream<StreamDone>('/ai/search/stream', data, handlers)
}

// 获取对话历史
export function getConversations(documentId?: number) {
  const params = documentId ? { document_id: documentId } : {}
  return request.get<AIConversation[]>('/ai/conversations', { params })
}

// 删除对话
export function deleteConversation(id: number) {
  return request.delete(`/ai/conversations/${id}`)
}

// 获取AI配置
export function getProviderConfig() {
  return request.get<AIProviderConfig>('/ai/provider-config')
}

// 更新AI配置
export function updateProviderConfig(data: {
  provider: string
  api_key?: string
  base_url?: string
  model?: string
}) {
  return request.put<AIProviderConfig>('/ai/provider-config', data)
}

// 旧API保留(兼容性)
export function explain(text: string) {
  return request.post<{ explanation: string }>('/ai/explain', { text })
}

export function translate(text: string, targetLang = 'zh') {
  return request.post<{ translation: string }>('/ai/translate', {
    text,
    target_lang: targetLang,
  })
}

async function postStream<T>(path: string, body: unknown, handlers: StreamHandlers): Promise<T> {
  const token = localStorage.getItem('token')
  const response = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(body),
    signal: handlers.signal,
  })

  if (!response.ok) {
    throw await responseError(response)
  }
  if (!response.body) {
    throw new Error('AI 流式响应不可用')
  }

  return readSSE<T>(response, handlers.onDelta)
}

async function responseError(response: Response) {
  try {
    const data = await response.json()
    return new Error(data?.detail || `请求失败: ${response.status}`)
  } catch {
    const text = await response.text().catch(() => '')
    return new Error(text || `请求失败: ${response.status}`)
  }
}

async function readSSE<T>(response: Response, onDelta?: (content: string) => void): Promise<T> {
  const reader = response.body!.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let donePayload = {} as T

  while (true) {
    const { value, done } = await reader.read()
    buffer += decoder.decode(value, { stream: !done })

    const blocks = buffer.split(/\r?\n\r?\n/)
    buffer = blocks.pop() || ''

    for (const block of blocks) {
      donePayload = handleSSEBlock<T>(block, onDelta) ?? donePayload
    }

    if (done) break
  }

  if (buffer.trim()) {
    donePayload = handleSSEBlock<T>(buffer, onDelta) ?? donePayload
  }

  return donePayload
}

function handleSSEBlock<T>(block: string, onDelta?: (content: string) => void): T | undefined {
  let event = 'message'
  const dataLines: string[] = []

  for (const line of block.split(/\r?\n/)) {
    if (line.startsWith('event:')) {
      event = line.slice(6).trim()
    } else if (line.startsWith('data:')) {
      dataLines.push(line.slice(5).trimStart())
    }
  }

  if (dataLines.length === 0) return undefined

  const payload = JSON.parse(dataLines.join('\n'))
  if (event === 'delta') {
    const content = typeof payload?.content === 'string' ? payload.content : ''
    if (content) onDelta?.(content)
    return undefined
  }
  if (event === 'error') {
    throw new Error(payload?.detail || 'AI 调用失败')
  }
  if (event === 'done') {
    return payload as T
  }

  return undefined
}
