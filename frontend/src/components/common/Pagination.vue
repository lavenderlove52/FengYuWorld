<script setup lang="ts">
import { NPagination } from 'naive-ui'

defineProps<{
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{ change: [page: number] }>()
</script>

<template>
  <div class="flex justify-center mt-8">
    <NPagination
      :page="page"
      :page-size="pageSize"
      :item-count="total"
      @update:page="(p: number) => emit('change', p)"
    />
  </div>
</template>
