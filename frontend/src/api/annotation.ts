import request from '../utils/request'
import type { Annotation } from '../types/annotation'

export function getAnnotations(documentId: number) {
  return request.get<Annotation[]>(`/documents/${documentId}/annotations`)
}

export function createAnnotation(data: Partial<Annotation>) {
  return request.post<Annotation>('/annotations', data)
}

export function deleteAnnotation(id: number) {
  return request.delete(`/annotations/${id}`)
}
