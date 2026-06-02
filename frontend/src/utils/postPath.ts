/** 文章详情页路径：纯数字 slug 会与文章 ID 冲突，链接统一用 ID */
export function getPostPath(post: { id: number; slug?: string }) {
  if (post.slug && !/^\d+$/.test(post.slug)) {
    return `/p/${post.slug}`
  }
  return `/p/${post.id}`
}
