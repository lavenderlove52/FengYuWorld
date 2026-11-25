<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { NSpin } from 'naive-ui'
import { getUserGridPlan, listUserGridSnapshots } from '@/api/gridPlans'
import { appBase } from '@/config/app'
import { computePlanMetrics, type PlanAssetStatus } from '@/core/gridplan/planMetrics'
import { formatPlanMoney, formatSignedMoney } from '@/core/gridplan/planFormat'
import { formatNumber, formatPercent, formatPlainPercent, FREQUENCY_LABEL } from '@/core/gridplan/format'
import LineChart, { type LinePoint } from '@/components/chart/LineChart.vue'
import DonutChart, { type DonutSlice } from '@/components/chart/DonutChart.vue'
import { copyToClipboard } from '@/utils/clipboard'
import type { GridPlanDetail, GridSnapshot } from '@/types/gridPlan'

/**
 * 个人主页里的投资计划详情：只读。
 * 数据全部来自 /users/:username/grid-plans/:id，可见性由后端判定，
 * 这里不做任何写操作，也不渲染输入框。
 * 视觉上跟随主页主题（--theme-* 变量），不再用 naive-ui 的浅色卡片。
 */

const props = withDefaults(defineProps<{ username: string; planId: number; backLabel?: string }>(), {
  backLabel: '返回计划列表',
})
const emit = defineEmits<{ (e: 'back'): void }>()

const plan = ref<GridPlanDetail | null>(null)
const snapshots = ref<GridSnapshot[]>([])
const loading = ref(true)
const error = ref('')

/** 分享链接固定指向公开直达页，而不是首页内联状态（那里地址栏不会变）。 */
const shareUrl = computed(
  () => `${window.location.origin}${appBase}u/${encodeURIComponent(props.username)}/plans/${props.planId}`,
)
const linkCopied = ref(false)
let copiedTimer: ReturnType<typeof setTimeout> | null = null

async function copyShareLink(): Promise<void> {
  try {
    await copyToClipboard(shareUrl.value)
    linkCopied.value = true
    if (copiedTimer) clearTimeout(copiedTimer)
    copiedTimer = setTimeout(() => {
      linkCopied.value = false
    }, 2000)
  } catch {
    // 复制失败就不提示成功，避免用户以为已经拿到链接
  }
}

onBeforeUnmount(() => {
  if (copiedTimer) clearTimeout(copiedTimer)
})

const metrics = computed(() => (plan.value ? computePlanMetrics(plan.value) : null))

const chartPoints = computed<LinePoint[]>(() =>
  snapshots.value.map((item) => ({
    date: item.snapshot_date,
    invested: item.invested,
    marketValue: item.market_value,
    pnl: item.pnl,
  })),
)

const lastSnapshotLabel = computed(() => {
  const last = snapshots.value[snapshots.value.length - 1]
  if (!last) return '还没有记录过净值'
  const days = Math.floor((Date.now() - new Date(`${last.snapshot_date}T00:00:00`).getTime()) / 86_400_000)
  if (days <= 0) return '最近记录：今天'
  return `最近记录：${days} 天前`
})

