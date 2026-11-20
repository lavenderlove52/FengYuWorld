/** 与后端 dto/grid_plan_dto.go 一一对应，字段名保持后端 snake_case。 */

export type Frequency = 'weekly' | 'biweekly' | 'monthly'
export type GateMode = 'alert' | 'grid-only' | 'block-all'
export type WeightMethod = 'target-weight' | 'invested-weight'
export type GridAggregation = 'highest' | 'cumulative'

export interface GridLayerDTO {
  key: string
  threshold: number
  allocation: number
  /** 已作废：不参与触发，也不计入资金比例合计 */
  voided: boolean
}

/** 读接口返回的标的。 */
export interface GridAssetDTO {
  id: number
  name: string
  code: string
  leveraged: boolean
  enabled: boolean
  current_point: number | null
  high_point: number | null
  weight: number
  invested: number
  market_value: number
  base_amount: number
  grid: GridLayerDTO[]
  executed_layers: string[]
}

/** 写接口的标的。id 给已有标的时后端原地更新（新标的省略）。 */
export interface GridAssetPayload {
  id?: number
  name: string
  code?: string
  leveraged?: boolean
  enabled?: boolean
  current_point?: number | null
  high_point?: number | null
  weight?: number
  invested?: number
  market_value?: number
  base_amount?: number
  grid?: GridLayerDTO[]
  executed_layers?: string[]
}

export interface GridPlanSummary {
  id: number
  name: string
  market: string
  currency: string
  currency_name: string
  total_budget: number
  invested: number
  market_value: number
  pnl: number
  asset_count: number
  sort_order: number
  updated_at: string
}

export interface GridPlanDetail {
  id: number
  name: string
  market: string
  currency: string
  currency_name: string
  unit: number
  total_budget: number
  frequency: Frequency
  /** yyyy-MM-dd，未设置时为空字符串 */
  next_date: string
  gate_threshold: number
  gate_mode: GateMode
  weight_method: WeightMethod
  grid_aggregation: GridAggregation
  unallocated_invested: number
  sort_order: number
  /** 乐观锁版本号：保存时原样带回，后端比对不上说明别处已改 */
  version: number
  assets: GridAssetDTO[]
  created_at: string
  updated_at: string
}

/** 新建与整体保存共用。标的列表是整体替换语义。 */
export interface SaveGridPlanPayload {
  name: string
  market?: string
  currency?: string
  currency_name?: string
  unit?: number
  total_budget?: number
  frequency?: Frequency
  next_date?: string
  gate_threshold?: number
  gate_mode?: GateMode
  weight_method?: WeightMethod
  grid_aggregation?: GridAggregation
  unallocated_invested?: number
  sort_order?: number
  /** 读到的版本号，后端据此判断是否被别人改过；省略表示不校验 */
  version?: number
  assets?: GridAssetPayload[]
}

export interface GridSnapshot {
  id: number
  /** yyyy-MM-dd */
  snapshot_date: string
  invested: number
  market_value: number
  pnl: number
  created_at: string
}

export interface GridExecution {
  id: number
  /** 标的名被删除后为 null */
  asset_id: number | null
  asset_name: string
  amount: number
  layer_keys: string[]
  was_regular: boolean
  prev_next_date: string
  note: string
  executed_at: string
}

export interface RecordExecutionPayload {
  /** 0 表示记到「已投入未分配」 */
  asset_id: number
  amount: number
  layer_keys?: string[]
  was_regular?: boolean
  note?: string
}

/** 账号下的自建标的，按市场分开，只用于下拉联想。 */
export interface GridAssetPreset {
  id: number
  market: string
  name: string
  code: string
  leveraged: boolean
}
