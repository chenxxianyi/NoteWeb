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
