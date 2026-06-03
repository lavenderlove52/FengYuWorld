<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NSwitch, NUpload, useMessage } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'
import { getUserTheme, updateMyTheme } from '@/api/theme'
import { uploadImage } from '@/api/uploads'
import { PRESETS, FONT_OPTIONS, DEFAULT_THEME } from '@/types/theme'
import type { UserTheme, ThemePreset, ThemeLayout } from '@/types/theme'

const auth = useAuthStore()
const message = useMessage()
const router = useRouter()

const myUrl = computed(() =>
  auth.user ? router.resolve({ name: 'user-home', params: { username: auth.user.username } }).href : '/'
)

const form = ref<UserTheme>({ ...DEFAULT_THEME })
const saving = ref(false)
const uploadingBanner = ref(false)

const selectedPreset = computed(() => PRESETS.find((p) => p.id === form.value.preset) ?? PRESETS[0])

const previewStyle = computed(() => {
  const p = selectedPreset.value
  return {
    '--theme-bg': p.bg,
    '--theme-accent': form.value.accent_color || p.accent,
    '--theme-text': p.text,
    '--theme-border': p.border,
    '--theme-card-bg': p.cardBg,
    backgroundColor: p.bg,
    color: p.text,
  }
})

onMounted(async () => {
  if (!auth.user?.username) return
  try {
    const res = await getUserTheme(auth.user.username)
    if (res.code === 0 && res.data) {
      form.value = { ...DEFAULT_THEME, ...res.data }
    }
  } catch {
    // 首次使用，使用默认值
  }
})

const selectPreset = (preset: ThemePreset) => {
  form.value.preset = preset
  // 切换预设时清空自定义强调色，使用预设默认色
  form.value.accent_color = ''
}

const selectLayout = (layout: ThemeLayout) => {
  form.value.layout = layout
}

const onBannerUpload = async ({ file, onFinish, onError }: any) => {
  uploadingBanner.value = true
  try {
    const res = await uploadImage(file.file as File)
    if (res.code === 0 && res.data) {
      form.value.banner_url = res.data.url
      onFinish()
    } else {
      onError()
    }
  } catch {
    onError()
  } finally {
    uploadingBanner.value = false
  }
}

