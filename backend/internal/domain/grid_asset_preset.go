package domain

import "time"

// GridAssetPreset 用户自建的标的联想项。按市场分开，只用于下拉，不参与计划计算。
type GridAssetPreset struct {
	ID        int64     `gorm:"primaryKey"                          json:"id"`
	UserID    int64     `gorm:"not null;index"                      json:"user_id"`
	Market    string    `gorm:"size:20;not null;default:cn"         json:"market"`
	Name      string    `gorm:"size:80;not null"                    json:"name"`
	Code      string    `gorm:"size:60;not null;default:''"         json:"code"`
	Leveraged bool      `gorm:"not null;default:false"              json:"leveraged"`
	CreatedAt time.Time `gorm:"not null;default:now()"              json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()"              json:"updated_at"`
	// DeletedAt 非空表示用户主动移除了这一条，自动收集时不再复活它。
	DeletedAt *time.Time `json:"-"`
}

func (GridAssetPreset) TableName() string { return "grid_asset_presets" }

// 预置指数名单，用来判断某个标的是不是用户自建的。
// 与前端 core/gridplan/templates.ts 的 CN_INDICES / US_INDICES 保持一致。
var presetIndexNames = map[string]map[string]struct{}{
	"cn": {
		"上证综指":    {},
		"中证 1000": {},
		"中证 2000": {},
		"创业板指":    {},
	},
	"us": {
		"标普 500":       {},
		"纳斯达克综合指数":     {},
		"道琼斯工业平均指数":    {},
		"标普 500 2 倍做多": {},
		"纳指 2 倍做多":     {},
		"可口可乐":         {},
	},
}

// IsPresetIndex 判断名称是否为内置预置指数（自建列表不重复收录它们）。
func IsPresetIndex(market, name string) bool {
	set, ok := presetIndexNames[market]
	if !ok {
		return false
	}
	_, ok = set[name]
	return ok
}
