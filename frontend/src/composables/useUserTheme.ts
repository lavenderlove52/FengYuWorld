import { ref, onUnmounted } from 'vue'
import { getUserTheme } from '@/api/theme'
import { DEFAULT_THEME, createDefaultEffectConfig, createDefaultHeroConfig } from '@/types/theme'
import { setActivePageTheme } from '@/stores/pageTheme'
import {
  buildThemeBackgroundLockCss,
  buildThemeColorLockCss,
  getPresetConfig,
  resolveThemeBackground,
  resolveThemeColor,
} from '@/utils/themeColor'
import type { UserTheme } from '@/types/theme'

const STYLE_ID_PREFIX = 'user-theme-css-'
const THEME_COLOR_LOCK_ID_PREFIX = 'user-theme-color-lock-'
const THEME_BG_LOCK_ID_PREFIX = 'user-theme-bg-lock-'

const jsCleanups = new Map<string, Array<() => void>>()

function defaultTheme(): UserTheme {
  return {
    ...DEFAULT_THEME,
    effect_config: createDefaultEffectConfig(),
    hero_config: createDefaultHeroConfig(),
  }
}

function normalizeTheme(theme?: UserTheme | null): UserTheme {
  return {
    ...defaultTheme(),
    ...theme,
    effect_config: theme?.effect_config ?? createDefaultEffectConfig(),
    hero_config: {
      ...createDefaultHeroConfig(),
      ...theme?.hero_config,
      image_ids: theme?.hero_config?.image_ids ?? [],
    },
  }
}

function buildCssVars(theme: UserTheme): Record<string, string> {
  const p = getPresetConfig(theme.preset)
  const locked = resolveThemeColor(theme)
  const lockedBg = resolveThemeBackground(theme)
  return {
    '--theme-bg': lockedBg || p.bg,
    '--theme-accent': locked || theme.accent_color || p.accent,
    '--theme-text': p.text,
    '--theme-border': p.border,
    '--theme-card-bg': p.cardBg,
    '--theme-font': theme.font_family || 'inherit',
  }
}

function injectCustomCss(username: string, css: string) {
  const id = STYLE_ID_PREFIX + username
  let el = document.getElementById(id) as HTMLStyleElement | null
  if (!el) {
    el = document.createElement('style')
    el.id = id
    document.head.appendChild(el)
  }

  const scope = `[data-theme-user="${CSS.escape(username)}"]`
  const scoped = css
    .split('}')
    .filter((block) => block.trim())
    .map((block) => {
      const [selector, ...rest] = block.split('{')
      if (!rest.length) return ''
      const selectors = selector
        .split(',')
        .map((s) => {
          const trimmed = s.trim()
          if (!trimmed) return ''
          // 用户已写 [data-theme-user] 时，替换为当前用户精确 scope，避免双重选择器导致样式不生效
          if (trimmed.includes('[data-theme-user')) {
            return trimmed.replace(/\[data-theme-user\]/g, scope)
          }
          return `${scope} ${trimmed}`
        })
        .filter(Boolean)
        .join(', ')
      return `${selectors} { ${rest.join('{')} }`
    })
    .join('\n')

  el.textContent = scoped
}

function removeCustomCss(username: string) {
  document.getElementById(STYLE_ID_PREFIX + username)?.remove()
}

function registerCleanup(username: string, cleanup: () => void) {
  const list = jsCleanups.get(username) ?? []
  list.push(cleanup)
  jsCleanups.set(username, list)
}

function runCleanup(username: string) {
  const list = jsCleanups.get(username) ?? []
  for (const cleanup of list.splice(0).reverse()) {
    try {
      cleanup()
    } catch (e) {
      console.warn('[useUserTheme] custom_js cleanup failed', e)
    }
  }
  jsCleanups.delete(username)
}

