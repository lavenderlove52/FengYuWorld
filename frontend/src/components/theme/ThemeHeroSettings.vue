<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { NButton, NUpload, useMessage } from 'naive-ui'
import {
  deleteMyThemeImage,
  listMyThemeImages,
  uploadMyThemeImage,
  type ThemeImageList,
} from '@/api/theme'
import {
  minHeroIntervalMs,
  type ThemeHeroConfig,
  type ThemeHeroTransition,
  type ThemeImage,
  type ThemeLayout,
} from '@/types/theme'

const config = defineModel<ThemeHeroConfig>({ required: true })

const message = useMessage()
const images = ref<ThemeImage[]>([])
const loading = ref(false)
const uploading = ref(false)
const deletingId = ref<number | null>(null)

const MAX_LIBRARY_IMAGES = 10
const MAX_HERO_IMAGES = 5

const transitions: Array<{ value: ThemeHeroTransition; label: string; description: string }> = [
  { value: 'off', label: '直接切换', description: '不播放过渡动画' },
  { value: 'fade', label: '淡入淡出', description: '柔和地交叉切换图片' },
  { value: 'slices', label: '切片', description: '纵向切片依次显现' },
  { value: 'interlaced', label: '交错切片', description: '两张图片按奇偶列上下交错切换' },
]
const articleLayouts: Array<{ value: ThemeLayout; label: string }> = [
  { value: 'list', label: '图文列表' },
  { value: 'grid', label: '四列网格' },
  { value: 'magazine', label: '杂志双栏' },
]

const selectedCount = computed(() => config.value.image_ids.length)
const libraryFull = computed(() => images.value.length >= MAX_LIBRARY_IMAGES)
const minimumInterval = computed(() => minHeroIntervalMs(config.value.transition))

watch(
  minimumInterval,
  (minimum) => {
    if (config.value.interval_ms < minimum) config.value.interval_ms = minimum
  },
  { immediate: true },
)

const normalizeList = (data?: ThemeImage[] | ThemeImageList) => {
  if (Array.isArray(data)) return data
  return data?.items ?? []
}

const loadImages = async () => {
  loading.value = true
  try {
    const res = await listMyThemeImages()
    if (res.code === 0) images.value = normalizeList(res.data)
  } catch (error: any) {
    message.error(error.message || '主页图库加载失败')
  } finally {
    loading.value = false
  }
}

const onUpload = async ({ file, onFinish, onError }: any) => {
  if (libraryFull.value) {
    message.warning('主页图库最多保存 10 张图片，请删除后再上传')
    onError()
    return
  }

  uploading.value = true
  try {
    const res = await uploadMyThemeImage(file.file as File)
    if (res.code === 0 && res.data) {
      const existingIndex = images.value.findIndex((item) => item.id === res.data!.id)
      if (existingIndex >= 0) images.value.splice(existingIndex, 1, res.data)
      else images.value.unshift(res.data)
      message.success('图片已加入主页图库')
      onFinish()
    } else {
      message.error(res.message || '上传失败')
      onError()
    }
  } catch (error: any) {
    message.error(error.message || '上传失败')
    onError()
  } finally {
    uploading.value = false
  }
}

const toggleImage = (id: number) => {
  const index = config.value.image_ids.indexOf(id)
  if (index >= 0) {
    config.value.image_ids.splice(index, 1)
    return
  }
  if (selectedCount.value >= MAX_HERO_IMAGES) {
    message.warning('全屏轮播最多选择 5 张图片')
    return
  }
  config.value.image_ids.push(id)
}

const removeImage = async (image: ThemeImage) => {
  if (!window.confirm('删除后服务器将立即移除这张主页图片，且无法恢复。确定删除吗？')) return

  deletingId.value = image.id
  try {
    const res = await deleteMyThemeImage(image.id)
    if (res.code === 0) {
      images.value = images.value.filter((item) => item.id !== image.id)
      config.value.image_ids = config.value.image_ids.filter((id) => id !== image.id)
      message.success('主页图片已删除')
    } else {
      message.error(res.message || '删除失败')
    }
  } catch (error: any) {
    message.error(error.message || '删除失败')
  } finally {
    deletingId.value = null
  }
}

onMounted(loadImages)
</script>

