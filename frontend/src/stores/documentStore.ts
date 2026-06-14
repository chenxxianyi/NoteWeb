import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Document } from '../types/document'
import * as documentApi from '../api/document'

export const useDocumentStore = defineStore('document', () => {
  const documents = ref<Document[]>([])
  const currentDocument = ref<Document | null>(null)
  const documentContent = ref<string>('')
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchDocuments(params?: { search?: string; type?: string }) {
    loading.value = true
    error.value = null
    try {
      const res = await documentApi.getDocuments(params)
      documents.value = res.data
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '获取文件列表失败'
      documents.value = []
    } finally {
      loading.value = false
    }
  }

  async function fetchDocument(id: number) {
    error.value = null
    try {
      const res = await documentApi.getDocument(id)
      currentDocument.value = res.data
      return res.data
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '获取文件详情失败'
      currentDocument.value = null
      throw e
    }
  }

  async function fetchDocumentContent(id: number) {
    error.value = null
    try {
      const res = await documentApi.getDocumentContent(id)
      documentContent.value = res.data.content || ''
      return res.data.content
    } catch (e: any) {
      error.value = e?.response?.data?.detail || e.message || '获取文件内容失败'
      documentContent.value = ''
      throw e
    }
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

  return { documents, currentDocument, documentContent, loading, error, fetchDocuments, fetchDocument, fetchDocumentContent, upload, remove, rename }
})
