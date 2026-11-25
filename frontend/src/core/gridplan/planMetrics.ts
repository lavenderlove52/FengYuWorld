import type { GateMode, GridAssetDTO, GridLayerDTO, GridPlanDetail, WeightMethod } from '@/types/gridPlan'

const EPS = 1e-9

export type PlanAssetStatus =
  | 'invalid'
  | 'gate-blocked'
  | 'budget-blocked'
  | 'grid-triggered'
  | 'regular-due'
  | 'waiting'

export interface PlanAlert {
  id: string
  level: 'error' | 'warn' | 'info' | 'ok'
  title: string
  detail: string
}

export type PlanLayerState = 'invested' | 'voided' | 'triggered' | 'waiting'

export interface PlanGridRow {
  layer: GridLayerDTO
  index: number
  point: number | null
  triggered: boolean
  executed: boolean
  /** 已作废：不参与触发与建议金额 */
  voided: boolean
  /** 一个层只展示一个状态：已投 > 作废 > 已触发 > 等待 */
  state: PlanLayerState
  amount: number
}

export interface PlanGateState {
  triggered: boolean
  threshold: number
  mode: GateMode
  drawdown: number | null
}

export interface PlanAssetMetrics {
  asset: GridAssetDTO
  valid: boolean
  dd: number | null
  targetBudget: number
  remainingIndex: number
  remainingMarket: number
  actualWeight: number | null
  weightDeviation: number | null
  rows: PlanGridRow[]
  pendingRows: PlanGridRow[]
  nextRow: PlanGridRow | null
  latestPending: PlanGridRow | null
  gridAmount: number
  regularAmount: number
  due: boolean
  gate: PlanGateState
  theoretical: number
  suggestion: number
  cappedBy: 'none' | 'gate' | 'index-budget' | 'market-budget' | 'unit' | 'invalid'
  status: PlanAssetStatus
  reason: string
  gridRatioTotal: number
  gridRatioValid: boolean
  /** 已作废的层数，界面用来提示「N 层不参与合计」 */
  voidedCount: number
  thresholdOrderValid: boolean
  strictPyramid: boolean
  /** 市值减去已投入；市值未填时仍按 0 计算，展示层用 hasMarketValue 区分 */
  pnl: number
  /** 收益率百分比，已投入为 0 时返回 null */
  returnRate: number | null
  /** 市值为 0 且已投入大于 0 视为「未填市值」 */
  hasMarketValue: boolean
}

export interface PlanMetrics {
  plan: GridPlanDetail
  invested: number
  marketValue: number
  pnl: number
  returnRate: number | null
  /** 未填市值但已有投入的金额，用于提示「还有多少钱没算进盈亏」 */
  unfilledInvested: number
  remaining: number
  overBudget: number
  weightSum: number
  weightValid: boolean
  drawdown: number | null
  gate: PlanGateState
  assets: PlanAssetMetrics[]
  enabledAssets: PlanAssetMetrics[]
  suggestionTotal: number
  theoreticalTotal: number
  alerts: PlanAlert[]
  /** 启用标的中已填市值的数量，用于饼图与「未填市值」提示 */
  marketValueFilled: number
}

export function assetPnl(asset: Pick<GridAssetDTO, 'invested' | 'market_value'>): number {
  return (Number(asset.market_value) || 0) - (Number(asset.invested) || 0)
}

/** 市值未填时返回 null，避免界面出现「-100%」这种假数据。 */
export function assetReturnRate(asset: Pick<GridAssetDTO, 'invested' | 'market_value'>): number | null {
  const invested = Number(asset.invested) || 0
  if (invested <= 0) return null
  if ((Number(asset.market_value) || 0) <= 0) return null
  return (assetPnl(asset) / invested) * 100
}

export function hasMarketValue(asset: Pick<GridAssetDTO, 'invested' | 'market_value'>): boolean {
  return (Number(asset.market_value) || 0) > 0
}

/**
 * 计划级汇总：只统计启用且填了市值的标的，未填的单独返回金额。
 * 把未填市值当成 0 会让总盈亏凭空多出「-已投入」，所以这里不算它们。
 * 停用的标的整体不参与账目，和饼图、净值曲线的口径保持一致。
 */
