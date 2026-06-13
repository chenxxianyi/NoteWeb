import request from '../utils/request'
import type { Note } from '../types/note'

export function getNotes(params?: { document_id?: number; tag?: string }) {
  return request.get<Note[]>('/notes', { params })
}

export function getNote(id: number) {
  return request.get<Note>(`/notes/${id}`)
}

export function createNote(data: Partial<Note>) {
  return request.post<Note>('/notes', data)
}

export function updateNote(id: number, data: Partial<Note>) {
  return request.patch(`/notes/${id}`, data)
}

export function deleteNote(id: number) {
  return request.delete(`/notes/${id}`)
}
