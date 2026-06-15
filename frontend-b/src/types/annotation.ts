export interface Annotation {
  id: number
  document_id: number
  page: number
  selected_text: string
  color: string
  type: 'highlight' | 'underline' | 'comment'
  note?: string
  position_data: Record<string, unknown>
  created_at: string
}
