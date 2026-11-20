import { createGridPlan } from '@/api/gridPlans'
import type { SaveGridPlanPayload, GridLayerDTO, GridAssetPayload } from '@/types/gridPlan'

/**
 * 旧版 localStorage 计划的迁移。
 * 旧结构只存在于浏览器里（gridplan.state.v1），新版已全部走后端，
 * 这里做一次性转换，成功后删掉旧键，并记下已导入的计划，重复点不会建重。
 */

export const LEGACY_STORAGE_KEY = 'gridplan.state.v1'
export const MIGRATION_MARKER_KEY = 'gridplan.migrated.v1'

export interface MigrationResult {
  ok: boolean
  created: number
  skipped: number
  message: string
}

interface LegacyLayer {
  id: string
  threshold: number
  allocation: number
}

interface LegacyAsset {
  id: string
  name: string
  code: string
  leveraged: boolean
  enabled: boolean
  current: number | null
  high: number | null
  weight: number
  invested: number
  base: number
  grid: LegacyLayer[]
  executedLayerIds: string[]
}

interface LegacyMarket {
  key: string
  name: string
  currency: string
  currencyName: string
  unit: number
  totalBudget: number
  frequency: string
  nextDate: string
  gateThreshold: number
  gateMode: string
  weightMethod: string
  gridAggregation: string
  unallocatedInvested: number
  assets: LegacyAsset[]
}

