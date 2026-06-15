import request from '../utils/request'
import type { Document, DocumentListParams } from '../types/document'

export function getDocuments(params?: DocumentListParams) {
  return request.get<Document[]>('/documents', { params })
}

export function getDocument(id: number) {
  return request.get<Document>(`/documents/${id}`)
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
