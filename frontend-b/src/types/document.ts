export interface Document {
  id: number
  title: string
  file_type: 'pdf' | 'md' | 'docx' | 'txt'
  file_size: number
  read_progress: number
  created_at: string
  updated_at: string
  parsed_content?: string
  page_count?: number
  word_count?: number
  file_name?: string
  mime_type?: string
}

export interface DocumentListParams {
  search?: string
  type?: string
  page?: number
  page_size?: number
}
