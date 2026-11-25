package repository

import (
	"encoding/json"
	"errors"
	"time"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

// ErrAssetGone 目标标的已被删除或不属于该计划。写入方据此中止，避免记录与金额脱节。
var ErrAssetGone = errors.New("grid asset not found")

// ErrVersionConflict 保存时版本号对不上：别处（另一标签页/另一设备）已经改过这份计划。
var ErrVersionConflict = errors.New("grid plan version conflict")

// GridPlanRepository 计划的持久化。所有查询都必须带 user_id，仓储层不信任调用方传入的 planID。
type GridPlanRepository struct {
	db *gorm.DB
}

func NewGridPlanRepository(db *gorm.DB) *GridPlanRepository {
	return &GridPlanRepository{db: db}
}

// WithTx 在一个事务里执行 fn。fn 收到的是绑定到该事务的仓储副本，
// 闭包里必须用它而不是外层仓储，否则写操作会跑到事务外面。
func (r *GridPlanRepository) WithTx(fn func(*GridPlanRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(&GridPlanRepository{db: tx})
	})
}

// Summaries 计划列表，带投入/市值/盈亏汇总。一次查询搞定，避免 N+1。
// viewerID == nil 表示未登录访客，此时只返回 public 计划；否则按可见性返回。
func (r *GridPlanRepository) Summaries(ownerID int64, viewerID *int64) ([]domain.PlanSummary, error) {
	var rows []domain.PlanSummary
	q := r.db.Model(&domain.GridPlan{}).
		Select(`grid_plans.id,
		        grid_plans.name,
		        grid_plans.market,
		        grid_plans.currency,
		        grid_plans.currency_name,
		        grid_plans.visibility,
		        grid_plans.total_budget,
		        grid_plans.sort_order,
		        grid_plans.updated_at,
		        COALESCE(SUM(CASE WHEN grid_assets.enabled THEN grid_assets.invested ELSE 0 END), 0) AS invested,
		        COALESCE(SUM(CASE WHEN grid_assets.enabled THEN grid_assets.market_value ELSE 0 END), 0) AS market_value,
		        COALESCE(SUM(CASE WHEN grid_assets.enabled THEN grid_assets.market_value - grid_assets.invested ELSE 0 END), 0) AS pnl,
		        COUNT(grid_assets.id) AS asset_count`).
		Joins("LEFT JOIN grid_assets ON grid_assets.plan_id = grid_plans.id").
		Where("grid_plans.user_id = ?", ownerID)
	if !isOwnerViewer(ownerID, viewerID) {
		if viewerID == nil {
			q = q.Where("grid_plans.visibility = ?", domain.PlanVisibilityPublic)
		} else {
			q = q.Where(visibleToViewerSQL, *viewerID, *viewerID, *viewerID)
		}
	}
	err := q.Group("grid_plans.id").
		Order("grid_plans.sort_order ASC, grid_plans.id ASC").
		Scan(&rows).Error
	return rows, err
}

// visibleToViewerSQL 非作者视角的可见性条件，参数依次是 viewerID 三次
// （好友分支两个占位、指定可见一个占位）。
// 好友关系与指定可见都走 EXISTS，和文章列表的判定保持一致。
const visibleToViewerSQL = `(
	grid_plans.visibility = 'public'
	OR grid_plans.visibility = 'login'
	OR (grid_plans.visibility = 'friends' AND EXISTS (
		SELECT 1 FROM friendships
		WHERE status = 'accepted'
		AND ((user_id = grid_plans.user_id AND friend_id = ?) OR (user_id = ? AND friend_id = grid_plans.user_id))
	))
	OR (grid_plans.visibility = 'specified' AND EXISTS (
		SELECT 1 FROM grid_plan_visible_users
		WHERE plan_id = grid_plans.id AND user_id = ?
	))
)`

func isOwnerViewer(ownerID int64, viewerID *int64) bool {
	return viewerID != nil && *viewerID == ownerID
}

// GetByID 读取计划详情（含标的）。找不到或不属于该用户都返回 (nil, nil)。
func (r *GridPlanRepository) GetByID(userID, planID int64) (*domain.GridPlan, error) {
	plan, err := r.loadPlan(r.db, planID)
	if err != nil || plan == nil {
		return nil, err
	}
	if plan.UserID != userID {
		return nil, nil
	}
	if err := r.loadVisibleUsers(r.db, []*domain.GridPlan{plan}); err != nil {
		return nil, err
	}
	return plan, nil
}

