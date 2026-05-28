<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { NButton, NInputNumber, NDatePicker, useMessage, useDialog } from 'naive-ui'
import { listInviteCodes, generateInviteCodes, deleteInviteCode } from '@/api/invite'
import type { InviteCode } from '@/api/invite'

const message = useMessage()
const dialog = useDialog()

const codes = ref<InviteCode[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const generating = ref(false)
const generateCount = ref(5)
const generateExpiry = ref<number | null>(null) // timestamp in ms

const load = async () => {
  loading.value = true
  try {
    const res = await listInviteCodes(page.value, 20)
    if (res.code === 0 && res.data) {
      codes.value = res.data.items
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

onMounted(load)

const generate = async () => {
  if (generateCount.value < 1 || generateCount.value > 50) {
    message.warning('生成数量需在 1-50 之间')
    return
  }
  generating.value = true
  try {
    const expiresAt = generateExpiry.value
      ? new Date(generateExpiry.value).toISOString()
      : undefined
    const res = await generateInviteCodes(generateCount.value, expiresAt)
    if (res.code === 0) {
      message.success(`已生成 ${generateCount.value} 个邀请码`)
      await load()
    } else {
      message.error(res.message || '生成失败')
    }
  } finally {
    generating.value = false
  }
}

const remove = (code: InviteCode) => {
  dialog.warning({
    title: '确认作废',
    content: `邀请码「${code.code}」作废后不可恢复，确定继续？`,
    positiveText: '作废',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res = await deleteInviteCode(code.id)
      if (res.code === 0) {
        message.success('已作废')
        load()
      } else {
        message.error(res.message || '操作失败')
      }
    },
  })
}

const copyCode = (code: string) => {
  navigator.clipboard.writeText(code)
  message.success(`已复制：${code}`)
}

const copyAll = () => {
  const unused = codes.value.filter((c) => !c.is_used).map((c) => c.code)
  if (!unused.length) { message.warning('没有未使用的邀请码'); return }
  navigator.clipboard.writeText(unused.join('\n'))
  message.success(`已复制 ${unused.length} 个未使用的邀请码`)
}

const formatDate = (d?: string) => d?.slice(0, 10) ?? '—'
const isExpired = (c: InviteCode) => !!c.expires_at && new Date(c.expires_at) < new Date()

const stats = computed(() => ({
  total: codes.value.length,
  unused: codes.value.filter((c) => !c.is_used && !isExpired(c)).length,
  used: codes.value.filter((c) => c.is_used).length,
  expired: codes.value.filter((c) => !c.is_used && isExpired(c)).length,
}))
</script>

<template>
  <div>
    <!-- 页头 -->
    <div class="flex items-start justify-between mb-6">
      <div>
        <div class="flex items-center gap-2 mb-1">
          <span class="text-xs font-semibold tracking-widest text-slate-400 uppercase">Admin</span>
        </div>
        <h1 class="text-2xl font-bold text-slate-800">邀请码管理</h1>
        <p class="text-sm text-slate-400 mt-1">注册新用户需要有效的邀请码，每个码只能使用一次</p>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-4 gap-4 mb-8">
      <div v-for="(item, key) in [
        { label: '总数', value: stats.total, cls: 'text-slate-700' },
        { label: '可用', value: stats.unused, cls: 'text-emerald-600' },
        { label: '已使用', value: stats.used, cls: 'text-blue-500' },
        { label: '已过期', value: stats.expired, cls: 'text-red-400' },
      ]" :key="key"
        class="bg-white border border-slate-100 rounded-xl p-4 text-center"
      >
        <div class="text-2xl font-bold" :class="item.cls">{{ item.value }}</div>
        <div class="text-xs text-slate-400 mt-1">{{ item.label }}</div>
      </div>
    </div>

    <!-- 生成区 -->
    <div class="bg-white border border-slate-100 rounded-xl p-5 mb-6">
      <h2 class="text-sm font-semibold text-slate-700 mb-4">批量生成邀请码</h2>
      <div class="flex items-end gap-4 flex-wrap">
        <div>
          <div class="text-xs text-slate-400 mb-1.5">生成数量（1-50）</div>
          <NInputNumber
            v-model:value="generateCount"
            :min="1"
            :max="50"
            style="width: 120px"
          />
        </div>
        <div>
          <div class="text-xs text-slate-400 mb-1.5">过期时间（可选，不填则永不过期）</div>
          <NDatePicker
            v-model:value="generateExpiry"
            type="date"
            clearable
            placeholder="选择过期日期"
            style="width: 180px"
          />
        </div>
        <NButton type="primary" :loading="generating" @click="generate">
          生成邀请码
        </NButton>
      </div>
    </div>

    <!-- 列表 -->
    <div class="bg-white border border-slate-100 rounded-xl overflow-hidden">
      <div class="flex items-center justify-between px-5 py-3 border-b border-slate-100">
        <h2 class="text-sm font-semibold text-slate-700">邀请码列表</h2>
        <button
          class="text-xs px-3 py-1.5 border border-slate-200 rounded-lg text-slate-500 hover:border-indigo-300 hover:text-indigo-500 transition"
          @click="copyAll"
        >复制全部可用</button>
      </div>

      <div v-if="loading" class="text-center py-12 text-slate-400 text-sm">加载中...</div>
      <div v-else-if="!codes.length" class="text-center py-16 text-slate-300">
        <p class="text-sm">暂无邀请码，点击上方按钮生成</p>
      </div>

      <div v-else>
        <div
          v-for="code in codes"
          :key="code.id"
          class="group flex items-center gap-4 px-5 py-3.5 border-b border-slate-50 hover:bg-slate-50/60 transition"
        >
          <!-- 状态点 -->
          <div class="w-2 h-2 rounded-full flex-shrink-0"
            :class="code.is_used ? 'bg-blue-400' : isExpired(code) ? 'bg-red-300' : 'bg-emerald-400'"
          />

          <!-- 邀请码 -->
          <div class="font-mono text-sm font-semibold text-slate-800 w-28 flex-shrink-0 tracking-widest">
            {{ code.code }}
          </div>

          <!-- 状态标签 -->
          <span class="text-[11px] px-2 py-0.5 rounded-full border flex-shrink-0"
            :class="code.is_used
              ? 'bg-blue-50 text-blue-500 border-blue-100'
              : isExpired(code)
                ? 'bg-red-50 text-red-400 border-red-100'
                : 'bg-emerald-50 text-emerald-600 border-emerald-100'"
          >
            {{ code.is_used ? '已使用' : isExpired(code) ? '已过期' : '可用' }}
          </span>

          <!-- 使用信息 -->
          <div class="flex-1 text-xs text-slate-400 min-w-0">
            <span v-if="code.is_used && code.used_by_user">
              由 <span class="text-slate-600 font-medium">{{ code.used_by_user.nickname }}</span>
              (@{{ code.used_by_user.username }}) 于 {{ formatDate(code.used_at) }} 使用
            </span>
            <span v-else-if="code.expires_at">
              过期：{{ formatDate(code.expires_at) }}
            </span>
            <span v-else>永不过期</span>
          </div>

          <!-- 创建时间 -->
          <span class="text-xs text-slate-300 flex-shrink-0">{{ formatDate(code.created_at) }}</span>

          <!-- 操作按钮 -->
          <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition flex-shrink-0">
            <button
              v-if="!code.is_used && !isExpired(code)"
              class="text-xs px-2.5 py-1 rounded-lg border border-slate-200 text-slate-500 hover:border-indigo-300 hover:text-indigo-500 transition"
              @click="copyCode(code.code)"
            >复制</button>
            <button
              v-if="!code.is_used"
              class="text-xs px-2.5 py-1 rounded-lg border border-red-100 text-red-400 hover:bg-red-50 hover:border-red-300 transition"
              @click="remove(code)"
            >作废</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
