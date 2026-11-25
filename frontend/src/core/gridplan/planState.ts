import { computed, reactive, readonly } from 'vue'
import {
  addGridExecution as apiAddExecution,
  getGridPlan,
  recordGridSnapshot as apiRecordSnapshot,
  undoLastGridExecution as apiUndoExecution,
  updateGridPlan as apiUpdatePlan,
} from '@/api/gridPlans'
import type {
  Frequency,
  GateMode,
  GridAggregation,
  GridAssetDTO,
  GridAssetPayload,
  GridExecution,
  GridLayerDTO,
  GridPlanDetail,
  GridSnapshot,
  PlanVisibility,
  SaveGridPlanPayload,
  WeightMethod,
} from '@/types/gridPlan'
import { computePlanMetrics } from './planMetrics'
import { gridForIndex, gridLayers, type PlanTemplate } from './templates'

/**
 * 计划详情页的单计划状态。
 * 所有编辑只改本地并置 dirty，点「保存」才整体 PUT。
 */

/** 版本冲突：别处已经改过这份计划，本地改动还在，由页面决定是刷新还是放弃。 */
export class PlanConflictError extends Error {
  constructor(message = '这份计划已在别处被修改') {
    super(message)
    this.name = 'PlanConflictError'
  }
}

export interface PlanState {
  plan: GridPlanDetail | null
  loading: boolean
  saving: boolean
  dirty: boolean
  error: string
  lastSavedAt: string
  conflict: boolean
}

export const planState = reactive<PlanState>({
  plan: null,
  loading: false,
  saving: false,
  dirty: false,
  error: '',
  lastSavedAt: '',
  conflict: false,
})

export const planReadonly = readonly(planState)

export const planMetrics = computed(() => (planState.plan ? computePlanMetrics(planState.plan) : null))

export function resetPlanState(): void {
  planState.plan = null
  planState.loading = false
  planState.saving = false
  planState.dirty = false
  planState.error = ''
  planState.lastSavedAt = ''
  planState.conflict = false
}

export async function loadPlan(id: number): Promise<void> {
  planState.loading = true
  planState.error = ''
  try {
    const res = await getGridPlan(id)
    if (res.code !== 0 || !res.data) throw new Error(res.message || '加载计划失败')
    planState.plan = normalizePlan(res.data)
    planState.dirty = false
    planState.conflict = false
    planState.lastSavedAt = res.data.updated_at
  } catch (err) {
    planState.plan = null
    planState.error = err instanceof Error ? err.message : '加载计划失败'
    throw err
  } finally {
    planState.loading = false
  }
}

export async function savePlan(): Promise<GridPlanDetail> {
  if (!planState.plan) throw new Error('没有可保存的计划')
  planState.saving = true
  planState.error = ''
  try {
    const res = await apiUpdatePlan(planState.plan.id, toPayload(planState.plan))
    if (res.code !== 0 || !res.data) throw new Error(res.message || '保存失败')
    // 后端会重建标的（id 变化），用返回值覆盖本地，避免之后记录执行时 asset_id 对不上
    planState.plan = normalizePlan(res.data)
    planState.dirty = false
    planState.conflict = false
    planState.lastSavedAt = res.data.updated_at
    return res.data
  } catch (err) {
    // dirty 保持为 true，本地改动不能丢
    const conflict = isConflictError(err)
    planState.conflict = conflict
    planState.error = err instanceof Error ? err.message : '保存失败'
    if (conflict) throw new PlanConflictError(planState.error)
    throw err
  } finally {
    planState.saving = false
  }
}

/** 后端 409 / 业务码 40900 都算版本冲突。 */
function isConflictError(err: unknown): boolean {
  const anyErr = err as { response?: { status?: number; data?: { code?: number } }; code?: number; name?: string }
  if (anyErr?.name === 'PlanConflictError') return true
  if (anyErr?.response?.status === 409) return true
  if (anyErr?.response?.data?.code === 40900) return true
  if (anyErr?.code === 40900) return true
  // http 拦截器把后端 message 包成了普通 Error，按文案兜底
  return err instanceof Error && err.message.includes('已在别处被修改')
}