export function planTotals(plan: Pick<GridPlanDetail, 'assets'>): {
  invested: number
  marketValue: number
  pnl: number
  returnRate: number | null
  unfilledInvested: number
  unfilledCount: number
} {
  let invested = 0
  let marketValue = 0
  let unfilledInvested = 0
  let unfilledCount = 0
  for (const asset of plan.assets) {
    if (!asset.enabled) continue
    const assetInvested = Number(asset.invested) || 0
    if (!hasMarketValue(asset)) {
      if (assetInvested > 0) unfilledInvested += assetInvested
      unfilledCount += 1
      continue
    }
    invested += assetInvested
    marketValue += Number(asset.market_value) || 0
  }
  return {
    invested,
    marketValue,
    pnl: marketValue - invested,
    returnRate: invested > 0 ? ((marketValue - invested) / invested) * 100 : null,
    unfilledInvested,
    unfilledCount,
  }
}

export function computeDrawdown(asset: Pick<GridAssetDTO, 'current_point' | 'high_point'>): number | null {
  const { current_point: current, high_point: high } = asset
  if (current === null || high === null) return null
  if (!Number.isFinite(current) || !Number.isFinite(high) || high <= 0) return null
  return Math.max(0, (1 - current / high) * 100)
}

export function gridRatioTotal(asset: Pick<GridAssetDTO, 'grid'>): number {
  return asset.grid.reduce((sum, layer) => (layer.voided ? sum : sum + (Number(layer.allocation) || 0)), 0)
}

export function gridRatioValid(asset: Pick<GridAssetDTO, 'grid'>): boolean {
  return gridRatioTotal(asset) <= 100 + EPS && asset.grid.every((layer) => (Number(layer.allocation) || 0) >= 0)
}

export function thresholdOrderValid(asset: Pick<GridAssetDTO, 'grid'>): boolean {
  // 作废层是历史快照，允许与新层同阈值，不参与顺序校验（与后端 validateAsset 一致）
  const values = asset.grid
    .filter((layer) => !layer.voided && (Number(layer.threshold) || 0) > 0)
    .map((layer) => Number(layer.threshold))
  return values.every((value, index) => index === 0 || value > values[index - 1]!)
}
export function isStrictPyramid(asset: Pick<GridAssetDTO, 'grid'>): boolean {
  // 作废层不参与，否则作废一个高比例层会被误报成「不是倒金字塔」。
  const active = asset.grid.filter((layer) => !layer.voided)
  return active.every(
    (layer, index) => index === 0 || (Number(layer.allocation) || 0) >= (Number(active[index - 1]!.allocation) || 0),
  )
}

/** 今天或更早即视为定投到期。 */
export function isRegularDue(plan: Pick<GridPlanDetail, 'next_date'>): boolean {
  return Boolean(plan.next_date) && plan.next_date <= todayISO()
}

export function totalInvested(plan: Pick<GridPlanDetail, 'assets' | 'unallocated_invested'>): number {
  // 停用的标的不记账：它的钱不再占预算，也不再出现在已投入里
  const assets = plan.assets
    .filter((asset) => asset.enabled)
    .reduce((sum, asset) => sum + (Number(asset.invested) || 0), 0)
  return assets + (Number(plan.unallocated_invested) || 0)
}

export function weightSum(plan: Pick<GridPlanDetail, 'assets'>): number {
  return plan.assets
    .filter((asset) => asset.enabled)
    .reduce((sum, asset) => sum + (Number(asset.weight) || 0), 0)
}

export function planDrawdown(plan: Pick<GridPlanDetail, 'assets' | 'weight_method'>): number | null {
  const rows = plan.assets
    .filter((asset) => asset.enabled)
    .map((asset) => ({ asset, dd: computeDrawdown(asset) }))
    .filter((row): row is { asset: GridAssetDTO; dd: number } => row.dd !== null)
  if (rows.length === 0) return null

  if (plan.weight_method === 'invested-weight') {
    const invested = rows.reduce((sum, row) => sum + Math.max(0, Number(row.asset.invested) || 0), 0)
    if (invested <= 0) return null
    return rows.reduce((sum, row) => sum + (row.dd * Math.max(0, row.asset.invested)) / invested, 0)
  }

  const total = rows.reduce((sum, row) => sum + Math.max(0, Number(row.asset.weight) || 0), 0)
  if (total <= 0) return null
  return rows.reduce((sum, row) => sum + (row.dd * Math.max(0, row.asset.weight)) / total, 0)
}

