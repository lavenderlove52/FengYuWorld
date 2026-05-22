export interface UserInfo {
  id: number
  username: string
  email?: string
  nickname: string
  avatar_url: string
  bio?: string
  role?: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  user: UserInfo
}

export interface ApiResponse<T> {
  code: number
  message: string
  data?: T
}
