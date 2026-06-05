export interface PublicUserInfo {
  username: string
  nickname: string
  avatar_url: string
  bio?: string
}

export interface UserInfo {
  id: number
  username: string
  email?: string
  nickname: string
  avatar_url: string
  bio?: string
  role?: 'user' | 'admin' | 'super_admin' | string
  can_site_pin?: boolean
  can_custom_js?: boolean
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
