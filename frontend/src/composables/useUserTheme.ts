import { ref, onUnmounted } from 'vue'
import { getUserTheme } from '@/api/theme'
import { PRESETS, DEFAULT_THEME } from '@/types/theme'
import { setActivePageTheme } from '@/stores/pageTheme'
import type { UserTheme } from '@/types/theme'

const STYLE_ID_PREFIX = 'user-theme-css-'

function getPresetConfig(preset: string) {
  return PRESETS.find((p) => p.id === preset) ?? PRESETS[0]
}

function buildCssVars(theme: UserTheme): Record<string, string> {
  const p = getPresetConfig(theme.preset)
  return {
    '--theme-bg': p.bg,
    '--theme-accent': theme.accent_color || p.accent,
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
  const scoped = css
    .split('}')
    .filter((block) => block.trim())
    .map((block) => {
      const [selector, ...rest] = block.split('{')
      if (!rest.length) return ''
      const selectors = selector
        .split(',')
        .map((s) => `[data-theme-user="${username}"] ${s.trim()}`)
        .join(', ')
      return `${selectors} { ${rest.join('{')} }`
    })
    .join('\n')
  el.textContent = scoped
}

function removeCustomCss(username: string) {
  document.getElementById(STYLE_ID_PREFIX + username)?.remove()
}

const JS_SCRIPT_ID_PREFIX = 'user-theme-js-'

function injectCustomJS(username: string, js: string) {
  removeCustomJS(username)
  if (!js.trim()) return

  // 包装：拦截 canvas 创建，记录以便导航后清理；同时提供停止 RAF 的钩子
  const cleanupKey = `__theme_cleanup_${username.replace(/[^a-z0-9]/gi, '_')}`
  const wrapper = `
(function(){
  var _canvases = [];
  var _origCE = document.createElement.bind(document);
  document.createElement = function(tag) {
    var el = _origCE(tag);
    if (typeof tag === 'string' && tag.toLowerCase() === 'canvas') _canvases.push(el);
    return el;
  };
  var _rafIds = [];
  var _origRAF = window.requestAnimationFrame.bind(window);
  window.requestAnimationFrame = function(cb) {
    var id = _origRAF(cb); _rafIds.push(id); return id;
  };
  try { ${js} } catch(e) { console.warn('[theme-js]', e); }
  document.createElement = _origCE;
  window.requestAnimationFrame = _origRAF;
  window['${cleanupKey}'] = function() {
    _canvases.forEach(function(c) { try { c.remove(); } catch(e){} });
    _rafIds.forEach(function(id) { try { cancelAnimationFrame(id); } catch(e){} });
    delete window['${cleanupKey}'];
  };
})();
`
  const script = document.createElement('script')
  script.id = JS_SCRIPT_ID_PREFIX + username
  script.textContent = wrapper
  document.body.appendChild(script)
}

function removeCustomJS(username: string) {
  const cleanupKey = `__theme_cleanup_${username.replace(/[^a-z0-9]/gi, '_')}`
  const cleanup = (window as any)[cleanupKey]
  if (cleanup) {
    try { cleanup() } catch {}
  }
  // 执行用户代码自注册的清理函数
  const bgCleanup = (window as any).__bgCleanup
  if (Array.isArray(bgCleanup)) {
    bgCleanup.forEach((fn: () => void) => { try { fn() } catch {} })
    delete (window as any).__bgCleanup
  }
  const el = document.getElementById(JS_SCRIPT_ID_PREFIX + username)
  el?.remove()
}

export function useUserTheme(containerRef: { value: HTMLElement | null }) {
  const theme = ref<UserTheme>({ ...DEFAULT_THEME })
  const loading = ref(false)
  let currentUsername = ''
  let originalBodyBg = ''
  let originalBodyColor = ''

  const applyToBody = (t: UserTheme) => {
    const p = getPresetConfig(t.preset)
    originalBodyBg = document.body.style.backgroundColor
    originalBodyColor = document.body.style.color
    document.body.style.backgroundColor = p.bg
    document.body.style.color = p.text
    if (t.font_family) {
      document.body.style.fontFamily = `'${t.font_family}', serif`
    }
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
    currentUsername = username
    loading.value = true
    try {
      const res = await getUserTheme(username)
      if (res.code === 0 && res.data) {
        theme.value = res.data
      } else {
        theme.value = { ...DEFAULT_THEME }
      }
    } catch {
      theme.value = { ...DEFAULT_THEME }
    } finally {
      loading.value = false
      // 用 try/catch 包裹所有 DOM 操作：若任意步骤抛出，不应影响页面内容显示
      try { applyToBody(theme.value) } catch (e) { console.warn('[useUserTheme] applyToBody failed', e) }
      try { applyToContainer(theme.value) } catch (e) { console.warn('[useUserTheme] applyToContainer failed', e) }
      if (theme.value.custom_css) {
        try { injectCustomCss(username, theme.value.custom_css) } catch (e) { console.warn('[useUserTheme] injectCustomCss failed', e) }
      }
      if (theme.value.custom_js) {
        try { injectCustomJS(username, theme.value.custom_js) } catch (e) { console.warn('[useUserTheme] injectCustomJS failed', e) }
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
      removeCustomJS(currentUsername)
    }
    restoreBody()
  })

  return { theme, loading, load, applyPreview }
}

// 用于在设置页获取预设对应的 CSS 变量（无需挂载 DOM）
export function getThemeCssVars(theme: UserTheme) {
  return buildCssVars(theme)
}
