import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Note } from '../types/note'
import * as noteApi from '../api/note'

export const useNoteStore = defineStore('note', () => {
  const notes = ref<Note[]>([])
  const currentNote = ref<Note | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchNotes(params?: { document_id?: number; tag?: string }) {
    loading.value = true
    error.value = null
    try {
      const res = await noteApi.getNotes(params)
      notes.value = res.data
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '获取笔记列表失败'
      notes.value = []
    } finally {
      loading.value = false
    }
  }

  async function fetchNote(id: number) {
    error.value = null
    try {
      const res = await noteApi.getNote(id)
      currentNote.value = res.data
      return res.data
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '获取笔记详情失败'
      currentNote.value = null
      throw e
    }
  }

  async function create(data: Partial<Note>) {
    const res = await noteApi.createNote(data)
    notes.value.unshift(res.data)
    currentNote.value = res.data
    return res.data
  }

  async function update(id: number, data: Partial<Note>) {
    await noteApi.updateNote(id, data)
    if (currentNote.value?.id === id) {
      Object.assign(currentNote.value, data)
    }
    const idx = notes.value.findIndex((n) => n.id === id)
    if (idx !== -1) Object.assign(notes.value[idx], data)
  }

  async function remove(id: number) {
    await noteApi.deleteNote(id)
    notes.value = notes.value.filter((n) => n.id !== id)
    if (currentNote.value?.id === id) currentNote.value = null
  }

  return { notes, currentNote, loading, error, fetchNotes, fetchNote, create, update, remove }
})
