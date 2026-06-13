import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Annotation } from '../types/annotation'
import * as annotationApi from '../api/annotation'

export const useAnnotationStore = defineStore('annotation', () => {
  const annotations = ref<Annotation[]>([])

  async function fetchAnnotations(documentId: number) {
    const res = await annotationApi.getAnnotations(documentId)
    annotations.value = res.data
  }

  async function create(data: Partial<Annotation>) {
    const res = await annotationApi.createAnnotation(data)
    annotations.value.push(res.data)
  }

  async function remove(id: number) {
    await annotationApi.deleteAnnotation(id)
    annotations.value = annotations.value.filter((a) => a.id !== id)
  }

  return { annotations, fetchAnnotations, create, remove }
})