function createScopedDocument(root: HTMLElement) {
  const scopedDocument = {
    querySelector: (selector: string) => root.querySelector(selector),
    querySelectorAll: (selector: string) => root.querySelectorAll(selector),
    getElementById: (id: string) => {
      const el = document.getElementById(id)
      return el && root.contains(el) ? el : null
    },
    getElementsByClassName: (className: string) => root.getElementsByClassName(className),
    getElementsByTagName: (tagName: string) => root.getElementsByTagName(tagName),
    createElement: document.createElement.bind(document),
    createTextNode: document.createTextNode.bind(document),
    addEventListener: root.addEventListener.bind(root),
    removeEventListener: root.removeEventListener.bind(root),
    body: root,
    documentElement: root,
  }

  return scopedDocument as unknown as Document
}

function createScopedWindow(username: string, root: HTMLElement) {
  const customState = new Map<PropertyKey, unknown>()
  const cleanupList: Array<() => void> = []

  customState.set('__themeRoot', root)
  customState.set('__themeCleanup', cleanupList)
  customState.set('__bgCleanup', cleanupList)
  registerCleanup(username, () => {
    for (const cleanup of cleanupList.splice(0).reverse()) {
      cleanup()
    }
  })

  return new Proxy(window, {
    get(target, prop, receiver) {
      if (prop === 'document') return createScopedDocument(root)
      if (prop === 'window' || prop === 'self' || prop === 'globalThis' || prop === 'top' || prop === 'parent') {
        return receiver
      }
      if (customState.has(prop)) return customState.get(prop)
      return Reflect.get(target, prop, target)
    },
    set(_target, prop, value) {
      customState.set(prop, value)
      return true
    },
  })
}

function injectCustomJs(username: string, js: string, root: HTMLElement | null) {
  removeCustomJs(username)
  if (!js.trim() || !root) return

  const scopedDocument = createScopedDocument(root)
  const scopedWindow = createScopedWindow(username, root)
  const beforeChildren = new Set(root.childNodes)

  // 勿在 strict 下声明 const eval / const arguments（会直接 SyntaxError，导致整段 custom_js 不执行）
  const runner = new Function(
    'root',
    'document',
    'window',
    'globalThis',
    'self',
    'parent',
    'top',
    'console',
    `${js}
//# sourceURL=custom-theme-${encodeURIComponent(username)}.js`,
  )

  runner(root, scopedDocument, scopedWindow, scopedWindow, scopedWindow, scopedWindow, scopedWindow, console)
  registerCleanup(username, () => {
    for (const node of Array.from(root.childNodes)) {
      if (!beforeChildren.has(node)) {
        node.parentNode?.removeChild(node)
      }
    }
  })
}

function removeCustomJs(username: string) {
  runCleanup(username)
}

function removeThemeColorLock(username: string) {
  document.getElementById(THEME_COLOR_LOCK_ID_PREFIX + username)?.remove()
}

function applyThemeColorLock(username: string, theme: UserTheme, root: HTMLElement | null) {
  removeThemeColorLock(username)
  root?.removeAttribute('data-theme-color-active')

  const color = resolveThemeColor(theme)
  if (!color || !root) return

  root.setAttribute('data-theme-color-active', 'true')
  const el = document.createElement('style')
  el.id = THEME_COLOR_LOCK_ID_PREFIX + username
  el.textContent = buildThemeColorLockCss(username, color)
  document.head.appendChild(el)
}

function removeThemeBackgroundLock(username: string) {
  document.getElementById(THEME_BG_LOCK_ID_PREFIX + username)?.remove()
}

function applyThemeBackgroundLock(username: string, theme: UserTheme, root: HTMLElement | null) {
  removeThemeBackgroundLock(username)
  root?.removeAttribute('data-theme-bg-active')

  const bg = resolveThemeBackground(theme)
  if (!bg || !root) return

  root.setAttribute('data-theme-bg-active', 'true')
  const el = document.createElement('style')
  el.id = THEME_BG_LOCK_ID_PREFIX + username
  el.textContent = buildThemeBackgroundLockCss(username, bg)
  document.head.appendChild(el)
}

