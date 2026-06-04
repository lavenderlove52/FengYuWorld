<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import type { PublicUserInfo } from '@/types/auth'
import type { ThemeHeroConfig, ThemeImage } from '@/types/theme'

const props = defineProps<{
  config: ThemeHeroConfig
  images: ThemeImage[]
  user: PublicUserInfo
}>()

defineEmits<{ scrollDown: [] }>()

const SLICE_COUNT = 8
const SLICE_STAGGER_MS = 70
const SLICE_ANIM_MS = 700

const activeIndex = ref(0)
const sliceCycle = ref(0)
const sliceAnimating = ref(false)
const reduceMotion = ref(false)
let timer: number | null = null
let sliceEndTimer: number | null = null
let motionQuery: MediaQueryList | null = null

const currentImage = computed(() => props.images[activeIndex.value] ?? props.images[0])
const transition = computed(() => reduceMotion.value ? 'off' : props.config.transition)
const overlayOpacity = computed(() => Math.min(0.85, Math.max(0, props.config.overlay_opacity)))
const heroStyle = computed(() => ({
  minHeight: `calc(${Math.min(100, Math.max(60, props.config.hero_height_vh))}svh - 3.5rem)`,
}))

const sliceAnimDelay = (index: number) =>
  sliceAnimating.value ? `${(index - 1) * SLICE_STAGGER_MS}ms` : '0ms'

const sliceImgLeft = (index: number) => {
  const i = index - 1
  return i === 0 ? '0' : `calc(-${i * 100}% - ${i}px)`
}

const stopTimer = () => {
  if (timer !== null) {
    window.clearInterval(timer)
    timer = null
  }
}

const clearSliceEndTimer = () => {
  if (sliceEndTimer !== null) {
    window.clearTimeout(sliceEndTimer)
    sliceEndTimer = null
  }
}

const advance = () => {
  if (props.images.length < 2 || sliceAnimating.value) return
  activeIndex.value = (activeIndex.value + 1) % props.images.length
  sliceCycle.value += 1
  sliceAnimating.value = true
  clearSliceEndTimer()
  const duration = SLICE_ANIM_MS + (SLICE_COUNT - 1) * SLICE_STAGGER_MS + 80
  sliceEndTimer = window.setTimeout(() => {
    sliceAnimating.value = false
    sliceEndTimer = null
  }, duration)
}

const startTimer = () => {
  stopTimer()
  if (!props.config.enabled || props.images.length < 2 || reduceMotion.value) return
  const interval = Math.min(30000, Math.max(3000, props.config.interval_ms))
  timer = window.setInterval(advance, interval)
}

const onMotionChange = (event: MediaQueryListEvent) => {
  reduceMotion.value = event.matches
  startTimer()
}

const onVisibilityChange = () => {
  if (document.hidden) stopTimer()
  else startTimer()
}

watch(
  () => props.images.map((image) => image.id).join(','),
  () => {
    activeIndex.value = 0
    sliceCycle.value = 0
    sliceAnimating.value = false
    clearSliceEndTimer()
    startTimer()
  },
)
watch(() => [props.config.enabled, props.config.interval_ms, props.config.transition], startTimer)

onMounted(() => {
  motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  reduceMotion.value = motionQuery.matches
  motionQuery.addEventListener('change', onMotionChange)
  document.addEventListener('visibilitychange', onVisibilityChange)
  startTimer()
})

onUnmounted(() => {
  stopTimer()
  clearSliceEndTimer()
  motionQuery?.removeEventListener('change', onMotionChange)
  document.removeEventListener('visibilitychange', onVisibilityChange)
})
</script>

