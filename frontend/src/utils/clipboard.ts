/**
 * 复制文本到剪贴板。
 * navigator.clipboard 仅在安全上下文可用（HTTPS 或 localhost），
 * 生产环境若为 http://IP 会失败，故提供 execCommand 回退。
 */
export async function copyToClipboard(text: string): Promise<void> {
  if (!text) {
    throw new Error('没有可复制的内容')
  }

  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text)
      return
    } catch {
      // 非安全上下文或权限被拒时继续走回退
    }
  }

  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  textarea.style.top = '0'
  document.body.appendChild(textarea)
  textarea.focus()
  textarea.select()

  try {
    const ok = document.execCommand('copy')
    if (!ok) throw new Error('copy failed')
  } finally {
    document.body.removeChild(textarea)
  }
}
