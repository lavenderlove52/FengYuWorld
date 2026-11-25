package domain

import "time"

// GridLayer 网格层参数。整体以 JSONB 存在 grid_assets.grid 里。
// Voided 表示这层已作废：不参与触发、不计入资金比例合计；老数据没有该字段视为生效。
type GridLayer struct {
	Key        string  `json:"key"`
	Threshold  float64 `json:"threshold"`
	Allocation float64 `json:"allocation"`
	Voided     bool    `json:"voided,omitempty"`
}

// 投资计划对外可见范围。默认 private：不主动改，个人主页就不会展示。
const (
	PlanVisibilityPrivate   = "private"
	PlanVisibilityLogin     = "login"
	PlanVisibilityFriends   = "friends"
	PlanVisibilitySpecified = "specified"
	PlanVisibilityPublic    = "public"
)

// GridPlan 一份投资计划，对应一个独立资金池。
type GridPlan struct {
	ID           int64    `gorm:"primaryKey"                        json:"id"`
	UserID       int64    `gorm:"not null;index"                    json:"user_id"`
	Name         string   `gorm:"size:80;not null"                  json:"name"`
	Market       string   `gorm:"size:20;not null;default:cn"       json:"market"`
	Currency     string   `gorm:"size:10;not null;default:¥"        json:"currency"`
	CurrencyName string   `gorm:"size:20;not null;default:人民币"    json:"currency_name"`
	Unit         float64  `gorm:"type:numeric(18,4);not null;default:100" json:"unit"`
	TotalBudget  float64  `gorm:"type:numeric(18,2);not null;default:0"   json:"total_budget"`
	Frequency    string   `gorm:"size:20;not null;default:monthly"  json:"frequency"`
	NextDate     DateOnly `gorm:"type:date"                         json:"next_date,omitempty"`
	// 不写 gorm default：否则 0（不设闸门）会被 GORM 当成零值跳过，落库变成 20
	GateThreshold       float64 `gorm:"type:numeric(6,2);not null" json:"gate_threshold"`
	GateMode            string  `gorm:"size:20;not null;default:grid-only" json:"gate_mode"`
	WeightMethod        string  `gorm:"size:20;not null;default:target-weight" json:"weight_method"`
	GridAggregation     string  `gorm:"size:20;not null;default:highest"  json:"grid_aggregation"`
	UnallocatedInvested float64 `gorm:"type:numeric(18,2);not null;default:0" json:"unallocated_invested"`
	SortOrder           int     `gorm:"not null;default:0"                json:"sort_order"`
	// Visibility 对外可见范围，见 PlanVisibility* 常量
	Visibility string `gorm:"size:20;not null;default:private" json:"visibility"`
	// Version 乐观锁：每次写库 +1，保存时要求与客户端读到的版本一致。
	Version   int64     `gorm:"not null;default:1"                      json:"version"`
	CreatedAt time.Time `gorm:"not null;default:now()"            json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()"            json:"updated_at"`

	Assets []GridAsset `gorm:"foreignKey:PlanID;constraint:OnDelete:CASCADE" json:"assets,omitempty"`
	// VisibleUserIDs specified 档下可见的用户；单独存 grid_plan_visible_users，不在这张表里
	VisibleUserIDs []int64 `gorm:"-" json:"visible_user_ids,omitempty"`
}

func (GridPlan) TableName() string { return "grid_plans" }

// GridAsset 计划下的一个标的。
type GridAsset struct {
	ID             int64       `gorm:"primaryKey"                            json:"id"`
	PlanID         int64       `gorm:"not null;index"                        json:"plan_id"`
	Name           string      `gorm:"size:80;not null"                      json:"name"`
	Code           string      `gorm:"size:60;not null;default:''"           json:"code"`
	Leveraged      bool        `gorm:"not null;default:false"                json:"leveraged"`
	Enabled        bool        `gorm:"not null;default:true"                 json:"enabled"`
	CurrentPoint   *float64    `gorm:"type:numeric(18,4)"                    json:"current_point"`
	HighPoint      *float64    `gorm:"type:numeric(18,4)"                    json:"high_point"`
	Weight         float64     `gorm:"type:numeric(9,6);not null;default:0"  json:"weight"`
	Invested       float64     `gorm:"type:numeric(18,2);not null;default:0" json:"invested"`
	MarketValue    float64     `gorm:"type:numeric(18,2);not null;default:0" json:"market_value"`
	// RealizedOut 累计卖出到账金额，只由记录卖出累加；已投入仍是累计买入，不会被卖出扣减。
	RealizedOut    float64     `gorm:"type:numeric(18,2);not null;default:0" json:"realized_out"`
	BaseAmount     float64     `gorm:"type:numeric(18,2);not null;default:0" json:"base_amount"`
	Grid           []GridLayer `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"grid"`
	ExecutedLayers []string    `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"executed_layers"`
	SortOrder      int         `gorm:"not null;default:0"                    json:"sort_order"`
	CreatedAt      time.Time   `gorm:"not null;default:now()"                json:"created_at"`
	UpdatedAt      time.Time   `gorm:"not null;default:now()"                json:"updated_at"`
}