export function readLegacyState(): Record<string, LegacyMarket> | null {
  try {
    const raw = window.localStorage.getItem(LEGACY_STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as { markets?: Record<string, LegacyMarket> }
    if (!parsed || typeof parsed !== 'object' || !parsed.markets) return null
    const markets: Record<string, LegacyMarket> = {}
    for (const [key, market] of Object.entries(parsed.markets)) {
      if (market && Array.isArray(market.assets)) markets[key] = { ...market, key: market.key || key }
    }
    return Object.keys(markets).length ? markets : null
  } catch {
    return null
  }
}

function readMarker(): string[] {
  try {
    const raw = window.localStorage.getItem(MIGRATION_MARKER_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as { imported?: string[] }
    return Array.isArray(parsed.imported) ? parsed.imported : []
  } catch {
    return []
  }
}

function writeMarker(imported: string[]): void {
  window.localStorage.setItem(MIGRATION_MARKER_KEY, JSON.stringify({ imported, at: new Date().toISOString() }))
}

const FREQUENCIES = new Set(['weekly', 'biweekly', 'monthly'])
const GATE_MODES = new Set(['alert', 'grid-only', 'block-all'])
const WEIGHT_METHODS = new Set(['target-weight', 'invested-weight'])
const AGGREGATIONS = new Set(['highest', 'cumulative'])

function num(value: unknown, fallback = 0): number {
  const n = Number(value)
  return Number.isFinite(n) ? n : fallback
}

/** 网格层改成 g1..gN 的位置键，同时把已执行标记翻译过去，避免旧 id 超过后端 40 字限制。 */
function convertGrid(asset: LegacyAsset): { grid: GridLayerDTO[]; executed: string[] } {
  const keyMap = new Map<string, string>()
  const grid: GridLayerDTO[] = []
  let prev = -1
  for (const layer of asset.grid ?? []) {
    const threshold = num(layer.threshold)
    const allocation = num(layer.allocation)
    // 后端要求阈值严格递增且大于 0，不合法的层直接丢弃，而不是整份计划导入失败
    if (threshold <= 0 || threshold <= prev) continue
    prev = threshold
    const key = `g${grid.length + 1}`
    if (layer.id) keyMap.set(layer.id, key)
    grid.push({ key, threshold, allocation: Math.max(0, Math.min(100, allocation)), voided: false })
  }
  const total = grid.reduce((sum, layer) => sum + layer.allocation, 0)
  if (total > 100) {
    const scale = 100 / total
    grid.forEach((layer) => {
      layer.allocation = Math.round(layer.allocation * scale * 100) / 100
    })
  }
  const executed = (asset.executedLayerIds ?? [])
    .map((id) => keyMap.get(id))
    .filter((key): key is string => Boolean(key))
  return { grid, executed }
}

function convertMarket(market: LegacyMarket): SaveGridPlanPayload {
  let weightLeft = 100
  const assets: GridAssetPayload[] = []
  for (const asset of market.assets) {
    const { grid, executed } = convertGrid(asset)
    const weight = Math.max(0, Math.min(weightLeft, num(asset.weight)))
    if (asset.enabled !== false) weightLeft -= weight
    assets.push({
      name: (asset.name || '').trim() || '未命名标的',
      code: asset.code ?? '',
      leveraged: asset.leveraged === true,
      enabled: asset.enabled !== false,
      current_point: asset.current === null || asset.current === undefined ? null : Math.max(0, num(asset.current)),
      high_point: asset.high === null || asset.high === undefined ? null : Math.max(0, num(asset.high)),
      weight,
      invested: Math.max(0, num(asset.invested)),
      // 旧版没有市值字段，导入后保持 0（界面显示「未填市值」），由用户自己补
      market_value: 0,
      base_amount: Math.max(0, num(asset.base)),
      grid,
      executed_layers: executed,
    })
  }

  return {
    name: (market.name || '旧版计划').trim() || '旧版计划',
    market: market.key === 'us' ? 'us' : 'cn',
    currency: market.currency || '¥',
    currency_name: market.currencyName || '人民币',
    unit: num(market.unit, 100) > 0 ? num(market.unit, 100) : 100,
    total_budget: Math.max(0, num(market.totalBudget)),
    frequency: (FREQUENCIES.has(market.frequency) ? market.frequency : 'monthly') as SaveGridPlanPayload['frequency'],
    next_date: typeof market.nextDate === 'string' ? market.nextDate : '',
    gate_threshold: Math.max(0, Math.min(100, num(market.gateThreshold, 20))),
    gate_mode: (GATE_MODES.has(market.gateMode) ? market.gateMode : 'grid-only') as SaveGridPlanPayload['gate_mode'],
    weight_method: (WEIGHT_METHODS.has(market.weightMethod)
      ? market.weightMethod
      : 'target-weight') as SaveGridPlanPayload['weight_method'],
    grid_aggregation: (AGGREGATIONS.has(market.gridAggregation)
      ? market.gridAggregation
      : 'highest') as SaveGridPlanPayload['grid_aggregation'],
    unallocated_invested: Math.max(0, num(market.unallocatedInvested)),
    assets,
  }
}

/**
 * 把旧数据导入成后端计划。
 * 已导入过的市场记在 marker 里，重复调用只补建失败的那部分，不会重复建计划。
 */
export async function migrateLegacyPlans(): Promise<MigrationResult> {
  const markets = readLegacyState()
  if (!markets) return { ok: false, created: 0, skipped: 0, message: '没有检测到旧版本地计划。' }

  const imported = readMarker()
  let created = 0
  let skipped = 0
  const errors: string[] = []

  for (const [key, market] of Object.entries(markets)) {
    if (imported.includes(key)) {
      skipped += 1
      continue
    }
    try {
      const res = await createGridPlan(convertMarket(market))
      if (res.code !== 0 || !res.data) {
        errors.push(`${market.name || key}：${res.message || '创建失败'}`)
        continue
      }
      imported.push(key)
      writeMarker(imported)
      created += 1
    } catch (err) {
      errors.push(`${market.name || key}：${err instanceof Error ? err.message : '创建失败'}`)
    }
  }

  if (errors.length) {
    // 有失败时保留旧键，用户修完还能再试；已成功的那部分靠 marker 去重
    return { ok: false, created, skipped, message: `部分计划导入失败：${errors.join('；')}` }
  }

  window.localStorage.removeItem(LEGACY_STORAGE_KEY)
  return {
    ok: true,
    created,
    skipped,
    message: created ? `已导入 ${created} 份计划。` : '旧计划已经导入过了。',
  }
}
