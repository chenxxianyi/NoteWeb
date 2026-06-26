import { defineStore } from 'pinia'
import { ref } from 'vue'
import type {
  Message,
  AIConversation,
  ChatResponse,
  SummaryResponse,
  SearchResponse,
} from '../types/ai'
import * as aiApi from '../api/ai'

export const useAIConversationStore = defineStore('aiConversation', () => {
  // State
  const conversations = ref<AIConversation[]>([])
  const currentConversation = ref<AIConversation | null>(null)
  const messages = ref<Message[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Actions
  function pushMessage(role: Message['role'], content: string) {
    messages.value.push({ role, content })
    return messages.value.length - 1
  }

  function appendMessageDelta(index: number, delta: string) {
    const current = messages.value[index]
    if (!current) return
    messages.value[index] = {
      ...current,
      content: current.content + delta,
    }
  }

  function removeEmptyAssistant(index: number) {
    const current = messages.value[index]
    if (current?.role === 'assistant' && !current.content.trim()) {
      messages.value.splice(index, 1)
    }
  }

  /**
   * 发送聊天消息
   */
  async function chat(
    documentId: number | null,
    question: string,
    type: 'chat' | 'search' | 'summary' = 'chat'
  ): Promise<ChatResponse | null> {
    loading.value = true
    error.value = null
    let assistantIndex = -1

    try {
      pushMessage('user', question)
      assistantIndex = pushMessage('assistant', '')

      const done = await aiApi.chatStream({
        document_id: documentId || undefined,
        question,
        conversation_type: type,
      }, {
        onDelta: (content) => appendMessageDelta(assistantIndex, content),
      })

      return {
        answer: messages.value[assistantIndex]?.content || '',
        conversation_id: done.conversation_id || 0,
      }
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '聊天失败'
      removeEmptyAssistant(assistantIndex)
      return null
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取文档总结
   */
  async function getSummary(documentId: number): Promise<SummaryResponse | null> {
    loading.value = true
    error.value = null
    let assistantIndex = -1

    try {
      assistantIndex = pushMessage('assistant', '文档总结\n\n')
      const done = await aiApi.getSummaryStream(documentId, {
        onDelta: (content) => appendMessageDelta(assistantIndex, content),
      })

      return {
        summary: messages.value[assistantIndex]?.content.replace(/^文档总结\n\n/, '') || '',
        conversation_id: done.conversation_id || 0,
      }
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '获取总结失败'
      removeEmptyAssistant(assistantIndex)
      return null
    } finally {
      loading.value = false
    }
  }

  /**
   * 联网搜索
   */
  async function search(query: string): Promise<SearchResponse | null> {
    loading.value = true
    error.value = null
    let assistantIndex = -1

    try {
      pushMessage('user', query)
      assistantIndex = pushMessage('assistant', '')

      const done = await aiApi.searchStream({ query }, {
        onDelta: (content) => appendMessageDelta(assistantIndex, content),
      })

      return {
        answer: messages.value[assistantIndex]?.content || '',
        sources: [],
        conversation_id: done.conversation_id || 0,
      }
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '搜索失败'
      removeEmptyAssistant(assistantIndex)
      return null
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取对话历史
   */
  async function fetchConversations(documentId?: number) {
    loading.value = true
    error.value = null

    try {
      const response = await aiApi.getConversations(documentId)
      conversations.value = response.data
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '获取对话历史失败'
      conversations.value = []
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除对话
   */
  async function deleteConversation(id: number) {
    try {
      await aiApi.deleteConversation(id)
      conversations.value = conversations.value.filter((c) => c.id !== id)
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '删除对话失败'
      throw e
    }
  }

  /**
   * 清空当前对话
   */
  function clearCurrentConversation() {
    currentConversation.value = null
    messages.value = []
  }

  return {
    // State
    conversations,
    currentConversation,
    messages,
    loading,
    error,
    // Actions
    chat,
    getSummary,
    search,
    fetchConversations,
    deleteConversation,
    clearCurrentConversation,
  }
})
