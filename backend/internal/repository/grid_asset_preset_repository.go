package repository

import (
	"time"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GridAssetPresetRepository 自建标的联想项的持久化。所有操作都带 user_id。
type GridAssetPresetRepository struct {
	db *gorm.DB
}

func NewGridAssetPresetRepository(db *gorm.DB) *GridAssetPresetRepository {
	return &GridAssetPresetRepository{db: db}
}

// ListByMarket 某市场下未删除的自建标的，按名称排序，便于下拉稳定展示。
func (r *GridAssetPresetRepository) ListByMarket(userID int64, market string) ([]domain.GridAssetPreset, error) {
	var items []domain.GridAssetPreset
	err := r.db.Where("user_id = ? AND market = ? AND deleted_at IS NULL", userID, market).
		Order("name ASC, id ASC").
		Find(&items).Error
	return items, err
}

// UpsertMany 批量写入，同名只更新代码与杠杆标记。
// 已软删除的同名项会被忽略——用户主动移除过，就不该再被自动收集复活。
func (r *GridAssetPresetRepository) UpsertMany(userID int64, market string, items []domain.GridAssetPreset) error {
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		items[i].UserID = userID
		items[i].Market = market
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "market"}, {Name: "name"}},
		// 另一个计划里同名标的可能没填代码/杠杆，不能让空值把已有的信息冲掉
		DoUpdates: clause.Assignments(map[string]interface{}{
			"code":       gorm.Expr("CASE WHEN EXCLUDED.code <> '' THEN EXCLUDED.code ELSE grid_asset_presets.code END"),
			"leveraged":  gorm.Expr("grid_asset_presets.leveraged OR EXCLUDED.leveraged"),
			"updated_at": gorm.Expr("now()"),
		}),
		// 软删除行同样命中唯一约束，跳过它们，否则用户移除过的项会立刻复活
		Where: clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "grid_asset_presets.deleted_at IS NULL"}}},
	}).Create(&items).Error
}

// Delete 软删除一条，找不到或不属于该用户都返回 false。
func (r *GridAssetPresetRepository) Delete(userID, id int64) (bool, error) {
	res := r.db.Model(&domain.GridAssetPreset{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Update("deleted_at", time.Now())
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
