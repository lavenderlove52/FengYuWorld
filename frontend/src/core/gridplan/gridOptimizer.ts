import type { GridAssetDTO, GridLayerDTO } from '@/types/gridPlan'
import { computeDrawdown } from './planMetrics'
import { gridForIndex } from './templates'

/**
 * 一键优化网格：把计划重新排成「按指数特性决定的档数」，让深跌时还有钱可买。
 *
 * 场景：按高点出的计划，浅层跌穿时没买，本金就堆在浅层上。
 * 优化后浅层作废（金额保留、可恢复），已执行层不动，钱重新排到更深的档位上。
 */

export type Volatility = 'low' | 'mid' | 'high'

/** 按指数习性给的默认档数：波动越大，越需要更深的档位。 */
const DEFAULT_LAYERS: Record<Volatility, number> = { low: 7, mid: 8, high: 9 }

const VOLATILITY_BY_NAME: Record<string, Volatility> = {
  上证综指: 'low',
  '标普 500': 'low',
  道琼斯工业平均指数: 'low',
  '中证 1000': 'mid',
  创业板指: 'mid',
  纳斯达克综合指数: 'mid',
  可口可乐: 'mid',
  '中证 2000': 'high',
  '标普 500 2 倍做多': 'high',
  '纳指 2 倍做多': 'high',
}

export const VOLATILITY_LABEL: Record<Volatility, string> = {
  low: '低波动',
  mid: '中波动',
  high: '高波动',
}

/** 最少档数：低于这个数，深跌时几乎没有加仓空间。 */
export const MIN_LAYERS = 7
/** 后端 validateAsset 的上限，作废层也占行数。 */
const MAX_GRID_ROWS = 50
/** 阈值上限：再深的跌幅在指数上已经很少见，排进去只会永远不触发。 */
const MAX_THRESHOLD = 95
/** 阈值下限：坡度为 0 时第一层不能落在 0（后端要求阈值 > 0）。 */
const MIN_THRESHOLD = 0.1

export function volatilityOf(asset: Pick<GridAssetDTO, 'name' | 'leveraged'>): Volatility {
  if (asset.leveraged) return 'high'
  return VOLATILITY_BY_NAME[asset.name] ?? 'mid'
}

/** 这个标的默认该有几层（给输入框的初始值，用户可以手动改）。 */
export function suggestLayerCount(asset: Pick<GridAssetDTO, 'name' | 'leveraged'>): number {
  return DEFAULT_LAYERS[volatilityOf(asset)]
}

export interface GridOptimizeResult {
  grid: GridLayerDTO[]
  /** 排新层的起点（回撤百分比） */
  anchor: number
  anchorSource: 'executed' | 'drawdown'
  volatility: Volatility
  defaultLayers: number
  /** 优化后的生效档数（已执行 + 新排） */
  targetLayers: number
  /** 最深一层的阈值 */
  maxDepth: number
  executedLayers: GridLayerDTO[]
  /** 本次需要作废的层：没买过、比起点浅、且原本是生效的 */
  voidedLayers: GridLayerDTO[]
  /** 保留不动的作废层：优化前就已经作废，只是没动它 */
  alreadyVoidedCount: number
  /** 整个网格的行数（含作废层），后端按行数校验 50 层上限 */
  rowCount: number
  /** 被新层取代的层：没买过、不比起点浅 */
  removedLayers: GridLayerDTO[]
  addedLayers: GridLayerDTO[]
}

/**
 * 计算优化结果，不改任何状态。
 * 锚点取「最深已执行层」和「当前回撤」里更深的一个：
 * 只看已执行层会把新层排到已经跌穿的区间，一上来就整片触发；
 * 只看当前回撤又会在还没到底时丢掉已买过的那个深度。
 */
