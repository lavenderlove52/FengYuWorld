export type ThemePreset = 'default' | 'minimal' | 'dark' | 'forest' | 'ocean' | 'warm'
export type ThemeLayout = 'grid' | 'list' | 'magazine'
export type ThemeEffectScope = 'home' | 'post' | 'all'
export type ThemeBackgroundEffectType = 'off' | 'floating-shapes'
export type ThemeCursorEffectType = 'off' | 'particle-trail'
export type ThemeClickEffectType = 'off' | 'heart' | 'star' | 'dot'

export interface ThemeEffectConfig {
  scope: ThemeEffectScope
  background: {
    type: ThemeBackgroundEffectType
    color: string
    count: number
    size: number
    opacity: number
    speed: number
  }
  cursor: {
    type: ThemeCursorEffectType
    color: string
    count: number
    size: number
    opacity: number
    duration: number
  }
  click: {
    type: ThemeClickEffectType
    color: string
    count: number
    size: number
    opacity: number
    duration: number
  }
}

export interface UserTheme {
  user_id: number
  preset: ThemePreset
  accent_color: string
  font_family: string
  banner_url: string
  layout: ThemeLayout
  show_pins_public: boolean
  custom_css: string
  custom_js: string
  effect_config: ThemeEffectConfig
}

export interface UpdateThemePayload {
  preset?: ThemePreset
  accent_color?: string
  font_family?: string
  banner_url?: string
  layout?: ThemeLayout
  show_pins_public?: boolean
  custom_css?: string
  custom_js?: string
  effect_config?: ThemeEffectConfig
}

export const createDefaultEffectConfig = (): ThemeEffectConfig => ({
  scope: 'all',
  background: {
    type: 'off',
    color: '#6366f1',
    count: 20,
    size: 20,
    opacity: 0.25,
    speed: 0.5,
  },
  cursor: {
    type: 'off',
    color: '#6366f1',
    count: 16,
    size: 5,
    opacity: 0.7,
    duration: 700,
  },
  click: {
    type: 'off',
    color: '#ec4899',
    count: 8,
    size: 16,
    opacity: 0.9,
    duration: 900,
  },
})

export const DEFAULT_THEME: UserTheme = {
  user_id: 0,
  preset: 'default',
  accent_color: '',
  font_family: '',
  banner_url: '',
  layout: 'grid',
  show_pins_public: false,
  custom_css: '',
  custom_js: '',
  effect_config: createDefaultEffectConfig(),
}

export interface PresetConfig {
  id: ThemePreset
  label: string
  bg: string
  accent: string
  text: string
  border: string
  cardBg: string
  desc: string
}

export const PRESETS: PresetConfig[] = [
  {
    id: 'default',
    label: '默认',
    bg: '#f8fafc',
    accent: '#6366f1',
    text: '#0f172a',
    border: '#e2e8f0',
    cardBg: '#ffffff',
    desc: '清爽蓝紫，当前风格',
  },
  {
    id: 'minimal',
    label: '极简',
    bg: '#ffffff',
    accent: '#0f172a',
    text: '#0f172a',
    border: '#e5e7eb',
    cardBg: '#f9fafb',
    desc: '纯白黑字，干净利落',
  },
  {
    id: 'dark',
    label: '暗色',
    bg: '#0f172a',
    accent: '#818cf8',
    text: '#e2e8f0',
    border: '#1e293b',
    cardBg: '#1e293b',
    desc: '深夜模式，护眼低调',
  },
  {
    id: 'forest',
    label: '森林',
    bg: '#f0fdf4',
    accent: '#16a34a',
    text: '#14532d',
    border: '#bbf7d0',
    cardBg: '#ffffff',
    desc: '自然绿意，清新宜人',
  },
  {
    id: 'ocean',
    label: '海洋',
    bg: '#eff6ff',
    accent: '#2563eb',
    text: '#1e3a5f',
    border: '#bfdbfe',
    cardBg: '#ffffff',
    desc: '深邃蓝调，沉静开阔',
  },
  {
    id: 'warm',
    label: '暖阳',
    bg: '#fefce8',
    accent: '#d97706',
    text: '#713f12',
    border: '#fde68a',
    cardBg: '#fffbeb',
    desc: '温暖金黄，治愈温馨',
  },
]

export const FONT_OPTIONS = [
  { label: '全局默认', value: '' },
  { label: '霞鹜文楷 (LXGW WenKai)', value: 'LXGW WenKai TC' },
  { label: 'Lora（衬线）', value: 'Lora' },
  { label: 'Noto Sans SC（思源黑体）', value: 'Noto Sans SC' },
  { label: 'Noto Serif SC（思源宋体）', value: 'Noto Serif SC' },
]
