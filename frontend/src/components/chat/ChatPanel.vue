<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { RouterLink } from 'vue-router'
import { NButton, NInput, NSpin } from 'naive-ui'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import * as uploadsApi from '@/api/uploads'
import type { FriendListItem } from '@/types/friend'
import type { Message } from '@/types/chat'

const props = defineProps<{
  friend: FriendListItem
  showBack?: boolean
}>()
const emit = defineEmits<{ close: [] }>()

const chatStore = useChatStore()
const authStore = useAuthStore()

const myId = computed(() => authStore.user?.id ?? 0)
const friendId = computed(() => props.friend.id)

const msgs = computed<Message[]>(() => chatStore.messages[friendId.value] ?? [])
const inputText = ref('')
const showEmoji = ref(false)
const loadingMore = ref(false)
const hasMore = ref(true)
const uploading = ref(false)
const scrollEl = ref<HTMLElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

const EMOJIS = ['😀', '😂', '❤️', '👍', '😭', '🙏', '💪', '🎉', '😊', '🔥', '😍', '🤣', '😅', '🥰', '😎', '🤔', '👏', '🎊', '💯', '✨']

function formatTime(dateStr: string) {
  return new Date(dateStr).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function formatDateLabel(dateStr: string) {
  const d = new Date(dateStr)
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const msgDay = new Date(d.getFullYear(), d.getMonth(), d.getDate())
  const diff = (today.getTime() - msgDay.getTime()) / 86400000
  if (diff === 0) return '今天'
  if (diff === 1) return '昨天'
  return d.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' })
}

function shouldShowTimestamp(idx: number): boolean {
  if (idx === 0) return true
  const prev = msgs.value[idx - 1]
  const curr = msgs.value[idx]
  const diff = new Date(curr.created_at).getTime() - new Date(prev.created_at).getTime()
  return diff > 5 * 60 * 1000
}

async function send() {
  const text = inputText.value.trim()
  if (!text) return
  chatStore.send(friendId.value, text, 'text')
  inputText.value = ''
  showEmoji.value = false
  await nextTick()
  scrollToBottom()
}

function insertEmoji(e: string) {
  inputText.value += e
  showEmoji.value = false
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send()
  }
}

async function onFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  uploading.value = true
  try {
    const res = await uploadsApi.uploadImage(file)
    if (res.code === 0 && res.data?.url) {
      chatStore.send(friendId.value, res.data.mobile_url || res.data.url, 'image')
      await nextTick()
      scrollToBottom()
    }
  } finally {
    uploading.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

async function loadMore() {
  if (loadingMore.value || !hasMore.value) return
  loadingMore.value = true
  const prevScrollHeight = scrollEl.value?.scrollHeight ?? 0
  const more = await chatStore.loadMoreHistory(friendId.value)
  hasMore.value = more
  loadingMore.value = false
  await nextTick()
  if (scrollEl.value) {
    scrollEl.value.scrollTop = scrollEl.value.scrollHeight - prevScrollHeight
  }
}

function onScroll() {
  if (scrollEl.value && scrollEl.value.scrollTop < 60) {
    loadMore()
  }
}

function scrollToBottom() {
  nextTick(() => {
    if (scrollEl.value) {
      scrollEl.value.scrollTop = scrollEl.value.scrollHeight
    }
  })
}

function openImageInNewTab(url: string) {
  window.open(url, '_blank')
}

async function initConversation() {
  await chatStore.loadHistory(friendId.value)
  chatStore.read(friendId.value)
  scrollToBottom()
}

watch(
  () => props.friend.id,
  () => {
    hasMore.value = true
    initConversation()
  },
  { immediate: true },
)

watch(
  msgs,
  () => {
    if (!scrollEl.value) return
    const el = scrollEl.value
    const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 120
    if (nearBottom) scrollToBottom()
  },
  { deep: true },
)

function onDocClick(e: MouseEvent) {
  if (showEmoji.value && !(e.target as Element).closest('.emoji-panel')) {
    showEmoji.value = false
  }
}
onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))
</script>