function buildRows(asset: GridAssetDTO, dd: number | null, targetBudget: number): PlanGridRow[] {
  const executed = new Set(asset.executed_layers)
  return asset.grid.map((layer, index) => {
    const threshold = Number(layer.threshold) || 0
    const high = asset.high_point
    const isExecuted = executed.has(layer.key)
    const voided = Boolean(layer.voided)
    const triggered = dd !== null && dd >= threshold - EPS
    return {
      layer,
      index,
      point: high !== null && high > 0 ? high * (1 - threshold / 100) : null,
      triggered,
      executed: isExecuted,
      voided,
      state: isExecuted ? 'invested' : voided ? 'voided' : triggered ? 'triggered' : 'waiting',
      amount: (targetBudget * (Number(layer.allocation) || 0)) / 100,
    }
  })
}

const STATUS_TEXT: Record<PlanAssetStatus, string> = {
  invalid: '数据或参数不完整',
  'gate-blocked': '被组合回撤闸门限制',
  'budget-blocked': '预算不足',
  'grid-triggered': '网格触发',
  'regular-due': '定投日到期',
  waiting: '等待',
}

function voidedRatioHint(asset: Pick<GridAssetDTO, 'grid'>, ratioTotal: number): string {
  const voided = asset.grid.filter((layer) => layer.voided).length
  const suffix = voided > 0 ? `（已作废 ${voided} 层不参与合计）` : ''
  return `网格资金比例合计 ${ratioTotal.toFixed(2)}%，超过 100%，请调整${suffix}`
}

