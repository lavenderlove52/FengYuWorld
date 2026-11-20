import type { GridPlanDetail } from '@/types/gridPlan'

/** 计划金额：A 股按整数，其它市场保留两位。 */
export function formatPlanMoney(value: number, plan: Pick<GridPlanDetail, 'currency' | 'market'>): string {
  const decimals = plan.market === 'cn' ? 0 : 2
  const safe = Number.isFinite(value) ? value : 0
  return `${plan.currency}${safe.toLocaleString('zh-CN', {
    minimumFractionDigits: 0,
    maximumFractionDigits: decimals,
  })}`
}

export function formatSignedMoney(
  value: number,
  plan: Pick<GridPlanDetail, 'currency' | 'market'>,
): string {
  const sign = value > 0 ? '+' : value < 0 ? '-' : ''
  return `${sign}${formatPlanMoney(Math.abs(value), plan)}`
}
