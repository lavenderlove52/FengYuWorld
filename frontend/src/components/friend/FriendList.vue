<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { NButton, NBadge } from 'naive-ui'
import type { FriendListItem } from '@/types/friend'
import type { UnreadCounts } from '@/types/chat'

defineProps<{
  items: FriendListItem[]
  selectedId?: number
  unreadCounts?: UnreadCounts
}>()

const emit = defineEmits<{
  remove: [id: number]
  select: [friend: FriendListItem]
}>()

const formatDate = (d: string) => d?.slice(0, 10) ?? ''
</script>

<template>
  <div v-if="!items.length" class="py-8 text-center text-sm text-slate-400">
    <div class="text-3xl mb-2">👥</div>
    暂无好友，快去添加吧！
  </div>
  <div v-else class="space-y-1.5">
    <div
      v-for="item in items"
      :key="item.id"
      class="group flex items-center gap-3 px-3 py-2.5 rounded-xl border cursor-pointer transition-all"
      :class="
        selectedId === item.id
          ? 'bg-indigo-50 border-indigo-200 shadow-sm'
          : 'bg-white border-transparent hover:border-slate-200 hover:bg-slate-50'
      "
      @click="emit('select', item)"
    >
      <!-- 头像 + 未读角标 -->
      <NBadge :value="unreadCounts?.[item.id] ?? 0" :max="99" :show="(unreadCounts?.[item.id] ?? 0) > 0">
        <div class="w-10 h-10 rounded-full flex-shrink-0 overflow-hidden">
          <img
            v-if="item.avatar_url"
            :src="item.avatar_url"
            :alt="item.nickname"
            class="w-full h-full object-cover"
          />
          <div
            v-else
            class="w-full h-full bg-indigo-100 flex items-center justify-center text-sm font-bold text-indigo-500"
          >
            {{ item.nickname?.[0]?.toUpperCase() || '?' }}
          </div>
        </div>
      </NBadge>

      <!-- 主信息 -->
      <div class="flex-1 min-w-0">
        <div class="flex items-baseline gap-1.5">
          <span class="font-semibold text-slate-800 text-sm truncate">{{ item.nickname }}</span>
          <span class="text-xs text-slate-400 flex-shrink-0">@{{ item.username }}</span>
        </div>
        <div class="text-xs text-slate-400 mt-0.5">成为好友于 {{ formatDate(item.friends_since) }}</div>
      </div>

      <!-- 操作按钮（hover 时显示） -->
      <div
        class="flex items-center gap-1 flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
        @click.stop
      >
        <RouterLink :to="`/u/${item.username}`">
          <NButton size="tiny" tertiary title="主页">主页</NButton>
        </RouterLink>
        <NButton size="tiny" type="error" tertiary title="删除好友" @click="emit('remove', item.id)">删除</NButton>
      </div>
    </div>
  </div>
</template>
