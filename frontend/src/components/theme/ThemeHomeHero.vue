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

const MOBILE_SLICE_COUNT = 4
const DESKTOP_SLICE_COUNT = 8
const MOBILE_BREAKPOINT = 768
const SLICE_STAGGER_MS = 70
const SLICE_ANIM_MS = 700
const MIN_INTERVAL_MS = 3000
const MIN_SLICE_INTERVAL_MS = 5000

const activeIndex = ref(0)
const previousIndex = ref(0)
const sliceCycle = ref(0)
const sliceAnimating = ref(false)
const reduceMotion = ref(false)
const sliceCount = ref(DESKTOP_SLICE_COUNT)
const mobileViewport = ref(window.innerWidth < MOBILE_BREAKPOINT)
let timer: number | null = null
let sliceEndTimer: number | null = null
let motionQuery: MediaQueryList | null = null
let sliceMediaQuery: MediaQueryList | null = null

const currentImage = computed(() => props.images[activeIndex.value] ?? props.images[0])
const previousImage = computed(() => props.images[previousIndex.value] ?? currentImage.value)
const imageURL = (image: ThemeImage) => mobileViewport.value && image.mobile_url ? image.mobile_url : image.url
const canRotate = computed(() => props.config.enabled && props.images.length >= 2 && !reduceMotion.value)
const transition = computed(() => canRotate.value ? props.config.transition : 'off')
const overlayOpacity = computed(() => Math.min(0.85, Math.max(0, props.config.overlay_opacity)))
const heroStyle = computed(() => ({
  minHeight: `calc(${Math.min(100, Math.max(60, props.config.hero_height_vh))}svh - 3.5rem)`,
}))
const sliceRevealDuration = () => {
  const staggerSteps = transition.value === 'interlaced'
    ? Math.floor((sliceCount.value - 1) / 2)
    : sliceCount.value - 1
  return SLICE_ANIM_MS + staggerSteps * SLICE_STAGGER_MS
}
const sliceFinalStyle = computed(() => ({
  '--slice-final-delay': `${sliceRevealDuration()}ms`,
}))

const sliceAnimDelay = (index: number) =>
  sliceAnimating.value ? `${(index - 1) * SLICE_STAGGER_MS}ms` : '0ms'

const interlacedAnimDelay = (index: number) =>
  sliceAnimating.value ? `${Math.floor((index - 1) / 2) * SLICE_STAGGER_MS}ms` : '0ms'

const sliceStyle = (index: number, delay: string) => ({
  '--slice-left': `calc(${((index - 1) / sliceCount.value) * 100}% - ${index === 1 ? 0 : 1}px)`,
  '--slice-right': `calc(${((sliceCount.value - index) / sliceCount.value) * 100}% - ${index === sliceCount.value ? 0 : 1}px)`,
  animationDelay: delay,
})

