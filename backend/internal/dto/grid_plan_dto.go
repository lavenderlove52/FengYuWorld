package dto

// ---------- 计划 ----------

type GridPlanSummary struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Market       string  `json:"market"`
	Currency     string  `json:"currency"`
	CurrencyName string  `json:"currency_name"`
	Visibility   string  `json:"visibility"`
	TotalBudget  float64 `json:"total_budget"`
	Invested     float64 `json:"invested"`
	MarketValue  float64 `json:"market_value"`
	Pnl          float64 `json:"pnl"`
	AssetCount   int     `json:"asset_count"`
	SortOrder    int     `json:"sort_order"`
	UpdatedAt    string  `json:"updated_at"`
}

type GridLayerDTO struct {
	Key        string  `json:"key" binding:"required,max=40"`
	Threshold  float64 `json:"threshold"`
	Allocation float64 `json:"allocation"`
	Voided     bool    `json:"voided"`
}

type GridAssetDTO struct {
	ID             int64          `json:"id"`
	Name           string         `json:"name" binding:"max=80"`
	Code           string         `json:"code" binding:"max=60"`
	Leveraged      bool           `json:"leveraged"`
	Enabled        *bool          `json:"enabled"`
	CurrentPoint   *float64       `json:"current_point"`
	HighPoint      *float64       `json:"high_point"`
	Weight         float64        `json:"weight"`
	Invested       float64        `json:"invested"`
	MarketValue    float64        `json:"market_value"`
	BaseAmount     float64        `json:"base_amount"`
	Grid           []GridLayerDTO `json:"grid" binding:"dive"`
	ExecutedLayers []string       `json:"executed_layers"`
}

type GridPlanDetail struct {
	ID                  int64          `json:"id"`
	Name                string         `json:"name"`
	Market              string         `json:"market"`
	Currency            string         `json:"currency"`
	CurrencyName        string         `json:"currency_name"`
	Unit                float64        `json:"unit"`
	TotalBudget         float64        `json:"total_budget"`
	Frequency           string         `json:"frequency"`
	NextDate            string         `json:"next_date"`
	GateThreshold       float64        `json:"gate_threshold"`
	GateMode            string         `json:"gate_mode"`
	WeightMethod        string         `json:"weight_method"`
	GridAggregation     string         `json:"grid_aggregation"`
	UnallocatedInvested float64        `json:"unallocated_invested"`
	SortOrder           int            `json:"sort_order"`
	Version             int64          `json:"version"`
	Visibility          string         `json:"visibility"`
	VisibleUserIDs      []int64        `json:"visible_user_ids"`
	Assets              []GridAssetDTO `json:"assets" binding:"dive"`
	CreatedAt           string         `json:"created_at"`
	UpdatedAt           string         `json:"updated_at"`
}

// SaveGridPlanRequest 新建和整体保存共用。
// 标的列表是整体替换语义：传什么就是什么，没传的会被删掉。
type SaveGridPlanRequest struct {
	Name         string  `json:"name" binding:"required,max=80"`
	Market       string  `json:"market" binding:"max=20"`
	Currency     string  `json:"currency" binding:"max=10"`
	CurrencyName string  `json:"currency_name" binding:"max=20"`
	Unit         float64 `json:"unit"`
	TotalBudget  float64 `json:"total_budget"`
	Frequency    string  `json:"frequency" binding:"max=20"`
	NextDate     string  `json:"next_date"`
	// 指针：0 是合法阈值（表示不设闸门），只有字段缺失时才用默认 20
	GateThreshold       *float64 `json:"gate_threshold"`
	GateMode            string   `json:"gate_mode" binding:"max=20"`
	WeightMethod        string   `json:"weight_method" binding:"max=20"`
	GridAggregation     string   `json:"grid_aggregation" binding:"max=20"`
	UnallocatedInvested float64  `json:"unallocated_invested"`
	SortOrder           int      `json:"sort_order"`
	// Version 保存时带上的读到的版本号；0 表示不校验（兼容老客户端）
	Version int64 `json:"version"`
	// Visibility 对外可见范围，缺省按 private
	Visibility     string         `json:"visibility" binding:"max=20"`
	VisibleUserIDs []int64        `json:"visible_user_ids"`
	Assets         []GridAssetDTO `json:"assets" binding:"dive"`
}

// ---------- 快照 ----------

type GridSnapshotDTO struct {
	ID           int64   `json:"id"`
	SnapshotDate string  `json:"snapshot_date"`
	Invested     float64 `json:"invested"`
	MarketValue  float64 `json:"market_value"`
	Pnl          float64 `json:"pnl"`
	CreatedAt    string  `json:"created_at"`
}

// RecordSnapshotRequest date 为空时按服务器当天算。
type RecordSnapshotRequest struct {
	Date string `json:"date"`
}

// ---------- 执行记录 ----------

type GridExecutionDTO struct {
	ID           int64    `json:"id"`
	AssetID      *int64   `json:"asset_id"`
	AssetName    string   `json:"asset_name"`
	Amount       float64  `json:"amount"`
	LayerKeys    []string `json:"layer_keys"`
	WasRegular   bool     `json:"was_regular"`
	PrevNextDate string   `json:"prev_next_date"`
	Note         string   `json:"note"`
	ExecutedAt   string   `json:"executed_at"`
}

// RecordExecutionRequest asset_id 为 0 表示记到「已投入未分配」。
type RecordExecutionRequest struct {
	AssetID    int64    `json:"asset_id"`
	Amount     float64  `json:"amount" binding:"required"`
	LayerKeys  []string `json:"layer_keys"`
	WasRegular bool     `json:"was_regular"`
	Note       string   `json:"note" binding:"max=200"`
}

// ---------- 自建标的 ----------

type GridAssetPresetDTO struct {
	ID        int64  `json:"id"`
	Market    string `json:"market"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Leveraged bool   `json:"leveraged"`
}
