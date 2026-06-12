<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { NButton, NInput, NSpin, NSwitch, NTabPane, NTabs, useMessage } from 'naive-ui'
import { getPushSettings, sendBarkPush, updateNtfyPublic, updateNtfyTopic, updatePushSettings } from '@/api/push'
import type { BarkMode } from '@/api/push'
import type { NtfyTopic } from '@/api/push'
import NtfyPushForm from '@/components/push/NtfyPushForm.vue'

const toast = useMessage()
const provider = ref<'bark' | 'ntfy'>('bark')
const mode = ref<BarkMode>('message')
const endpoint = ref('')
const savedEndpoint = ref('')
const title = ref('')
const content = ref('')
const loading = ref(true)
const saving = ref(false)
const sending = ref(false)
const ntfyBaseURL = ref('')
const ntfyTopics = ref<NtfyTopic[]>([])
const selectedNtfyTopicID = ref<number | null>(null)
const ntfyTopicSuffix = ref('')
const ntfySaving = ref(false)
const ntfyPublicSaving = ref(false)

const selectedNtfyTopic = computed(() => ntfyTopics.value.find((item) => item.id === selectedNtfyTopicID.value) || null)
const ntfySubscribeURL = computed(() => selectedNtfyTopic.value ? `${ntfyBaseURL.value.replace(/\/+$/, '')}/${selectedNtfyTopic.value.topic}` : '')
const ntfyPublicURL = computed(() => selectedNtfyTopic.value?.public_token ? `${window.location.origin}/push/${selectedNtfyTopic.value.public_token}` : '')

const modes: Array<{ key: BarkMode; label: string; description: string }> = [
  { key: 'message', label: '普通消息', description: '发送一条只包含正文的通知' },
  { key: 'title', label: '标题消息', description: '发送带标题和正文的通知' },
  { key: 'sound', label: '推送铃声', description: '使用 minuet 铃声发送通知' },
  { key: 'call', label: '持续响铃', description: '持续响铃，直到打开通知或系统停止' },
  { key: 'critical', label: '重要警告', description: '以 critical 级别、音量 5 发送重要警告' },
]

const currentMode = computed(() => modes.find((item) => item.key === mode.value) ?? modes[0])
const previewURL = computed(() => {
  const base = endpoint.value.trim().replace(/\/+$/, '') || 'https://你的Bark服务/bark/设备Key'
  const messagePart = encodeURIComponent(content.value.trim() || '推送内容')
  if (mode.value === 'title') return `${base}/${encodeURIComponent(title.value.trim() || '推送标题')}/${messagePart}`
  const query = mode.value === 'sound' ? '?sound=minuet' : mode.value === 'call' ? '?call=1' : mode.value === 'critical' ? '?level=critical&volume=5' : ''
  return `${base}/${messagePart}${query}`
})

onMounted(async () => {
  try {
    const response = await getPushSettings()
    endpoint.value = response.data?.bark_endpoint || ''
    savedEndpoint.value = endpoint.value
    ntfyBaseURL.value = response.data?.ntfy_base_url || ''
    ntfyTopics.value = response.data?.ntfy_topics || []
    selectedNtfyTopicID.value = ntfyTopics.value[0]?.id || null
  } catch (error: any) {
    toast.error(error.message || '加载推送配置失败')
  } finally {
    loading.value = false
  }
})

const saveEndpoint = async () => {
  if (!endpoint.value.trim()) {
    toast.warning('请填写 Bark 通知地址')
    return false
  }
  saving.value = true
  try {
    const response = await updatePushSettings(endpoint.value.trim())
    endpoint.value = response.data?.bark_endpoint || endpoint.value.trim()
    savedEndpoint.value = endpoint.value
    toast.success('Bark 通知地址已保存')
    return true
  } catch (error: any) {
    toast.error(error.message || '保存失败')
    return false
  } finally {
    saving.value = false
  }
}

const saveNtfyTopic = async () => {
  if (!ntfyTopicSuffix.value.trim()) { toast.warning('请输入话题后缀'); return }
  ntfySaving.value = true
  try {
    const response = await updateNtfyTopic(ntfyTopicSuffix.value.trim())
    const topic = response.data as NtfyTopic
    ntfyTopics.value.push(topic)
    selectedNtfyTopicID.value = topic.id
    ntfyTopicSuffix.value = ''
    toast.success('ntfy 话题已保存，blog 用户读写权限已授予')
  } catch (error: any) {
    toast.error(error.message || '保存 ntfy 话题失败')
  } finally {
    ntfySaving.value = false
  }
}

