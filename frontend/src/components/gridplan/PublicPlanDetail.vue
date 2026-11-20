<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { NButton, NCard, NSpin, NTag } from 'naive-ui'
import { getUserGridPlan, listUserGridSnapshots } from '@/api/gridPlans'
import { computePlanMetrics } from '@/core/gridplan/planMetrics'
import { formatPlanMoney, formatSignedMoney } from '@/core/gridplan/planFormat'
import { formatNumber, formatPercent, formatPlainPercent, FREQUENCY_LABEL } from '@/core/gridplan/format'
import LineChart, { type LinePoint } from '@/components/chart/LineChart.vue'
import type { GridPlanDetail, GridSnapshot } from '@/types/gridPlan'

/**
 * 个人主页里的投资计划详情：只读。
 * 数据全部来自 /users/:username/grid-plans/:id，可见性由后端判定，
 * 这里不做任何写操作，也不渲染输入框。
 */

const props = defineProps<{ username: string; planId: number }>()
const emit = defineEmits<{ (e: 'back'): void }>()

const plan = ref<GridPlanDetail | null>(null)
const snapshots = ref<GridSnapshot[]>([])
const loading = ref(true)
const error = ref('')

const metrics = computed(() => (plan.value ? computePlanMetrics(plan.value) : null))

const chartPoints = computed<LinePoint[]>(() =>
  snapshots.value.map((item) => ({
    date: item.snapshot_date,
    invested: item.invested,
    marketValue: item.market_value,
    pnl: item.pnl,
  })),
)

const STATUS_LABEL = {
  invested: '已投',
  voided: '已作废',
  triggered: '已触发',
  waiting: '等待',
} as const

const STATUS_TYPE = {
  invested: 'info',
  voided: 'default',
  triggered: 'success',
  waiting: 'default',
} as const

function pnlClass(value: number): string {
  if (value > 0) return 'text-red-600'
  if (value < 0) return 'text-emerald-600'
  return 'text-slate-500'
}

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const res = await getUserGridPlan(props.username, props.planId)
    if (res.code !== 0 || !res.data) throw new Error(res.message || '计划不存在或无权查看')
    plan.value = res.data
    const snap = await listUserGridSnapshots(props.username, props.planId, 180).catch(() => null)
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
      <NButton size="small" quaternary @click="emit('back')">← 返回计划列表</NButton>
      <NTag size="tiny" type="info">只读</NTag>
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

        <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-5">
          <NCard size="small">
            <div class="text-xs text-slate-500">总预算</div>
            <div class="text-lg font-semibold mt-1">{{ formatPlanMoney(plan.total_budget, plan) }}</div>
            <div class="text-xs text-slate-500 mt-1">已投入 {{ formatPlanMoney(metrics.invested, plan) }}</div>
          </NCard>
          <NCard size="small">
            <div class="text-xs text-slate-500">当前市值</div>
            <div class="text-lg font-semibold mt-1">
              {{ metrics.marketValueFilled ? formatPlanMoney(metrics.marketValue, plan) : '未填' }}
            </div>
          </NCard>
          <NCard size="small">
            <div class="text-xs text-slate-500">总盈亏</div>
            <div class="text-lg font-semibold mt-1" :class="pnlClass(metrics.pnl)">
              {{ metrics.marketValueFilled ? formatSignedMoney(metrics.pnl, plan) : '—' }}
            </div>
            <div class="text-xs mt-1" :class="pnlClass(metrics.pnl)">
              收益率 {{ formatPercent(metrics.returnRate) }}
            </div>
          </NCard>
          <NCard size="small">
            <div class="text-xs text-slate-500">组合加权回撤</div>
            <div class="text-lg font-semibold mt-1">{{ formatPlainPercent(metrics.drawdown) }}</div>
            <div class="text-xs text-slate-500 mt-1">闸门 {{ plan.gate_threshold }}%</div>
          </NCard>
        </div>

        <div class="flex flex-col gap-3">
          <NCard v-for="row in metrics.assets" :key="row.asset.id" size="small">
            <template #header>
              <div class="flex items-center gap-2 flex-wrap">
                <span class="font-medium">{{ row.asset.name }}</span>
                <span v-if="row.asset.code" class="text-xs text-slate-400">{{ row.asset.code }}</span>
                <NTag v-if="row.asset.leveraged" size="tiny" type="warning">2×</NTag>
              </div>
            </template>

            <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mb-3">
              <div>
                <div class="text-xs text-slate-500">目标权重</div>
                <div class="text-sm font-medium mt-0.5">{{ row.asset.weight }}%</div>
              </div>
              <div>
                <div class="text-xs text-slate-500">当前回撤</div>
                <div class="text-sm font-medium mt-0.5">{{ formatPlainPercent(row.dd) }}</div>
              </div>
              <div>
                <div class="text-xs text-slate-500">已投入</div>
                <div class="text-sm font-medium mt-0.5">{{ formatPlanMoney(row.asset.invested, plan) }}</div>
              </div>
              <div>
                <div class="text-xs text-slate-500">当前市值</div>
                <div class="text-sm font-medium mt-0.5" :class="pnlClass(row.pnl)">
                  {{ row.hasMarketValue ? formatPlanMoney(row.asset.market_value, plan) : '未填' }}
                </div>
              </div>
            </div>

            <!-- 只读网格：不放 NInputNumber，避免看起来像能改 -->
            <div class="overflow-x-auto">
              <table class="w-full text-xs" style="border-collapse: collapse">
                <thead>
                  <tr class="text-slate-500">
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
                      <NTag size="tiny" :type="STATUS_TYPE[layer.state]">{{ STATUS_LABEL[layer.state] }}</NTag>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </NCard>

          <NCard v-if="!metrics.assets.length" size="small">
            <div class="text-center text-sm text-slate-500 py-6">这份计划还没有标的。</div>
          </NCard>
        </div>

        <NCard size="small" class="mt-4" title="收益变化曲线">
          <LineChart :points="chartPoints" :height="260" />
          <div class="text-xs text-slate-400 mt-2">
            由计划所有者手动记录，最近 180 天。
          </div>
        </NCard>
      </template>
    </NSpin>
  </div>
</template>