<template>
  <section class="home-hero" :style="heroStyle" aria-label="主页封面">
    <div class="absolute inset-0 bg-slate-900">
      <template v-if="transition === 'fade'">
        <Transition name="hero-fade">
          <img
            :key="currentImage.id"
            :src="currentImage.url"
            alt=""
            class="absolute inset-0 h-full w-full object-cover"
          />
        </Transition>
      </template>

      <img
        v-else-if="transition === 'off'"
        :src="currentImage.url"
        alt=""
        class="absolute inset-0 h-full w-full object-cover"
      />

      <template v-else>
        <img
          v-if="!sliceAnimating"
          :key="currentImage.id"
          :src="currentImage.url"
          alt=""
          class="absolute inset-0 h-full w-full object-cover"
        />
        <div v-else class="absolute inset-0 flex" aria-hidden="true">
          <div
            v-for="index in SLICE_COUNT"
            :key="`${sliceCycle}-${index}`"
            class="hero-slice-clip"
            :class="{ 'hero-slice-clip--reveal': sliceAnimating, 'hero-slice-clip--overlap': index > 1 }"
            :style="{ animationDelay: sliceAnimDelay(index) }"
          >
            <img
              :src="currentImage.url"
              alt=""
              class="hero-slice-img"
              :style="{ left: sliceImgLeft(index) }"
            />
          </div>
        </div>
      </template>
    </div>

    <div class="absolute inset-0 bg-black" :style="{ opacity: overlayOpacity }" />

    <div class="absolute inset-0 z-[1] flex flex-col items-center justify-center px-5 py-16 text-center text-white">
      <div class="mb-5 h-20 w-20 overflow-hidden rounded-full border-2 border-white/70 bg-white/15 shadow-xl backdrop-blur-sm">
        <img v-if="user.avatar_url" :src="user.avatar_url" :alt="user.nickname" class="h-full w-full object-cover" />
        <div v-else class="flex h-full w-full items-center justify-center text-3xl font-bold">
          {{ user.nickname?.[0]?.toUpperCase() }}
        </div>
      </div>
      <h1 class="text-4xl font-bold tracking-tight drop-shadow-lg sm:text-6xl">{{ user.nickname }}</h1>
      <p class="mt-3 text-sm text-white/75 sm:text-base">@{{ user.username }}</p>
      <p v-if="user.bio" class="mt-5 max-w-2xl text-sm leading-7 text-white/90 drop-shadow sm:text-base">
        {{ user.bio }}
      </p>
    </div>

    <button
      type="button"
      class="hero-down absolute bottom-6 left-1/2 z-[2] rounded-full p-3 text-white/80 transition-colors hover:text-white"
      aria-label="向下查看文章"
      @click="$emit('scrollDown')"
    >
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" class="h-7 w-7" stroke-width="1.8">
        <path d="m5 9 7 7 7-7" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </button>
  </section>
</template>

<style scoped>
.home-hero {
  position: relative;
  width: 100vw;
  min-height: calc(100svh - 3.5rem);
  margin-top: -1.5rem;
  margin-left: calc(50% - 50vw);
  overflow: hidden;
  background: #0f172a;
}

.hero-fade-enter-active,
.hero-fade-leave-active {
  transition: opacity 1s ease;
}

.hero-fade-enter-from,
.hero-fade-leave-to {
  opacity: 0;
}

.hero-slice-clip {
  flex: 1 1 0;
  min-width: 0;
  height: 100%;
  overflow: hidden;
  position: relative;
  opacity: 1;
  clip-path: inset(0 0 0 0);
}

.hero-slice-clip--overlap {
  margin-left: -1px;
}

.hero-slice-img {
  position: absolute;
  top: 0;
  height: 100%;
  width: calc(800% + 7px);
  max-width: none;
  object-fit: cover;
  object-position: center;
  pointer-events: none;
  user-select: none;
}

.hero-slice-clip--reveal {
  opacity: 0;
  clip-path: inset(100% 0 0 0);
  animation: reveal-slice 700ms ease-out forwards;
}

.hero-down {
  transform: translateX(-50%);
  animation: hero-bounce 1.8s ease-in-out infinite;
}

@keyframes reveal-slice {
  from {
    opacity: 0;
    clip-path: inset(100% 0 0 0);
  }
  to {
    opacity: 1;
    clip-path: inset(0 0 0 0);
  }
}

@keyframes hero-bounce {
  0%,
  100% {
    transform: translate(-50%, 0);
  }
  50% {
    transform: translate(-50%, 0.45rem);
  }
}

@media (min-width: 640px) {
  .home-hero {
    margin-top: -2rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .hero-fade-enter-active,
  .hero-fade-leave-active,
  .hero-slice-clip,
  .hero-down {
    transition: none;
    animation: none;
  }
}
</style>
