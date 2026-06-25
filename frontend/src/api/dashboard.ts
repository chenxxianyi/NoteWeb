import request from '../utils/request'

export interface ContinueReading {
  id: number
  title: string
  fileType: string
  readProgress: number
  updatedAt: string
}

export interface RecentActivity {
  type: string // "highlight", "note", "upload"
  id: number
  documentId: number
  title?: string
  text?: string
  document: string
  page: string
  createdAt: string
}

export interface DashboardSummary {
  documentCount: number
  annotationCount: number
  noteCount: number
  weeklyReadingSeconds: number
  continueReading: ContinueReading[]
  recentActivities: RecentActivity[]
}

export function getDashboardSummary() {
  return request.get<DashboardSummary>('/dashboard/summary')
}