const toggleNtfyPublic = async (enabled: boolean) => {
  if (!selectedNtfyTopic.value) return
  ntfyPublicSaving.value = true
  try {
    const response = await updateNtfyPublic(selectedNtfyTopic.value.id, enabled)
    const updated = response.data as NtfyTopic
    const index = ntfyTopics.value.findIndex((item) => item.id === updated.id)
    if (index >= 0) ntfyTopics.value[index] = updated
    toast.success(enabled ? '公开推送地址已开启' : '公开推送地址已关闭')
  } catch (error: any) {
    toast.error(error.message || '更新公开推送设置失败')
  } finally {
    ntfyPublicSaving.value = false
  }
}

const copyText = async (value: string, label: string) => {
  if (!value) return
  await navigator.clipboard.writeText(value)
  toast.success(`${label}已复制`)
}

const send = async () => {
  if (!endpoint.value.trim()) {
    toast.warning('请先配置 Bark 通知地址')
    return
  }
  if (!content.value.trim()) {
    toast.warning('请输入推送内容')
    return
  }
  if (mode.value === 'title' && !title.value.trim()) {
    toast.warning('请输入推送标题')
    return
  }

  if (endpoint.value.trim() !== savedEndpoint.value) {
    const saved = await saveEndpoint()
    if (!saved) return
  }

  sending.value = true
  try {
    await sendBarkPush({ mode: mode.value, title: title.value.trim(), message: content.value.trim() })
    toast.success('通知已发送到 Bark')
  } catch (error: any) {
    toast.error(error.message || '推送失败')
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <div class="max-w-4xl">
    <div class="mb-7">
      <h1 class="text-2xl font-bold text-slate-800">手机推送</h1>
      <p class="mt-1 text-sm text-slate-400">从用户中心向你的 iOS 或 Android 设备发送通知</p>
    </div>

    <NTabs v-model:value="provider" type="segment" class="mb-6">
      <NTabPane name="bark" tab="iOS · Bark" />
      <NTabPane name="ntfy" tab="Android · ntfy" />
    </NTabs>

    <NSpin v-if="provider === 'ntfy'" :show="loading">
      <div class="space-y-6">
        <section class="rounded-2xl border border-slate-100 bg-white p-5 md:p-6">
          <div class="mb-4">
            <h2 class="font-semibold text-slate-800">ntfy 订阅话题</h2>
            <p class="mt-1 text-sm text-slate-400">话题固定以 blog- 开头。保存时服务器会给 blog 用户授予该话题的读取和发送权限。</p>
            <p class="mt-1 text-xs text-amber-600">在 ntfy 手机 App 订阅时，还需要填写服务器地址，并使用 blog 用户的账号密码或访问令牌登录。</p>
          </div>
          <div class="flex flex-col gap-3 sm:flex-row">
            <NInput v-model:value="ntfyTopicSuffix" placeholder="my-phone" :allow-input="(value) => /^[A-Za-z0-9_-]*$/.test(value)" maxlength="64">
              <template #prefix>blog-</template>
            </NInput>
            <NButton type="primary" :loading="ntfySaving" @click="saveNtfyTopic">保存话题</NButton>
          </div>

          <div v-if="ntfyTopics.length" class="mt-6">
            <div class="mb-3 text-sm font-medium text-slate-700">已保存的话题</div>
            <div class="grid gap-3 sm:grid-cols-2">
              <button
                v-for="item in ntfyTopics"
                :key="item.id"
                type="button"
                class="rounded-xl border p-4 text-left transition-colors"
                :class="selectedNtfyTopicID === item.id ? 'border-emerald-400 bg-emerald-50' : 'border-slate-200 bg-white hover:border-slate-300'"
                @click="selectedNtfyTopicID = item.id"
              >
                <div class="flex items-center justify-between gap-3">
                  <span class="font-mono text-sm text-slate-700">{{ item.topic }}</span>
                  <span v-if="selectedNtfyTopicID === item.id" class="text-xs font-medium text-emerald-600">当前发送</span>
                </div>
                <div class="mt-2 text-xs text-slate-400">{{ item.public_enabled ? '公开代理已开启' : '公开代理未开启' }}</div>
              </button>
            </div>
          </div>

          <div v-if="selectedNtfyTopic" class="mt-5 space-y-3">
            <div class="flex flex-col gap-2 rounded-xl bg-slate-50 p-4 sm:flex-row sm:items-center">
              <div class="min-w-0 flex-1">
                <div class="text-xs text-slate-400">完整订阅地址</div>
                <div class="break-all font-mono text-sm text-slate-600">{{ ntfySubscribeURL }}</div>
              </div>
              <NButton size="small" @click="copyText(ntfySubscribeURL, '订阅地址')">复制地址</NButton>
            </div>
            <div class="flex flex-col gap-2 rounded-xl bg-slate-50 p-4 sm:flex-row sm:items-center">
              <div class="min-w-0 flex-1">
                <div class="text-xs text-slate-400">话题名称</div>
                <div class="font-mono text-sm text-slate-600">{{ selectedNtfyTopic.topic }}</div>
              </div>
              <NButton size="small" @click="copyText(selectedNtfyTopic.topic, '话题')">复制话题</NButton>
            </div>
          </div>
        </section>

        <section v-if="selectedNtfyTopic" class="rounded-2xl border border-slate-100 bg-white p-5 md:p-6">
          <div class="flex items-center justify-between gap-4">
            <div>
              <h2 class="font-semibold text-slate-800">公开网页代理</h2>
              <p class="mt-1 text-sm text-slate-400">开启后，持有随机地址的人无需登录博客即可发送通知。每个地址限制每分钟 10 次。</p>
            </div>
            <NSwitch :value="selectedNtfyTopic.public_enabled" :loading="ntfyPublicSaving" @update:value="toggleNtfyPublic" />
          </div>
          <div v-if="selectedNtfyTopic.public_enabled" class="mt-4 flex flex-col gap-2 rounded-xl bg-amber-50 p-4 sm:flex-row sm:items-center">
            <div class="min-w-0 flex-1 break-all font-mono text-sm text-amber-800">{{ ntfyPublicURL }}</div>
            <NButton size="small" @click="copyText(ntfyPublicURL, '公开地址')">复制公开地址</NButton>
          </div>
        </section>

        <div v-if="!selectedNtfyTopic" class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
          发送功能已准备好，请先在上方保存话题并完成 blog 用户授权。
        </div>
        <NtfyPushForm :topic-id="selectedNtfyTopic?.id" :disabled="!selectedNtfyTopic" />
      </div>
    </NSpin>

    <NSpin v-else :show="loading">
      <div class="space-y-6">
        <section class="rounded-2xl border border-slate-100 bg-white p-5 md:p-6">
          <div class="mb-4">
            <h2 class="font-semibold text-slate-800">Bark 通知地址</h2>
            <p class="mt-1 text-sm text-slate-400">填写到设备 Key 为止，不要包含具体推送内容。</p>
          </div>
          <div class="flex flex-col gap-3 sm:flex-row">
            <NInput v-model:value="endpoint" placeholder="http://124.223.116.8/bark/你的设备Key" clearable />
            <NButton type="primary" :loading="saving" @click="saveEndpoint">保存地址</NButton>
          </div>
        </section>

        <section class="rounded-2xl border border-slate-100 bg-white p-5 md:p-6">
          <h2 class="mb-4 font-semibold text-slate-800">选择推送方式</h2>
          <NTabs v-model:value="mode" type="card" animated>
            <NTabPane v-for="item in modes" :key="item.key" :name="item.key" :tab="item.label" />
          </NTabs>

          <div class="mt-5 space-y-4">
            <div class="rounded-lg bg-indigo-50 px-4 py-3 text-sm text-indigo-700">
              {{ currentMode.description }}
            </div>
            <div v-if="mode === 'title'">
              <label class="mb-2 block text-sm font-medium text-slate-600">推送标题</label>
              <NInput v-model:value="title" maxlength="200" show-count placeholder="请输入推送标题" />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-slate-600">推送内容</label>
              <NInput
                v-model:value="content"
                type="textarea"
                :autosize="{ minRows: 4, maxRows: 9 }"
                maxlength="2000"
                show-count
                placeholder="请输入要发送到手机的消息"
              />
            </div>
            <div>
              <div class="mb-2 text-xs font-medium uppercase tracking-wide text-slate-400">请求预览</div>
              <div class="break-all rounded-lg bg-slate-50 px-4 py-3 font-mono text-xs text-slate-500">{{ previewURL }}</div>
            </div>
            <div class="flex justify-end">
              <NButton type="primary" size="large" :loading="sending" @click="send">发送通知</NButton>
            </div>
          </div>
        </section>
      </div>
    </NSpin>
  </div>
</template>
