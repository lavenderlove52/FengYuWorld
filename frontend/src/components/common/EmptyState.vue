<script setup lang="ts">
import { NEmpty } from 'naive-ui'

defineProps<{ message?: string }>()
</script>

<template>
  <NEmpty :description="message || '暂无内容'" class="py-12" />
</template>
