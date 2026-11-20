package service

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrGridPlanNotFound    = errors.New("grid plan not found")
	ErrGridPlanInvalid     = errors.New("grid plan invalid")
	ErrGridPlanNoExecution = errors.New("no execution to undo")
	// ErrGridPlanConflict 保存时版本号过期：别处已经改过这份计划，本次保存被拒绝。
	ErrGridPlanConflict = errors.New("grid plan version conflict")
)

const (
	// 金额/比例比较的容差，避免 100.000000001 这种浮点误差被判超限。
	gridEpsilon   = 1e-6
	maxGridLayers = 50
)

// ValidationError 带上具体标的，前端可以直接定位到哪一张卡片。
type ValidationError struct {
	AssetIndex int
	AssetName  string
	Field      string
	Reason     string
}

func (e *ValidationError) Error() string {
	if e.AssetName != "" {
		return fmt.Sprintf("标的「%s」%s：%s", e.AssetName, e.Field, e.Reason)
	}
	return fmt.Sprintf("%s：%s", e.Field, e.Reason)
}

func (e *ValidationError) Unwrap() error { return ErrGridPlanInvalid }

// GridPlanService 计划相关的业务规则。
type GridPlanService struct {
	repo *repository.GridPlanRepository
	// collector 可选：保存计划后把自定义标的收进账号的自建联想列表。
	collector *GridAssetPresetService
	// friends 可见性校验要确认「指定好友」确实是好友
	friends *repository.FriendshipRepository
	// policy 对外可见性；nil 时按「只有作者可见」处理
	policy *GridPlanPolicy
}

func NewGridPlanService(repo *repository.GridPlanRepository) *GridPlanService {
	return &GridPlanService{repo: repo}
}

func (s *GridPlanService) SetPresetCollector(c *GridAssetPresetService) {
	s.collector = c
}

// SetFriendship 注入好友仓储，用于校验「指定可见」的用户确实是好友。
func (s *GridPlanService) SetFriendship(f *repository.FriendshipRepository) {
	s.friends = f
}

func (s *GridPlanService) SetPolicy(p *GridPlanPolicy) {
	s.policy = p
}

// List 计划列表 + 汇总。
func (s *GridPlanService) List(userID int64) ([]domain.PlanSummary, error) {
	return s.repo.Summaries(userID, &userID)
}

// ListForViewer 某位用户对外展示的计划汇总；viewerID 为 nil 表示未登录访客。
func (s *GridPlanService) ListForViewer(ownerID int64, viewerID *int64) ([]domain.PlanSummary, error) {
	return s.repo.Summaries(ownerID, viewerID)
}

// GetForViewer 读取他人计划详情，可见性不通过时返回 ErrGridPlanNotFound（不暴露存在性）。
func (s *GridPlanService) GetForViewer(viewerID *int64, planID int64) (*domain.GridPlan, error) {
	plan, err := s.repo.GetAny(planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrGridPlanNotFound
	}
	canRead, err := s.canRead(viewerID, plan)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, ErrGridPlanNotFound
	}
	return plan, nil
}

// SnapshotsForViewer 曲线数据，同样先过可见性。
func (s *GridPlanService) SnapshotsForViewer(viewerID *int64, planID int64, days int) ([]domain.GridPlanSnapshot, error) {
	if _, err := s.GetForViewer(viewerID, planID); err != nil {
		return nil, err
	}
	return s.snapshotsByID(planID, days)
}

// GetForViewerWithSnapshots 详情 + 曲线一次取回，避免对同一份计划做两遍读取和可见性判定。
func (s *GridPlanService) GetForViewerWithSnapshots(viewerID *int64, planID int64, days int) (*domain.GridPlan, []domain.GridPlanSnapshot, error) {
	plan, err := s.GetForViewer(viewerID, planID)
	if err != nil {
		return nil, nil, err
	}
	snapshots, err := s.snapshotsByID(planID, days)
	if err != nil {
		return nil, nil, err
	}
	return plan, snapshots, nil
}

func (s *GridPlanService) canRead(viewerID *int64, plan *domain.GridPlan) (bool, error) {
	if s.policy == nil {
		return viewerID != nil && *viewerID == plan.UserID, nil
	}
	return s.policy.CanReadPlan(viewerID, plan)
}