export async function discardChanges(): Promise<void> {
  if (!planState.plan) return
  await loadPlan(planState.plan.id)
}

export async function recordSnapshot(date?: string): Promise<GridSnapshot> {
  if (!planState.plan) throw new Error('没有可记录的计划')
  if (planState.dirty) throw new Error('请先保存计划，再记录净值')
  const res = await apiRecordSnapshot(planState.plan.id, date)
  if (res.code !== 0 || !res.data) throw new Error(res.message || '记录净值失败')
  return res.data
}

export async function recordExecution(payload: {
  assetId: number
  amount: number
  layerKeys: string[]
  wasRegular: boolean
  /** 成交点位快照，用于买卖点图 */
  point?: number | null
  note?: string
}): Promise<GridExecution> {
  if (!planState.plan) throw new Error('没有可记录的计划')
  if (planState.dirty) throw new Error('请先保存计划，再记录执行')
  const res = await apiAddExecution(planState.plan.id, {
    asset_id: payload.assetId,
    amount: payload.amount,
    point: payload.point ?? null,
    layer_keys: payload.layerKeys,
    was_regular: payload.wasRegular,
    note: payload.note ?? '',
  })
  if (res.code !== 0 || !res.data) throw new Error(res.message || '记录执行失败')
  await loadPlan(planState.plan.id)
  return res.data
}

export async function undoLastExecution(): Promise<GridExecution> {
  if (!planState.plan) throw new Error('没有可撤销的计划')
  const res = await apiUndoExecution(planState.plan.id)
  if (res.code !== 0 || !res.data) throw new Error(res.message || '撤销失败')
  await loadPlan(planState.plan.id)
  return res.data
}

/**
 * 记录一次卖出：只累加「累计卖出到账」，后端不会动已投入和网格层。
 * restoreLayerKeys 是本地编辑——把这些层恢复成未执行，调用方需要提示用户保存。
 */
export async function recordSell(payload: {
  assetId: number
  amount: number
  point?: number | null
  note?: string
  restoreLayerKeys?: string[]
}): Promise<{ exec: GridExecution; restored: boolean }> {
  if (!planState.plan) throw new Error('没有可记录的计划')
  if (planState.dirty) throw new Error('请先保存计划，再记录卖出')
  const res = await apiAddExecution(planState.plan.id, {
    asset_id: payload.assetId,
    amount: payload.amount,
    side: 'sell',
    point: payload.point ?? null,
    note: payload.note ?? '',
  })
  if (res.code !== 0 || !res.data) throw new Error(res.message || '记录卖出失败')
  await loadPlan(planState.plan.id)

  const restore = payload.restoreLayerKeys ?? []
  if (restore.length === 0) return { exec: res.data, restored: false }
  const asset = findAsset(payload.assetId)
  if (!asset) return { exec: res.data, restored: false }
  const kept = new Set(restore)
  asset.executed_layers = asset.executed_layers.filter((key) => !kept.has(key))
  touch()
  return { exec: res.data, restored: true }
}

// ---------- 本地编辑 ----------

function touch(): void {
  planState.dirty = true
}

function findAsset(assetId: number): GridAssetDTO | undefined {
  return planState.plan?.assets.find((a) => a.id === assetId)
}

export function updatePlanName(name: string): void {
  if (!planState.plan) return
  planState.plan.name = name
  touch()
}

export function updateBudget(value: number): void {
  if (!planState.plan) return
  planState.plan.total_budget = Math.max(0, value)
  touch()
}

export function updateGateThreshold(value: number): void {
  if (!planState.plan) return
  planState.plan.gate_threshold = Math.min(100, Math.max(0, value))
  touch()
}

export function updateFrequency(value: Frequency): void {
  if (!planState.plan) return
  planState.plan.frequency = value
  touch()
}

