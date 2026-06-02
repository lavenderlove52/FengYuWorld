<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { NButton, NModal, NSlider } from 'naive-ui'

const props = defineProps<{
  show: boolean
  file: File | null
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  confirm: [blob: Blob]
}>()

const FRAME_SIZE = 280
const OUTPUT_SIZE = 200

const imageUrl = ref('')
const imageEl = ref<HTMLImageElement | null>(null)
const naturalWidth = ref(0)
const naturalHeight = ref(0)
const zoom = ref(1)
const offsetX = ref(0)
const offsetY = ref(0)
const loading = ref(false)

let dragStartX = 0
let dragStartY = 0
let dragOriginX = 0
let dragOriginY = 0

const hasImage = computed(() => !!imageEl.value && naturalWidth.value > 0 && naturalHeight.value > 0)
const baseScale = computed(() => {
  if (!naturalWidth.value || !naturalHeight.value) return 1
  return Math.max(FRAME_SIZE / naturalWidth.value, FRAME_SIZE / naturalHeight.value)
})
const displayScale = computed(() => baseScale.value * zoom.value)
const displayWidth = computed(() => naturalWidth.value * displayScale.value)
const displayHeight = computed(() => naturalHeight.value * displayScale.value)

const clampOffsets = () => {
  const maxOffsetX = Math.max(0, (displayWidth.value - FRAME_SIZE) / 2)
  const maxOffsetY = Math.max(0, (displayHeight.value - FRAME_SIZE) / 2)
  offsetX.value = Math.min(maxOffsetX, Math.max(-maxOffsetX, offsetX.value))
  offsetY.value = Math.min(maxOffsetY, Math.max(-maxOffsetY, offsetY.value))
}

const resetState = () => {
  zoom.value = 1
  offsetX.value = 0
  offsetY.value = 0
  naturalWidth.value = 0
  naturalHeight.value = 0
  imageEl.value = null
  loading.value = false
}

const revokePreview = () => {
  if (imageUrl.value) {
    URL.revokeObjectURL(imageUrl.value)
    imageUrl.value = ''
  }
}

const loadFile = async (file: File | null) => {
  revokePreview()
  resetState()
  if (!file) return

  loading.value = true
  imageUrl.value = URL.createObjectURL(file)

  const img = new Image()
  img.src = imageUrl.value
  await new Promise<void>((resolve, reject) => {
    img.onload = () => resolve()
    img.onerror = () => reject(new Error('load image failed'))
  })

  imageEl.value = img
  naturalWidth.value = img.naturalWidth
  naturalHeight.value = img.naturalHeight
  clampOffsets()
  loading.value = false
}

watch(
  () => [props.file, props.show] as const,
  async ([file, show]) => {
    if (!show) {
      revokePreview()
      resetState()
      return
    }
    try {
      await loadFile(file)
    } catch {
      revokePreview()
      resetState()
    }
  },
  { immediate: true },
)

watch(zoom, () => {
  clampOffsets()
})

const imageStyle = computed(() => ({
  width: `${displayWidth.value}px`,
  height: `${displayHeight.value}px`,
  left: `${(FRAME_SIZE - displayWidth.value) / 2 + offsetX.value}px`,
  top: `${(FRAME_SIZE - displayHeight.value) / 2 + offsetY.value}px`,
}))

const stopDragging = () => {
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('pointerup', stopDragging)
}

const onPointerMove = (event: PointerEvent) => {
  offsetX.value = dragOriginX + (event.clientX - dragStartX)
  offsetY.value = dragOriginY + (event.clientY - dragStartY)
  clampOffsets()
}

const onPointerDown = (event: PointerEvent) => {
  if (!hasImage.value) return
  dragStartX = event.clientX
  dragStartY = event.clientY
  dragOriginX = offsetX.value
  dragOriginY = offsetY.value
  window.addEventListener('pointermove', onPointerMove)
  window.addEventListener('pointerup', stopDragging)
}

const close = () => {
  stopDragging()
  emit('update:show', false)
}

const confirmCrop = () => {
  if (!imageEl.value) return

  const canvas = document.createElement('canvas')
  canvas.width = OUTPUT_SIZE
  canvas.height = OUTPUT_SIZE
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const left = (FRAME_SIZE - displayWidth.value) / 2 + offsetX.value
  const top = (FRAME_SIZE - displayHeight.value) / 2 + offsetY.value
  const sourceX = (0 - left) / displayScale.value
  const sourceY = (0 - top) / displayScale.value
  const sourceSize = FRAME_SIZE / displayScale.value

  ctx.drawImage(
    imageEl.value,
    sourceX,
    sourceY,
    sourceSize,
    sourceSize,
    0,
    0,
    OUTPUT_SIZE,
    OUTPUT_SIZE,
  )

  canvas.toBlob((blob) => {
    if (!blob) return
    emit('confirm', blob)
  }, 'image/jpeg', 0.92)
}

onBeforeUnmount(() => {
  stopDragging()
  revokePreview()
})
</script>

<template>
  <NModal
    :show="show"
    preset="card"
    title="裁剪头像"
    style="max-width: 560px"
    @update:show="(value) => emit('update:show', value)"
  >
    <div class="space-y-5">
      <div class="text-sm text-slate-500">
        头像会保存为固定的 `200 × 200` 尺寸。拖动图片调整位置，滑动滑块进行放大缩小。
      </div>

      <div class="flex flex-col items-center gap-4">
        <div
          class="relative overflow-hidden rounded-2xl border border-slate-200 bg-slate-50 touch-none"
          :style="{ width: `${FRAME_SIZE}px`, height: `${FRAME_SIZE}px` }"
          @pointerdown.prevent="onPointerDown"
        >
          <img
            v-if="imageUrl && hasImage"
            :src="imageUrl"
            alt="头像裁剪预览"
            class="absolute select-none max-w-none"
            draggable="false"
            :style="imageStyle"
          />

          <div
            class="absolute inset-0 rounded-2xl ring-1 ring-inset ring-white/70 shadow-[inset_0_0_0_9999px_rgba(15,23,42,0.16)] pointer-events-none"
          />
          <div class="absolute inset-0 pointer-events-none grid grid-cols-3 grid-rows-3">
            <div v-for="i in 9" :key="i" class="border border-white/20" />
          </div>

          <div v-if="loading" class="absolute inset-0 flex items-center justify-center text-sm text-slate-400 bg-white/70">
            加载中...
          </div>
        </div>

        <div class="w-full max-w-sm">
          <div class="flex items-center justify-between text-xs text-slate-400 mb-2">
            <span>缩小</span>
            <span>放大</span>
          </div>
          <NSlider v-model:value="zoom" :min="1" :max="3" :step="0.01" />
        </div>
      </div>

      <div class="flex justify-end gap-2">
        <NButton @click="close">取消</NButton>
        <NButton type="primary" :disabled="!hasImage" @click="confirmCrop">确定裁剪</NButton>
      </div>
    </div>
  </NModal>
</template>