const updateSliceCount = () => {
  mobileViewport.value = sliceMediaQuery?.matches ?? window.innerWidth < MOBILE_BREAKPOINT
  const nextCount = mobileViewport.value ? MOBILE_SLICE_COUNT : DESKTOP_SLICE_COUNT
  if (sliceCount.value === nextCount) return
  sliceCount.value = nextCount
  sliceCycle.value += 1
  clearSliceEndTimer()
  sliceAnimating.value = false
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

const resetRotation = () => {
  stopTimer()
  clearSliceEndTimer()
  activeIndex.value = 0
  previousIndex.value = 0
  sliceCycle.value = 0
  sliceAnimating.value = false
}

const advance = () => {
  if (!canRotate.value || sliceAnimating.value) return
  previousIndex.value = activeIndex.value
  activeIndex.value = (activeIndex.value + 1) % props.images.length
  sliceCycle.value += 1
  sliceAnimating.value = true
  clearSliceEndTimer()
  const duration = sliceRevealDuration() + 80
  sliceEndTimer = window.setTimeout(() => {
    sliceAnimating.value = false
    sliceEndTimer = null
  }, duration)
}

const startTimer = () => {
  stopTimer()
  if (!canRotate.value) {
    clearSliceEndTimer()
    sliceAnimating.value = false
    return
  }
  const minInterval = transition.value === 'slices' || transition.value === 'interlaced'
    ? MIN_SLICE_INTERVAL_MS
    : MIN_INTERVAL_MS
  const interval = Math.min(30000, Math.max(minInterval, props.config.interval_ms))
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

const onSliceMediaChange = () => {
  updateSliceCount()
}

watch(
  () => props.images.map((image) => image.id).join(','),
  () => {
    resetRotation()
    startTimer()
  },
)
watch(() => [props.config.enabled, props.config.interval_ms, props.config.transition], startTimer)

onMounted(() => {
  motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  reduceMotion.value = motionQuery.matches
  motionQuery.addEventListener('change', onMotionChange)
  document.addEventListener('visibilitychange', onVisibilityChange)
  sliceMediaQuery = window.matchMedia(`(max-width: ${MOBILE_BREAKPOINT - 1}px)`)
  sliceMediaQuery.addEventListener('change', onSliceMediaChange)
  updateSliceCount()
  startTimer()
})

onUnmounted(() => {
  stopTimer()
  clearSliceEndTimer()
  motionQuery?.removeEventListener('change', onMotionChange)
  sliceMediaQuery?.removeEventListener('change', onSliceMediaChange)
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
            :src="imageURL(currentImage)"
            alt=""
            class="absolute inset-0 h-full w-full object-cover"
          />
        </Transition>
      </template>

      <img
        v-else-if="transition === 'off'"
        :src="imageURL(currentImage)"
        alt=""
        class="absolute inset-0 h-full w-full object-cover"
      />

      <template v-else-if="transition === 'slices'">
        <img
          v-if="!sliceAnimating"
          :key="currentImage.id"
          :src="imageURL(currentImage)"
          alt=""
          class="absolute inset-0 h-full w-full object-cover"
        />
        <div v-else class="absolute inset-0" aria-hidden="true">
          <img
            :src="imageURL(previousImage)"
            alt=""
            class="absolute inset-0 h-full w-full object-cover"
          />
          <img
            :src="imageURL(currentImage)"
            alt=""
            class="hero-slice-final-img"
            :style="sliceFinalStyle"
          />
          <div
            v-for="index in sliceCount"
            :key="`${sliceCycle}-${index}`"
            class="hero-slice-clip hero-slice-clip--reveal"
            :style="sliceStyle(index, sliceAnimDelay(index))"
          >
            <img
              :src="imageURL(currentImage)"
              alt=""
              class="hero-slice-img"
            />
          </div>
        </div>
      </template>

      <template v-else>
        <img
          v-if="!sliceAnimating"
          :key="currentImage.id"
          :src="imageURL(currentImage)"
          alt=""
          class="absolute inset-0 h-full w-full object-cover"
        />
        <div v-else class="absolute inset-0" aria-hidden="true">
          <img
            :src="imageURL(previousImage)"
            alt=""
            class="absolute inset-0 h-full w-full object-cover"
          />
          <img
            :src="imageURL(currentImage)"
            alt=""
            class="hero-slice-final-img"
            :style="sliceFinalStyle"
          />
          <div class="absolute inset-0">
            <div
              v-for="index in sliceCount"
              :key="`${sliceCycle}-interlaced-${index}`"
              class="hero-slice-clip hero-interlaced-clip"
              :class="index % 2 === 0 ? 'hero-interlaced-clip--bottom' : 'hero-interlaced-clip--top'"
              :style="sliceStyle(index, interlacedAnimDelay(index))"
            >
              <img
                :src="imageURL(currentImage)"
                alt=""
                class="hero-slice-img"
              />
            </div>
          </div>
        </div>
      </template>
    </div>

    <div class="absolute inset-0 bg-black" :style="{ opacity: overlayOpacity }" />

    <div class="absolute inset-0 z-[1] flex flex-col items-center justify-center px-5 py-16 text-center text-white">
      <div class="mb-5 h-20 w-20 overflow-hidden rounded-full border-2 border-white/70 bg-black/25 shadow-xl">
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
  content-visibility: auto;
  contain-intrinsic-size: auto calc(100svh - 3.5rem);
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
  position: absolute;
  inset: 0;
  opacity: 1;
  clip-path: inset(0 var(--slice-right) 0 var(--slice-left));
}

.hero-slice-img {
  position: absolute;
  inset: 0;
  height: 100%;
  width: 100%;
  object-fit: cover;
  object-position: center;
  pointer-events: none;
  user-select: none;
}

.hero-slice-final-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  opacity: 0;
  animation: show-slice-final 1ms linear forwards;
  animation-delay: var(--slice-final-delay);
}

.hero-slice-clip--reveal {
  opacity: 0;
  animation: reveal-slice 700ms ease-out forwards;
}

.hero-interlaced-clip {
  opacity: 0;
}

.hero-interlaced-clip--top {
  animation: reveal-interlaced-top 700ms cubic-bezier(0.22, 0.75, 0.25, 1) forwards;
}

.hero-interlaced-clip--bottom {
  animation: reveal-interlaced-bottom 700ms cubic-bezier(0.22, 0.75, 0.25, 1) forwards;
}

.hero-down {
  transform: translateX(-50%);
  animation: hero-bounce 1.8s ease-in-out infinite;
}

@keyframes reveal-slice {
  from {
    opacity: 0;
    clip-path: inset(100% var(--slice-right) 0 var(--slice-left));
  }
  to {
    opacity: 1;
    clip-path: inset(0 var(--slice-right) 0 var(--slice-left));
  }
}

@keyframes reveal-interlaced-top {
  from {
    opacity: 0;
    clip-path: inset(100% var(--slice-right) 0 var(--slice-left));
  }
  to {
    opacity: 1;
    clip-path: inset(0 var(--slice-right) 0 var(--slice-left));
  }
}

@keyframes reveal-interlaced-bottom {
  from {
    opacity: 0;
    clip-path: inset(0 var(--slice-right) 100% var(--slice-left));
  }
  to {
    opacity: 1;
    clip-path: inset(0 var(--slice-right) 0 var(--slice-left));
  }
}

@keyframes show-slice-final {
  to {
    opacity: 1;
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
  .hero-interlaced-clip,
  .hero-down {
    transition: none;
    animation: none;
  }
}
</style>