const DONUT_COLORS = ['#4f46e5', '#0ea5e9', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#14b8a6', '#f97316']

/** 与主人页口径一致：停用标的不计入；未填市值但有投入的标的单独一块。 */
const donutSlices = computed<DonutSlice[]>(() => {
  const rows = metrics.value?.assets ?? []
  const slices: DonutSlice[] = []
  let colorIndex = 0
  let filled = 0
  let missing = 0
  let unfilledInvested = 0
  for (const row of rows) {
    if (!row.asset.enabled) continue
    if (!row.hasMarketValue) {
      missing += 1
      unfilledInvested += Number(row.asset.invested) || 0
      continue
    }
    const value = Number(row.asset.market_value) || 0
    filled += value
    slices.push({
      label: row.asset.name || '未命名标的',
      value,
      color: DONUT_COLORS[colorIndex % DONUT_COLORS.length]!,
    })
    colorIndex += 1
  }
  if (filled === 0) return []
  if (missing > 0 && unfilledInvested > 0) {
    slices.push({ label: `未填市值（${missing} 个）`, value: unfilledInvested, color: '#cbd5e1' })
  }
  return slices
})

const donutCenter = computed(() => {
  const m = metrics.value
  if (!m || !m.marketValueFilled || !plan.value) return { value: '未填市值', label: '' }
  return { value: formatPlanMoney(m.marketValue, plan.value), label: '当前市值' }
})

const ASSET_STATUS_META: Record<PlanAssetStatus, { label: string; color: string }> = {
  invalid: { label: '参数异常', color: '#ef4444' },
  'gate-blocked': { label: '闸门限制', color: '#f59e0b' },
  'budget-blocked': { label: '预算受限', color: '#f59e0b' },
  'grid-triggered': { label: '网格触发', color: '#10b981' },
  'regular-due': { label: '定投到期', color: '#6366f1' },
  waiting: { label: '等待', color: 'currentColor' },
}

const STATUS_LABEL = {
  invested: '已投',
  voided: '已作废',
  triggered: '已触发',
  waiting: '等待',
} as const

const STATUS_COLOR = {
  invested: '#6366f1',
  voided: 'currentColor',
  triggered: '#10b981',
  waiting: 'currentColor',
} as const

const SURFACE = 'background: var(--theme-card-bg, #ffffff); border: 1px solid var(--theme-border, #e2e8f0)'

/** 涨跌色：红涨绿跌，选中深色/浅色卡都能读的中间色 */
function pnlStyle(value: number): Record<string, string> {
  if (value > 0) return { color: '#ef4444' }
  if (value < 0) return { color: '#10b981' }
  return { color: 'var(--theme-text, #64748b)', opacity: '0.7' }
}

/** 权重偏离：超配红、低配绿，和主人页口径一致。 */
function deviationStyle(value: number): Record<string, string> {
  if (value > 0) return { color: '#ef4444' }
  if (value < 0) return { color: '#10b981' }
  return { color: 'var(--theme-text, #64748b)', opacity: '0.7' }
}

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const res = await getUserGridPlan(props.username, props.planId)
    if (res.code !== 0 || !res.data) throw new Error(res.message || '计划不存在或无权查看')
    plan.value = res.data
    // days=0 表示全部历史，和主人页曲线默认口径一致
    const snap = await listUserGridSnapshots(props.username, props.planId, 0).catch(() => null)
    snapshots.value = snap?.code === 0 ? snap.data ?? [] : []
  } catch (err) {
    plan.value = null
    error.value = err instanceof Error ? err.message : '计划加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center gap-2 mb-4">
      <button type="button" class="back-link text-sm px-2 py-1 rounded-md transition-colors"
        style="color: var(--theme-text, #64748b)" @click="emit('back')">
        ← {{ backLabel }}
      </button>
      <span class="text-xs px-2 py-0.5 rounded"
        style="border: 1px solid var(--theme-border, #e2e8f0); color: var(--theme-text, #64748b); opacity: 0.7">
        只读
      </span>
      <button v-if="plan" type="button" class="back-link text-sm px-2 py-1 rounded-md transition-colors ml-auto"
        style="color: var(--theme-text, #64748b)" @click="copyShareLink">
        {{ linkCopied ? '已复制链接' : '复制分享链接' }}
      </button>
    </div>

    <NSpin :show="loading">
      <div v-if="error" class="text-sm py-10 text-center" style="color: var(--theme-text, #64748b); opacity: 0.7">
        {{ error }}
      </div>

      <template v-else-if="plan && metrics">
        <h2 class="text-lg sm:text-xl font-bold mb-1" style="color: var(--theme-text, #0f172a)">
          {{ plan.name }}
        </h2>
        <p class="text-sm mb-5" style="color: var(--theme-text, #64748b); opacity: 0.7">
          {{ plan.currency_name }}计价 · {{ plan.assets.length }} 个标的 ·
          {{ FREQUENCY_LABEL[plan.frequency] ?? plan.frequency }}
        </p>

        <div class="grid grid-cols-2 lg:grid-cols-5 gap-3 mb-5">
          <div class="rounded-xl p-4" :style="SURFACE">
            <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">总预算</div>
            <div class="text-lg font-semibold mt-1" style="color: var(--theme-text, #0f172a)">
              {{ formatPlanMoney(plan.total_budget, plan) }}
            </div>
            <div class="text-xs mt-1" style="color: var(--theme-text, #64748b); opacity: 0.65">
              已投入 {{ formatPlanMoney(metrics.invested, plan) }}
            </div>
          </div>
          <div class="rounded-xl p-4" :style="SURFACE">
            <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">当前市值</div>
            <div class="text-lg font-semibold mt-1" style="color: var(--theme-text, #0f172a)">
              {{ metrics.marketValueFilled ? formatPlanMoney(metrics.marketValue, plan) : '未填' }}
            </div>
          </div>
          <div class="rounded-xl p-4" :style="SURFACE">
            <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">总盈亏</div>
            <div class="text-lg font-semibold mt-1" :style="pnlStyle(metrics.pnl)">
              {{ metrics.marketValueFilled ? formatSignedMoney(metrics.pnl, plan) : '—' }}
            </div>
            <div class="text-xs mt-1" :style="pnlStyle(metrics.pnl)">
              收益率 {{ formatPercent(metrics.returnRate) }}
            </div>
          </div>
          <div class="rounded-xl p-4" :style="SURFACE">
            <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">组合加权回撤</div>
            <div class="text-lg font-semibold mt-1" style="color: var(--theme-text, #0f172a)">
              {{ formatPlainPercent(metrics.drawdown) }}
            </div>
            <div class="text-xs mt-1" style="color: var(--theme-text, #64748b); opacity: 0.65">
              闸门 {{ plan.gate_threshold }}%
            </div>
          </div>
          <div class="rounded-xl p-4" :style="SURFACE">
            <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">本期建议合计</div>
            <div class="text-lg font-semibold mt-1" style="color: var(--theme-text, #0f172a)">
              {{ formatPlanMoney(metrics.suggestionTotal, plan) }}
            </div>
            <div class="text-xs mt-1" style="color: var(--theme-text, #64748b); opacity: 0.65">
              理论 {{ formatPlanMoney(metrics.theoreticalTotal, plan) }}
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-[280px_minmax(0,1fr)] gap-3 mb-5 items-start">
          <div class="rounded-xl p-4" :style="SURFACE">
            <div class="font-medium mb-3" style="color: var(--theme-text, #0f172a)">资产占比</div>
            <DonutChart :slices="donutSlices" :center-value="donutCenter.value"
              :center-label="donutCenter.label" :size="180" />
            <div v-if="!donutSlices.length" class="text-xs mt-3"
              style="color: var(--theme-text, #64748b); opacity: 0.6">
              计划所有者填入标的「当前市值」后，这里会显示各标的的占比。
            </div>
          </div>

          <div class="rounded-xl p-4" :style="SURFACE">
            <div class="flex flex-wrap items-baseline justify-between gap-2 mb-3">
              <div class="font-medium" style="color: var(--theme-text, #0f172a)">收益变化曲线</div>
              <span class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.6">
                {{ lastSnapshotLabel }}
              </span>
            </div>
            <LineChart :points="chartPoints" :height="240" empty-hint="计划所有者还没有记录过净值" />
            <div class="text-xs mt-2" style="color: var(--theme-text, #64748b); opacity: 0.6">
              由计划所有者手动记录，全部历史。
            </div>
          </div>
        </div>

        <div class="flex flex-col gap-3">
          <div v-for="row in metrics.assets" :key="row.asset.id" class="rounded-xl p-4" :style="SURFACE">
            <div class="flex items-center gap-2 flex-wrap mb-3">
              <span class="font-medium" style="color: var(--theme-text, #0f172a)">{{ row.asset.name }}</span>
              <span v-if="row.asset.code" class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.6">
                {{ row.asset.code }}
              </span>
              <span v-if="row.asset.leveraged" class="text-xs font-medium" style="color: #ef4444">2×</span>
              <span class="inline-flex items-center gap-1 text-xs"
                :style="{ color: ASSET_STATUS_META[row.status].color, opacity: row.status === 'waiting' ? 0.65 : 1 }">
                <i class="inline-block w-1.5 h-1.5 rounded-full" style="background: currentColor" />
                {{ ASSET_STATUS_META[row.status].label }}
              </span>
              <span v-if="!row.asset.enabled" class="text-xs px-1.5 py-0.5 rounded"
                style="border: 1px solid var(--theme-border, #e2e8f0); color: var(--theme-text, #64748b); opacity: 0.75">
                已停用
              </span>
              <span v-if="!row.hasValueOrRealized" class="text-xs px-1.5 py-0.5 rounded"
                style="border: 1px solid color-mix(in srgb, #f59e0b 45%, transparent); color: #f59e0b">
                未填市值
              </span>
            </div>

            <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mb-3">
              <div>
                <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">当前 / 参考高点</div>
                <div class="text-sm font-medium mt-0.5" style="color: var(--theme-text, #0f172a)">
                  {{ formatNumber(row.asset.current_point) }} / {{ formatNumber(row.asset.high_point) }}
                </div>
              </div>
              <div>
                <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">当前回撤</div>
                <div class="text-sm font-medium mt-0.5" style="color: var(--theme-text, #0f172a)">
                  {{ formatPlainPercent(row.dd) }}
                </div>
              </div>
              <div>
                <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">目标权重</div>
                <div class="text-sm font-medium mt-0.5" style="color: var(--theme-text, #0f172a)">
                  {{ formatPlainPercent(row.asset.weight) }}
                </div>
                <div class="text-xs mt-0.5" style="color: var(--theme-text, #64748b); opacity: 0.6">
                  应投 {{ formatPlanMoney(row.targetBudget, plan) }}
                </div>
              </div>
              <div>
                <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">已投入占比</div>
                <div class="text-sm font-medium mt-0.5" style="color: var(--theme-text, #0f172a)">
                  {{ formatPlainPercent(row.actualWeight) }}
                </div>
                <div v-if="row.weightDeviation !== null" class="text-xs mt-0.5" :style="deviationStyle(row.weightDeviation)">
                  {{ row.weightDeviation > 0 ? '超配' : row.weightDeviation < 0 ? '低配' : '持平' }}
                  {{ formatPlainPercent(Math.abs(row.weightDeviation)) }}
                </div>
              </div>
              <div>
                <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">已投入 / 市值</div>
                <div class="text-sm font-medium mt-0.5" style="color: var(--theme-text, #0f172a)">
                  {{ formatPlanMoney(row.asset.invested, plan) }} /
                  {{ row.hasMarketValue ? formatPlanMoney(row.asset.market_value, plan) : '—' }}
                </div>
                <div v-if="row.asset.realized_out > 0" class="text-xs mt-0.5"
                  style="color: var(--theme-text, #64748b); opacity: 0.6">
                  累计卖出 {{ formatPlanMoney(row.asset.realized_out, plan) }}
                </div>
              </div>
              <div>
                <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">盈亏</div>
                <div class="text-sm font-medium mt-0.5"
                  :style="row.hasValueOrRealized ? pnlStyle(row.pnl) : { color: 'var(--theme-text, #64748b)', opacity: '0.7' }">
                  {{ row.hasValueOrRealized ? formatSignedMoney(row.pnl, plan) : '未填市值' }}
                </div>
                <div v-if="row.hasValueOrRealized" class="text-xs mt-0.5" :style="pnlStyle(row.pnl)">
                  {{ formatPercent(row.returnRate) }}
                </div>
              </div>
              <div>
                <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">目标预算余量</div>
                <div class="text-sm font-medium mt-0.5" style="color: var(--theme-text, #0f172a)">
                  {{ formatPlanMoney(row.remainingIndex, plan) }}
                </div>
              </div>
            </div>

            <div class="flex flex-wrap items-start justify-between gap-3 mb-3">
              <div>
                <div class="text-xs" style="color: var(--theme-text, #64748b); opacity: 0.65">当前建议投入</div>
                <div class="text-lg font-semibold mt-0.5" style="color: var(--theme-accent, #6366f1)">
                  {{ formatPlanMoney(row.suggestion, plan) }}
                </div>
              </div>
              <div class="text-xs max-w-[420px]" style="color: var(--theme-text, #64748b); opacity: 0.75; line-height: 1.6">
                {{ row.reason }}
              </div>
            </div>

            <!-- 只读网格：不放 NInputNumber，避免看起来像能改 -->
            <div class="overflow-x-auto">
              <table class="w-full text-xs" style="border-collapse: collapse; color: var(--theme-text, #0f172a)">
                <thead>
                  <tr style="opacity: 0.6">
                    <th class="text-left py-1 pr-3 font-medium">层级</th>
                    <th class="text-right py-1 pr-3 font-medium">回撤阈值</th>
                    <th class="text-right py-1 pr-3 font-medium">触发点位</th>
                    <th class="text-right py-1 pr-3 font-medium">资金比例</th>
                    <th class="text-right py-1 pr-3 font-medium">计划金额</th>
                    <th class="text-left py-1 font-medium">状态</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="layer in row.rows" :key="layer.layer.key"
                    :class="layer.voided ? 'opacity-40 line-through' : ''"
                    style="border-top: 1px solid var(--theme-border, rgba(148,163,184,.2))">
                    <td class="py-1 pr-3">{{ layer.layer.key }}</td>
                    <td class="py-1 pr-3 text-right">{{ Number(layer.layer.threshold) }}%</td>
                    <td class="py-1 pr-3 text-right">{{ formatNumber(layer.point) }}</td>
                    <td class="py-1 pr-3 text-right">{{ Number(layer.layer.allocation) }}%</td>
                    <td class="py-1 pr-3 text-right">{{ formatPlanMoney(layer.amount, plan) }}</td>
                    <td class="py-1">
                      <span class="inline-flex items-center gap-1.5">
                        <i class="inline-block w-1.5 h-1.5 rounded-full"
                          :style="{ background: STATUS_COLOR[layer.state], opacity: layer.state === 'invested' || layer.state === 'triggered' ? 1 : 0.5 }" />
                        <span :style="{ opacity: layer.state === 'invested' || layer.state === 'triggered' ? 1 : 0.6 }">
                          {{ STATUS_LABEL[layer.state] }}
                        </span>
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div v-if="!metrics.assets.length" class="rounded-xl p-4 text-center text-sm"
            :style="`${SURFACE}; color: var(--theme-text, #64748b); opacity: 0.7`">
            这份计划还没有标的。
          </div>
        </div>
      </template>
    </NSpin>
  </div>
</template>

<style scoped>
.back-link:hover {
  color: var(--theme-accent, #6366f1);
  background: color-mix(in srgb, var(--theme-accent, #6366f1) 12%, transparent);
}
</style>
