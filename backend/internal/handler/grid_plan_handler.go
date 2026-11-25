package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"
	"my-blog/backend/internal/util"

	"github.com/gin-gonic/gin"
)

type GridPlanHandler struct {
	plans *service.GridPlanService
	users *service.UserService
}

func NewGridPlanHandler(plans *service.GridPlanService, users *service.UserService) *GridPlanHandler {
	return &GridPlanHandler{plans: plans, users: users}
}

// List GET /api/grid-plans
func (h *GridPlanHandler) List(c *gin.Context) {
	userID := middleware.MustUserID(c)
	items, err := h.plans.List(userID)
	if err != nil {
		h.fail(c, err)
		return
	}
	out := make([]dto.GridPlanSummary, 0, len(items))
	for _, item := range items {
		out = append(out, toPlanSummaryDTO(item))
	}
	c.JSON(http.StatusOK, dto.OK(out))
}

// Get GET /api/grid-plans/:id
func (h *GridPlanHandler) Get(c *gin.Context) {
	userID := middleware.MustUserID(c)
	planID, ok := parseID(c, "id")
	if !ok {
		return
	}
	plan, err := h.plans.Get(userID, planID)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(toPlanDetailDTO(plan)))
}

// Create POST /api/grid-plans
func (h *GridPlanHandler) Create(c *gin.Context) {
	userID := middleware.MustUserID(c)
	var req dto.SaveGridPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	plan, err := h.plans.Create(userID, toPlanDomain(&req))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(toPlanDetailDTO(plan)))
}

// Update PUT /api/grid-plans/:id
func (h *GridPlanHandler) Update(c *gin.Context) {
	userID := middleware.MustUserID(c)
	planID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req dto.SaveGridPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	plan, err := h.plans.Update(userID, planID, toPlanDomain(&req))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(toPlanDetailDTO(plan)))
}

// Delete DELETE /api/grid-plans/:id
func (h *GridPlanHandler) Delete(c *gin.Context) {
	userID := middleware.MustUserID(c)
	planID, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.plans.Delete(userID, planID); err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

// ListSnapshots GET /api/grid-plans/:id/snapshots?days=180
func (h *GridPlanHandler) ListSnapshots(c *gin.Context) {
	userID := middleware.MustUserID(c)
	planID, ok := parseID(c, "id")
	if !ok {
		return
	}
	// days=0 表示全部历史；参数缺失或非法时才回退到默认 180 天
	days, err := strconv.Atoi(c.DefaultQuery("days", "180"))
	if err != nil || days < 0 {
		days = 180
	}
	items, err := h.plans.Snapshots(userID, planID, days)
	if err != nil {
		h.fail(c, err)
		return
	}
	out := make([]dto.GridSnapshotDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toSnapshotDTO(item))
	}
	c.JSON(http.StatusOK, dto.OK(out))
}

// RecordSnapshot POST /api/grid-plans/:id/snapshots
func (h *GridPlanHandler) RecordSnapshot(c *gin.Context) {
	userID := middleware.MustUserID(c)
	planID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req dto.RecordSnapshotRequest
	// 允许空 body，此时按今天算
	_ = c.ShouldBindJSON(&req)

	date := domain.DateOnly{}
	if req.Date == "" {
		now := time.Now()
		date = domain.NewDateOnly(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local))
	} else {
		date = domain.ParseDateOnly(req.Date)
		if !date.Valid {
			c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "date 格式应为 yyyy-MM-dd"))
			return
		}
	}

	snap, err := h.plans.RecordSnapshot(userID, planID, date)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(toSnapshotDTO(*snap)))
}

// AddExecution POST /api/grid-plans/:id/executions
func (h *GridPlanHandler) AddExecution(c *gin.Context) {
	userID := middleware.MustUserID(c)
	planID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req dto.RecordExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	exec, err := h.plans.RecordExecution(userID, planID, req.AssetID, req.Amount, req.LayerKeys, req.WasRegular, req.Note)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(toExecutionDTO(*exec)))
}

// UndoLastExecution DELETE /api/grid-plans/:id/executions/last
func (h *GridPlanHandler) UndoLastExecution(c *gin.Context) {
	userID := middleware.MustUserID(c)
	planID, ok := parseID(c, "id")
	if !ok {
		return
	}
	exec, err := h.plans.UndoLastExecution(userID, planID)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(toExecutionDTO(*exec)))
}

// ---------- 对外展示（个人主页） ----------

// PublicList GET /api/users/:username/grid-plans
// 只返回该用户对外可见的计划汇总，未登录访客只能看到 public。
func (h *GridPlanHandler) PublicList(c *gin.Context) {
	username := c.Param("username")
	owner, err := h.users.GetByUsername(username)
	viewerID, _ := middleware.GetUserID(c)
	if err != nil || owner == nil || !ownerVisibleTo(owner, viewerID) {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "用户不存在"))
		return
	}
	// 作者看自己的主页时，展示全部计划（含 private），和文章列表一致
	summaries, err := h.plans.ListForViewer(owner.ID, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	out := make([]dto.GridPlanSummary, 0, len(summaries))
	for _, item := range summaries {
		out = append(out, toPlanSummaryDTO(item))
	}
	c.JSON(http.StatusOK, dto.OK(out))
}

