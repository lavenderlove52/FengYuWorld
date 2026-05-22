<script setup lang="ts">
import { NSelect } from 'naive-ui'
import type { PostVisibility } from '@/types/post'

const model = defineModel<PostVisibility>({ required: true })

const options = [
  { label: '公开', value: 'public' },
  { label: '仅自己', value: 'private' },
  { label: '全部好友', value: 'friends' },
  { label: '指定好友', value: 'specified' },
]
</script>

<template>
  <NSelect v-model:value="model" :options="options" placeholder="可见性" />
</template>
