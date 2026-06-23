import request from '../utils/request'
import type { Annotation } from '../types/annotation'

export interface AnnotationReplacementCreate {
  page: number
  selected_text: string
  color: string
  type: Annotation['type']
  note?: string
  position_data: Record<string, unknown>
}

export interface AnnotationReplacePayload {
  document_id: number
  delete_ids: number[]
  creates: AnnotationReplacementCreate[]
}

export function getAnnotations(documentId: number) {
  return request.get<Annotation[]>(`/documents/${documentId}/annotations`)
}

export function createAnnotation(data: Partial<Annotation>) {
  return request.post<Annotation>('/annotations', data)
}

export function deleteAnnotation(id: number) {
  return request.delete(`/annotations/${id}`)
}

export function replaceAnnotations(data: AnnotationReplacePayload) {
  return request.post<Annotation[]>('/annotations/replace', data)
}
