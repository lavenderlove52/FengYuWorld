<script setup lang="ts">
import { NButton, NSpace } from 'naive-ui'
import type { FriendshipItem } from '@/types/friend'

defineProps<{ items: FriendshipItem[] }>()
const emit = defineEmits<{ accept: [id: number]; reject: [id: number] }>()
</script>

<template>
  <div v-if="!items.length" class="text-sm text-slate-500">暂无待处理申请</div>
  <div v-else class="space-y-3">
    <div
      v-for="item in items"
      :key="item.id"
      class="flex items-center justify-between p-3 bg-white border rounded-lg"
    >
      <div>
        <div class="font-medium">{{ item.user.nickname }} (@{{ item.user.username }})</div>
        <div class="text-xs text-slate-500">{{ item.created_at }}</div>
      </div>
      <NSpace>
        <NButton size="small" type="primary" @click="emit('accept', item.id)">接受</NButton>
        <NButton size="small" @click="emit('reject', item.id)">拒绝</NButton>
      </NSpace>
    </div>
  </div>
</template>