export function updateNextDate(value: string): void {
  if (!planState.plan) return
  planState.plan.next_date = value
  touch()
}

export function updateGateMode(value: GateMode): void {
  if (!planState.plan) return
  planState.plan.gate_mode = value
  touch()
}

export function updateWeightMethod(value: WeightMethod): void {
  if (!planState.plan) return
  planState.plan.weight_method = value
  touch()
}

export function updateAggregation(value: GridAggregation): void {
  if (!planState.plan) return
  planState.plan.grid_aggregation = value
  touch()
}

export function updateUnallocated(value: number): void {
  if (!planState.plan) return
  planState.plan.unallocated_invested = Math.max(0, value)
  touch()
}

/** 改对外可见范围；只切档位，不动已选好友，来回切换不会丢选择。 */
export function updateVisibility(visibility: PlanVisibility): void {
  if (!planState.plan) return
  planState.plan.visibility = visibility
  if (visibility !== 'specified') planState.plan.visible_user_ids = []
  touch()
}

export function updateVisibleUsers(ids: number[]): void {
  if (!planState.plan) return
  planState.plan.visible_user_ids = [...ids]
  touch()
}

export function updateCurrency(currency: string, currencyName: string): void {
  if (!planState.plan) return
  planState.plan.currency = currency
  planState.plan.currency_name = currencyName
  touch()
}

export type NumericAssetField = 'current_point' | 'high_point' | 'weight' | 'invested' | 'market_value' | 'base_amount'

/** 权重保留 6 位小数，和后端 numeric(9,6) 对齐。 */
const WEIGHT_SCALE = 1e6

export function roundWeight(value: number): number {
  return Math.round(value * WEIGHT_SCALE) / WEIGHT_SCALE
}

/** 权重是唯一落库的字段，目标预算由「总预算 × 权重」算出来。 */
export function targetBudgetOf(asset: Pick<GridAssetDTO, 'weight'>, totalBudget: number): number {
  return ((Number(totalBudget) || 0) * (Number(asset.weight) || 0)) / 100
}

/**
 * 改目标预算：反算权重。
 * 若结果会让所有生效标的的权重合计超过 100%，按 100% 取还能给这个标的的最大金额
 * （否则用户填个大数只会得到一句「超过 100%」的报错，什么也没省下）。
 */
export function updateTargetBudget(assetId: number, amount: number): void {
  const asset = findAsset(assetId)
  const plan = planState.plan
  if (!asset || !plan) return
  const totalBudget = Number(plan.total_budget) || 0
  if (totalBudget <= 0) return

  let wanted = roundWeight(((Math.max(0, Number(amount) || 0)) / totalBudget) * 100)
  const others = plan.assets
    .filter((item) => item !== asset && item.enabled)
    .reduce((sum, item) => sum + (Number(item.weight) || 0), 0)
  const remaining = Math.max(0, 100 - others)
  if (asset.enabled) wanted = Math.min(wanted, roundWeight(remaining))
  else wanted = Math.min(wanted, 100)

  asset.weight = wanted
  touch()
}

export function updateAssetNumber(assetId: number, field: NumericAssetField, value: number | null): void {
  const asset = findAsset(assetId)
  const plan = planState.plan
  if (!asset) return
  if (field === 'current_point' || field === 'high_point') {
    asset[field] = value === null || !Number.isFinite(value) ? null : Math.max(0, value)
    touch()
    return
  }
  if (field === 'weight' && plan) {
    // 直接改权重时也守住 100%：其余标的不动，这个最多占满剩下的部分
    const totalBudget = Number(plan.total_budget) || 0
    if (totalBudget > 0) {
      updateTargetBudget(assetId, (Math.max(0, Number(value) || 0) * totalBudget) / 100)
      return
    }
    asset.weight = roundWeight(Math.max(0, Math.min(100, Number(value) || 0)))
    touch()
    return
  }
  asset[field] = Math.max(0, Number(value) || 0)
  touch()
}

