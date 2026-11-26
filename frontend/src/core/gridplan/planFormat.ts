import type { GridPlanDetail } from '@/types/gridPlan'

/**
 * 计划金额：最多两位小数（后端 numeric(18,2) 的口径）。
 * 整数不带 .00，有分/厘的金额（基金净值、小数买入）也能显示出来。
 */
export function formatPlanMoney(value: number, plan: Pick<GridPlanDetail, 'currency' | 'market'>): string {
  const safe = Number.isFinite(value) ? value : 0
  return `${plan.currency}${safe.toLocaleString('zh-CN', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  })}`
}

export function formatSignedMoney(
  value: number,
  plan: Pick<GridPlanDetail, 'currency' | 'market'>,
): string {
  const sign = value > 0 ? '+' : value < 0 ? '-' : ''
  return `${sign}${formatPlanMoney(Math.abs(value), plan)}`
}