export function computeAssetMetrics(
  plan: GridPlanDetail,
  asset: GridAssetDTO,
  ctx?: { investedNow?: number; planDD?: number | null; due?: boolean },
): PlanAssetMetrics {
  const dd = computeDrawdown(asset)
  const totalBudget = Number(plan.total_budget) || 0
  const targetBudget = (totalBudget * (Number(asset.weight) || 0)) / 100
  // 计划级数字由调用方算一次传进来；单独调用时退回自己算，结果一致。
  const investedNow = ctx?.investedNow ?? totalInvested(plan)
  const remainingMarket = Math.max(0, totalBudget - investedNow)
  const remainingIndex = Math.max(0, targetBudget - (Number(asset.invested) || 0))

  const rows = buildRows(asset, dd, targetBudget)
  const pendingRows = rows.filter((row) => row.triggered && !row.executed && !row.voided)
  const nextRow = rows.find((row) => !row.voided && Number(row.layer.threshold) > (dd ?? 0) + EPS) ?? null

  const planDD = ctx && ctx.planDD !== undefined ? ctx.planDD : planDrawdown(plan)
  const gate: PlanGateState = {
    triggered: planDD !== null && planDD >= (Number(plan.gate_threshold) || 0) - EPS,
    threshold: Number(plan.gate_threshold) || 0,
    mode: plan.gate_mode,
    drawdown: planDD,
  }

  const latestPending = pendingRows.length ? pendingRows[pendingRows.length - 1]! : null
  const gridAmount =
    plan.grid_aggregation === 'cumulative'
      ? pendingRows.reduce((sum, row) => sum + row.amount, 0)
      : latestPending?.amount ?? 0
  const due = ctx?.due ?? isRegularDue(plan)
  const regularAmount = due ? Number(asset.base_amount) || 0 : 0

  let theoretical = gridAmount + regularAmount
  let cappedBy: PlanAssetMetrics['cappedBy'] = 'none'
  let reason = STATUS_TEXT[latestPending ? 'grid-triggered' : due ? 'regular-due' : 'waiting']
  if (latestPending) reason += `（G${latestPending.index + 1}）`

  const invalidEndpoint =
    asset.current_point === null ||
    asset.high_point === null ||
    !Number.isFinite(asset.current_point) ||
    !Number.isFinite(asset.high_point) ||
    asset.high_point <= 0 ||
    asset.current_point < 0
  const hasName = asset.name.trim().length > 0
  const valid = !invalidEndpoint && hasName

  if (!valid) {
    theoretical = 0
    cappedBy = 'invalid'
    reason = !hasName ? '缺少标的名称' : '当前点位或参考高点无效（参考高点需大于 0）'
  } else if (gate.triggered && gate.mode === 'block-all') {
    theoretical = 0
    cappedBy = 'gate'
    reason = `组合回撤 ${gate.drawdown?.toFixed(2)}% 达到闸门，已暂停全部新增投入`
  } else if (gate.triggered && gate.mode === 'grid-only' && gridAmount <= 0) {
    theoretical = 0
    cappedBy = 'gate'
    reason = `组合回撤 ${gate.drawdown?.toFixed(2)}% 达到闸门，当前只允许网格加仓`
  }

  const unit = Number(plan.unit) || 1
  const affordable = Math.min(theoretical, remainingIndex, remainingMarket)
  let suggestion = roundDownToUnit(affordable, unit)

  if (valid && theoretical > 0) {
    if (remainingIndex <= 0) {
      suggestion = 0
      cappedBy = 'index-budget'
      reason = '该标的目标预算已用尽'
    } else if (remainingMarket <= 0) {
      suggestion = 0
      cappedBy = 'market-budget'
      reason = '计划总预算已用尽'
    } else if (suggestion <= 0) {
      suggestion = 0
      cappedBy = 'unit'
      reason = `可用金额低于最小交易单位 ${unit}`
    } else if (suggestion + EPS < theoretical && affordable < theoretical - EPS) {
      cappedBy = theoretical > remainingIndex ? 'index-budget' : 'market-budget'
      reason =
        theoretical > remainingIndex
          ? '理论金额已按标的目标预算余量封顶'
          : '理论金额已按计划总预算余量封顶'
    } else if (suggestion + EPS < theoretical) {
      // 预算够用，差额只是最小交易单位向下取整，不算被限制
      reason += `（已按最小交易单位 ${unit} 向下取整）`
    }
  }

  let status: PlanAssetStatus
  if (!valid) status = 'invalid'
  else if (cappedBy === 'gate') status = 'gate-blocked'
  else if (cappedBy === 'index-budget' || cappedBy === 'market-budget' || cappedBy === 'unit') status = 'budget-blocked'
  else if (latestPending) status = 'grid-triggered'
  else if (due && suggestion > 0) status = 'regular-due'
  else status = 'waiting'

  const actualWeight = investedNow > 0 ? ((Number(asset.invested) || 0) / investedNow) * 100 : null
  const weightDeviation = actualWeight === null ? null : actualWeight - (Number(asset.weight) || 0)
  const ratioTotal = gridRatioTotal(asset)
  const ratioValid = gridRatioValid(asset)
  const orderValid = thresholdOrderValid(asset)

  if (valid && !ratioValid) {
    suggestion = 0
    cappedBy = 'invalid'
    status = 'invalid'
    reason = voidedRatioHint(asset, ratioTotal)
  } else if (valid && !orderValid && asset.grid.length > 0) {
    suggestion = 0
    cappedBy = 'invalid'
    status = 'invalid'
    reason = '网格回撤阈值必须严格递增且大于 0'
  }

  return {
    asset,
    valid,
    dd,
    targetBudget,
    remainingIndex,
    remainingMarket,
    actualWeight,
    weightDeviation,
    rows,
    pendingRows,
    nextRow,
    latestPending,
    gridAmount,
    regularAmount,
    due,
    gate,
    theoretical,
    suggestion,
    cappedBy,
    status,
    reason,
    gridRatioTotal: ratioTotal,
    gridRatioValid: ratioValid,
    voidedCount: asset.grid.filter((layer) => layer.voided).length,
    thresholdOrderValid: orderValid,
    strictPyramid: isStrictPyramid(asset),
    pnl: assetPnl(asset),
    returnRate: assetReturnRate(asset),
    hasMarketValue: hasMarketValue(asset),
  }
}

