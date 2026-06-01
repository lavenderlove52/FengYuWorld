<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { RouterLink } from 'vue-router'
import type { PinnedPost } from '@/types/post'

const props = defineProps<{
  pins: PinnedPost[]
  autoPlayMs?: number
}>()

const current = ref(0)
const transitioning = ref(false)
let timer: ReturnType<typeof setInterval> | null = null

const post = computed(() => props.pins[current.value]?.post)

const goTo = (index: number) => {
  if (transitioning.value || index === current.value) return
  transitioning.value = true
  current.value = index
  setTimeout(() => (transitioning.value = false), 400)
}

const prev = () => goTo((current.value - 1 + props.pins.length) % props.pins.length)
const next = () => goTo((current.value + 1) % props.pins.length)

const formatDate = (d: string) => d?.slice(0, 10) ?? ''

const startTimer = () => {
  if (props.pins.length <= 1) return
  timer = setInterval(next, props.autoPlayMs ?? 4500)
}
const stopTimer = () => { if (timer) clearInterval(timer) }

onMounted(startTimer)
onUnmounted(stopTimer)
</script>

<template>
  <div
    v-if="pins.length"
    class="pin-carousel relative overflow-hidden rounded-xl bg-[#f1f5f9] border border-slate-200"
    @mouseenter="stopTimer"
    @mouseleave="startTimer"
  >
    <RouterLink :to="`/p/${post?.id}`" class="block">
      <transition name="content-fade" mode="out-in">
        <div :key="current" class="flex flex-col sm:flex-row sm:items-stretch">

          <!-- 移动端封面图（顶部） / 桌面端右侧 -->
          <div class="sm:hidden h-40 relative overflow-hidden">
            <img
              v-if="post?.cover_image_url"
              :src="post.cover_image_url"
              :alt="post?.title"
              class="absolute inset-0 w-full h-full object-cover"
            />
            <div
              v-else
              class="absolute inset-0 flex flex-col items-center justify-center bg-slate-200 gap-2"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="m2.25 15.75 5.159-5.159a2.25 2.25 0 0 1 3.182 0l5.159 5.159m-1.5-1.5 1.409-1.409a2.25 2.25 0 0 1 3.182 0l2.909 2.909m-18 3.75h16.5a1.5 1.5 0 0 0 1.5-1.5V6a1.5 1.5 0 0 0-1.5-1.5H3.75A1.5 1.5 0 0 0 2.25 6v12a1.5 1.5 0 0 0 1.5 1.5Zm10.5-11.25h.008v.008h-.008V8.25Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />
              </svg>
              <span class="text-xs text-slate-300">暂无封面</span>
            </div>
          </div>

          <!-- 文字内容 -->
          <div class="flex-1 flex flex-col justify-center px-5 py-5 sm:px-10 sm:py-8 min-w-0">
            <div class="flex items-center gap-2 mb-3 sm:mb-4">
              <span class="text-[11px] font-semibold tracking-widest text-slate-400 uppercase">置顶</span>
              <span class="w-4 h-px bg-slate-300" />
              <span class="text-[11px] text-slate-400">{{ post?.author?.nickname }}</span>
              <span class="text-[11px] text-slate-300">·</span>
              <span class="text-[11px] text-slate-400">{{ formatDate(post?.published_at || post?.created_at || '') }}</span>
            </div>

            <h2 class="text-lg sm:text-2xl font-bold text-slate-800 leading-snug mb-2 sm:mb-3 line-clamp-2">
              {{ post?.title }}
            </h2>

            <p class="text-slate-500 text-sm leading-relaxed line-clamp-2 mb-4 sm:mb-5 sm:max-w-lg">
              {{ post?.summary }}
            </p>

            <div class="flex items-center gap-4">
              <span class="text-xs text-slate-400">{{ post?.view_count }} 阅读</span>
              <!-- 圆点导航 -->
              <div v-if="pins.length > 1" class="flex gap-1.5">
                <button
                  v-for="(_, i) in pins"
                  :key="i"
                  class="rounded-full transition-all duration-300"
                  :class="i === current
                    ? 'w-4 h-2 bg-slate-600'
                    : 'w-2 h-2 bg-slate-300 hover:bg-slate-400'"
                  @click.prevent="goTo(i)"
                />
              </div>
            </div>
          </div>

          <!-- 桌面端右侧封面图（sm 以上显示） -->
          <div class="hidden sm:block w-72 flex-shrink-0 relative overflow-hidden">
            <img
              v-if="post?.cover_image_url"
              :src="post.cover_image_url"
              :alt="post?.title"
              class="absolute inset-0 w-full h-full object-cover"
            />
            <div
              v-else
              class="absolute inset-0 flex flex-col items-center justify-center bg-slate-200 gap-2"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-10 h-10 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="m2.25 15.75 5.159-5.159a2.25 2.25 0 0 1 3.182 0l5.159 5.159m-1.5-1.5 1.409-1.409a2.25 2.25 0 0 1 3.182 0l2.909 2.909m-18 3.75h16.5a1.5 1.5 0 0 0 1.5-1.5V6a1.5 1.5 0 0 0-1.5-1.5H3.75A1.5 1.5 0 0 0 2.25 6v12a1.5 1.5 0 0 0 1.5 1.5Zm10.5-11.25h.008v.008h-.008V8.25Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />
              </svg>
              <span class="text-xs text-slate-300">暂无封面</span>
            </div>
          </div>
        </div>
      </transition>
    </RouterLink>

    <!-- 左右箭头 -->
    <button v-if="pins.length > 1" class="arrow-btn left-3" @click.prevent="prev">‹</button>
    <button v-if="pins.length > 1" class="arrow-btn right-3" @click.prevent="next">›</button>
  </div>
</template>

<style scoped>
.pin-carousel { }

.arrow-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  z-index: 20;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.85);
  color: #475569;
  font-size: 20px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #e2e8f0;
  cursor: pointer;
  transition: all 0.2s;
}
.arrow-btn:hover {
  background: #fff;
  color: #1e293b;
  border-color: #cbd5e1;
}

.content-fade-enter-active,
.content-fade-leave-active { transition: opacity 0.35s ease; }
.content-fade-enter-from,
.content-fade-leave-to { opacity: 0; }
</style>
