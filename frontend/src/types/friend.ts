import type { UserInfo } from './auth'

export interface FriendshipItem {
  id: number
  status: string
  created_at: string
  user: UserInfo
}

export interface FriendListItem extends UserInfo {
  friends_since: string
}