<template>
  <section class="space-y-5">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h2 class="text-base font-semibold text-slate-700">全屏主页大图</h2>
        <p class="text-xs text-slate-400 mt-1">
          使用独立主页图库，不会与文章/头像混用；新上传会同时生成清晰桌面版和最长边 1280px 的移动版。
        </p>
      </div>
      <label class="inline-flex items-center gap-2 text-sm text-slate-600 cursor-pointer">
        <input v-model="config.enabled" type="checkbox" class="accent-indigo-500" />
        启用全屏大图
      </label>
    </div>

    <div class="rounded-xl border border-slate-200 bg-slate-50/60 p-4 space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h3 class="text-sm font-medium text-slate-700">专用主页图库</h3>
          <p class="text-xs text-slate-400 mt-0.5">
            已保存 {{ images.length }} / {{ MAX_LIBRARY_IMAGES }} 张，已选轮播 {{ selectedCount }} / {{ MAX_HERO_IMAGES }} 张
          </p>
        </div>
        <NUpload
          :show-file-list="false"
          accept="image/*"
          :disabled="libraryFull"
          :custom-request="onUpload"
        >
          <NButton size="small" :loading="uploading" :disabled="libraryFull">
            {{ libraryFull ? '请先删除图片' : '上传主页图片' }}
          </NButton>
        </NUpload>
      </div>

      <div v-if="loading" class="py-8 text-center text-sm text-slate-400">图库加载中...</div>
      <div v-else-if="images.length" class="grid grid-cols-2 sm:grid-cols-3 gap-3">
        <div
          v-for="image in images"
          :key="image.id"
          class="group relative aspect-video overflow-hidden rounded-lg border-2 bg-slate-100 transition-colors"
          :class="config.image_ids.includes(image.id) ? 'border-indigo-500' : 'border-transparent'"
        >
          <button type="button" class="block h-full w-full" @click="toggleImage(image.id)">
            <img
              :src="image.mobile_url || image.url"
              alt=""
              loading="lazy"
              decoding="async"
              class="h-full w-full object-cover"
            />
            <span
              v-if="config.image_ids.includes(image.id)"
              class="absolute left-2 top-2 rounded-full bg-indigo-500 px-2 py-0.5 text-[11px] font-medium text-white"
            >
              第 {{ config.image_ids.indexOf(image.id) + 1 }} 张
            </span>
          </button>
          <button
            type="button"
            class="absolute bottom-2 right-2 rounded-md bg-black/65 px-2 py-1 text-xs text-white opacity-0 transition-opacity group-hover:opacity-100 focus:opacity-100"
            :disabled="deletingId === image.id"
            @click.stop="removeImage(image)"
          >
            {{ deletingId === image.id ? '删除中' : '删除' }}
          </button>
        </div>
      </div>
      <div v-else class="rounded-lg border border-dashed border-slate-300 py-8 text-center text-sm text-slate-400">
        暂无主页图片，上传后点击图片即可加入轮播。
      </div>
    </div>

    <div class="grid gap-4 sm:grid-cols-2">
      <label class="hero-field">
        <span>轮换方式</span>
        <select v-model="config.transition">
          <option v-for="item in transitions" :key="item.value" :value="item.value">
            {{ item.label }} - {{ item.description }}
          </option>
        </select>
      </label>

      <label class="hero-field">
        <span>轮换间隔 {{ Math.round(config.interval_ms / 1000) }} 秒</span>
        <input v-model.number="config.interval_ms" type="range" :min="minimumInterval" max="30000" step="500" />
        <small v-if="minimumInterval === 5000">切片动画最低间隔为 5 秒，以降低连续 GPU 合成压力。</small>
      </label>

      <label class="hero-field">
        <span>遮罩透明度 {{ Math.round(config.overlay_opacity * 100) }}%</span>
        <input v-model.number="config.overlay_opacity" type="range" min="0" max="0.9" step="0.05" />
      </label>

      <label class="hero-field">
        <span>大图高度 {{ config.hero_height_vh }}vh</span>
        <input v-model.number="config.hero_height_vh" type="range" min="60" max="100" step="5" />
      </label>

      <label class="hero-field">
        <span>个人主页文章内容宽度 {{ config.page_width_px }}px</span>
        <input v-model.number="config.page_width_px" type="range" min="600" max="1800" step="20" />
      </label>

      <label class="hero-field">
        <span>文章详情正文宽度 {{ config.article_list_width_px }}px</span>
        <input
          v-model.number="config.article_list_width_px"
          type="range"
          min="600"
          max="1600"
          step="20"
        />
      </label>

      <label class="hero-field">
        <span>大图下方文章布局</span>
        <select v-model="config.article_layout">
          <option v-for="item in articleLayouts" :key="item.value" :value="item.value">{{ item.label }}</option>
        </select>
      </label>
    </div>
  </section>
</template>

<style scoped>
.hero-field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  color: #64748b;
  font-size: 0.75rem;
}

.hero-field select {
  width: 100%;
  padding: 0.55rem 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  outline: none;
  background: white;
  color: #475569;
}

.hero-field input[type='range'] {
  width: 100%;
  accent-color: #6366f1;
}

.hero-field small {
  color: #94a3b8;
  line-height: 1.4;
}
</style>