export function computePlanMetrics(plan: GridPlanDetail): PlanMetrics {
  const invested = totalInvested(plan)
  const remaining = (Number(plan.total_budget) || 0) - invested
  const weights = weightSum(plan)
  const drawdown = planDrawdown(plan)
  const gate: PlanGateState = {
    triggered: drawdown !== null && drawdown >= (Number(plan.gate_threshold) || 0) - EPS,
    threshold: Number(plan.gate_threshold) || 0,
    mode: plan.gate_mode,
    drawdown,
  }
  // 计划级数字只算一次，避免每个标的都重扫全表（标的 × 10 层时是 O(n²)）
  const due = isRegularDue(plan)
  const ctx = { investedNow: invested, planDD: drawdown, due }
  const assets = plan.assets.map((asset) => computeAssetMetrics(plan, asset, ctx))
  const enabledAssets = assets.filter((row) => row.asset.enabled)
  const totals = planTotals(plan)
  const metrics: PlanMetrics = {
    plan,
    invested,
    marketValue: totals.marketValue,
    pnl: totals.pnl,
    returnRate: totals.returnRate,
    unfilledInvested: totals.unfilledInvested,
    remaining,
    overBudget: Math.max(0, -remaining),
    weightSum: weights,
    weightValid: Math.abs(weights - 100) < 0.01,
    drawdown,
    gate,
    assets,
    enabledAssets,
    suggestionTotal: enabledAssets.reduce((sum, row) => sum + row.suggestion, 0),
    theoreticalTotal: enabledAssets.reduce((sum, row) => sum + row.theoretical, 0),
    alerts: [],
    marketValueFilled: enabledAssets.filter((row) => row.hasMarketValue).length,
  }
  metrics.alerts = buildAlerts(metrics)
  return metrics
}