// PublicGet GET /api/users/:username/grid-plans/:id
func (h *GridPlanHandler) PublicGet(c *gin.Context) {
	username := c.Param("username")
	owner, err := h.users.GetByUsername(username)
	viewerID, _ := middleware.GetUserID(c)
	if err != nil || owner == nil || !ownerVisibleTo(owner, viewerID) {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "用户不存在"))
		return
	}
	planID, ok := parseID(c, "id")
	if !ok {
		return
	}
	plan, err := h.plans.GetForViewer(viewerID, planID)
	if err != nil {
		h.fail(c, err)
		return
	}
	if plan.UserID != owner.ID {
		// 防串号：URL 里的用户和计划作者必须一致，否则等同于泄露别人的计划
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "计划不存在"))
		return
	}
	detail := toPlanDetailDTO(plan)
	if plan.UserID != viewerIDOf(c) {
		// 指定可见的好友名单只给作者看，别把别人都授权给了谁暴露出去
		detail.VisibleUserIDs = nil
	}
	c.JSON(http.StatusOK, dto.OK(detail))
}

// viewerIDOf 当前登录用户；未登录返回 0，不会和任何真实用户 id 撞上。
func viewerIDOf(c *gin.Context) int64 {
	if id, ok := middleware.GetUserID(c); ok {
		return *id
	}
	return 0
}

// ownerVisibleTo 被锁定/停用的账号，对外一律当作不存在，和文章列表的口径一致。
// 作者本人仍然能看自己的主页。
func ownerVisibleTo(owner *domain.User, viewerID *int64) bool {
	if viewerID != nil && *viewerID == owner.ID {
		return true
	}
	return owner.Status == domain.UserStatusActive
}

// PublicSnapshots GET /api/users/:username/grid-plans/:id/snapshots?days=180
func (h *GridPlanHandler) PublicSnapshots(c *gin.Context) {
	username := c.Param("username")
	owner, err := h.users.GetByUsername(username)
	viewerID, _ := middleware.GetUserID(c)
	if err != nil || owner == nil || !ownerVisibleTo(owner, viewerID) {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "用户不存在"))
		return
	}
	planID, ok := parseID(c, "id")
	if !ok {
		return
	}
	// days=0 表示全部历史；参数缺失或非法时才回退到默认 180 天
	days, err := strconv.Atoi(c.DefaultQuery("days", "180"))
	if err != nil || days < 0 {
		days = 180
	}
	// 一次拿回详情和曲线，可见性只判定一次
	plan, items, err := h.plans.GetForViewerWithSnapshots(viewerID, planID, days)
	if err != nil {
		h.fail(c, err)
		return
	}
	if plan.UserID != owner.ID {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "计划不存在"))
		return
	}
	out := make([]dto.GridSnapshotDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toSnapshotDTO(item))
	}
	c.JSON(http.StatusOK, dto.OK(out))
}

// ---------- 错误映射 ----------

func (h *GridPlanHandler) fail(c *gin.Context, err error) {
	var verr *service.ValidationError
	switch {
	case errors.As(err, &verr):
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, verr.Error()))
	case errors.Is(err, service.ErrGridPlanNotFound):
		// 越权访问也走 404，不暴露计划是否存在
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "计划不存在"))
	case errors.Is(err, service.ErrGridPlanNoExecution):
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "没有可撤销的执行记录"))
	case errors.Is(err, service.ErrGridPlanConflict):
		c.JSON(http.StatusConflict, dto.Err(dto.CodeConflict, "这份计划已在别处被修改，请刷新后重试"))
	default:
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
	}
}

func parseID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "无效的 ID"))
		return 0, false
	}
	return id, true
}

// ---------- 转换 ----------

func toPlanSummaryDTO(item domain.PlanSummary) dto.GridPlanSummary {
	return dto.GridPlanSummary{
		ID:           item.ID,
		Name:         item.Name,
		Market:       item.Market,
		Currency:     item.Currency,
		CurrencyName: item.CurrencyName,
		Visibility:   item.Visibility,
		TotalBudget:  item.TotalBudget,
		Invested:     item.Invested,
		MarketValue:  item.MarketValue,
		Pnl:          item.Pnl,
		AssetCount:   item.AssetCount,
		SortOrder:    item.SortOrder,
		UpdatedAt:    util.FormatTime(item.UpdatedAt),
	}
}