// Get 计划详情。找不到或不属于该用户都返回 ErrGridPlanNotFound，避免暴露存在性。
func (s *GridPlanService) Get(userID, planID int64) (*domain.GridPlan, error) {
	plan, err := s.repo.GetByID(userID, planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrGridPlanNotFound
	}
	return plan, nil
}

// Create 新建计划。标的列表可以为空。
func (s *GridPlanService) Create(userID int64, plan *domain.GridPlan) (*domain.GridPlan, error) {
	plan.UserID = userID
	plan.ID = 0
	applyPlanDefaults(plan)
	for i := range plan.Assets {
		plan.Assets[i].ID = 0
		plan.Assets[i].PlanID = 0
		plan.Assets[i].SortOrder = i
		applyAssetDefaults(&plan.Assets[i])
	}
	if err := validatePlan(plan); err != nil {
		return nil, err
	}
	if err := s.validateVisibility(userID, plan, ""); err != nil {
		return nil, err
	}
	if err := s.repo.Create(plan); err != nil {
		return nil, err
	}
	return s.Get(userID, plan.ID)
}

// Update 整体保存。校验在写库之前做，失败时不落任何数据。
// incoming.Version > 0 时按乐观锁处理：与库里的版本不一致说明客户端读到的是旧数据，返回 ErrGridPlanConflict。
func (s *GridPlanService) Update(userID, planID int64, incoming *domain.GridPlan) (*domain.GridPlan, error) {
	existing, err := s.repo.GetByID(userID, planID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrGridPlanNotFound
	}

	incoming.ID = planID
	incoming.UserID = userID
	incoming.CreatedAt = existing.CreatedAt
	applyPlanDefaults(incoming)
	for i := range incoming.Assets {
		// 保留客户端带来的 id：仓储层按 id 原地更新，删掉重建会让执行记录的 asset_id 变成 NULL。
		// 负数是前端给新标的的临时占位，落库时按新增处理。
		if incoming.Assets[i].ID < 0 {
			incoming.Assets[i].ID = 0
		}
		incoming.Assets[i].PlanID = planID
		incoming.Assets[i].SortOrder = i
		applyAssetDefaults(&incoming.Assets[i])
	}
	if err := validatePlan(incoming); err != nil {
		return nil, err
	}
	if err := s.validateVisibility(userID, incoming, existing.Visibility); err != nil {
		return nil, err
	}
	if err := s.repo.Update(incoming, incoming.Version); err != nil {
		if errors.Is(err, repository.ErrVersionConflict) {
			return nil, ErrGridPlanConflict
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGridPlanNotFound
		}
		return nil, err
	}
	// 收集失败不影响保存结果：联想列表只是便利功能，不能因为它让整次保存报错。
	if s.collector != nil {
		_ = s.collector.CollectFromAssets(userID, incoming.Market, incoming.Assets)
	}
	return s.Get(userID, planID)
}

// Delete 删除计划，级联清理标的/快照/执行记录。
func (s *GridPlanService) Delete(userID, planID int64) error {
	ok, err := s.repo.Delete(userID, planID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrGridPlanNotFound
	}
	return nil
}

// Snapshots 曲线数据。
func (s *GridPlanService) Snapshots(userID, planID int64, days int) ([]domain.GridPlanSnapshot, error) {
	if _, err := s.Get(userID, planID); err != nil {
		return nil, err
	}
	return s.snapshotsByID(planID, days)
}

func (s *GridPlanService) snapshotsByID(planID int64, days int) ([]domain.GridPlanSnapshot, error) {
	if days < 0 {
		days = 0
	}
	if days > 3650 {
		days = 3650
	}
	return s.repo.ListSnapshots(planID, days)
}