export function updateAssetText(assetId: number, field: 'name' | 'code', value: string): void {
  const asset = findAsset(assetId)
  if (!asset) return
  asset[field] = value
  touch()
}

export function updateAssetEnabled(assetId: number, enabled: boolean): void {
  const asset = findAsset(assetId)
  if (!asset) return
  asset.enabled = enabled
  touch()
}

export function updateAssetLeveraged(assetId: number, leveraged: boolean): void {
  const asset = findAsset(assetId)
  if (!asset) return
  asset.leveraged = leveraged
  touch()
}

/** 新标的还没有后端 id，用负数占位，保存后以后端返回值为准。 */
let tempAssetId = -1

export function addAsset(defaults: Partial<GridAssetPayload> = {}): GridAssetDTO | null {
  if (!planState.plan) return null
  const asset: GridAssetDTO = {
    id: tempAssetId--,
    name: defaults.name ?? '新标的',
    code: defaults.code ?? '',
    leveraged: defaults.leveraged ?? false,
    enabled: defaults.enabled ?? true,
    current_point: defaults.current_point ?? null,
    high_point: defaults.high_point ?? null,
    weight: defaults.weight ?? 0,
    invested: defaults.invested ?? 0,
    market_value: defaults.market_value ?? 0,
    realized_out: 0,
    base_amount: defaults.base_amount ?? 0,
    grid: (defaults.grid ?? gridLayers()).map((layer) => ({ ...layer, voided: Boolean(layer.voided) })),
    executed_layers: [],
  }
  planState.plan.assets.push(asset)
  touch()
  return asset
}

export function removeAsset(assetId: number): boolean {
  if (!planState.plan) return false
  const index = planState.plan.assets.findIndex((a) => a.id === assetId)
  if (index < 0) return false
  planState.plan.assets.splice(index, 1)
  touch()
  return true
}

/**
 * 套用模板：整体替换标的结构与计划参数（币种、预算、频率、闸门）。
 * 已投入、市值、点位这些「真实持仓数据」会被清零，所以界面上必须二次确认。
 */
export function applyTemplate(template: PlanTemplate): void {
  if (!planState.plan) return
  const assets: GridAssetDTO[] = template.assets.map((asset) => ({
    id: tempAssetId--,
    name: asset.name,
    code: asset.code,
    leveraged: asset.leveraged ?? false,
    enabled: true,
    current_point: null,
    high_point: null,
    weight: asset.weight,
    invested: 0,
    market_value: 0,
    realized_out: 0,
    base_amount: asset.baseAmount,
    grid: gridLayers(gridForIndex(asset.name, template.market)),
    executed_layers: [],
  }))
  planState.plan.assets = assets
  planState.plan.market = template.market
  planState.plan.currency = template.currency
  planState.plan.currency_name = template.currencyName
  planState.plan.unit = template.unit
  planState.plan.total_budget = template.totalBudget
  planState.plan.frequency = template.frequency
  planState.plan.gate_threshold = template.gateThreshold
  planState.plan.gate_mode = template.gateMode
  planState.plan.weight_method = template.weightMethod
  planState.plan.grid_aggregation = template.gridAggregation
  touch()
}

/**
 * 新增一层：阈值按该指数的平均档距往下延伸，直接就是合法值。
 * 原来写死 0，用户忘了填就保存会整份被后端拒绝（要求阈值在 0~100 之间）。
 */
export function addGridLayer(assetId: number): void {
  const asset = findAsset(assetId)
  const plan = planState.plan
  if (!asset || !plan) return
  const index = asset.grid.length + 1
  const rows = gridForIndex(asset.name, plan.market === 'us' ? 'us' : 'cn')
  const first = rows[0]?.[0] ?? 5
  const last = rows[rows.length - 1]?.[0] ?? 60
  const gap = rows.length > 1 ? (last - first) / (rows.length - 1) : 8
  // 后端要求所有层阈值严格递增且 < 100，取整后至少要比当前最深一层大 0.1
  const deepest = asset.grid.reduce((max, layer) => Math.max(max, Number(layer.threshold) || 0), 0)
  const threshold = Math.min(99, roundThreshold(Math.max(deepest + gap, deepest + 0.1)))
  asset.grid.push({ key: nextLayerKey(asset.grid, index), threshold, allocation: 0, voided: false })
  touch()
}

