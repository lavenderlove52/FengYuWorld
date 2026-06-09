import type { PostSummary } from './post'

export type PostTreeNodeType = 'folder' | 'post'

export interface PostTreeNode {
  id: number
  parent_id?: number | null
  node_type: PostTreeNodeType
  name: string
  post_id?: number
  sort_order: number
  post?: PostSummary
  children?: PostTreeNode[]
}

export interface CreatePostTreeFolderPayload {
  parent_id?: number | null
  name: string
  sort_order?: number
}

export interface AddPostTreePostPayload {
  parent_id?: number | null
  post_id: number
  sort_order?: number
}

export interface UpdatePostTreeNodePayload {
  parent_id?: number | null
  name?: string
  sort_order?: number
}