// RecordSnapshot 记录今日净值。同一天重复调用只覆盖当天那条。
// 投入和市值由标的实时汇总，不接受前端传入，避免两边口径不一致。
func (s *GridPlanService) RecordSnapshot(userID, planID int64, date domain.DateOnly) (*domain.GridPlanSnapshot, error) {
	plan, err := s.Get(userID, planID)
	if err != nil {
		return nil, err
	}
	if !date.Valid {
		return nil, &ValidationError{Field: "日期", Reason: "必须是 yyyy-MM-dd 格式"}
	}

	invested, marketValue := 0.0, 0.0
	for _, asset := range plan.Assets {
		invested += asset.Invested
		marketValue += asset.MarketValue
	}

	snap := &domain.GridPlanSnapshot{
		PlanID:       planID,
		SnapshotDate: date,
		Invested:     round2(invested),
		MarketValue:  round2(marketValue),
		Pnl:          round2(marketValue - invested),
	}
	if err := s.repo.UpsertSnapshot(snap); err != nil {
		return nil, err
	}
	return snap, nil
}

// RecordExecution 记录一次执行：写执行记录、累加已投入、标记网格层、必要时推进定投日。
// assetID 为 0 表示「已投入未分配」，只累加 market.unallocated_invested。
func (s *GridPlanService) RecordExecution(
	userID, planID int64,
	assetID int64,
	amount float64,
	layerKeys []string,
	wasRegular bool,
	note string,
) (*domain.GridPlanExecution, error) {
	var exec *domain.GridPlanExecution
	err := s.repo.WithTx(func(tx *repository.GridPlanRepository) error {
		// 必须用事务句柄读：用 s.Get 会走另一条池化连接，读到的 invested 可能已被别的事务改掉，
		// 之后按绝对值写回就会丢掉一次并发累加。
		plan, err := tx.GetByID(userID, planID)
		if err != nil {
			return err
		}
		if plan == nil {
			return ErrGridPlanNotFound
		}
		if amount <= 0 {
			return &ValidationError{Field: "金额", Reason: "必须大于 0"}
		}

		var asset *domain.GridAsset
		if assetID > 0 {
			for i := range plan.Assets {
				if plan.Assets[i].ID == assetID {
					asset = &plan.Assets[i]
					break
				}
			}
			if asset == nil {
				return &ValidationError{Field: "标的", Reason: "不属于该计划"}
			}
		}

		exec = &domain.GridPlanExecution{
			PlanID:     planID,
			AssetName:  "未分配",
			Amount:     round2(amount),
			LayerKeys:  filterKeys(layerKeys),
			WasRegular: wasRegular,
			Note:       strings.TrimSpace(note),
		}
		if asset != nil {
			exec.AssetName = asset.Name
			exec.AssetID = &asset.ID

			// 层级 key 必须真实存在：凭空写进去的 key 会让下一次保存被 validateAsset 拒绝，
			// 整个计划从此保存不了。作废层也不允许被标记已投。
			if err := checkLayerKeys(asset, exec.LayerKeys); err != nil {
				return err
			}

			// 锁住这一行再算增量：并发记录执行时，读到的 invested 必须是当前值。
			lockedInvested, lockedLayers, err := tx.LockAssetsInvested(planID, asset.ID)
			if err != nil {
				return err
			}
			asset.Invested = round2(lockedInvested + amount)
			asset.ExecutedLayers = mergeKeys(lockedLayers, exec.LayerKeys)
			if err := tx.UpdateAssetsInvested(planID, asset.ID, asset.Invested, asset.ExecutedLayers); err != nil {
				return err
			}
		} else {
			if err := tx.UpdatePlanUnallocated(planID, round2(plan.UnallocatedInvested+amount)); err != nil {
				return err
			}
		}

		if wasRegular {
			exec.PrevNextDate = plan.NextDate
			// NextDate 也可能是空（用户没设过定投日），此时从今天起算，别把日期推进整个跳过。
			if err := tx.UpdatePlanNextDate(planID, advanceByFrequency(plan.NextDate, plan.Frequency)); err != nil {
				return err
			}
		}

		if err := tx.BumpPlanVersion(planID); err != nil {
			return err
		}
		return tx.AddExecution(exec)
	})
	if err != nil {
		return nil, err
	}
	return exec, nil
}

