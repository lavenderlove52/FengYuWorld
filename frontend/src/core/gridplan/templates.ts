import type { GridAssetPayload, GridLayerDTO, SaveGridPlanPayload } from '@/types/gridPlan'

/** 预置指数：用户从下拉里挑一个，自动填名称与代码，不用手输。 */
export interface IndexPreset {
  name: string
  code: string
  leveraged?: boolean
  market: 'cn' | 'us'
  /** 该指数适用的网格；没有配的用 GRID_10 兜底。 */
  grid?: GridRows
}

export type GridRows = Array<[number, number]>

/**
 * 各指数一套网格：阈值按「历史上这指数跌多少算罕见」来定，比例按越跌越买加厚。
 * 都是经验档位，不是实时行情算出来的，可以在详情页自行调整。
 */
const GRID_SH_INDEX: GridRows = [
  [5, 8], [10, 9], [15, 10], [22, 12], [30, 14], [40, 15], [50, 16], [55, 16],
]

const GRID_CSI_1000: GridRows = [
  [7, 8], [14, 9], [21, 10], [29, 12], [38, 14], [48, 15], [57, 16], [65, 16],
]

const GRID_CSI_2000: GridRows = [
  [8, 7], [16, 8], [25, 9], [34, 10], [43, 11], [52, 13], [61, 14], [69, 14], [75, 14],
]

const GRID_CHINEXT: GridRows = [
  [8, 9], [15, 10], [23, 11], [31, 12], [40, 13], [49, 14], [58, 15], [70, 16],
]

const GRID_SP500: GridRows = [
  [5, 8], [10, 9], [15, 10], [21, 12], [28, 14], [36, 15], [44, 16], [50, 16],
]

const GRID_NASDAQ: GridRows = [
  [6, 8], [12, 9], [19, 10], [27, 12], [36, 14], [46, 15], [54, 16], [60, 16],
]

const GRID_DJI: GridRows = [
  [4, 10], [9, 11], [14, 12], [20, 13], [27, 13], [35, 13], [40, 14], [45, 14],
]

/** 2 倍杠杆：同样跌幅对应的净值损耗更快，档位拉宽、尾部更厚。 */
const GRID_LEVERAGED: GridRows = [
  [10, 6], [20, 7], [30, 8], [40, 10], [50, 12], [60, 14], [70, 18], [80, 25],
]

const GRID_KO: GridRows = [
  [6, 8], [12, 9], [19, 10], [27, 12], [36, 14], [45, 15], [53, 16], [60, 16],
]

export const CN_INDICES: IndexPreset[] = [
  { name: '上证综指', code: '000001.SH', market: 'cn', grid: GRID_SH_INDEX },
  { name: '中证 1000', code: '000852.CSI', market: 'cn', grid: GRID_CSI_1000 },
  { name: '中证 2000', code: '932000.CSI', market: 'cn', grid: GRID_CSI_2000 },
  { name: '创业板指', code: '399006.SZ', market: 'cn', grid: GRID_CHINEXT },
]

export const US_INDICES: IndexPreset[] = [
  { name: '标普 500', code: 'S&P 500 / SPY', market: 'us', grid: GRID_SP500 },
  { name: '纳斯达克综合指数', code: 'NASDAQ / QQQ', market: 'us', grid: GRID_NASDAQ },
  { name: '道琼斯工业平均指数', code: 'DJI / DIA', market: 'us', grid: GRID_DJI },
  { name: '标普 500 2 倍做多', code: 'SSO / 2× S&P 500', leveraged: true, market: 'us', grid: GRID_LEVERAGED },
  { name: '纳指 2 倍做多', code: 'QLD / 2× NASDAQ', leveraged: true, market: 'us', grid: GRID_LEVERAGED },
  { name: '可口可乐', code: 'KO', market: 'us', grid: GRID_KO },
]

export const CURRENCIES: Array<{ value: string; name: string; label: string }> = [
  { value: '¥', name: '人民币', label: '¥ 人民币' },
  { value: '$', name: '美元', label: '$ 美元' },
  { value: 'HK$', name: '港币', label: 'HK$ 港币' },
  { value: '€', name: '欧元', label: '€ 欧元' },
]

/** 10 档网格：阈值严格递增，比例合计正好 100。 */
export const GRID_10: Array<[number, number]> = [
  [5, 4],
  [10, 5],
  [16, 6],
  [22, 7],
  [28, 8],
  [35, 9],
  [42, 10],
  [50, 13],
  [58, 18],
  [68, 20],
]

export function gridLayers(rows: GridRows = GRID_10): GridLayerDTO[] {
  return rows.map(([threshold, allocation], index) => ({
    key: `g${index + 1}`,
    threshold,
    allocation,
    voided: false,
  }))
}

