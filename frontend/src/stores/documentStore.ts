import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Document } from '../types/document'
import * as documentApi from '../api/document'

export const useDocumentStore = defineStore('document', () => {
  const documents = ref<Document[]>([])
  const currentDocument = ref<Document | null>(null)
  const loading = ref(false)

  async function fetchDocuments(params?: { search?: string; type?: string }) {
    loading.value = true
    try {
      const res = await documentApi.getDocuments(params)
      documents.value = res.data
    } finally {
      loading.value = false
    }
  }

  async function fetchDocument(id: number) {
    const res = await documentApi.getDocument(id)
    currentDocument.value = res.data
    return res.data
  }

  async function upload(file: File) {
    const res = await documentApi.uploadDocument(file)
    documents.value.unshift(res.data)
    return res.data
  }

  async function remove(id: number) {
    await documentApi.deleteDocument(id)
    documents.value = documents.value.filter((d) => d.id !== id)
  }

  async function rename(id: number, title: string) {
    await documentApi.renameDocument(id, title)
    const doc = documents.value.find((d) => d.id === id)
    if (doc) doc.title = title
  }

  return { documents, currentDocument, loading, fetchDocuments, fetchDocument, upload, remove, rename }
})
