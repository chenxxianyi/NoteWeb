// AI 对话消息
export interface Message {
  role: 'user' | 'assistant' | 'system'
  content: string
}

// AI 对话
export interface AIConversation {
  id: number
  user_id: number
  document_id?: number
  title: string
  conversation_type: 'chat' | 'search' | 'summary'
  messages: string // JSON字符串
  created_at: string
  updated_at: string
}

// AI 提供商配置
export interface AIProviderConfig {
  provider: string
  base_url: string
  model: string
  has_api_key?: boolean
}

// 聊天响应
export interface ChatResponse {
  answer: string
  conversation_id: number
}

// 总结响应
export interface SummaryResponse {
  summary: string
  conversation_id: number
}

// 搜索响应
export interface SearchResponse {
  answer: string
  sources: SearchSource[]
  conversation_id: number
}

// 搜索来源
export interface SearchSource {
  title: string
  url: string
  snippet: string
}
