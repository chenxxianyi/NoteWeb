import request from '../utils/request'

export interface UserSettings {
  id: number
  user_id: number
  theme: string
  font: string
  reading_mode: boolean
  created_at: string
  updated_at: string
}

export interface UpdateSettingsPayload {
  theme?: string
  font?: string
  reading_mode?: boolean
}

export function getSettings() {
  return request.get<UserSettings>('/settings')
}

export function updateSettings(data: UpdateSettingsPayload) {
  return request.patch<{ detail: string }>('/settings', data)
}
