export interface Message {
  id: number
  from_user: number
  to_user: number
  content: string
  type: 'text' | 'image' | 'emoji'
  is_read: boolean
  created_at: string
}

export interface UnreadCounts {
  [friendId: number]: number
}

export type WsOutMsgType = 'msg' | 'unread_counts' | 'unread_update' | 'error'

export interface WsOutMsg {
  type: WsOutMsgType
  payload: unknown
}