// UndoLastExecution 撤销最近一条执行：回退已投入、解除网格层标记、恢复定投日，然后删记录。
func (s *GridPlanService) UndoLastExecution(userID, planID int64) (*domain.GridPlanExecution, error) {
	var undone *domain.GridPlanExecution
	err := s.repo.WithTx(func(tx *repository.GridPlanRepository) error {
		plan, err := tx.GetByID(userID, planID)
		if err != nil {
			return err
		}
		if plan == nil {
			return ErrGridPlanNotFound
		}
		exec, err := tx.LastExecution(planID)
		if err != nil {
			return err
		}
		if exec == nil {
			return ErrGridPlanNoExecution
		}

		if exec.AssetID != nil {
			var asset *domain.GridAsset
			for i := range plan.Assets {
				if plan.Assets[i].ID == *exec.AssetID {
					asset = &plan.Assets[i]
					break
				}
			}
			if asset != nil {
				// 同样先锁行再减，避免和并发的记录执行互相覆盖
				lockedInvested, lockedLayers, err := tx.LockAssetsInvested(planID, asset.ID)
				if err != nil {
					return err
				}
				invested := math.Max(0, round2(lockedInvested-exec.Amount))
				remaining := removeKeys(lockedLayers, exec.LayerKeys)
				if err := tx.UpdateAssetsInvested(planID, asset.ID, invested, remaining); err != nil {
					return err
				}
			}
			// 标的已被删除（asset_id 置空）时，已投入无处可退，只删记录。
		} else if exec.AssetName == "未分配" {
			unallocated := math.Max(0, round2(plan.UnallocatedInvested-exec.Amount))
			if err := tx.UpdatePlanUnallocated(planID, unallocated); err != nil {
				return err
			}
		}

		// 定投日只在 wasRegular 的执行里被推进过，所以只对这类记录回滚。
		// PrevNextDate 为空表示原本就没设定投日，要明确写回 NULL，否则会留下一个凭空多出的日期。
		if exec.WasRegular {
			if err := tx.UpdatePlanNextDate(planID, exec.PrevNextDate); err != nil {
				return err
			}
		}

		if err := tx.DeleteExecution(planID, exec.ID); err != nil {
			return err
		}
		if err := tx.BumpPlanVersion(planID); err != nil {
			return err
		}
		undone = exec
		return nil
	})
	if err != nil {
		return nil, err
	}
	return undone, nil
}

// ---------- 校验 ----------

var (
	allowedPlanVisibility = map[string]struct{}{
		domain.PlanVisibilityPrivate:   {},
		domain.PlanVisibilityLogin:     {},
		domain.PlanVisibilityFriends:   {},
		domain.PlanVisibilitySpecified: {},
		domain.PlanVisibilityPublic:    {},
	}
)