func (GridAsset) TableName() string { return "grid_assets" }

// GridPlanSnapshot 计划某一天的净值快照，(plan_id, snapshot_date) 唯一。
type GridPlanSnapshot struct {
	ID           int64     `gorm:"primaryKey"                                json:"id"`
	PlanID       int64     `gorm:"not null;index"                            json:"plan_id"`
	SnapshotDate DateOnly  `gorm:"type:date;not null"                        json:"snapshot_date"`
	Invested     float64   `gorm:"type:numeric(18,2);not null"               json:"invested"`
	MarketValue  float64   `gorm:"type:numeric(18,2);not null"               json:"market_value"`
	Pnl          float64   `gorm:"type:numeric(18,2);not null"               json:"pnl"`
	CreatedAt    time.Time `gorm:"not null;default:now()"                    json:"created_at"`
}

func (GridPlanSnapshot) TableName() string { return "grid_plan_snapshots" }

// GridPlanExecution 一次执行记录。删标的后 asset_id 置空，记录仍保留。
const (
	ExecutionSideBuy  = "buy"
	ExecutionSideSell = "sell"
)

type GridPlanExecution struct {
	ID           int64     `gorm:"primaryKey"                             json:"id"`
	PlanID       int64     `gorm:"not null;index"                         json:"plan_id"`
	AssetID      *int64    `                                             json:"asset_id"`
	AssetName    string    `gorm:"size:80;not null"                        json:"asset_name"`
	Amount       float64   `gorm:"type:numeric(18,2);not null"             json:"amount"`
	// Side 执行方向：buy / sell。老记录默认 buy。
	Side         string    `gorm:"size:8;not null;default:'buy'"            json:"side"`
	// Point 成交时的点位快照，空表示没记（老记录都没有）。
	Point        *float64  `gorm:"type:numeric(18,4)"                      json:"point"`
	LayerKeys    []string  `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"layer_keys"`
	WasRegular   bool      `gorm:"not null;default:false"                  json:"was_regular"`
	PrevNextDate DateOnly  `gorm:"type:date"                              json:"prev_next_date,omitempty"`
	Note         string    `gorm:"size:200;not null;default:''"            json:"note"`
	ExecutedAt   time.Time `gorm:"not null;default:now()"                  json:"executed_at"`
}

func (GridPlanExecution) TableName() string { return "grid_plan_executions" }

// GridPlanVisibleUser specified 档下能看到该计划的用户。
type GridPlanVisibleUser struct {
	PlanID    int64     `gorm:"primaryKey"                  json:"plan_id"`
	UserID    int64     `gorm:"primaryKey"                  json:"user_id"`
	CreatedAt time.Time `gorm:"not null;default:now()"      json:"created_at"`
}

func (GridPlanVisibleUser) TableName() string { return "grid_plan_visible_users" }

// ---------- 计划级汇总（列表页用，不落库） ----------

// PlanSummary 计划列表里的汇总数据。
type PlanSummary struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Market       string    `json:"market"`
	Currency     string    `json:"currency"`
	CurrencyName string    `json:"currency_name"`
	Visibility   string    `json:"visibility"`
	TotalBudget  float64   `json:"total_budget"`
	Invested     float64   `json:"invested"`
	MarketValue  float64   `json:"market_value"`
	Pnl          float64   `json:"pnl"`
	AssetCount   int       `json:"asset_count"`
	SortOrder    int       `json:"sort_order"`
	UpdatedAt    time.Time `json:"updated_at"`
}
