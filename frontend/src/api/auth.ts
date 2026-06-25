import request from '../utils/request'
import type { LoginPayload, RegisterPayload, AuthResponse, User } from '../types/auth'

export function login(data: LoginPayload) {
  return request.post<AuthResponse>('/auth/login', data)
}

export function register(data: RegisterPayload) {
  return request.post<AuthResponse>('/auth/register', data)
}

export function getMe() {
  return request.get<User>('/auth/me')
}

// Change password
export interface ChangePasswordPayload {
  old_password: string
  new_password: string
}

export function changePassword(data: ChangePasswordPayload) {
  return request.post<{ detail: string }>('/auth/change-password', data)
}

// Update profile
export interface UpdateProfilePayload {
  username?: string
  email?: string
}

export function updateProfile(data: UpdateProfilePayload) {
  return request.patch<User>('/auth/profile', data)
}

// Upload avatar
export function uploadAvatar(file: File) {
  const form = new FormData()
  form.append('file', file)
  return request.post<{ url: string; user: User }>('/auth/avatar', form)
}

// Delete account
export interface DeleteAccountPayload {
  password: string
}

export function deleteAccount(data: DeleteAccountPayload) {
  return request.delete<{ detail: string }>('/auth/account', { data })
}
