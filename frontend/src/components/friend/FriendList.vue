<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { NButton, NBadge, NDropdown } from 'naive-ui'
import type { DropdownOption } from 'naive-ui'
import type { FriendListItem } from '@/types/friend'
import type { UnreadCounts } from '@/types/chat'

defineProps<{
  items: FriendListItem[]
  unreadCounts?: UnreadCounts
}>()

const emit = defineEmits<{
  remove: [id: number]
  chat: [friend: FriendListItem]
}>()

const moreMenuOptions: DropdownOption[] = [
  { label: '删除好友', key: 'delete' },
]

function onMoreSelect(key: string | number, item: FriendListItem) {
  if (key === 'delete') emit('remove', item.id)
}
</script>

<template>
  <div v-if="!items.length" class="py-12 text-center text-sm text-slate-400">
    <div class="text-4xl mb-3">👥</div>
    <p>暂无好友</p>
    <p class="text-xs mt-1">在上方输入用户名添加好友吧</p>
  </div>
  <ul v-else class="divide-y divide-slate-100">
    <li
      v-for="item in items"
      :key="item.id"
      class="flex items-center gap-3 px-3 py-3 cursor-pointer hover:bg-slate-50 active:bg-slate-100 transition-colors"
      @click="emit('chat', item)"
    >
      <!-- 头像 + 未读 -->
      <NBadge
        :value="unreadCounts?.[item.id] ?? 0"
        :max="99"
        :show="(unreadCounts?.[item.id] ?? 0) > 0"
        :offset="[-2, 2]"
      >
        <div class="w-12 h-12 rounded-full flex-shrink-0 overflow-hidden bg-slate-100">
          <img
            v-if="item.avatar_url"
            :src="item.avatar_url"
            :alt="item.nickname"
            class="w-full h-full object-cover"
          />
          <div
            v-else
            class="w-full h-full flex items-center justify-center text-base font-semibold text-[#12b7f5]"
          >
            {{ item.nickname?.[0]?.toUpperCase() || '?' }}
          </div>
        </div>
      </NBadge>

      <!-- 信息 -->
      <div class="flex-1 min-w-0">
        <div class="font-medium text-slate-800 truncate">{{ item.nickname || item.username }}</div>
        <div class="text-xs text-slate-400 truncate mt-0.5">@{{ item.username }}</div>
      </div>

      <!-- 操作 -->
      <div class="flex items-center gap-0.5 flex-shrink-0" @click.stop>
        <RouterLink :to="`/u/${item.username}`">
          <NButton size="small" quaternary>主页</NButton>
        </RouterLink>
        <NDropdown
          trigger="click"
          :options="moreMenuOptions"
          @select="(key) => onMoreSelect(key, item)"
        >
          <NButton size="small" quaternary title="更多" class="!px-2">⋯</NButton>
        </NDropdown>
      </div>
    </li>
  </ul>
</template>
