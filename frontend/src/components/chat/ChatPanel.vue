<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { NButton, NInput, NSpin } from 'naive-ui'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import * as uploadsApi from '@/api/uploads'
import type { FriendListItem } from '@/types/friend'
import type { Message } from '@/types/chat'

const props = defineProps<{ friend: FriendListItem }>()
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
      chatStore.send(friendId.value, res.data.url, 'image')
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
  // 保持滚动位置
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

// 进入对话时：加载历史 + 标记已读 + 滚到底部
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

// 新消息到达时自动滚到底部（如果已在底部附近）
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

// 点击面板外关闭表情面板
function onDocClick(e: MouseEvent) {
  if (showEmoji.value && !(e.target as Element).closest('.emoji-panel')) {
    showEmoji.value = false
  }
}
onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))
</script>

<template>
  <div class="flex flex-col h-full bg-white rounded-2xl border border-slate-200 shadow-lg overflow-hidden">
    <!-- 顶部 -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-slate-100 bg-slate-50">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-full bg-indigo-100 flex items-center justify-center text-sm font-bold text-indigo-500 flex-shrink-0">
          {{ friend.nickname?.[0]?.toUpperCase() || '?' }}
        </div>
        <div>
          <div class="font-semibold text-slate-800">{{ friend.nickname }}</div>
          <div class="text-xs text-slate-400">@{{ friend.username }}</div>
        </div>
      </div>
      <NButton text @click="emit('close')">
        <span class="text-xl text-slate-400 hover:text-slate-600">✕</span>
      </NButton>
    </div>

    <!-- 消息列表 -->
    <div
      ref="scrollEl"
      class="flex-1 overflow-y-auto px-4 py-3 space-y-1 bg-slate-50/50"
      @scroll="onScroll"
    >
      <!-- 加载更多 -->
      <div class="flex justify-center py-2">
        <NButton
          v-if="hasMore"
          size="tiny"
          tertiary
          :loading="loadingMore"
          @click="loadMore"
        >
          加载更多
        </NButton>
        <span v-else class="text-xs text-slate-400">已加载全部消息</span>
      </div>

      <template v-for="(msg, idx) in msgs" :key="msg.id">
        <!-- 时间戳 -->
        <div v-if="shouldShowTimestamp(idx)" class="text-center text-xs text-slate-400 py-1">
          {{ formatTime(msg.created_at) }}
        </div>

        <!-- 自己的消息（右侧） -->
        <div v-if="msg.from_user === myId" class="flex justify-end">
          <div class="max-w-[70%]">
            <!-- 图片 -->
            <img
              v-if="msg.type === 'image'"
              :src="msg.content"
              class="max-w-full rounded-xl cursor-pointer shadow-sm border border-slate-100"
              @click="openImageInNewTab(msg.content)"
            />
            <!-- 表情 / 文字 -->
            <div
              v-else
              class="px-3 py-2 bg-indigo-500 text-white rounded-2xl rounded-tr-sm text-sm leading-relaxed break-words whitespace-pre-wrap shadow-sm"
              :class="msg.type === 'emoji' ? 'text-2xl bg-transparent !shadow-none' : ''"
            >
              {{ msg.content }}
            </div>
          </div>
        </div>

        <!-- 对方的消息（左侧） -->
        <div v-else class="flex justify-start gap-2">
          <div class="w-8 h-8 rounded-full bg-slate-200 flex items-center justify-center text-xs font-bold text-slate-500 flex-shrink-0 mt-1">
            {{ friend.nickname?.[0]?.toUpperCase() || '?' }}
          </div>
          <div class="max-w-[70%]">
            <img
              v-if="msg.type === 'image'"
              :src="msg.content"
              class="max-w-full rounded-xl cursor-pointer shadow-sm border border-slate-100"
              @click="openImageInNewTab(msg.content)"
            />
            <div
              v-else
              class="px-3 py-2 bg-white border border-slate-200 rounded-2xl rounded-tl-sm text-sm leading-relaxed break-words whitespace-pre-wrap shadow-sm text-slate-800"
              :class="msg.type === 'emoji' ? 'text-2xl bg-transparent border-transparent !shadow-none' : ''"
            >
              {{ msg.content }}
            </div>
          </div>
        </div>
      </template>

      <div v-if="msgs.length === 0 && !loadingMore" class="text-center text-sm text-slate-400 py-8">
        开始聊天吧 👋
      </div>
    </div>

    <!-- 底部输入区 -->
    <div class="border-t border-slate-100 bg-white">
      <!-- 表情面板 -->
      <div v-if="showEmoji" class="emoji-panel flex flex-wrap gap-1 px-3 pt-3 pb-1">
        <button
          v-for="e in EMOJIS"
          :key="e"
          class="text-2xl hover:bg-slate-100 rounded-lg p-1 transition"
          @click.stop="insertEmoji(e)"
        >
          {{ e }}
        </button>
      </div>

      <div class="flex items-end gap-2 px-3 py-2">
        <!-- 表情按钮 -->
        <button
          class="emoji-panel text-2xl leading-none flex-shrink-0 mb-1 opacity-70 hover:opacity-100 transition"
          title="表情"
          @click.stop="showEmoji = !showEmoji"
        >
          😊
        </button>

        <!-- 图片上传 -->
        <button
          class="flex-shrink-0 mb-1 opacity-70 hover:opacity-100 transition text-slate-500"
          title="发送图片"
          :disabled="uploading"
          @click="fileInput?.click()"
        >
          <NSpin v-if="uploading" size="small" />
          <span v-else class="text-xl">🖼️</span>
        </button>
        <input
          ref="fileInput"
          type="file"
          accept="image/*"
          class="hidden"
          @change="onFileChange"
        />

        <!-- 输入框 -->
        <NInput
          v-model:value="inputText"
          type="textarea"
          placeholder="输入消息…"
          :autosize="{ minRows: 1, maxRows: 4 }"
          class="flex-1"
          @keydown="onKeydown"
        />

        <!-- 发送 -->
        <NButton
          type="primary"
          :disabled="!inputText.trim()"
          class="flex-shrink-0 mb-1"
          @click="send"
        >
          发送
        </NButton>
      </div>
    </div>
  </div>
</template>