const save = async () => {
  saving.value = true
  try {
    const res = await updateMyTheme({
      preset: form.value.preset,
      accent_color: form.value.accent_color,
      font_family: form.value.font_family,
      banner_url: form.value.banner_url,
      layout: form.value.layout,
      show_pins_public: form.value.show_pins_public,
      custom_css: form.value.custom_css,
      custom_js: form.value.custom_js,
    })
    if (res.code === 0) {
      message.success('主题已保存，刷新个人主页即可看到效果')
    } else {
      message.error(res.message || '保存失败')
    }
  } catch (e: any) {
    message.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

const LAYOUTS = [
  { id: 'grid', label: '4列网格', icon: '▦' },
  { id: 'list', label: '图文列表', icon: '≡' },
  { id: 'magazine', label: '杂志双栏', icon: '⊟' },
] as const
</script>

<template>
  <div>
    <div class="flex items-start justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">主页定制</h1>
        <p class="text-sm text-slate-400 mt-1">自定义你的个人主页和文章详情页风格</p>
      </div>
      <div class="flex items-center gap-3">
        <a
          :href="myUrl"
          target="_blank"
          class="text-sm text-slate-500 hover:text-indigo-500 transition-colors flex items-center gap-1"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="w-3.5 h-3.5">
            <path d="M6.22 8.72a.75.75 0 0 0 1.06 1.06l5.22-5.22v1.69a.75.75 0 0 0 1.5 0v-3.5a.75.75 0 0 0-.75-.75h-3.5a.75.75 0 0 0 0 1.5h1.69L6.22 8.72Z" />
            <path d="M3.5 6.75c0-.69.56-1.25 1.25-1.25H7A.75.75 0 0 0 7 4H4.75A2.75 2.75 0 0 0 2 6.75v4.5A2.75 2.75 0 0 0 4.75 14h4.5A2.75 2.75 0 0 0 12 11.25V9a.75.75 0 0 0-1.5 0v2.25c0 .69-.56 1.25-1.25 1.25h-4.5c-.69 0-1.25-.56-1.25-1.25v-4.5Z" />
          </svg>
          查看主页
        </a>
        <NButton type="primary" :loading="saving" @click="save">保存</NButton>
      </div>
    </div>

    <div class="grid grid-cols-5 gap-8">
      <!-- 左侧设置区 -->
      <div class="col-span-3 space-y-8">

        <!-- 预设主题 -->
        <section>
          <h2 class="text-base font-semibold text-slate-700 mb-4">预设主题</h2>
          <div class="grid grid-cols-3 gap-3">
            <button
              v-for="preset in PRESETS"
              :key="preset.id"
              class="relative rounded-xl overflow-hidden border-2 transition-all text-left"
              :class="form.preset === preset.id ? 'border-indigo-500 ring-2 ring-indigo-200' : 'border-slate-200 hover:border-slate-400'"
              @click="selectPreset(preset.id)"
            >
              <!-- 主题缩略图 -->
              <div class="h-20 p-3 flex flex-col justify-between" :style="{ backgroundColor: preset.bg }">
                <div class="flex gap-1">
                  <div class="w-6 h-6 rounded-full" :style="{ backgroundColor: preset.accent }" />
                  <div class="flex-1 flex flex-col gap-1 ml-1">
                    <div class="h-2 rounded" :style="{ backgroundColor: preset.text, opacity: 0.8, width: '70%' }" />
                    <div class="h-1.5 rounded" :style="{ backgroundColor: preset.text, opacity: 0.4, width: '90%' }" />
                    <div class="h-1.5 rounded" :style="{ backgroundColor: preset.text, opacity: 0.3, width: '60%' }" />
                  </div>
                </div>
                <div class="flex gap-1">
                  <div v-for="i in 3" :key="i" class="flex-1 h-1 rounded" :style="{ backgroundColor: preset.border }" />
                </div>
              </div>
              <div class="px-3 py-2" :style="{ backgroundColor: preset.cardBg }">
                <div class="text-xs font-semibold" :style="{ color: preset.text }">{{ preset.label }}</div>
                <div class="text-[10px] mt-0.5" :style="{ color: preset.text, opacity: 0.5 }">{{ preset.desc }}</div>
              </div>
              <!-- 选中标记 -->
              <div v-if="form.preset === preset.id" class="absolute top-1.5 right-1.5 w-4 h-4 rounded-full bg-indigo-500 flex items-center justify-center">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 12 12" fill="white" class="w-2.5 h-2.5">
                  <path fill-rule="evenodd" d="M10.22 2.47a.75.75 0 0 1 .06 1.06l-5.5 6a.75.75 0 0 1-1.06.06L1.22 6.84a.75.75 0 1 1 1.06-1.06l2 1.94L9.16 2.53a.75.75 0 0 1 1.06-.06Z" />
                </svg>
              </div>
            </button>
          </div>
        </section>

        <!-- 强调色 -->
        <section>
          <h2 class="text-base font-semibold text-slate-700 mb-3">强调色
            <span class="text-xs font-normal text-slate-400 ml-2">覆盖预设默认色，留空则用预设色</span>
          </h2>
          <div class="flex items-center gap-3">
            <input
              type="color"
              :value="form.accent_color || selectedPreset.accent"
              class="w-10 h-10 rounded-lg border border-slate-200 cursor-pointer p-0.5"
              @input="(e) => form.accent_color = (e.target as HTMLInputElement).value"
            />
            <input
              v-model="form.accent_color"
              type="text"
              placeholder="#6366f1"
              class="flex-1 px-3 py-2 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-slate-400"
            />
            <button
              v-if="form.accent_color"
              class="text-xs text-slate-400 hover:text-slate-600 transition-colors"
              @click="form.accent_color = ''"
            >重置</button>
          </div>
        </section>

        <!-- 字体 -->
        <section>
          <h2 class="text-base font-semibold text-slate-700 mb-3">字体</h2>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="font in FONT_OPTIONS"
              :key="font.value"
              class="px-3 py-2 text-sm rounded-lg border transition-all"
              :class="form.font_family === font.value
                ? 'border-indigo-400 bg-indigo-50 text-indigo-600'
                : 'border-slate-200 text-slate-600 hover:border-slate-400'"
              :style="font.value ? { fontFamily: `'${font.value}', serif` } : {}"
              @click="form.font_family = font.value"
            >{{ font.label }}</button>
          </div>
        </section>

        <!-- Banner 图 -->
        <section>
          <h2 class="text-base font-semibold text-slate-700 mb-3">Banner 图
            <span class="text-xs font-normal text-slate-400 ml-2">显示在个人主页顶部</span>
          </h2>
          <div v-if="form.banner_url" class="mb-3 rounded-lg overflow-hidden h-28 bg-slate-100">
            <img :src="form.banner_url" class="w-full h-full object-cover" />
          </div>
          <div class="flex gap-3">
            <NUpload :show-file-list="false" accept="image/*" :custom-request="onBannerUpload">
              <NButton size="small" :loading="uploadingBanner">上传图片</NButton>
            </NUpload>
            <input
              v-model="form.banner_url"
              type="text"
              placeholder="或粘贴图片 URL"
              class="flex-1 px-3 py-1.5 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-slate-400"
            />
            <button
              v-if="form.banner_url"
              class="text-xs text-slate-400 hover:text-red-400 transition-colors"
              @click="form.banner_url = ''"
            >移除</button>
          </div>
        </section>

        <!-- 布局 -->
        <section>
          <h2 class="text-base font-semibold text-slate-700 mb-3">文章列表布局</h2>
          <div class="flex gap-3">
            <button
              v-for="layout in LAYOUTS"
              :key="layout.id"
              class="flex items-center gap-2 px-4 py-2.5 rounded-lg border text-sm transition-all"
              :class="form.layout === layout.id
                ? 'border-indigo-400 bg-indigo-50 text-indigo-600'
                : 'border-slate-200 text-slate-600 hover:border-slate-400'"
              @click="selectLayout(layout.id)"
            >
              <span class="text-base">{{ layout.icon }}</span>
              {{ layout.label }}
            </button>
          </div>
        </section>

        <!-- 置顶对外显示 -->
        <section>
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-base font-semibold text-slate-700">对所有人展示置顶</h2>
              <p class="text-xs text-slate-400 mt-0.5">开启后访客也能看到置顶文章轮播</p>
            </div>
            <NSwitch v-model:value="form.show_pins_public" />
          </div>
        </section>

        <!-- 自定义 CSS -->
        <section>
          <div class="flex items-baseline gap-2 mb-3">
            <h2 class="text-base font-semibold text-slate-700">自定义 CSS</h2>
            <span class="text-xs text-slate-400">高级 · 最多 10000 字符</span>
          </div>
          <div class="relative">
            <textarea
              v-model="form.custom_css"
              rows="10"
              placeholder=".user-card { border-radius: 16px; }&#10;h1 { letter-spacing: -1px; }"
              class="w-full px-4 py-3 text-sm font-mono border border-slate-200 rounded-xl focus:outline-none focus:border-slate-400 resize-y bg-slate-50"
              :maxlength="10000"
            />
            <span class="absolute bottom-3 right-3 text-xs text-slate-300">{{ form.custom_css.length }} / 10000</span>
          </div>
          <p class="text-xs text-slate-400 mt-2">CSS 自动限定在个人主页/文章页容器内。可直接写 <code class="bg-slate-100 px-1 rounded">[data-theme-user] &gt; div:first-child</code> 或 <code class="bg-slate-100 px-1 rounded">.min-h-screen</code> 等选择器</p>
        </section>

        <!-- 自定义 JS（需权限） -->
        <section>
          <div class="flex items-baseline gap-2 mb-3">
            <h2 class="text-base font-semibold text-slate-700">自定义 JS</h2>
            <span class="text-xs text-slate-400">高级 · 需要管理员授权 · 最大 10000 字符</span>
          </div>
          <div v-if="auth.user?.role === 'admin' || auth.user?.can_custom_js" class="relative">
            <textarea
              v-model="form.custom_js"
              rows="8"
              placeholder="console.log('hello from my page')"
              class="w-full px-4 py-3 text-sm font-mono border border-slate-200 rounded-xl focus:outline-none focus:border-slate-400 resize-y bg-slate-50"
              :maxlength="10000"
            />
            <span class="absolute bottom-3 right-3 text-xs text-slate-300">{{ form.custom_js.length }} / 10000</span>
          </div>
          <div v-else class="flex items-start gap-3 p-4 bg-slate-50 border border-slate-200 rounded-xl">
            <div>
              <div class="text-sm font-medium text-slate-600">自定义 JS 需要管理员授权</div>
              <p class="text-xs text-slate-400 mt-1">可在账号设置中复制申请文案，管理员授权后这里会显示 JS 编辑框。</p>
            </div>
          </div>
        </section>

      </div>

      <!-- 右侧预览 -->
      <div class="col-span-2 sticky top-6 self-start">
        <h2 class="text-base font-semibold text-slate-700 mb-4">实时预览</h2>
        <div
          class="rounded-2xl overflow-hidden border"
          :style="{ borderColor: selectedPreset.border }"
        >
          <!-- 模拟导航 -->
          <div class="px-4 py-2.5 flex items-center gap-2 border-b" :style="{ backgroundColor: selectedPreset.bg, borderColor: selectedPreset.border }">
            <div class="w-3 h-3 rounded-full bg-red-400" />
            <div class="w-3 h-3 rounded-full bg-yellow-400" />
            <div class="w-3 h-3 rounded-full bg-green-400" />
            <span class="text-xs ml-2" :style="{ color: selectedPreset.text, opacity: 0.4 }">
              fengyu.blog/u/{{ auth.user?.username }}
            </span>
          </div>

          <!-- 模拟页面 -->
          <div class="p-5" :style="previewStyle">
            <!-- Banner -->
            <div v-if="form.banner_url" class="h-16 rounded-lg overflow-hidden mb-4">
              <img :src="form.banner_url" class="w-full h-full object-cover" />
            </div>
            <div v-else-if="!form.banner_url" class="h-3 rounded mb-4" :style="{ background: selectedPreset.accent, opacity: 0.15 }" />

            <!-- 用户头 -->
            <div class="flex items-center gap-3 mb-5 pb-4" :style="{ borderBottom: `1px solid ${selectedPreset.border}` }">
              <div class="w-10 h-10 rounded-full flex items-center justify-center text-sm font-bold text-white" :style="{ backgroundColor: form.accent_color || selectedPreset.accent }">
                {{ auth.user?.nickname?.[0]?.toUpperCase() || 'U' }}
              </div>
              <div>
                <div class="text-sm font-bold" :style="{ color: selectedPreset.text }">{{ auth.user?.nickname }}</div>
                <div class="text-xs" :style="{ color: selectedPreset.text, opacity: 0.5 }">@{{ auth.user?.username }}</div>
              </div>
            </div>

            <!-- 模拟文章卡片 -->
            <div v-if="form.layout === 'grid'" class="grid grid-cols-2 gap-2">
              <div v-for="i in 4" :key="i" class="rounded-lg p-2.5" :style="{ background: selectedPreset.cardBg, border: `1px solid ${selectedPreset.border}` }">
                <div class="h-1.5 rounded mb-1.5 w-3/4" :style="{ backgroundColor: selectedPreset.text, opacity: 0.7 }" />
                <div class="h-1 rounded mb-1 w-full" :style="{ backgroundColor: selectedPreset.text, opacity: 0.3 }" />
                <div class="h-1 rounded w-2/3" :style="{ backgroundColor: selectedPreset.text, opacity: 0.2 }" />
              </div>
            </div>
            <div v-else-if="form.layout === 'list'" class="space-y-2">
              <div v-for="i in 3" :key="i" class="flex gap-2 items-center rounded-lg p-2" :style="{ background: selectedPreset.cardBg }">
                <div class="w-12 h-8 rounded flex-shrink-0" :style="{ backgroundColor: selectedPreset.border }" />
                <div class="flex-1">
                  <div class="h-1.5 rounded mb-1 w-3/4" :style="{ backgroundColor: selectedPreset.text, opacity: 0.7 }" />
                  <div class="h-1 rounded w-1/2" :style="{ backgroundColor: selectedPreset.text, opacity: 0.3 }" />
                </div>
              </div>
            </div>
            <div v-else class="grid grid-cols-2 gap-2">
              <div v-for="i in 2" :key="i" class="rounded-lg overflow-hidden" :style="{ background: selectedPreset.cardBg, border: `1px solid ${selectedPreset.border}` }">
                <div class="h-8" :style="{ backgroundColor: selectedPreset.border }" />
                <div class="p-2">
                  <div class="h-1.5 rounded mb-1" :style="{ backgroundColor: selectedPreset.text, opacity: 0.7 }" />
                  <div class="h-1 rounded w-2/3" :style="{ backgroundColor: selectedPreset.text, opacity: 0.3 }" />
                </div>
              </div>
            </div>

            <!-- 强调色标注 -->
            <div class="mt-4 flex items-center gap-2">
              <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: form.accent_color || selectedPreset.accent }" />
              <span class="text-xs" :style="{ color: form.accent_color || selectedPreset.accent }">
                {{ form.accent_color || selectedPreset.accent }}
              </span>
              <span class="text-xs" :style="{ color: selectedPreset.text, opacity: 0.4 }">
                — {{ FONT_OPTIONS.find(f => f.value === form.font_family)?.label || '全局默认字体' }}
              </span>
            </div>
          </div>
        </div>

        <p class="text-xs text-slate-400 mt-3 text-center">保存后在个人主页查看完整效果</p>
      </div>
    </div>
  </div>
</template>
