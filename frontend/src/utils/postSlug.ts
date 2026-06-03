const SLUG_PATTERN = /^[a-z0-9-]+$/

/** 校验用户填写的 slug，空字符串表示留空由后端生成 */
export function validatePostSlug(slug: string): string {
  const s = slug.trim()
  if (!s) return ''
  if (!SLUG_PATTERN.test(s)) return '只允许小写字母、数字和连字符'
  if (/^\d+$/.test(s)) return '不能为纯数字（会与文章 ID 链接冲突，请包含字母）'
  return ''
}
