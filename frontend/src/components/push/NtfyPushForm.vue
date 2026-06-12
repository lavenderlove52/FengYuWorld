<script setup lang="ts">
import { computed, ref } from 'vue'
import { NButton, NInput, NInputNumber, NTabPane, NTabs, useMessage } from 'naive-ui'
import { sendNtfyPush, sendPublicNtfyPush } from '@/api/push'
import type { NtfyMode } from '@/api/push'

const props = defineProps<{ publicToken?: string; topicId?: number; disabled?: boolean }>()
const toast = useMessage()
const mode = ref<NtfyMode>('message')
const title = ref('')
const content = ref('')
const priority = ref(3)
const tags = ref('warning,rotating_light')
const sending = ref(false)

const modes: Array<{ key: NtfyMode; label: string; description: string }> = [
  { key: 'message', label: '普通通知', description: '发送一条普通 ntfy 通知' },
  { key: 'title', label: '标题通知', description: '发送带标题和正文的通知' },
  { key: 'urgent', label: '高优先级响铃', description: '使用最高优先级 5，Android 可响铃、长振动并弹出通知' },
  { key: 'priority', label: '自定义优先级', description: '选择 1 到 5 的 ntfy 消息优先级' },
  { key: 'tags', label: '标签警告', description: '添加 ntfy 标签或 Emoji 标签，多个标签使用英文逗号分隔' },
]
const currentMode = computed(() => modes.find((item) => item.key === mode.value) ?? modes[0])

const send = async () => {
  if (props.disabled) { toast.warning('请先保存 ntfy 话题'); return }
  if (!content.value.trim()) { toast.warning('请输入推送内容'); return }
  if (mode.value === 'title' && !title.value.trim()) { toast.warning('请输入推送标题'); return }
  sending.value = true
  const payload = {
    topic_id: props.topicId,
    mode: mode.value,
    title: title.value.trim(),
    message: content.value.trim(),
    priority: priority.value,
    tags: tags.value.trim(),
  }
  try {
    if (props.publicToken) await sendPublicNtfyPush(props.publicToken, payload)
    else await sendNtfyPush(payload)
    toast.success('通知已发送到 ntfy')
  } catch (error: any) {
    toast.error(error.message || '推送失败')
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <section class="rounded-2xl border border-slate-100 bg-white p-5 md:p-6">
    <h2 class="mb-4 font-semibold text-slate-800">选择推送方式</h2>
    <NTabs v-model:value="mode" type="card" animated>
      <NTabPane v-for="item in modes" :key="item.key" :name="item.key" :tab="item.label" />
    </NTabs>
    <div class="mt-5 space-y-4">
      <div class="rounded-lg bg-emerald-50 px-4 py-3 text-sm text-emerald-700">{{ currentMode.description }}</div>
      <div v-if="mode === 'title'">
        <label class="mb-2 block text-sm font-medium text-slate-600">推送标题</label>
        <NInput v-model:value="title" maxlength="200" show-count placeholder="请输入推送标题" />
      </div>
      <div v-if="mode === 'priority'">
        <label class="mb-2 block text-sm font-medium text-slate-600">消息优先级</label>
        <NInputNumber v-model:value="priority" :min="1" :max="5" class="w-full" />
      </div>
      <div v-if="mode === 'tags'">
        <label class="mb-2 block text-sm font-medium text-slate-600">消息标签</label>
        <NInput v-model:value="tags" maxlength="200" placeholder="warning,rotating_light" />
      </div>
      <div>
        <label class="mb-2 block text-sm font-medium text-slate-600">推送内容</label>
        <NInput v-model:value="content" type="textarea" :autosize="{ minRows: 4, maxRows: 9 }" maxlength="2000" show-count placeholder="请输入要发送到手机的消息" />
      </div>
      <div class="flex justify-end">
        <NButton type="primary" size="large" :loading="sending" :disabled="disabled" @click="send">发送通知</NButton>
      </div>
    </div>
  </section>
</template>
