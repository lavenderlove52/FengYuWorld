<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { NInput, NButton, useMessage } from 'naive-ui'
import * as friendsApi from '@/api/friends'
import { useFriendStore } from '@/stores/friend'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import { getAccessToken } from '@/api/http'
import FriendRequestList from '@/components/friend/FriendRequestList.vue'
import FriendList from '@/components/friend/FriendList.vue'
import ChatPanel from '@/components/chat/ChatPanel.vue'
import type { FriendListItem } from '@/types/friend'

const message = useMessage()
const friendStore = useFriendStore()
const chatStore = useChatStore()
const authStore = useAuthStore()
const username = ref('')
const selectedFriend = ref<FriendListItem | null>(null)
const showRequests = ref(true)

const requestCount = computed(() => friendStore.received.length)

const load = async () => {
  await Promise.all([friendStore.loadFriends(), friendStore.loadReceived()])
}

onMounted(async () => {
  await load()
  const token = getAccessToken()
  if (token && authStore.user) {
    chatStore.initWs(authStore.user.id, token)
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
    selectedFriend.value = null
    chatStore.setActiveFriend(null)
  }
  load()
}

const selectFriend = (friend: FriendListItem) => {
  selectedFriend.value = friend
  chatStore.setActiveFriend(friend.id)
}
</script>

<template>
  <div class="flex gap-4 h-[calc(100vh-8rem)] min-h-0">
    <!-- 左侧：好友管理 -->
    <div class="w-80 flex-shrink-0 flex flex-col gap-3 min-h-0">
      <!-- 标题 + 添加好友 -->
      <div class="bg-white rounded-2xl border border-slate-200 p-4 flex-shrink-0">
        <h1 class="text-lg font-bold mb-3 text-slate-800">好友管理</h1>
        <div class="flex gap-2">
          <NInput
            v-model:value="username"
            placeholder="输入用户名添加好友"
            size="small"
            class="flex-1"
            @keydown.enter="send"
          />
          <NButton type="primary" size="small" @click="send">添加</NButton>
        </div>
      </div>

      <!-- 收到的申请（可折叠） -->
      <div class="bg-white rounded-2xl border border-slate-200 flex-shrink-0">
        <button
          class="w-full flex items-center justify-between px-4 py-3 text-sm font-semibold text-slate-700 hover:bg-slate-50 rounded-2xl transition"
          @click="showRequests = !showRequests"
        >
          <div class="flex items-center gap-2">
            <span>收到的申请</span>
            <span
              v-if="requestCount > 0"
              class="bg-red-500 text-white text-xs font-bold px-1.5 py-0.5 rounded-full min-w-[18px] text-center leading-none"
            >{{ requestCount }}</span>
          </div>
          <span class="text-slate-400 text-xs transition-transform" :class="showRequests ? 'rotate-180' : ''">▼</span>
        </button>
        <div v-if="showRequests" class="px-3 pb-3">
          <FriendRequestList :items="friendStore.received" @accept="accept" @reject="reject" />
        </div>
      </div>

      <!-- 好友列表 -->
      <div class="bg-white rounded-2xl border border-slate-200 flex flex-col min-h-0 flex-1">
        <div class="px-4 py-3 border-b border-slate-100 flex-shrink-0">
          <h2 class="text-sm font-semibold text-slate-700">
            好友列表
            <span class="text-slate-400 font-normal ml-1">{{ friendStore.friends.length }}</span>
          </h2>
        </div>
        <div class="flex-1 overflow-y-auto px-2 py-2">
          <FriendList
            :items="friendStore.friends"
            :selected-id="selectedFriend?.id"
            :unread-counts="chatStore.unreadCounts"
            @remove="remove"
            @select="selectFriend"
          />
        </div>
      </div>
    </div>

    <!-- 右侧：聊天面板 -->
    <div class="flex-1 min-w-0">
      <ChatPanel
        v-if="selectedFriend"
        :friend="selectedFriend"
        @close="() => { selectedFriend = null; chatStore.setActiveFriend(null) }"
      />
      <div
        v-else
        class="h-full flex flex-col items-center justify-center text-slate-400 bg-white rounded-2xl border border-slate-200"
      >
        <div class="text-5xl mb-4">💬</div>
        <div class="text-base font-medium text-slate-500">选择一位好友开始聊天</div>
        <div class="text-sm text-slate-400 mt-1">从左侧列表中点击好友即可</div>
      </div>
    </div>
  </div>
</template>