export function optimizeGrid(
  asset: GridAssetDTO,
  market: string,
  requestedLayers?: number,
): GridOptimizeResult {
  const volatility = volatilityOf(asset)
  const defaultLayers = DEFAULT_LAYERS[volatility]
  const executedKeys = new Set(asset.executed_layers)
  const executed = asset.grid.filter((layer) => executedKeys.has(layer.key))

  const dd = computeDrawdown(asset)
  const deepestExecuted = executed.length > 0
    ? Math.max(...executed.map((layer) => Number(layer.threshold) || 0))
    : null
  if (deepestExecuted === null && dd === null) {
    throw new Error('请先填「当前点位」和「参考高点」，或者先点「已投」标出起点')
  }
  let anchor = round1(Math.max(deepestExecuted ?? 0, dd ?? 0))
  if (anchor < MIN_THRESHOLD) anchor = MIN_THRESHOLD
  // drawdown 表示起点由当前回撤决定（新层第一档今天就能买）；否则由已执行层决定
  const anchorSource: 'executed' | 'drawdown' =
    dd !== null && dd >= (deepestExecuted ?? -1) ? 'drawdown' : 'executed'

  const requested = Math.floor(Number(requestedLayers) || defaultLayers)
  const targetLayers = Math.max(requested, executed.length + 1, MIN_LAYERS)
  const newCount = targetLayers - executed.length

  // 比起点浅又没买过的层作废（金额保留）；不比起点浅的由新层取代。
  const voidedLayers: GridLayerDTO[] = []
  const removedLayers: GridLayerDTO[] = []
  let alreadyVoidedCount = 0
  for (const layer of asset.grid) {
    if (executedKeys.has(layer.key)) continue
    const threshold = Number(layer.threshold) || 0
    if (threshold < anchor - 0.05) {
      if (layer.voided) {
        // 本来就作废的层原样留着，不算「本次作废」
        alreadyVoidedCount += 1
        voidedLayers.push(layer)
      } else {
        voidedLayers.push({ ...layer, voided: true })
      }
    } else {
      removedLayers.push(layer)
    }
  }

  const preset = [...gridForIndex(asset.name, market === 'us' ? 'us' : 'cn')]
    .filter(([threshold]) => Number(threshold) > 0)
    .sort((a, b) => a[0] - b[0])
  if (preset.length === 0) throw new Error('这个标的没有可用的预设网格')
  const presetFirst = preset[0]![0]
  const presetLast = preset[preset.length - 1]![0]
  let maxDepth = presetLast
  if (maxDepth <= anchor + 0.5) {
    // 起点已经超过该指数的预设最深档：按它自己的平均档距继续往下延伸
    const gap = preset.length > 1 ? (presetLast - presetFirst) / (preset.length - 1) : 8
    maxDepth = Math.min(MAX_THRESHOLD, anchor + gap * newCount)
  }
  if (maxDepth <= anchor + 0.5) {
    throw new Error(`起点已经在 ${anchor.toFixed(1)}% 的回撤上，再往下排也不会触发，请先更新参考高点`)
  }

  const curve = makeDepthCurve(preset)
  const thresholds: number[] = []
  const keyOf = makeKeyAllocator(asset.grid)
  let previous = Math.max(anchor, deepestExecuted ?? 0)
  for (let i = 0; i < newCount; i += 1) {
    // 第一档落在锚点上（今天即可买入）；之后按指数自身的档距形状往深处排。
    const progress = i / Math.max(1, newCount - 1)
    let value = round1(anchor + (maxDepth - anchor) * curve(progress))
    // 取整后可能和锚点或上一层撞车，逐层加 0.1 保证严格递增（后端会拒绝非递增的网格）
    if (value <= previous + 0.05) value = round1(previous + 0.1)
    if (value > MAX_THRESHOLD) {
      throw new Error(`档数太多，排到 ${MAX_THRESHOLD}% 以后已无意义，请减少层数`)
    }
    thresholds.push(value)
    previous = value
  }

  const allocations = splitAllocation(executed, newCount)
  const addedLayers: GridLayerDTO[] = thresholds.map((threshold, index) => ({
    key: keyOf.next(),
    threshold,
    allocation: allocations[index]!,
    voided: false,
  }))

  const rowCount = voidedLayers.length + executed.length + addedLayers.length
  if (rowCount > MAX_GRID_ROWS) {
    throw new Error(
      `优化后表格会是 ${rowCount} 行（作废层也算行），超过后端上限 ${MAX_GRID_ROWS} 行；`
      + `请先删掉一些已作废的浅层再优化`,
    )
  }

  const grid = [...voidedLayers, ...executed, ...addedLayers].sort(
    (a, b) => (Number(a.threshold) || 0) - (Number(b.threshold) || 0),
  )
  // 已执行层 + 新层才是真正有动作的档数，作废层不计入
  const activeLayers = executed.length + addedLayers.length
  const deepest = grid[grid.length - 1]

  return {
    grid,
    anchor,
    anchorSource,
    volatility,
    defaultLayers,
    targetLayers: activeLayers,
    maxDepth: deepest ? round1(Number(deepest.threshold) || 0) : 0,
    executedLayers: executed,
    voidedLayers,
    removedLayers,
    addedLayers,
    alreadyVoidedCount,
    rowCount,
  }
}

