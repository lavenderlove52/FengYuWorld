import { PRESETS } from '@/types/theme'
import type { UserTheme } from '@/types/theme'

export function getPresetConfig(preset: string) {
  return PRESETS.find((p) => p.id === preset) ?? PRESETS[0]
}

/** 开启主题色开关后的有效颜色（自定义 hex 或当前预设强调色） */
export function resolveThemeColor(theme: UserTheme): string | null {
  if (!theme.theme_color_enabled) return null
  const preset = getPresetConfig(theme.preset)
  return theme.accent_color?.trim() || preset.accent
}

/** 开启背景色开关后的有效颜色（自定义 hex 或当前预设背景） */
export function resolveThemeBackground(theme: UserTheme): string | null {
  if (!theme.background_color_enabled) return null
  const preset = getPresetConfig(theme.preset)
  return theme.background_color?.trim() || preset.bg
}

export function buildThemeBackgroundLockCss(username: string, bg: string): string {
  const scope = `[data-theme-user="${CSS.escape(username)}"][data-theme-bg-active="true"]`
  const bodyHas = `html body:has(${scope})`

  return `
${bodyHas} {
  background-color: ${bg} !important;
}
${bodyHas} main {
  background-color: ${bg} !important;
}
${bodyHas} header {
  background-color: ${bg} !important;
}
${scope} {
  --theme-bg: ${bg} !important;
  background-color: ${bg} !important;
}
${scope} .home-content-shell,
${scope} .home-content-inner,
${scope} .theme-main-below,
${scope} .article-detail-root,
${scope} .article-detail-layout {
  background-color: ${bg} !important;
  background: ${bg} !important;
}
`.trim()
}

export function buildThemeColorLockCss(username: string, color: string): string {
  const scope = `[data-theme-user="${CSS.escape(username)}"][data-theme-color-active="true"]`
  const bodyHas = `html body:has(${scope})`

  return `
${scope} {
  --theme-accent: ${color} !important;
  --theme-color: ${color} !important;
  --md-theme-link-color: ${color} !important;
}
${scope} a,
${scope} .theme-color-accent,
${scope} .toc-link:hover {
  color: ${color} !important;
  border-left-color: ${color} !important;
}
${scope} .blog-markdown-viewer {
  --md-theme-link-color: ${color} !important;
}
${scope} .blog-markdown-viewer a,
${scope} .blog-markdown-viewer .md-editor-preview a {
  color: ${color} !important;
}
${scope} .blog-markdown-viewer blockquote {
  border-left-color: ${color} !important;
  background: color-mix(in srgb, ${color} 10%, transparent) !important;
}
${scope} .blog-markdown-viewer code:not(pre code) {
  color: ${color} !important;
  background: color-mix(in srgb, ${color} 12%, transparent) !important;
}
${scope} .post-list-title:hover,
${scope} .post-card-title:hover {
  color: ${color} !important;
  opacity: 1 !important;
}
${bodyHas} header .article-detail-home-link:hover {
  color: ${color} !important;
  border-color: ${color} !important;
}
`.trim()
}
