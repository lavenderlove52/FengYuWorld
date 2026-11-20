import type { Frequency } from '@/types/gridPlan'

const LOCALE = 'zh-CN'

export function formatNumber(value: number | null, maximumFractionDigits = 2): string {
  if (value === null || !Number.isFinite(value)) return '—'
  return value.toLocaleString(LOCALE, { maximumFractionDigits })
}

export function formatPercent(value: number | null, digits = 2): string {
  if (value === null || !Number.isFinite(value)) return '—'
  const sign = value > 0 ? '+' : ''
  return `${sign}${value.toFixed(digits)}%`
}

export function formatPlainPercent(value: number | null, digits = 2): string {
  if (value === null || !Number.isFinite(value)) return '—'
  return `${value.toFixed(digits)}%`
}

export const FREQUENCY_LABEL: Record<Frequency, string> = {
  weekly: '每周',
  biweekly: '双周',
  monthly: '每月',
}

export function todayISO(): string {
  const now = new Date()
  const offset = now.getTimezoneOffset() * 60_000
  return new Date(now.getTime() - offset).toISOString().slice(0, 10)
}

export function roundDownToUnit(value: number, unit: number): number {
  const safeUnit = unit > 0 ? unit : 1
  if (!Number.isFinite(value) || value <= 0) return 0
  return Math.floor(value / safeUnit + 1e-9) * safeUnit
}
