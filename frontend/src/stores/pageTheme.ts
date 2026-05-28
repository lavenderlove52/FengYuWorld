import { reactive } from 'vue'
import { PRESETS } from '@/types/theme'
import type { UserTheme } from '@/types/theme'

// 当前页面激活的主题（用于 AppHeader 等全局组件读取）
export const activePageTheme = reactive({
  enabled: false,
  bg: '',
  text: '',
  accent: '',
  border: '',
})

export function setActivePageTheme(theme: UserTheme | null) {
  if (!theme) {
    activePageTheme.enabled = false
    activePageTheme.bg = ''
    activePageTheme.text = ''
    activePageTheme.accent = ''
    activePageTheme.border = ''
    return
  }
  const p = PRESETS.find((pr) => pr.id === theme.preset) ?? PRESETS[0]
  activePageTheme.enabled = true
  activePageTheme.bg = p.bg
  activePageTheme.text = p.text
  activePageTheme.accent = theme.accent_color || p.accent
  activePageTheme.border = p.border
}