function roundThreshold(value: number): number {
  return Math.round(value * 10) / 10
}

export function removeGridLayer(assetId: number, layerKey: string): boolean {
  const asset = findAsset(assetId)
  if (!asset) return false
  // 已投的层删掉会让执行记录对不上、净值曲线断裂，必须先在别处处理这笔投入。
  if (asset.executed_layers.includes(layerKey)) return false
  asset.grid = asset.grid.filter((layer) => layer.key !== layerKey)
  asset.executed_layers = asset.executed_layers.filter((key) => key !== layerKey)
  touch()
  return true
}

/** 作废：不再触发、不计入比例合计。已投的层不允许作废。 */
export function voidLayer(assetId: number, layerKey: string): boolean {
  const asset = findAsset(assetId)
  const layer = asset?.grid.find((l) => l.key === layerKey)
  if (!asset || !layer) return false
  if (asset.executed_layers.includes(layerKey)) return false
  layer.voided = true
  touch()
  return true
}

/** 生效：撤销作废，回到正常触发。 */
export function restoreLayer(assetId: number, layerKey: string): boolean {
  const layer = findAsset(assetId)?.grid.find((l) => l.key === layerKey)
  if (!layer) return false
  layer.voided = false
  touch()
  return true
}

/** 一键优化：整体替换该标的的网格（含作废层与新层），只改本地，仍需保存才落库。 */
export function applyOptimizedGrid(assetId: number, grid: GridLayerDTO[]): boolean {
  const asset = findAsset(assetId)
  if (!asset) return false
  asset.grid = grid.map((layer) => ({ ...layer, voided: Boolean(layer.voided) }))
  // 被优化合并掉的层不能再留在执行状态里，否则后端会以「引用了不存在的网格层」拒绝保存
  const keys = new Set(asset.grid.map((layer) => layer.key))
  asset.executed_layers = asset.executed_layers.filter((key) => keys.has(key))
  touch()
  return true
}

/**
 * 改某一层的计划金额：该层比例按目标预算换算，其余未作废层等比缩放到剩余比例，
 * 保证生效层合计始终 100%。作废层保留原比例，撤销作废后能回到原份额。
 */
export function setLayerAmount(assetId: number, layerKey: string, amount: number): void {
  const asset = findAsset(assetId)
  const plan = planState.plan
  const layer = asset?.grid.find((l) => l.key === layerKey)
  if (!asset || !plan || !layer) return

  const targetBudget = ((Number(plan.total_budget) || 0) * (Number(asset.weight) || 0)) / 100
  if (targetBudget <= 0) return

  const nextAmount = Math.max(0, Number(amount) || 0)
  layer.allocation = roundAllocation(Math.min(100, Math.max(0, (nextAmount / targetBudget) * 100)))

  const others = asset.grid.filter((l) => l !== layer && !l.voided)
  const remaining = Math.max(0, Math.round((100 - layer.allocation) * 100) / 100)
  if (others.length > 0) {
    const othersSum = others.reduce((sum, l) => sum + (Number(l.allocation) || 0), 0)
    if (othersSum > 0) {
      const scale = remaining / othersSum
      others.forEach((l) => { l.allocation = (Number(l.allocation) || 0) * scale })
    } else {
      // 其它层都是 0 时平均分配，避免合计只剩改动的这一层。
      others.forEach((l) => { l.allocation = remaining / others.length })
    }
    // 先四舍五入再补尾差：把剩下的零头加到最后一行非作废层上，
    // 否则每层独立取整会让合计出现 100.01 这种值，后端直接拒绝整份保存。
    others.forEach((l) => { l.allocation = Math.floor(l.allocation * 100) / 100 })
    const used = others.reduce((sum, l) => sum + l.allocation, 0)
    const stub = Math.round((remaining - used) * 100) / 100
    others[others.length - 1]!.allocation = Math.round((others[others.length - 1]!.allocation + stub) * 100) / 100
  } else {
    layer.allocation = 100
  }

  touch()
}

