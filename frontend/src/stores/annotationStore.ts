import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Annotation } from '../types/annotation'
import * as annotationApi from '../api/annotation'

export const useAnnotationStore = defineStore('annotation', () => {
  const annotations = ref<Annotation[]>([])
  const error = ref<string | null>(null)

  async function fetchAnnotations(documentId: number) {
    error.value = null
    try {
      const res = await annotationApi.getAnnotations(documentId)
      annotations.value = res.data
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '获取批注失败'
      annotations.value = []
    }
  }

  async function create(data: Partial<Annotation>) {
    const res = await annotationApi.createAnnotation(data)
    annotations.value.push(res.data)
    return res.data
  }

  async function remove(id: number) {
    await annotationApi.deleteAnnotation(id)
    annotations.value = annotations.value.filter((a) => a.id !== id)
  }

  async function replace(data: annotationApi.AnnotationReplacePayload) {
    const res = await annotationApi.replaceAnnotations(data)
    const deleted = new Set(data.delete_ids)
    annotations.value = [
      ...annotations.value.filter((annotation) => !deleted.has(annotation.id)),
      ...res.data,
    ]
    return res.data
  }

  return { annotations, error, fetchAnnotations, create, remove, replace }
})
