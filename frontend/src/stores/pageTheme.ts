import { reactive } from 'vue'
import { PRESETS } from '@/types/theme'
import { resolveThemeBackground, resolveThemeColor } from '@/utils/themeColor'
import type { UserTheme } from '@/types/theme'

// 当前页面激活的主题（用于 AppHeader 等全局组件读取）
export const activePageTheme = reactive({
  enabled: false,
  bg: '',
  text: '',
  accent: '',
  border: '',
  themeColorActive: false,
  themeBgActive: false,
})

export function setActivePageTheme(theme: UserTheme | null) {
  if (!theme) {
    activePageTheme.enabled = false
    activePageTheme.bg = ''
    activePageTheme.text = ''
    activePageTheme.accent = ''
    activePageTheme.border = ''
    activePageTheme.themeColorActive = false
    activePageTheme.themeBgActive = false
    return
  }
  const p = PRESETS.find((pr) => pr.id === theme.preset) ?? PRESETS[0]
  const locked = resolveThemeColor(theme)
  const lockedBg = resolveThemeBackground(theme)
  activePageTheme.enabled = true
  activePageTheme.bg = lockedBg || p.bg
  activePageTheme.text = p.text
  activePageTheme.accent = locked || theme.accent_color || p.accent
  activePageTheme.border = p.border
  activePageTheme.themeColorActive = !!locked
  activePageTheme.themeBgActive = !!lockedBg
}