func toPlanDetailDTO(plan *domain.GridPlan) dto.GridPlanDetail {
	assets := make([]dto.GridAssetDTO, 0, len(plan.Assets))
	for _, asset := range plan.Assets {
		grid := make([]dto.GridLayerDTO, 0, len(asset.Grid))
		for _, layer := range asset.Grid {
			grid = append(grid, dto.GridLayerDTO{Key: layer.Key, Threshold: layer.Threshold, Allocation: layer.Allocation, Voided: layer.Voided})
		}
		executed := asset.ExecutedLayers
		if executed == nil {
			executed = []string{}
		}
		enabled := asset.Enabled
		assets = append(assets, dto.GridAssetDTO{
			ID:             asset.ID,
			Name:           asset.Name,
			Code:           asset.Code,
			Leveraged:      asset.Leveraged,
			Enabled:        &enabled,
			CurrentPoint:   asset.CurrentPoint,
			HighPoint:      asset.HighPoint,
			Weight:         asset.Weight,
			Invested:       asset.Invested,
			MarketValue:    asset.MarketValue,
			BaseAmount:     asset.BaseAmount,
			Grid:           grid,
			ExecutedLayers: executed,
		})
	}
	return dto.GridPlanDetail{
		ID:                  plan.ID,
		Name:                plan.Name,
		Market:              plan.Market,
		Currency:            plan.Currency,
		CurrencyName:        plan.CurrencyName,
		Unit:                plan.Unit,
		TotalBudget:         plan.TotalBudget,
		Frequency:           plan.Frequency,
		NextDate:            plan.NextDate.String(),
		GateThreshold:       plan.GateThreshold,
		GateMode:            plan.GateMode,
		WeightMethod:        plan.WeightMethod,
		GridAggregation:     plan.GridAggregation,
		UnallocatedInvested: plan.UnallocatedInvested,
		SortOrder:           plan.SortOrder,
		Version:             plan.Version,
		Visibility:          plan.Visibility,
		VisibleUserIDs:      plan.VisibleUserIDs,
		Assets:              assets,
		CreatedAt:           util.FormatTime(plan.CreatedAt),
		UpdatedAt:           util.FormatTime(plan.UpdatedAt),
	}
}

func toPlanDomain(req *dto.SaveGridPlanRequest) *domain.GridPlan {
	assets := make([]domain.GridAsset, 0, len(req.Assets))
	for i, asset := range req.Assets {
		grid := make([]domain.GridLayer, 0, len(asset.Grid))
		for _, layer := range asset.Grid {
			grid = append(grid, domain.GridLayer{Key: layer.Key, Threshold: layer.Threshold, Allocation: layer.Allocation, Voided: layer.Voided})
		}
		executed := asset.ExecutedLayers
		if executed == nil {
			executed = []string{}
		}
		enabled := true
		if asset.Enabled != nil {
			enabled = *asset.Enabled
		}
		assets = append(assets, domain.GridAsset{
			ID:             asset.ID,
			PlanID:         0,
			Name:           asset.Name,
			Code:           asset.Code,
			Leveraged:      asset.Leveraged,
			Enabled:        enabled,
			CurrentPoint:   asset.CurrentPoint,
			HighPoint:      asset.HighPoint,
			Weight:         asset.Weight,
			Invested:       asset.Invested,
			MarketValue:    asset.MarketValue,
			BaseAmount:     asset.BaseAmount,
			Grid:           grid,
			ExecutedLayers: executed,
			SortOrder:      i,
		})
	}
	gateThreshold := 20.0
	if req.GateThreshold != nil {
		gateThreshold = *req.GateThreshold
	}
	return &domain.GridPlan{
		Name:                req.Name,
		Market:              req.Market,
		Currency:            req.Currency,
		CurrencyName:        req.CurrencyName,
		Unit:                req.Unit,
		TotalBudget:         req.TotalBudget,
		Frequency:           req.Frequency,
		NextDate:            domain.ParseDateOnly(req.NextDate),
		GateThreshold:       gateThreshold,
		GateMode:            req.GateMode,
		WeightMethod:        req.WeightMethod,
		GridAggregation:     req.GridAggregation,
		UnallocatedInvested: req.UnallocatedInvested,
		SortOrder:           req.SortOrder,
		Version:             req.Version,
		Visibility:          req.Visibility,
		VisibleUserIDs:      req.VisibleUserIDs,
		Assets:              assets,
	}
}

func toSnapshotDTO(snap domain.GridPlanSnapshot) dto.GridSnapshotDTO {
	return dto.GridSnapshotDTO{
		ID:           snap.ID,
		SnapshotDate: snap.SnapshotDate.String(),
		Invested:     snap.Invested,
		MarketValue:  snap.MarketValue,
		Pnl:          snap.Pnl,
		CreatedAt:    util.FormatTime(snap.CreatedAt),
	}
}

func toExecutionDTO(exec domain.GridPlanExecution) dto.GridExecutionDTO {
	keys := exec.LayerKeys
	if keys == nil {
		keys = []string{}
	}
	return dto.GridExecutionDTO{
		ID:           exec.ID,
		AssetID:      exec.AssetID,
		AssetName:    exec.AssetName,
		Amount:       exec.Amount,
		LayerKeys:    keys,
		WasRegular:   exec.WasRegular,
		PrevNextDate: exec.PrevNextDate.String(),
		Note:         exec.Note,
		ExecutedAt:   util.FormatTime(exec.ExecutedAt),
	}
}
