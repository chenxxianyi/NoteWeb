export interface Note {
  id: number
  title: string
  content: string
  document_id?: number
  document_title?: string
  tags: string[]
  created_at: string
  updated_at: string
}