// validateVisibility 只放行已知档位；specified 必须至少选一个好友，且确实都是好友。
// visibility 为空表示客户端没带这个字段：沿用 current（库里的现有值），
// 这样老客户端保存一次不会把共享悄悄撤回。
func (s *GridPlanService) validateVisibility(ownerID int64, plan *domain.GridPlan, current string) error {
	if plan.Visibility == "" {
		plan.Visibility = current
	}
	if plan.Visibility == "" {
		plan.Visibility = domain.PlanVisibilityPrivate
	}
	if _, ok := allowedPlanVisibility[plan.Visibility]; !ok {
		return &ValidationError{Field: "可见性", Reason: "只能是 private / login / friends / specified / public"}
	}
	if plan.Visibility != domain.PlanVisibilitySpecified {
		// 非 specified 档清掉明细，避免切换档位后留下旧授权。
		// 但客户端完全没带 visible_user_ids 时（老客户端）不清，保持库里的授权不变。
		if plan.VisibleUserIDs != nil {
			plan.VisibleUserIDs = []int64{}
		}
		return nil
	}
	if len(plan.VisibleUserIDs) == 0 {
		return &ValidationError{Field: "可见性", Reason: "选择「指定好友」时至少要选一个人"}
	}
	if s.friends == nil {
		return &ValidationError{Field: "可见性", Reason: "指定可见暂不可用"}
	}
	friendIDs, err := s.friends.ListAcceptedFriendIDs(ownerID)
	if err != nil {
		return err
	}
	friendSet := make(map[int64]struct{}, len(friendIDs))
	for _, id := range friendIDs {
		friendSet[id] = struct{}{}
	}
	seen := make(map[int64]struct{}, len(plan.VisibleUserIDs))
	kept := make([]int64, 0, len(plan.VisibleUserIDs))
	for _, id := range plan.VisibleUserIDs {
		if id == ownerID {
			return &ValidationError{Field: "可见性", Reason: "计划本来就对你自己可见，不用选自己"}
		}
		if _, ok := friendSet[id]; !ok {
			return &ValidationError{Field: "可见性", Reason: "只能选择已经互为好友的用户"}
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		kept = append(kept, id)
	}
	plan.VisibleUserIDs = kept
	return nil
}

func validatePlan(plan *domain.GridPlan) error {
	if strings.TrimSpace(plan.Name) == "" {
		return &ValidationError{Field: "计划名称", Reason: "不能为空"}
	}
	plan.Name = strings.TrimSpace(plan.Name)
	if len([]rune(plan.Name)) > 80 {
		return &ValidationError{Field: "计划名称", Reason: "不能超过 80 个字符"}
	}

	switch plan.Market {
	case "cn", "us":
	default:
		return &ValidationError{Field: "市场", Reason: "只能是 cn / us"}
	}
	switch plan.Frequency {
	case "weekly", "biweekly", "monthly":
	default:
		return &ValidationError{Field: "定投频率", Reason: "只能是 weekly / biweekly / monthly"}
	}
	switch plan.GateMode {
	case "alert", "grid-only", "block-all":
	default:
		return &ValidationError{Field: "闸门模式", Reason: "只能是 alert / grid-only / block-all"}
	}
	switch plan.WeightMethod {
	case "target-weight", "invested-weight":
	default:
		return &ValidationError{Field: "回撤加权口径", Reason: "只能是 target-weight / invested-weight"}
	}
	switch plan.GridAggregation {
	case "highest", "cumulative":
	default:
		return &ValidationError{Field: "网格聚合方式", Reason: "只能是 highest / cumulative"}
	}

	if plan.Unit <= 0 {
		return &ValidationError{Field: "最小交易单位", Reason: "必须大于 0"}
	}
	if plan.TotalBudget < 0 || plan.UnallocatedInvested < 0 {
		return &ValidationError{Field: "金额", Reason: "不能为负数"}
	}
	if plan.GateThreshold < 0 || plan.GateThreshold > 100 {
		return &ValidationError{Field: "闸门阈值", Reason: "必须在 0~100 之间"}
	}

	weightSum := 0.0
	for i := range plan.Assets {
		asset := &plan.Assets[i]
		if err := validateAsset(asset, i); err != nil {
			return err
		}
		if asset.Enabled {
			weightSum += asset.Weight
		}
	}
	if weightSum > 100+gridEpsilon {
		return &ValidationError{Field: "目标权重", Reason: fmt.Sprintf("合计 %.2f%%，超过 100%%", weightSum)}
	}
	return nil
}

func validateAsset(asset *domain.GridAsset, index int) error {
	wrap := func(field, reason string) error {
		return &ValidationError{AssetIndex: index, AssetName: asset.Name, Field: field, Reason: reason}
	}
	if strings.TrimSpace(asset.Name) == "" {
		return wrap("名称", "不能为空")
	}
	asset.Name = strings.TrimSpace(asset.Name)
	if len([]rune(asset.Name)) > 80 {
		return wrap("名称", "不能超过 80 个字符")
	}
	if asset.Weight < 0 || asset.Invested < 0 || asset.MarketValue < 0 || asset.BaseAmount < 0 {
		return wrap("金额", "不能为负数")
	}
	if asset.CurrentPoint != nil && *asset.CurrentPoint < 0 {
		return wrap("当前点位", "不能为负数")
	}
	if asset.HighPoint != nil && *asset.HighPoint < 0 {
		return wrap("参考高点", "不能为负数")
	}
	if len(asset.Grid) > maxGridLayers {
		return wrap("网格", fmt.Sprintf("最多 %d 层", maxGridLayers))
	}

	ratioSum := 0.0
	prevThreshold := -1.0
	seenKeys := make(map[string]struct{}, len(asset.Grid))
	for j := range asset.Grid {
		layer := &asset.Grid[j]
		layer.Key = strings.TrimSpace(layer.Key)
		if layer.Key == "" {
			return wrap(fmt.Sprintf("第 %d 层", j+1), "缺少层标识")
		}
		if _, dup := seenKeys[layer.Key]; dup {
			return wrap(fmt.Sprintf("第 %d 层", j+1), "层标识重复")
		}
		seenKeys[layer.Key] = struct{}{}

		if layer.Threshold <= 0 || layer.Threshold > 100 {
			return wrap(fmt.Sprintf("第 %d 层回撤阈值", j+1), "必须在 0~100 之间")
		}
		if layer.Threshold <= prevThreshold+gridEpsilon {
			return wrap(fmt.Sprintf("第 %d 层回撤阈值", j+1), "必须严格大于上一层")
		}
		prevThreshold = layer.Threshold

		if layer.Allocation < 0 || layer.Allocation > 100 {
			return wrap(fmt.Sprintf("第 %d 层资金比例", j+1), "必须在 0~100 之间")
		}
		// 作废层不参与配额，但仍保留 allocation，撤销作废时能回到原份额。
		if !layer.Voided {
			ratioSum += layer.Allocation
		}
	}
	if ratioSum > 100+gridEpsilon {
		return wrap("网格资金比例", fmt.Sprintf("合计 %.2f%%，超过 100%%", ratioSum))
	}

	// 已执行的层必须真实存在，避免脏数据让撤销逻辑对不上。
	layerKeys := make(map[string]struct{}, len(asset.Grid))
	for _, layer := range asset.Grid {
		layerKeys[layer.Key] = struct{}{}
	}
	for _, key := range asset.ExecutedLayers {
		if _, ok := layerKeys[key]; !ok {
			return wrap("执行状态", fmt.Sprintf("引用了不存在的网格层 %q", key))
		}
	}
	return nil
}

// ---------- 默认值与日期推进 ----------

func applyPlanDefaults(plan *domain.GridPlan) {
	if plan.Market == "" {
		plan.Market = "cn"
	}
	if plan.Currency == "" {
		plan.Currency = "¥"
	}
	if plan.CurrencyName == "" {
		plan.CurrencyName = "人民币"
	}
	if plan.Unit <= 0 {
		plan.Unit = 100
	}
	if plan.Frequency == "" {
		plan.Frequency = "monthly"
	}
	if plan.GateMode == "" {
		plan.GateMode = "grid-only"
	}
	if plan.WeightMethod == "" {
		plan.WeightMethod = "target-weight"
	}
	if plan.GridAggregation == "" {
		plan.GridAggregation = "highest"
	}
	// GateThreshold 不在这里兜底：0 是合法值（表示不设闸门），
	// 缺省值由请求转换层按「字段是否出现」区分。
}

func applyAssetDefaults(asset *domain.GridAsset) {
	if asset.Grid == nil {
		asset.Grid = []domain.GridLayer{}
	}
	if asset.ExecutedLayers == nil {
		asset.ExecutedLayers = []string{}
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// checkLayerKeys 确认执行记录引用的层都在该标的的网格里，且没有被作废。
func checkLayerKeys(asset *domain.GridAsset, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	byKey := make(map[string]domain.GridLayer, len(asset.Grid))
	for _, layer := range asset.Grid {
		byKey[layer.Key] = layer
	}
	for _, key := range keys {
		layer, ok := byKey[key]
		if !ok {
			return &ValidationError{AssetName: asset.Name, Field: "网格层", Reason: fmt.Sprintf("不存在 %q", key)}
		}
		if layer.Voided {
			return &ValidationError{AssetName: asset.Name, Field: "网格层", Reason: fmt.Sprintf("%q 已作废，不能标记已投", key)}
		}
	}
	return nil
}

func filterKeys(keys []string) []string {
	out := make([]string, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

func mergeKeys(base, extra []string) []string {
	seen := make(map[string]struct{}, len(base)+len(extra))
	out := make([]string, 0, len(base)+len(extra))
	for _, key := range append(append([]string{}, base...), extra...) {
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

func removeKeys(base, remove []string) []string {
	drop := make(map[string]struct{}, len(remove))
	for _, key := range remove {
		drop[key] = struct{}{}
	}
	out := make([]string, 0, len(base))
	for _, key := range base {
		if _, ok := drop[key]; ok {
			continue
		}
		out = append(out, key)
	}
	return out
}
