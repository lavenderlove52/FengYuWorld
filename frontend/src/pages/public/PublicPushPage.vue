<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NSpin, useMessage } from 'naive-ui'
import { useRoute } from 'vue-router'
import { getPublicNtfyPush } from '@/api/push'
import NtfyPushForm from '@/components/push/NtfyPushForm.vue'

const route = useRoute()
const toast = useMessage()
const token = String(route.params.token || '')
const topic = ref('')
const loading = ref(true)
const notFound = ref(false)

onMounted(async () => {
  try {
    const response = await getPublicNtfyPush(token)
    topic.value = response.data?.topic || ''
  } catch (error: any) {
    notFound.value = true
    toast.error(error.message || '公开推送地址不可用')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <main class="min-h-screen bg-slate-50 px-4 py-10">
    <div class="mx-auto max-w-3xl">
      <div class="mb-7 text-center">
        <h1 class="text-2xl font-bold text-slate-800">发送手机通知</h1>
        <p v-if="topic" class="mt-2 text-sm text-slate-400">ntfy 话题：{{ topic }}</p>
      </div>
      <NSpin :show="loading">
        <div v-if="notFound" class="rounded-2xl bg-white p-12 text-center text-slate-500">该公开推送地址不存在或已关闭。</div>
        <NtfyPushForm v-else-if="topic" :public-token="token" />
      </NSpin>
    </div>
  </main>
</template>
