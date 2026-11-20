<script setup lang="ts">
import { NTag } from 'naive-ui'
import { formatPlanMoney, formatSignedMoney } from '@/core/gridplan/planFormat'
import type { GridPlanSummary } from '@/types/gridPlan'

/** 个人主页里的计划汇总卡片：只读，点击进入详情。 */
const props = defineProps<{ plan: GridPlanSummary }>()
const emit = defineEmits<{ (e: 'open'): void }>()

const VISIBILITY_LABEL: Record<string, string> = {
  public: '公开',
  login: '登录可见',
  friends: '好友可见',
  specified: '指定查看',
  private: '仅自己',
}

function pnlClass(value: number): string {
  if (value > 0) return 'text-red-600'
  if (value < 0) return 'text-emerald-600'
  return 'text-slate-500'
}
</script>

<template>
  <button
    type="button"
    class="theme-plan-card w-full text-left rounded-xl p-4 transition"
    style="
      border: 1px solid var(--theme-border, #e2e8f0);
      background: var(--theme-card-bg, #ffffff);
    "
    @click="emit('open')"
  >
    <div class="flex items-center gap-2 min-w-0 mb-3">
      <span class="font-semibold truncate" style="color: var(--theme-text, #0f172a)">
        {{ props.plan.name }}
      </span>
      <NTag size="tiny">{{ props.plan.currency_name }}</NTag>
      <NTag size="tiny" type="info">{{ VISIBILITY_LABEL[props.plan.visibility] ?? props.plan.visibility }}</NTag>
    </div>

    <div class="grid grid-cols-2 gap-3 text-sm">
      <div>
        <div class="text-xs" style="color: var(--theme-text, #94a3b8); opacity: 0.8">总预算</div>
        <div class="font-medium mt-0.5" style="color: var(--theme-text, #0f172a)">
          {{ formatPlanMoney(props.plan.total_budget, props.plan) }}
        </div>
      </div>
      <div>
        <div class="text-xs" style="color: var(--theme-text, #94a3b8); opacity: 0.8">已投入</div>
        <div class="font-medium mt-0.5" style="color: var(--theme-text, #0f172a)">
          {{ formatPlanMoney(props.plan.invested, props.plan) }}
        </div>
      </div>
      <div>
        <div class="text-xs" style="color: var(--theme-text, #94a3b8); opacity: 0.8">当前市值</div>
        <div class="font-medium mt-0.5" style="color: var(--theme-text, #0f172a)">
          {{ props.plan.market_value > 0 ? formatPlanMoney(props.plan.market_value, props.plan) : '未填' }}
        </div>
      </div>
      <div>
        <div class="text-xs" style="color: var(--theme-text, #94a3b8); opacity: 0.8">盈亏</div>
        <div class="font-medium mt-0.5" :class="pnlClass(props.plan.pnl)">
          {{ props.plan.market_value > 0 ? formatSignedMoney(props.plan.pnl, props.plan) : '—' }}
        </div>
      </div>
    </div>

    <div class="text-xs mt-3" style="color: var(--theme-text, #94a3b8); opacity: 0.7">
      {{ props.plan.asset_count }} 个标的 ·
      {{ props.plan.market === 'us' ? '美股' : 'A 股' }}
    </div>
  </button>
</template>

<style scoped>
.theme-plan-card:hover {
  border-color: var(--theme-accent, #6366f1);
  transform: translateY(-1px);
}
</style>
