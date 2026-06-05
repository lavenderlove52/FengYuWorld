<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { NInput, NButton, useMessage } from 'naive-ui'
import * as friendsApi from '@/api/friends'
import { useFriendStore } from '@/stores/friend'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'
import FriendRequestList from '@/components/friend/FriendRequestList.vue'
import FriendList from '@/components/friend/FriendList.vue'
import ChatPanel from '@/components/chat/ChatPanel.vue'
import type { FriendListItem } from '@/types/friend'

type ViewMode = 'list' | 'chat'

const message = useMessage()
const friendStore = useFriendStore()
const chatStore = useChatStore()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const username = ref('')
const selectedFriend = ref<FriendListItem | null>(null)
const viewMode = ref<ViewMode>('list')
const showRequests = ref(false)
const showAddFriend = ref(false)

const requestCount = computed(() => friendStore.received.length)
const totalUnread = computed(() =>
  settingsStore.friendChatEnabled
    ? Object.values(chatStore.unreadCounts).reduce((sum, n) => sum + (n || 0), 0)
    : 0,
)

const load = async () => {
  await Promise.all([friendStore.loadFriends(), friendStore.loadReceived()])
}

onMounted(async () => {
  await settingsStore.loadPublicSettings()
  await load()
  if (authStore.user && settingsStore.friendChatEnabled) {
    chatStore.initWs(authStore.user.id)
  }
})

onUnmounted(() => {
  chatStore.closeWs()
  chatStore.setActiveFriend(null)
})

const send = async () => {
  if (!username.value.trim()) return
  const res = await friendsApi.sendFriendRequest(username.value.trim())
  if (res.code === 0) {
    message.success('申请已发送')
    username.value = ''
    showAddFriend.value = false
    load()
  } else {
    message.error(res.message)
  }
}

const accept = async (id: number) => {
  await friendsApi.acceptRequest(id)
  message.success('已接受')
  load()
}

const reject = async (id: number) => {
  await friendsApi.rejectRequest(id)
  message.info('已拒绝')
  load()
}

const remove = async (id: number) => {
  await friendsApi.deleteFriend(id)
  message.success('已删除好友')
  if (selectedFriend.value?.id === id) {
    closeChat()
  }
  load()
}

const openChat = (friend: FriendListItem) => {
  if (!settingsStore.friendChatEnabled) {
    message.warning('好友聊天已关闭')
    return
  }
  selectedFriend.value = friend
  viewMode.value = 'chat'
  chatStore.setActiveFriend(friend.id)
}

const closeChat = () => {
  selectedFriend.value = null
  viewMode.value = 'list'
  chatStore.setActiveFriend(null)
}
</script>

<template>
  <div class="friends-page h-[calc(100vh-5.5rem)] sm:h-[calc(100vh-6rem)] min-h-[420px] max-w-2xl mx-auto flex flex-col">
    <!-- 好友列表视图 -->
    <div v-show="viewMode === 'list'" class="flex flex-col h-full min-h-0 bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
      <!-- 顶栏 -->
      <header class="flex items-center justify-between px-4 py-3 border-b border-slate-100 bg-[#f5f5f5] flex-shrink-0">
        <h1 class="text-lg font-semibold text-slate-800">好友</h1>
        <div class="flex items-center gap-2">
          <span
            v-if="!settingsStore.friendChatEnabled"
            class="text-xs text-slate-400 bg-slate-100 px-2 py-0.5 rounded-full"
          >聊天已关闭</span>
          <span
            v-if="totalUnread > 0"
            class="text-xs text-white bg-red-500 px-2 py-0.5 rounded-full"
          >{{ totalUnread > 99 ? '99+' : totalUnread }} 条未读</span>
          <NButton size="small" quaternary @click="showAddFriend = !showAddFriend">
            {{ showAddFriend ? '收起' : '添加' }}
          </NButton>
        </div>
      </header>

      <!-- 添加好友 -->
      <div v-if="showAddFriend" class="px-4 py-3 border-b border-slate-100 bg-slate-50 flex-shrink-0">
        <div class="flex gap-2">
          <NInput
            v-model:value="username"
            placeholder="输入用户名添加好友"
            size="small"
            class="flex-1"
            @keydown.enter="send"
          />
          <NButton type="primary" size="small" class="!bg-[#12b7f5] !border-[#12b7f5]" @click="send">
            发送申请
          </NButton>
        </div>
      </div>

      <!-- 收到的申请 -->
      <div class="flex-shrink-0 border-b border-slate-100">
        <button
          type="button"
          class="w-full flex items-center justify-between px-4 py-3 text-sm text-slate-700 hover:bg-slate-50 transition"
          @click="showRequests = !showRequests"
        >
          <span class="flex items-center gap-2">
            <span class="font-medium">新的朋友</span>
            <span
              v-if="requestCount > 0"
              class="bg-red-500 text-white text-xs font-bold min-w-[18px] h-[18px] flex items-center justify-center rounded-full px-1"
            >{{ requestCount }}</span>
          </span>
          <span class="text-slate-400 text-xs" :class="showRequests ? 'rotate-180 inline-block' : ''">▼</span>
        </button>
        <div v-if="showRequests" class="px-3 pb-3 bg-slate-50/80">
          <FriendRequestList :items="friendStore.received" @accept="accept" @reject="reject" />
        </div>
      </div>

      <!-- 好友列表 -->
      <div class="flex-1 min-h-0 overflow-y-auto">
        <div class="px-4 py-2 text-xs text-slate-400 font-medium">
          我的好友 ({{ friendStore.friends.length }})
        </div>
        <FriendList
          :items="friendStore.friends"
          :unread-counts="chatStore.unreadCounts"
          :chat-enabled="settingsStore.friendChatEnabled"
          @remove="remove"
          @chat="openChat"
        />
      </div>
    </div>

    <!-- 聊天视图（全屏覆盖） -->
    <div v-show="viewMode === 'chat' && selectedFriend" class="h-full min-h-0 flex flex-col">
      <ChatPanel
        v-if="selectedFriend"
        :friend="selectedFriend"
        show-back
        @close="closeChat"
      />
    </div>
  </div>
</template>

<style scoped>
.friends-page {
  -webkit-tap-highlight-color: transparent;
}
</style>