function buildAlerts(m: PlanMetrics): PlanAlert[] {
  const alerts: PlanAlert[] = []
  const { plan } = m

  if (m.weightSum > 100 + EPS) {
    alerts.push({
      id: 'weight-over',
      level: 'error',
      title: `目标权重合计 ${m.weightSum.toFixed(2)}%，已超过 100%`,
      detail: '请先下调权重，当前计划不可直接执行。',
    })
  } else if (m.weightSum < 100 - EPS) {
    alerts.push({
      id: 'weight-under',
      level: 'info',
      title: `目标权重合计 ${m.weightSum.toFixed(2)}%，尚有 ${(100 - m.weightSum).toFixed(2)}% 未分配`,
      detail: '未分配部分不会自动落到任何标的，可保留为现金或补充权重。',
    })
  } else {
    alerts.push({
      id: 'weight-ok',
      level: 'ok',
      title: '目标权重合计为 100%',
      detail: '计划预算已完整分配到各标的。',
    })
  }

  if (m.remaining < 0) {
    alerts.push({
      id: 'market-over',
      level: 'error',
      title: `总投入超出预算 ${Math.abs(m.remaining).toFixed(2)}`,
      detail: '所有新增建议都会被预算上限归零，请核对已投入金额。',
    })
  }

  if (m.gate.triggered) {
    const modeText =
      plan.gate_mode === 'block-all' ? '暂停全部新增投入' : plan.gate_mode === 'grid-only' ? '仅允许网格加仓' : '仅提醒，不限制建议'
    alerts.push({
      id: 'gate',
      level: plan.gate_mode === 'alert' ? 'warn' : 'error',
      title: `组合回撤 ${m.drawdown?.toFixed(2)}% 已越过闸门 ${plan.gate_threshold}%`,
      detail: `当前闸门模式：${modeText}。`,
    })
  } else if (m.drawdown !== null) {
    alerts.push({
      id: 'gate-ok',
      level: 'info',
      title: `组合回撤 ${m.drawdown.toFixed(2)}%，闸门阈值 ${plan.gate_threshold}%`,
      detail: `距离闸门还有 ${(plan.gate_threshold - m.drawdown).toFixed(2)} 个百分点。`,
    })
  }

  if (m.theoreticalTotal > m.suggestionTotal + EPS) {
    alerts.push({
      id: 'capped',
      level: 'warn',
      title: `理论建议 ${m.theoreticalTotal.toFixed(0)}，实际可执行 ${m.suggestionTotal.toFixed(0)}`,
      detail: '多个标的同时触发或受预算限制，请人工决定执行顺序。',
    })
  }

  if (m.suggestionTotal > m.remaining + EPS) {
    alerts.push({
      id: 'cash-short',
      level: 'error',
      title: '本期建议总额超过计划剩余预算',
      detail: `剩余 ${m.remaining.toFixed(0)}，建议 ${m.suggestionTotal.toFixed(0)}；请降低权重、增加预算或分批执行。`,
    })
  }

  m.enabledAssets.forEach((row) => {
    if (!row.gridRatioValid) {
      alerts.push({
        id: `ratio-${row.asset.id}`,
        level: 'error',
        title: `${row.asset.name} 网格资金比例合计 ${row.gridRatioTotal.toFixed(2)}%`,
        detail: '比例超过 100% 时该标的建议金额归零，请调整各层资金比例。',
      })
    }
    if (!row.thresholdOrderValid && row.asset.grid.length > 0) {
      alerts.push({
        id: `order-${row.asset.id}`,
        level: 'error',
        title: `${row.asset.name} 的网格阈值未严格递增`,
        detail: '阈值必须逐层增大，否则外层网格无法按预期触发。',
      })
    }
    if (!row.strictPyramid) {
      alerts.push({
        id: `pyramid-${row.asset.id}`,
        level: 'warn',
        title: `${row.asset.name} 的资金比例不是严格倒金字塔`,
        detail: '后层资金比例低于前层，页面仍按你的比例计算，但不符合越跌越买的结构。',
      })
    }
    if (row.weightDeviation !== null && Math.abs(row.weightDeviation) >= 5) {
      const over = row.weightDeviation > 0
      alerts.push({
        id: `drift-${row.asset.id}`,
        level: 'info',
        title: `${row.asset.name} ${over ? '超配' : '低配'} ${Math.abs(row.weightDeviation).toFixed(2)} 个百分点`,
        detail: `实际权重约 ${row.actualWeight?.toFixed(2)}%，目标权重 ${row.asset.weight}%。可用新增资金优先纠偏。`,
      })
    }
  })

  if (plan.assets.some((a) => a.leveraged && a.enabled)) {    alerts.push({
      id: 'leveraged',
      level: 'warn',
      title: '组合包含 2 倍杠杆标的',
      detail: '杠杆 ETF 存在每日再平衡与波动损耗，长期持有可能显著偏离指数倍数，请单独设置更低上限。',
    })
  }

  if (plan.assets.length === 0) {
    alerts.push({ id: 'no-assets', level: 'warn', title: '当前计划没有标的', detail: '请新增标的后再计算。' })
  }

  // 停用的标的没市值是正常的，不算「未填」
  const enabledMissing = m.enabledAssets.length - m.marketValueFilled
  if (enabledMissing > 0) {
    alerts.push({
      id: 'market-value-missing',
      level: 'info',
      title: `${enabledMissing} 个标的未填当前市值`,
      detail: m.unfilledInvested > 0
        ? `其中 ${m.unfilledInvested.toFixed(2)} 的已投入没有计入盈亏，填入市值后才算完整。`
        : '未填市值的标的不会计入总市值和总盈亏。',
    })
  }

  return alerts
}

export function roundDownToUnit(value: number, unit: number): number {
  const safeUnit = unit > 0 ? unit : 1
  if (!Number.isFinite(value) || value <= 0) return 0
  return Math.floor(value / safeUnit + 1e-9) * safeUnit
}

export function todayISO(): string {
  const now = new Date()
  const offset = now.getTimezoneOffset() * 60_000
  return new Date(now.getTime() - offset).toISOString().slice(0, 10)
}

export type { WeightMethod }
