import request from '../utils/request'
import type { Document, DocumentListParams } from '../types/document'

export function getDocuments(params?: DocumentListParams) {
  return request.get<Document[]>('/documents', { params })
}

export function getDocument(id: number) {
  return request.get<Document>(`/documents/${id}`)
}

export function getDocumentContent(id: number) {
  return request.get<{ content: string }>(`/documents/${id}/content`)
}

export function uploadDocument(file: File) {
  const form = new FormData()
  form.append('file', file)
  return request.post<Document>('/documents/upload', form)
}

export function deleteDocument(id: number) {
  return request.delete(`/documents/${id}`)
}

export function renameDocument(id: number, title: string) {
  return request.patch(`/documents/${id}`, { title })
}

export function markDocumentAsRead(id: number) {
  return request.patch(`/documents/${id}/read-progress`)
}

export function updateReadProgress(id: number, progress: number) {
  return request.patch(`/documents/${id}/read-progress`, { progress })
}