/** 比例保留两位小数。 */
function roundAllocation(value: number): number {
  return Math.round(value * 100) / 100
}

export function updateLayer(
  assetId: number,
  layerKey: string,
  field: 'threshold' | 'allocation',
  value: number,
): void {
  const asset = findAsset(assetId)
  const layer = asset?.grid.find((l) => l.key === layerKey)
  if (!layer) return
  layer[field] = Math.max(0, Number(value) || 0)
  touch()
}

export function setReferenceHighToCurrent(assetId: number): void {
  const asset = findAsset(assetId)
  if (!asset || asset.current_point === null) return
  asset.high_point = asset.current_point
  touch()
}

export function clearExecutedLayers(assetId: number): void {
  const asset = findAsset(assetId)
  if (!asset) return
  asset.executed_layers = []
  touch()
}

// ---------- 内部工具 ----------

/** 补齐后端可能返回 null 的字段，模板里少写一堆判空。 */
function normalizePlan(plan: GridPlanDetail): GridPlanDetail {
  return {
    ...plan,
    next_date: plan.next_date ?? '',
    version: Number(plan.version) || 0,
    visibility: plan.visibility ?? 'private',
    visible_user_ids: plan.visible_user_ids ?? [],
    assets: (plan.assets ?? []).map((asset) => ({
      ...asset,
      grid: (asset.grid ?? []).map((layer) => ({ ...layer, voided: Boolean(layer.voided) })),
      executed_layers: asset.executed_layers ?? [],
    })),
  }
}

function toPayload(plan: GridPlanDetail): SaveGridPlanPayload {
  return {
    name: plan.name,
    market: plan.market,
    currency: plan.currency,
    currency_name: plan.currency_name,
    unit: plan.unit,
    total_budget: plan.total_budget,
    frequency: plan.frequency,
    next_date: plan.next_date,
    gate_threshold: plan.gate_threshold,
    gate_mode: plan.gate_mode,
    weight_method: plan.weight_method,
    grid_aggregation: plan.grid_aggregation,
    unallocated_invested: plan.unallocated_invested,
    sort_order: plan.sort_order,
    version: plan.version,
    visibility: plan.visibility,
    visible_user_ids: plan.visibility === 'specified' ? [...(plan.visible_user_ids ?? [])] : [],
    assets: plan.assets.map((asset) => ({
      id: asset.id > 0 ? asset.id : undefined,
      name: asset.name,
      code: asset.code,
      leveraged: asset.leveraged,
      enabled: asset.enabled,
      current_point: asset.current_point,
      high_point: asset.high_point,
      weight: asset.weight,
      invested: asset.invested,
      market_value: asset.market_value,
      base_amount: asset.base_amount,
      grid: asset.grid
        // 阈值为 0 的层是「新增网格层」留下的空行，后端不接受阈值 0，保存时一并跳过
        .filter((layer: GridLayerDTO) => (Number(layer.threshold) || 0) > 0)
        .map((layer: GridLayerDTO) => ({ ...layer, voided: Boolean(layer.voided) })),
      executed_layers: asset.executed_layers.filter((key) => asset.grid.some(
        (layer) => layer.key === key && (Number(layer.threshold) || 0) > 0,
      )),
    })),
  }
}

function nextLayerKey(grid: GridLayerDTO[], index: number): string {
  const used = new Set(grid.map((layer) => layer.key))
  let n = index
  while (used.has(`g${n}`)) n += 1
  return `g${n}`
}