function applyThemeLocks(username: string, theme: UserTheme, root: HTMLElement | null) {
  applyThemeColorLock(username, theme, root)
  applyThemeBackgroundLock(username, theme, root)
}

export function useUserTheme(containerRef: { value: HTMLElement | null }) {
  const theme = ref<UserTheme>(defaultTheme())
  const loading = ref(false)
  let currentUsername = ''
  let originalBodyBg = ''
  let originalBodyColor = ''

  const applyToBody = (t: UserTheme) => {
    const p = getPresetConfig(t.preset)
    originalBodyBg = document.body.style.backgroundColor
    originalBodyColor = document.body.style.color
    document.body.style.backgroundColor = resolveThemeBackground(t) || p.bg
    document.body.style.color = p.text
    document.body.style.fontFamily = t.font_family ? `'${t.font_family}', serif` : ''
    setActivePageTheme(t)
  }

  const restoreBody = () => {
    document.body.style.backgroundColor = originalBodyBg
    document.body.style.color = originalBodyColor
    document.body.style.fontFamily = ''
    setActivePageTheme(null)
  }

  const applyToContainer = (t: UserTheme) => {
    const el = containerRef.value
    if (!el) return
    const vars = buildCssVars(t)
    Object.entries(vars).forEach(([k, v]) => el.style.setProperty(k, v))
    if (t.font_family) {
      el.style.setProperty('font-family', `'${t.font_family}', serif`)
    } else {
      el.style.removeProperty('font-family')
    }
    el.style.setProperty('color', 'var(--theme-text)')
    el.style.setProperty('min-height', '100%')
  }

  const load = async (username: string) => {
    if (currentUsername && currentUsername !== username) {
      removeCustomCss(currentUsername)
      removeCustomJs(currentUsername)
      removeThemeColorLock(currentUsername)
      removeThemeBackgroundLock(currentUsername)
    }
    currentUsername = username
    loading.value = true
    try {
      const res = await getUserTheme(username)
      theme.value = res.code === 0 && res.data ? normalizeTheme(res.data) : defaultTheme()
    } catch {
      theme.value = defaultTheme()
    } finally {
      loading.value = false
      try { applyToBody(theme.value) } catch (e) { console.warn('[useUserTheme] applyToBody failed', e) }
      try { applyToContainer(theme.value) } catch (e) { console.warn('[useUserTheme] applyToContainer failed', e) }
      if (theme.value.custom_css) {
        try { injectCustomCss(username, theme.value.custom_css) } catch (e) { console.warn('[useUserTheme] injectCustomCss failed', e) }
      }
      if (theme.value.custom_js) {
        try { injectCustomJs(username, theme.value.custom_js, containerRef.value) } catch (e) { console.warn('[useUserTheme] injectCustomJs failed', e) }
      }
      const applyLock = () => {
        try {
          applyThemeLocks(username, theme.value, containerRef.value)
        } catch (e) {
          console.warn('[useUserTheme] applyThemeLocks failed', e)
        }
      }
      applyLock()
      if (theme.value.custom_js) {
        requestAnimationFrame(applyLock)
      }
    }
  }

  const applyPreview = (t: UserTheme) => {
    theme.value = t
    applyToContainer(t)
    if (t.custom_css) {
      injectCustomCss(currentUsername, t.custom_css)
    } else {
      removeCustomCss(currentUsername)
    }
  }

  onUnmounted(() => {
    if (currentUsername) {
      removeCustomCss(currentUsername)
      removeCustomJs(currentUsername)
      removeThemeColorLock(currentUsername)
      removeThemeBackgroundLock(currentUsername)
      containerRef.value?.removeAttribute('data-theme-color-active')
      containerRef.value?.removeAttribute('data-theme-bg-active')
    }
    restoreBody()
  })

  return { theme, loading, load, applyPreview }
}

export function getThemeCssVars(theme: UserTheme) {
  return buildCssVars(theme)
}
