import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Annotation } from '../types/annotation'
import * as annotationApi from '../api/annotation'

export const useAnnotationStore = defineStore('annotation', () => {
  const annotations = ref<Annotation[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  function sortByCreatedAtDesc(items: Annotation[]) {
    return [...items].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  }

  async function fetchAnnotations(documentId: number) {
    loading.value = true
    error.value = null
    try {
      const res = await annotationApi.getAnnotations(documentId)
      annotations.value = sortByCreatedAtDesc(res.data)
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '获取批注失败'
      annotations.value = []
    } finally {
      loading.value = false
    }
  }

  async function fetchAnnotationsForDocuments(documentIds: number[]) {
    const ids = Array.from(new Set(documentIds.filter((id) => Number.isFinite(id))))
    loading.value = true
    error.value = null

    if (ids.length === 0) {
      annotations.value = []
      loading.value = false
      return
    }

    try {
      const results = await Promise.allSettled(ids.map((id) => annotationApi.getAnnotations(id)))
      const fetched = results.flatMap((result) =>
        result.status === 'fulfilled' ? result.value.data : []
      )
      annotations.value = sortByCreatedAtDesc(fetched)
      if (results.some((result) => result.status === 'rejected')) {
        error.value = '部分批注获取失败'
      }
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '获取批注失败'
      annotations.value = []
    } finally {
      loading.value = false
    }
  }

  async function create(data: Partial<Annotation>) {
    const res = await annotationApi.createAnnotation(data)
    annotations.value = sortByCreatedAtDesc([...annotations.value, res.data])
    return res.data
  }

  async function remove(id: number) {
    await annotationApi.deleteAnnotation(id)
    annotations.value = annotations.value.filter((a) => a.id !== id)
  }

  async function replace(data: annotationApi.AnnotationReplacePayload) {
    const res = await annotationApi.replaceAnnotations(data)
    const deleted = new Set(data.delete_ids)
    annotations.value = sortByCreatedAtDesc([
      ...annotations.value.filter((annotation) => !deleted.has(annotation.id)),
      ...res.data,
    ])
    return res.data
  }

  return {
    annotations,
    loading,
    error,
    fetchAnnotations,
    fetchAnnotationsForDocuments,
    create,
    remove,
    replace,
  }
})
