import http from './http'
import type { ApiResponse } from '@/types/auth'
import type { FriendListItem, FriendshipItem } from '@/types/friend'

export const sendFriendRequest = (friend_username: string) =>
  http.post<ApiResponse<FriendshipItem>>('/friends/requests', { friend_username }).then((r) => r.data)

export const listReceivedRequests = () =>
  http.get<ApiResponse<FriendshipItem[]>>('/friends/requests/received').then((r) => r.data)

export const listSentRequests = () =>
  http.get<ApiResponse<FriendshipItem[]>>('/friends/requests/sent').then((r) => r.data)

export const acceptRequest = (id: number) =>
  http.post<ApiResponse<null>>(`/friends/requests/${id}/accept`).then((r) => r.data)

export const rejectRequest = (id: number) =>
  http.post<ApiResponse<null>>(`/friends/requests/${id}/reject`).then((r) => r.data)

export const listFriends = () =>
  http.get<ApiResponse<FriendListItem[]>>('/friends').then((r) => r.data)

export const deleteFriend = (userId: number) =>
  http.delete<ApiResponse<null>>(`/friends/${userId}`).then((r) => r.data)
