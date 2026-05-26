import type { UserInfo } from './auth'

export interface PostSummary {
  id: number
  title: string
  slug: string
  cover_image_url: string
  summary: string
  status: string
  visibility: string
  view_count: number
  published_at?: string
  created_at: string
  author: UserInfo
  tags?: string[]
}

export interface PostDetail extends PostSummary {
  content_md?: string
  content_html: string
  category_id?: number
  allow_comment: boolean
  visible_user_ids?: number[]
}

export interface PageResult<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface PinnedPost {
  id: number
  post_id: number
  display_order: number
  post: PostSummary
}

export type PostVisibility = 'public' | 'private' | 'friends' | 'specified'
export type PostStatus = 'draft' | 'published'

export interface CreatePostPayload {
  title: string
  slug?: string
  cover_image_url?: string
  content_md: string
  category_id?: number
  tag_names?: string[]
  status: PostStatus
  visibility: PostVisibility
  visible_user_ids?: number[]
  allow_comment?: boolean
}
