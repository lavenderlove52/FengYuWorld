<script setup lang="ts">
import { NButton } from 'naive-ui'
import type { FriendListItem } from '@/types/friend'

defineProps<{ items: FriendListItem[] }>()
const emit = defineEmits<{ remove: [id: number] }>()
</script>

<template>
  <div v-if="!items.length" class="text-sm text-slate-500">暂无好友</div>
  <div v-else class="space-y-3">
    <div
      v-for="item in items"
      :key="item.id"
      class="flex items-center justify-between p-3 bg-white border rounded-lg"
    >
      <div>
        <div class="font-medium">{{ item.nickname }} (@{{ item.username }})</div>
        <div class="text-xs text-slate-500">成为好友于 {{ item.friends_since }}</div>
      </div>
      <NButton size="small" type="error" tertiary @click="emit('remove', item.id)">删除</NButton>
    </div>
  </div>
</template>
