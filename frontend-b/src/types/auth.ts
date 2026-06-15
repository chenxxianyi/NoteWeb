export interface User {
  id: number
  username: string
  email: string
  avatar?: string
  storage_used: number
  storage_limit: number
}

export interface LoginPayload {
  email: string
  password: string
}

export interface RegisterPayload {
  username: string
  email: string
  password: string
  confirm_password: string
}

export interface AuthResponse {
  token: string
  user: User
}