/** 按指数名称取它自己那套网格；查不到（含自定义标的）用默认 10 档。 */
export function gridForIndex(name: string, market: 'cn' | 'us'): GridRows {
  const preset = indicesFor(market).find((p) => p.name === name)
  return preset?.grid ?? GRID_10
}

interface TemplateAsset extends IndexPreset {
  weight: number
  baseAmount: number
}

export interface PlanTemplate {
  key: 'cn' | 'us'
  label: string
  description: string
  market: 'cn' | 'us'
  currency: string
  currencyName: string
  unit: number
  totalBudget: number
  frequency: 'weekly' | 'biweekly' | 'monthly'
  gateThreshold: number
  gateMode: 'alert' | 'grid-only' | 'block-all'
  weightMethod: 'target-weight' | 'invested-weight'
  gridAggregation: 'highest' | 'cumulative'
  assets: TemplateAsset[]
}

export const PLAN_TEMPLATES: PlanTemplate[] = [
  {
    key: 'cn',
    label: 'A 股模板',
    description: '上证综指 / 中证1000 / 中证2000 / 创业板指，每个指数一套网格',
    market: 'cn',
    currency: '¥',
    currencyName: '人民币',
    unit: 100,
    totalBudget: 300_000,
    frequency: 'monthly',
    gateThreshold: 18,
    gateMode: 'grid-only',
    weightMethod: 'target-weight',
    gridAggregation: 'highest',
    assets: [
      { name: '上证综指', code: '000001.SH', market: 'cn', weight: 30, baseAmount: 3000 },
      { name: '中证 1000', code: '000852.CSI', market: 'cn', weight: 30, baseAmount: 3000 },
      { name: '中证 2000', code: '932000.CSI', market: 'cn', weight: 15, baseAmount: 1500 },
      { name: '创业板指', code: '399006.SZ', market: 'cn', weight: 25, baseAmount: 2500 },
    ],
  },
  {
    key: 'us',
    label: '美股模板',
    description: '标普500 / 纳指 / 道指 / SSO / QLD / KO，每个标的单独一套网格',
    market: 'us',
    currency: '$',
    currencyName: '美元',
    unit: 1,
    totalBudget: 50_000,
    frequency: 'biweekly',
    gateThreshold: 20,
    gateMode: 'grid-only',
    weightMethod: 'target-weight',
    gridAggregation: 'highest',
    assets: [
      { name: '标普 500', code: 'S&P 500 / SPY', market: 'us', weight: 25, baseAmount: 500 },
      { name: '纳斯达克综合指数', code: 'NASDAQ / QQQ', market: 'us', weight: 20, baseAmount: 400 },
      { name: '道琼斯工业平均指数', code: 'DJI / DIA', market: 'us', weight: 15, baseAmount: 300 },
      { name: '标普 500 2 倍做多', code: 'SSO / 2× S&P 500', market: 'us', weight: 10, baseAmount: 120, leveraged: true },
      { name: '纳指 2 倍做多', code: 'QLD / 2× NASDAQ', market: 'us', weight: 10, baseAmount: 120, leveraged: true },
      { name: '可口可乐', code: 'KO', market: 'us', weight: 20, baseAmount: 250 },
    ],
  },
]

export function findTemplate(key: string): PlanTemplate | undefined {
  return PLAN_TEMPLATES.find((t) => t.key === key)
}

export function indicesFor(market: string): IndexPreset[] {
  return market === 'us' ? US_INDICES : CN_INDICES
}

/**
 * 把模板转成标的 payload。
 * 参考高点、点位、已投入、当前市值一律留空：模板只给结构和比例，
 * 用旧高点当参考位会误导判断。
 */
export function templateAssets(template: PlanTemplate): GridAssetPayload[] {
  return template.assets.map((asset) => ({
    name: asset.name,
    code: asset.code,
    leveraged: asset.leveraged ?? false,
    enabled: true,
    current_point: null,
    high_point: null,
    weight: asset.weight,
    invested: 0,
    market_value: 0,
    base_amount: asset.baseAmount,
    grid: gridLayers(gridForIndex(asset.name, template.market)),
    executed_layers: [],
  }))
}

/** 新建计划时直接用的完整请求体。 */
export function templatePlanPayload(template: PlanTemplate, name: string): SaveGridPlanPayload {
  return {
    name,
    market: template.market,
    currency: template.currency,
    currency_name: template.currencyName,
    unit: template.unit,
    total_budget: template.totalBudget,
    frequency: template.frequency,
    next_date: '',
    gate_threshold: template.gateThreshold,
    gate_mode: template.gateMode,
    weight_method: 'target-weight',
    grid_aggregation: 'highest',
    unallocated_invested: 0,
    sort_order: 0,
    assets: templateAssets(template),
  }
}