<template>
  <div class="chat-panel flex flex-col h-full bg-[#ebebeb] rounded-xl border border-slate-200 shadow-sm overflow-hidden">
    <!-- QQ 风格顶栏 -->
    <header class="flex items-center gap-2 px-2 py-2.5 bg-gradient-to-r from-[#4a9eff] to-[#12b7f5] text-white flex-shrink-0 shadow-sm">
      <button
        v-if="showBack !== false"
        type="button"
        class="w-9 h-9 flex items-center justify-center rounded-lg hover:bg-white/15 active:bg-white/25 transition flex-shrink-0"
        aria-label="返回"
        @click="emit('close')"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>

      <div class="w-10 h-10 rounded-full overflow-hidden bg-white/20 flex-shrink-0">
        <img
          v-if="friend.avatar_url"
          :src="friend.avatar_url"
          :alt="friend.nickname"
          class="w-full h-full object-cover"
        />
        <div v-else class="w-full h-full flex items-center justify-center text-sm font-bold">
          {{ friend.nickname?.[0]?.toUpperCase() || '?' }}
        </div>
      </div>

      <div class="flex-1 min-w-0">
        <div class="font-semibold truncate">{{ friend.nickname || friend.username }}</div>
        <div class="text-xs text-white/80 truncate">@{{ friend.username }}</div>
      </div>

      <RouterLink
        :to="`/u/${friend.username}`"
        class="text-xs px-2.5 py-1 rounded-md bg-white/20 hover:bg-white/30 transition flex-shrink-0"
      >
        主页
      </RouterLink>
    </header>

    <!-- 消息区 -->
    <div
      ref="scrollEl"
      class="flex-1 overflow-y-auto px-3 py-4 space-y-3 chat-messages"
      @scroll="onScroll"
    >
      <div class="flex justify-center py-1">
        <button
          v-if="hasMore"
          type="button"
          class="text-xs text-slate-500 bg-white/60 px-3 py-1 rounded-full hover:bg-white transition"
          :disabled="loadingMore"
          @click="loadMore"
        >
          {{ loadingMore ? '加载中…' : '查看更多消息' }}
        </button>
        <span v-else-if="msgs.length" class="text-xs text-slate-400">没有更多消息了</span>
      </div>

      <template v-for="(msg, idx) in msgs" :key="msg.id">
        <div v-if="shouldShowTimestamp(idx)" class="flex justify-center">
          <span class="text-xs text-slate-500 bg-black/5 px-3 py-1 rounded-full">
            {{ formatDateLabel(msg.created_at) }} {{ formatTime(msg.created_at) }}
          </span>
        </div>

        <!-- 自己 -->
        <div v-if="msg.from_user === myId" class="flex justify-end items-end gap-2">
          <div class="max-w-[75%] sm:max-w-[65%]">
            <img
              v-if="msg.type === 'image'"
              :src="msg.content"
              class="max-w-full rounded-lg cursor-pointer shadow-md border-2 border-white"
              @click="openImageInNewTab(msg.content)"
            />
            <div
              v-else
              class="bubble bubble-self px-3.5 py-2 text-sm leading-relaxed break-words whitespace-pre-wrap text-white"
              :class="msg.type === 'emoji' ? '!bg-transparent !shadow-none text-3xl px-0 py-0' : ''"
            >
              {{ msg.content }}
            </div>
          </div>
        </div>

        <!-- 对方 -->
        <div v-else class="flex justify-start items-end gap-2">
          <div class="w-9 h-9 rounded-full overflow-hidden bg-slate-200 flex-shrink-0">
            <img
              v-if="friend.avatar_url"
              :src="friend.avatar_url"
              class="w-full h-full object-cover"
            />
            <div
              v-else
              class="w-full h-full flex items-center justify-center text-xs font-bold text-slate-500"
            >
              {{ friend.nickname?.[0]?.toUpperCase() || '?' }}
            </div>
          </div>
          <div class="max-w-[75%] sm:max-w-[65%]">
            <img
              v-if="msg.type === 'image'"
              :src="msg.content"
              class="max-w-full rounded-lg cursor-pointer shadow-md border-2 border-white"
              @click="openImageInNewTab(msg.content)"
            />
            <div
              v-else
              class="bubble bubble-other px-3.5 py-2 text-sm leading-relaxed break-words whitespace-pre-wrap text-slate-800"
              :class="msg.type === 'emoji' ? '!bg-transparent !shadow-none text-3xl px-0 py-0' : ''"
            >
              {{ msg.content }}
            </div>
          </div>
        </div>
      </template>

      <div v-if="msgs.length === 0 && !loadingMore" class="text-center text-sm text-slate-500 py-16">
        <p class="text-4xl mb-2">💬</p>
        <p>和 {{ friend.nickname || friend.username }} 打个招呼吧</p>
      </div>
    </div>

    <!-- 输入区 -->
    <footer class="flex-shrink-0 bg-[#f5f5f5] border-t border-slate-200/80">
      <div v-if="showEmoji" class="emoji-panel flex flex-wrap gap-1 px-3 pt-2 pb-1 max-h-28 overflow-y-auto">
        <button
          v-for="e in EMOJIS"
          :key="e"
          type="button"
          class="text-2xl hover:bg-white rounded-lg p-1.5 transition"
          @click.stop="insertEmoji(e)"
        >
          {{ e }}
        </button>
      </div>

      <div class="flex items-end gap-2 px-2 py-2">
        <button
          type="button"
          class="emoji-panel w-9 h-9 flex items-center justify-center text-xl rounded-lg hover:bg-white transition flex-shrink-0"
          title="表情"
          @click.stop="showEmoji = !showEmoji"
        >
          😊
        </button>

        <button
          type="button"
          class="w-9 h-9 flex items-center justify-center rounded-lg hover:bg-white transition flex-shrink-0 text-slate-600"
          title="发送图片"
          :disabled="uploading"
          @click="fileInput?.click()"
        >
          <NSpin v-if="uploading" size="small" />
          <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </button>
        <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onFileChange" />

        <NInput
          v-model:value="inputText"
          type="textarea"
          placeholder="输入消息，Enter 发送"
          :autosize="{ minRows: 1, maxRows: 4 }"
          class="flex-1 !rounded-lg"
          @keydown="onKeydown"
        />

        <NButton
          type="primary"
          :disabled="!inputText.trim()"
          class="flex-shrink-0 !bg-[#12b7f5] !border-[#12b7f5] !rounded-lg"
          @click="send"
        >
          发送
        </NButton>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.chat-messages {
  background-image: radial-gradient(circle at 1px 1px, rgba(0, 0, 0, 0.04) 1px, transparent 0);
  background-size: 16px 16px;
}

.bubble {
  position: relative;
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
}

.bubble-self {
  background: linear-gradient(135deg, #4a9eff 0%, #12b7f5 100%);
  border-top-right-radius: 4px;
}

.bubble-other {
  background: #fff;
  border-top-left-radius: 4px;
}
</style>