/** 作废层沿用原有标识；新层必须避开所有现存标识（作废层和被取代的层都可能占着 gN）。 */
function makeKeyAllocator(grid: GridLayerDTO[]): { next: () => string } {
  const used = new Set(grid.map((layer) => layer.key))
  let index = 1
  return {
    next: () => {
      while (used.has(`g${index}`)) index += 1
      const key = `g${index}`
      used.add(key)
      return key
    },
  }
}

/**
 * 已执行层比例不动，剩余比例按越跌越厚（权重 = 层序）分给新层。
 * 先向下取整到分，尾差补到最后一层，保证合计正好 100。
 */
function splitAllocation(executed: GridLayerDTO[], newCount: number): number[] {
  const executedSum = executed.reduce((sum, layer) => sum + (Number(layer.allocation) || 0), 0)
  const remaining = round2(100 - executedSum)
  if (remaining <= 0.01) {
    throw new Error(
      `已执行层的资金比例合计 ${formatPercent(executedSum)}，没有余额分给新层，请先调低它们的比例`,
    )
  }
  const weights = Array.from({ length: newCount }, (_, index) => index + 1)
  const weightSum = weights.reduce((sum, value) => sum + value, 0)
  const allocations = weights.map((weight) => Math.floor(((remaining * weight) / weightSum) * 100) / 100)
  const used = allocations.reduce((sum, value) => sum + value, 0)
  const last = allocations.length - 1
  allocations[last] = round2(allocations[last]! + (remaining - used))
  return allocations
}

/**
 * 该指数自己的档位形状：浅层密、深层疏。
 * progress∈[0,1] 均匀 → 返回落在 [0,1] 的深度比例，用它采样就保留了原始形状。
 */
function makeDepthCurve(rows: Array<[number, number]>): (progress: number) => number {
  if (rows.length < 2) return (progress) => progress
  const first = rows[0]![0]
  const last = rows[rows.length - 1]![0]
  if (last <= first) return (progress) => progress
  // 每一档在深度区间里的归一化位置
  const positions = rows.map(([threshold]) => (threshold - first) / (last - first))
  return (progress) => {
    const pos = Math.min(1, Math.max(0, progress)) * (positions.length - 1)
    const low = Math.floor(pos)
    const high = Math.min(positions.length - 1, low + 1)
    const ratio = pos - low
    return positions[low]! + (positions[high]! - positions[low]!) * ratio
  }
}

function round1(value: number): number {
  return Math.round(value * 10) / 10
}

function round2(value: number): number {
  return Math.round(value * 100) / 100
}

function formatPercent(value: number): string {
  return `${Math.round(value * 100) / 100}%`
}