// GetAny 按 id 读取计划，不校验归属：可见性判定交给 service 层，判定通过后再用。
func (r *GridPlanRepository) GetAny(planID int64) (*domain.GridPlan, error) {
	plan, err := r.loadPlan(r.db, planID)
	if err != nil || plan == nil {
		return nil, err
	}
	if err := r.loadVisibleUsers(r.db, []*domain.GridPlan{plan}); err != nil {
		return nil, err
	}
	return plan, nil
}

func (r *GridPlanRepository) loadPlan(db *gorm.DB, planID int64) (*domain.GridPlan, error) {
	var plan domain.GridPlan
	err := db.
		Preload("Assets", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Where("id = ?", planID).
		First(&plan).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// loadVisibleUsers 一次性把多份计划的指定可见用户填进去，避免逐条查询。
func (r *GridPlanRepository) loadVisibleUsers(db *gorm.DB, plans []*domain.GridPlan) error {
	ids := make([]int64, 0, len(plans))
	for _, plan := range plans {
		if plan != nil && plan.Visibility == domain.PlanVisibilitySpecified {
			ids = append(ids, plan.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var rows []struct {
		PlanID int64
		UserID int64
	}
	if err := db.Table("grid_plan_visible_users").
		Select("plan_id, user_id").
		Where("plan_id IN ?", ids).
		Scan(&rows).Error; err != nil {
		return err
	}
	byPlan := make(map[int64][]int64, len(ids))
	for _, row := range rows {
		byPlan[row.PlanID] = append(byPlan[row.PlanID], row.UserID)
	}
	for _, plan := range plans {
		if plan == nil || plan.Visibility != domain.PlanVisibilitySpecified {
			continue
		}
		plan.VisibleUserIDs = byPlan[plan.ID]
		if plan.VisibleUserIDs == nil {
			plan.VisibleUserIDs = []int64{}
		}
	}
	return nil
}

// Create 新建计划。assets 里的 PlanID 会被覆盖为新建的 id。
func (r *GridPlanRepository) Create(plan *domain.GridPlan) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(plan).Error; err != nil {
			return err
		}
		return syncPlanVisibleUsers(tx, plan)
	})
}

// syncPlanVisibleUsers 只在 specified 档写明细，其它档一律清空。
// 显式带空数组 [] 表示「用户清空了选择」：同样删除全部授权。
func syncPlanVisibleUsers(tx *gorm.DB, plan *domain.GridPlan) error {
	if plan.Visibility != domain.PlanVisibilitySpecified {
		return tx.Where("plan_id = ?", plan.ID).Delete(&domain.GridPlanVisibleUser{}).Error
	}
	if plan.VisibleUserIDs == nil {
		// nil = 老客户端没带这个字段，保留现有授权不动
		return nil
	}
	if err := tx.Where("plan_id = ?", plan.ID).Delete(&domain.GridPlanVisibleUser{}).Error; err != nil {
		return err
	}
	if len(plan.VisibleUserIDs) == 0 {
		return nil
	}
	rows := make([]domain.GridPlanVisibleUser, 0, len(plan.VisibleUserIDs))
	seen := make(map[int64]struct{}, len(plan.VisibleUserIDs))
	for _, userID := range plan.VisibleUserIDs {
		if userID <= 0 {
			continue
		}
		if _, dup := seen[userID]; dup {
			continue
		}
		seen[userID] = struct{}{}
		rows = append(rows, domain.GridPlanVisibleUser{PlanID: plan.ID, UserID: userID})
	}
	if len(rows) == 0 {
		return nil
	}
	return tx.Create(&rows).Error
}

// IsVisibleToUser 判断 specified 档是否放行给某个登录用户。
func (r *GridPlanRepository) IsVisibleToUser(planID, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&domain.GridPlanVisibleUser{}).
		Where("plan_id = ? AND user_id = ?", planID, userID).
		Count(&count).Error
	return count > 0, err
}

// PurgeVisibleUsers 取消好友时回收授权：双向都清，谁给谁授权过都算。
// 只删 grid_plan_visible_users，计划本身不动；下次编辑时 validateVisibility 也会重新校验好友关系。
func (r *GridPlanRepository) PurgeVisibleUsers(userA, userB int64) error {
	return r.db.Exec(`DELETE FROM grid_plan_visible_users
		WHERE (user_id = ? AND plan_id IN (SELECT id FROM grid_plans WHERE user_id = ?))
		   OR (user_id = ? AND plan_id IN (SELECT id FROM grid_plans WHERE user_id = ?))`,
		userA, userB, userB, userA).Error
}

// Update 整体更新：计划参数 + 标的列表（先删后插），同一事务。
// expectedVersion > 0 时做乐观锁校验：版本不一致说明别处已经改过，返回 ErrVersionConflict。
func (r *GridPlanRepository) Update(plan *domain.GridPlan, expectedVersion int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		q := tx.Model(&domain.GridPlan{}).Where("id = ? AND user_id = ?", plan.ID, plan.UserID)
		if expectedVersion > 0 {
			q = q.Where("version = ?", expectedVersion)
		}
		res := q.
			Select("name", "market", "currency", "currency_name", "unit", "total_budget",
				"frequency", "next_date", "gate_threshold", "gate_mode", "weight_method",
				"grid_aggregation", "unallocated_invested", "sort_order", "visibility", "version", "updated_at").
			Updates(map[string]interface{}{
				"name":                 plan.Name,
				"market":               plan.Market,
				"currency":             plan.Currency,
				"currency_name":        plan.CurrencyName,
				"unit":                 plan.Unit,
				"total_budget":         plan.TotalBudget,
				"frequency":            plan.Frequency,
				"next_date":            nextDateValue(plan.NextDate),
				"gate_threshold":       plan.GateThreshold,
				"gate_mode":            plan.GateMode,
				"weight_method":        plan.WeightMethod,
				"grid_aggregation":     plan.GridAggregation,
				"unallocated_invested": plan.UnallocatedInvested,
				"sort_order":           plan.SortOrder,
				"visibility":           plan.Visibility,
				"version":              gorm.Expr("version + 1"),
				"updated_at":           time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			if expectedVersion > 0 {
				return ErrVersionConflict
			}
			return gorm.ErrRecordNotFound
		}
		if err := syncPlanVisibleUsers(tx, plan); err != nil {
			return err
		}

		// 标的按 id 增量同步：已有 id 的原地更新，没有的插入，本次没出现的删除。
		// 绝对不能整体删除重建——grid_plan_executions.asset_id 是 ON DELETE SET NULL，
		// 每次保存都会把历史执行记录的归属清空，撤销时也就退不回已投入了。
		var existingIDs []int64
		if err := tx.Model(&domain.GridAsset{}).Where("plan_id = ?", plan.ID).Pluck("id", &existingIDs).Error; err != nil {
			return err
		}
		owned := make(map[int64]struct{}, len(existingIDs))
		for _, id := range existingIDs {
			owned[id] = struct{}{}
		}
		kept := make(map[int64]struct{}, len(plan.Assets))
		for i := range plan.Assets {
			asset := plan.Assets[i]
			asset.PlanID = plan.ID
			asset.SortOrder = i
			if asset.ID > 0 {
				if _, ok := owned[asset.ID]; ok {
					kept[asset.ID] = struct{}{}
					// Select 显式列：gorm 对 struct Updates 默认跳过零值，加上 Select 才会连 0/null 一起写。
					if err := tx.Model(&domain.GridAsset{}).
						Where("id = ? AND plan_id = ?", asset.ID, plan.ID).
						Select("name", "code", "leveraged", "enabled", "current_point", "high_point",
							"weight", "invested", "market_value", "base_amount", "grid",
							"executed_layers", "sort_order", "updated_at").
						Updates(&asset).Error; err != nil {
						return err
					}
					continue
				}
				// 不是本计划的 id，按新增处理，避免跨计划改别人的标的
				asset.ID = 0
			}
			asset.CreatedAt = time.Time{}
			asset.UpdatedAt = time.Time{}
			if err := tx.Create(&asset).Error; err != nil {
				return err
			}
		}
		removed := make([]int64, 0, len(existingIDs))
		for _, id := range existingIDs {
			if _, ok := kept[id]; !ok {
				removed = append(removed, id)
			}
		}
		if len(removed) > 0 {
			if err := tx.Where("plan_id = ? AND id IN ?", plan.ID, removed).Delete(&domain.GridAsset{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete 删除计划，标的/快照/执行记录由外键级联清理。
func (r *GridPlanRepository) Delete(userID, planID int64) (bool, error) {
	res := r.db.Where("id = ? AND user_id = ?", planID, userID).Delete(&domain.GridPlan{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// UpsertSnapshot 写入当天快照；同一天重复写入只保留最后一次。
func (r *GridPlanRepository) UpsertSnapshot(snap *domain.GridPlanSnapshot) error {
	return r.db.Exec(`INSERT INTO grid_plan_snapshots (plan_id, snapshot_date, invested, market_value, pnl)
		VALUES (?, ?::date, ?, ?, ?)
		ON CONFLICT (plan_id, snapshot_date) DO UPDATE
		SET invested = EXCLUDED.invested,
		    market_value = EXCLUDED.market_value,
		    pnl = EXCLUDED.pnl`,
		snap.PlanID, snap.SnapshotDate.String(), snap.Invested, snap.MarketValue, snap.Pnl).Error
}

// ListSnapshots 按日期升序返回最近 days 天的快照。
func (r *GridPlanRepository) ListSnapshots(planID int64, days int) ([]domain.GridPlanSnapshot, error) {
	q := r.db.Where("plan_id = ?", planID)
	if days > 0 {
		q = q.Where("snapshot_date >= CURRENT_DATE - ?::int", days)
	}
	var snapshots []domain.GridPlanSnapshot
	err := q.Order("snapshot_date ASC").Find(&snapshots).Error
	return snapshots, err
}

// AddExecution 写入执行记录。
func (r *GridPlanRepository) AddExecution(exec *domain.GridPlanExecution) error {
	return r.db.Create(exec).Error
}

// LastExecution 最近一条执行记录，没有则返回 (nil, nil)。
func (r *GridPlanRepository) LastExecution(planID int64) (*domain.GridPlanExecution, error) {
	var exec domain.GridPlanExecution
	err := r.db.Where("plan_id = ?", planID).Order("executed_at DESC, id DESC").First(&exec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &exec, nil
}

// DeleteExecution 删除指定执行记录。
func (r *GridPlanRepository) DeleteExecution(planID, execID int64) error {
	return r.db.Where("id = ? AND plan_id = ?", execID, planID).Delete(&domain.GridPlanExecution{}).Error
}

// UpdateAssetsInvested 批量写回标的的已投入金额与已执行网格层。
func (r *GridPlanRepository) UpdateAssetsInvested(planID int64, assetID int64, invested float64, executed []string) error {
	// executed_layers 是 jsonb，map[string]any 里的 []string 会被 pgx 当成 record/text[] 处理，
	// 必须显式序列化成 JSON 字符串再 ::jsonb。
	if executed == nil {
		executed = []string{}
	}
	payload, err := json.Marshal(executed)
	if err != nil {
		return err
	}
	res := r.db.Exec(`UPDATE grid_assets
		SET invested = ?, executed_layers = ?::jsonb, updated_at = now()
		WHERE id = ? AND plan_id = ?`,
		invested, string(payload), assetID, planID)
	if res.Error != nil {
		return res.Error
	}
	// 标的可能已被另一处保存删掉；静默成功会让执行记录和金额对不上。
	if res.RowsAffected == 0 {
		return ErrAssetGone
	}
	return nil
}

// LockAssetsInvested 用 SELECT ... FOR UPDATE 锁住标的行并返回已投入与已执行层。
// 执行/撤销都基于这两个值做累加，锁住才能避免并发时的读-改-写丢更新。
func (r *GridPlanRepository) LockAssetsInvested(planID, assetID int64) (float64, []string, error) {
	var row struct {
		Invested       float64
		ExecutedLayers []byte
	}
	err := r.db.Raw(`SELECT invested, executed_layers FROM grid_assets
		WHERE id = ? AND plan_id = ? FOR UPDATE`, assetID, planID).Scan(&row).Error
	if err != nil {
		return 0, nil, err
	}
	var layers []string
	if len(row.ExecutedLayers) > 0 {
		if err := json.Unmarshal(row.ExecutedLayers, &layers); err != nil {
			return 0, nil, err
		}
	}
	if layers == nil {
		layers = []string{}
	}
	return row.Invested, layers, nil
}

// UpdatePlanUnallocated 更新「已投入未分配」金额。
func (r *GridPlanRepository) UpdatePlanUnallocated(planID int64, value float64) error {
	res := r.db.Exec(`UPDATE grid_plans SET unallocated_invested = ?, updated_at = now() WHERE id = ?`,
		value, planID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdatePlanNextDate 更新下一次定投日（撤销/推进时使用）。
func (r *GridPlanRepository) UpdatePlanNextDate(planID int64, nextDate domain.DateOnly) error {
	return r.db.Exec(`UPDATE grid_plans SET next_date = ?, updated_at = now() WHERE id = ?`,
		nextDateValue(nextDate), planID).Error
}

// nextDateValue 空日期要写 NULL，不能写空字符串，否则 date 列直接报错。
func nextDateValue(d domain.DateOnly) any {
	if !d.Valid {
		return nil
	}
	return d.String()
}

// BumpPlanVersion 执行/撤销后递增版本号：这些操作改动了标的的已投入与已执行层，
// 手里还拿着保存前版本的编辑页不该再整体覆盖回写。
func (r *GridPlanRepository) BumpPlanVersion(planID int64) error {
	return r.db.Exec(`UPDATE grid_plans SET version = version + 1 WHERE id = ?`, planID).Error
}
