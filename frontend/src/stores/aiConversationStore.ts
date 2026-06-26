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

  const STREAM_DISPLAY_INTERVAL_MS = 24

  function getDisplayChunkSize(queuedLength: number) {
    if (queuedLength > 160) return 12
    if (queuedLength > 80) return 8
    if (queuedLength > 24) return 4
    return 2
  }

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

  function createMessageDisplayStream(index: number) {
    let queued = ''
    let timer: ReturnType<typeof window.setTimeout> | null = null
    let cancelled = false
    const idleResolvers: Array<() => void> = []

    const clearTimer = () => {
      if (timer === null) return
      window.clearTimeout(timer)
      timer = null
    }

    const resolveIdle = () => {
      if (queued || timer !== null) return
      while (idleResolvers.length > 0) {
        idleResolvers.shift()?.()
      }
    }

    const schedule = () => {
      if (timer !== null || cancelled) return
      timer = window.setTimeout(tick, STREAM_DISPLAY_INTERVAL_MS)
    }

    const tick = () => {
      timer = null

      if (cancelled) {
        queued = ''
        resolveIdle()
        return
      }

      if (queued) {
        const chunkSize = Math.min(queued.length, getDisplayChunkSize(queued.length))
        const chunk = queued.slice(0, chunkSize)
        queued = queued.slice(chunkSize)
        appendMessageDelta(index, chunk)
      }

      if (queued) {
        schedule()
      } else {
        resolveIdle()
      }
    }

    return {
      push(delta: string) {
        if (!delta || cancelled) return
        queued += delta
        schedule()
      },
      finish() {
        if (!queued && timer === null) return Promise.resolve()
        return new Promise<void>((resolve) => {
          idleResolvers.push(resolve)
          schedule()
        })
      },
      cancel() {
        cancelled = true
        queued = ''
        clearTimer()
        resolveIdle()
      },
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
    let displayStream: ReturnType<typeof createMessageDisplayStream> | null = null

    try {
      pushMessage('user', question)
      assistantIndex = pushMessage('assistant', '')
      displayStream = createMessageDisplayStream(assistantIndex)

      const done = await aiApi.chatStream({
        document_id: documentId || undefined,
        question,
        conversation_type: type,
      }, {
        onDelta: (content) => displayStream?.push(content),
      })
      await displayStream.finish()

      return {
        answer: messages.value[assistantIndex]?.content || '',
        conversation_id: done.conversation_id || 0,
      }
    } catch (e: any) {
      displayStream?.cancel()
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
    let displayStream: ReturnType<typeof createMessageDisplayStream> | null = null

    try {
      assistantIndex = pushMessage('assistant', '文档总结\n\n')
      displayStream = createMessageDisplayStream(assistantIndex)
      const done = await aiApi.getSummaryStream(documentId, {
        onDelta: (content) => displayStream?.push(content),
      })
      await displayStream.finish()

      return {
        summary: messages.value[assistantIndex]?.content.replace(/^文档总结\n\n/, '') || '',
        conversation_id: done.conversation_id || 0,
      }
    } catch (e: any) {
      displayStream?.cancel()
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
    let displayStream: ReturnType<typeof createMessageDisplayStream> | null = null

    try {
      pushMessage('user', query)
      assistantIndex = pushMessage('assistant', '')
      displayStream = createMessageDisplayStream(assistantIndex)

      const done = await aiApi.searchStream({ query }, {
        onDelta: (content) => displayStream?.push(content),
      })
      await displayStream.finish()

      return {
        answer: messages.value[assistantIndex]?.content || '',
        sources: [],
        conversation_id: done.conversation_id || 0,
      }
    } catch (e: any) {
      displayStream?.cancel()
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
