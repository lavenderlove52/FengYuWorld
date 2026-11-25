package service

import (
	"errors"
	"strings"
	"testing"

	"my-blog/backend/internal/domain"
)

// 服务层的校验全是纯函数，不碰数据库，所以这里的单测不需要连 PG。
// 编排逻辑（执行/撤销/快照）依赖仓储，放在 repository 包的集成测试里覆盖。

func basePlan() *domain.GridPlan {
	return &domain.GridPlan{
		Name:            "测试计划",
		Market:          "cn",
		Currency:        "¥",
		CurrencyName:    "人民币",
		Unit:            100,
		TotalBudget:     100000,
		Frequency:       "monthly",
		GateThreshold:   18,
		GateMode:        "grid-only",
		WeightMethod:    "target-weight",
		GridAggregation: "highest",
		Assets: []domain.GridAsset{
			{
				Name: "上证综指", Weight: 50, Enabled: true,
				Grid: []domain.GridLayer{
					{Key: "g1", Threshold: 8, Allocation: 20},
					{Key: "g2", Threshold: 15, Allocation: 30},
				},
			},
			{
				Name: "创业板指", Weight: 50, Enabled: true,
				Grid: []domain.GridLayer{
					{Key: "g1", Threshold: 12, Allocation: 25},
				},
			},
		},
	}
}

func expectValidation(t *testing.T, err error, contains string) {
	t.Helper()
	if err == nil {
		t.Fatalf("期望校验失败（含 %q），实际通过了", contains)
	}
	if !errors.Is(err, ErrGridPlanInvalid) {
		t.Fatalf("错误应可 unwrap 到 ErrGridPlanInvalid，实际 %v", err)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("错误信息应含 %q，实际 %q", contains, err.Error())
	}
}

func TestValidatePlan_AcceptsNormal(t *testing.T) {
	if err := validatePlan(basePlan()); err != nil {
		t.Fatalf("正常计划不该被拒: %v", err)
	}
}

func TestValidatePlan_PlanLevelFields(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*domain.GridPlan)
		contain string
	}{
		{"空名称", func(p *domain.GridPlan) { p.Name = "  " }, "计划名称"},
		{"名称超长", func(p *domain.GridPlan) { p.Name = strings.Repeat("名", 81) }, "80"},
		{"频率非法", func(p *domain.GridPlan) { p.Frequency = "daily" }, "定投频率"},
		{"闸门模式非法", func(p *domain.GridPlan) { p.GateMode = "stop" }, "闸门模式"},
		{"加权口径非法", func(p *domain.GridPlan) { p.WeightMethod = "equal" }, "回撤加权口径"},
		{"聚合方式非法", func(p *domain.GridPlan) { p.GridAggregation = "all" }, "网格聚合方式"},
		{"单位为零", func(p *domain.GridPlan) { p.Unit = 0 }, "最小交易单位"},
		{"预算为负", func(p *domain.GridPlan) { p.TotalBudget = -1 }, "不能为负数"},
		{"闸门阈值超 100", func(p *domain.GridPlan) { p.GateThreshold = 101 }, "闸门阈值"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := basePlan()
			tc.mutate(p)
			expectValidation(t, validatePlan(p), tc.contain)
		})
	}
}

func TestValidatePlan_WeightSum(t *testing.T) {
	t.Run("合计正好 100 通过", func(t *testing.T) {
		p := basePlan()
		p.Assets[0].Weight = 60
		p.Assets[1].Weight = 40
		if err := validatePlan(p); err != nil {
			t.Fatalf("60+40 应通过: %v", err)
		}
	})

	t.Run("合计不足 100 通过", func(t *testing.T) {
		p := basePlan()
		p.Assets[0].Weight = 20
		p.Assets[1].Weight = 30
		if err := validatePlan(p); err != nil {
			t.Fatalf("50 应通过: %v", err)
		}
	})

	t.Run("合计超过 100 失败", func(t *testing.T) {
		p := basePlan()
		p.Assets[0].Weight = 60
		p.Assets[1].Weight = 50
		expectValidation(t, validatePlan(p), "超过 100%")
	})

	t.Run("停用的标的权重不计入", func(t *testing.T) {
		p := basePlan()
		p.Assets[0].Weight = 100
		p.Assets[1].Weight = 100
		p.Assets[1].Enabled = false
		if err := validatePlan(p); err != nil {
			t.Fatalf("停用标的不该计入权重和: %v", err)
		}
	})

	t.Run("浮点容差", func(t *testing.T) {
		p := basePlan()
		p.Assets[0].Weight = 33.3333333
		p.Assets[1].Weight = 66.6666667
		if err := validatePlan(p); err != nil {
			t.Fatalf("接近 100 的浮点值应通过: %v", err)
		}
	})
}

func TestValidateAsset_Fields(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*domain.GridAsset)
		contain string
	}{
		{"空名称", func(a *domain.GridAsset) { a.Name = " " }, "名称"},
		{"权重为负", func(a *domain.GridAsset) { a.Weight = -1 }, "不能为负数"},
		{"已投入为负", func(a *domain.GridAsset) { a.Invested = -1 }, "不能为负数"},
		{"市值为负", func(a *domain.GridAsset) { a.MarketValue = -1 }, "不能为负数"},
		{"点位为负", func(a *domain.GridAsset) { v := -1.0; a.CurrentPoint = &v }, "当前点位"},
		{"高点为负", func(a *domain.GridAsset) { v := -1.0; a.HighPoint = &v }, "参考高点"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := basePlan()
			tc.mutate(&p.Assets[0])
			expectValidation(t, validatePlan(p), tc.contain)
		})
	}
}

func TestValidateAsset_GridLayers(t *testing.T) {
	cases := []struct {
		name    string
		grid    []domain.GridLayer
		contain string
	}{
		{
			"阈值必须递增",
			[]domain.GridLayer{{Key: "g1", Threshold: 20, Allocation: 10}, {Key: "g2", Threshold: 10, Allocation: 20}},
			"必须严格大于上一层",
		},
		{
			"阈值相等也不合法",
			[]domain.GridLayer{{Key: "g1", Threshold: 10, Allocation: 10}, {Key: "g2", Threshold: 10, Allocation: 20}},
			"必须严格大于上一层",
		},
		{
			"阈值为 0",
			[]domain.GridLayer{{Key: "g1", Threshold: 0, Allocation: 10}},
			"必须在 0~100 之间",
		},
		{
			"阈值超 100",
			[]domain.GridLayer{{Key: "g1", Threshold: 101, Allocation: 10}},
			"必须在 0~100 之间",
		},
		{
			"比例合计超 100",
			[]domain.GridLayer{{Key: "g1", Threshold: 10, Allocation: 60}, {Key: "g2", Threshold: 20, Allocation: 50}},
			"超过 100%",
		},
		{
			"缺少层标识",
			[]domain.GridLayer{{Key: "", Threshold: 10, Allocation: 10}},
			"缺少层标识",
		},
		{
			"层标识重复",
			[]domain.GridLayer{{Key: "g1", Threshold: 10, Allocation: 10}, {Key: "g1", Threshold: 20, Allocation: 10}},
			"层标识重复",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := basePlan()
			p.Assets[0].Grid = tc.grid
			p.Assets[0].ExecutedLayers = nil
			expectValidation(t, validatePlan(p), tc.contain)
		})
	}
}

func TestValidateAsset_GridRatioOK(t *testing.T) {
	p := basePlan()
	p.Assets[0].Grid = []domain.GridLayer{
		{Key: "g1", Threshold: 8, Allocation: 30},
		{Key: "g2", Threshold: 15, Allocation: 30},
		{Key: "g3", Threshold: 25, Allocation: 40},
	}
	if err := validatePlan(p); err != nil {
		t.Fatalf("合计 100 应通过: %v", err)
	}
}

func TestValidateAsset_ExecutedLayersMustExist(t *testing.T) {
	p := basePlan()
	p.Assets[0].ExecutedLayers = []string{"g1", "g9"}
	err := validatePlan(p)
	expectValidation(t, err, "不存在的网格层")
}

func TestCheckLayerKeys(t *testing.T) {
	asset := &domain.GridAsset{
		Name: "上证综指",
		Grid: []domain.GridLayer{
			{Key: "g1", Threshold: 5, Allocation: 50},
			{Key: "g2", Threshold: 10, Allocation: 50, Voided: true},
		},
	}
	if err := checkLayerKeys(asset, []string{"g1"}); err != nil {
		t.Fatalf("存在的生效层应通过: %v", err)
	}
	if err := checkLayerKeys(asset, nil); err != nil {
		t.Fatalf("空 keys 应通过: %v", err)
	}
	if err := checkLayerKeys(asset, []string{"g9"}); err == nil || !strings.Contains(err.Error(), "不存在") {
		t.Fatalf("未知层应被拒，实际 %v", err)
	}
	if err := checkLayerKeys(asset, []string{"g2"}); err == nil || !strings.Contains(err.Error(), "已作废") {
		t.Fatalf("作废层应被拒，实际 %v", err)
	}
}

// 作废层不占配额：作废层比例再大，只要生效层合计不超过 100% 就应放行。
func TestValidateAsset_VoidedLayersSkipRatioSum(t *testing.T) {
	p := basePlan()
	p.Assets[0].Grid = []domain.GridLayer{
		{Key: "g1", Threshold: 8, Allocation: 90, Voided: true},
		{Key: "g2", Threshold: 15, Allocation: 60},
		{Key: "g3", Threshold: 25, Allocation: 40},
	}
	if err := validatePlan(p); err != nil {
		t.Fatalf("作废层不应计入合计: %v", err)
	}
}

// 作废层是历史快照：允许保留旧阈值，也能和新层重叠，不参与递增校验。
// 生效层之间的顺序仍然必须严格递增。
func TestValidateAsset_VoidedLayersKeepHistoricalThresholds(t *testing.T) {
	p := basePlan()
	p.Assets[0].Grid = []domain.GridLayer{
		{Key: "g1", Threshold: 20, Allocation: 10, Voided: true},
		{Key: "g2", Threshold: 10, Allocation: 90},
	}
	if err := validatePlan(p); err != nil {
		t.Fatalf("作废层不应参与递增校验: %v", err)
	}

	// 作废层和新层阈值相同也允许（一键优化后新旧层重叠）
	p.Assets[0].Grid = []domain.GridLayer{
		{Key: "g1", Threshold: 8, Allocation: 9, Voided: true},
		{Key: "g2", Threshold: 8, Allocation: 91},
	}
	if err := validatePlan(p); err != nil {
		t.Fatalf("作废层与新层允许同阈值: %v", err)
	}

	// 生效层之间的倒序仍然拒绝
	p.Assets[0].Grid = []domain.GridLayer{
		{Key: "g1", Threshold: 20, Allocation: 10},
		{Key: "g2", Threshold: 10, Allocation: 90},
	}
	expectValidation(t, validatePlan(p), "必须严格大于上一层")
}

func TestValidateAsset_ErrorPointsToAsset(t *testing.T) {
	p := basePlan()
	p.Assets[1].Grid = []domain.GridLayer{{Key: "g1", Threshold: 0, Allocation: 10}}
	err := validatePlan(p)
	var verr *ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("应是 *ValidationError，实际 %T", err)
	}
	if verr.AssetIndex != 1 {
		t.Errorf("AssetIndex = %d, 期望 1", verr.AssetIndex)
	}
	if verr.AssetName != "创业板指" {
		t.Errorf("AssetName = %q", verr.AssetName)
	}
	if !strings.Contains(err.Error(), "创业板指") {
		t.Errorf("错误信息应含标的名称，实际 %q", err.Error())
	}
}

func TestApplyPlanDefaults(t *testing.T) {
	p := &domain.GridPlan{}
	applyPlanDefaults(p)
	if p.Market != "cn" || p.Currency != "¥" || p.CurrencyName != "人民币" {
		t.Errorf("市场默认值不对: %+v", p)
	}
	if p.Unit != 100 || p.Frequency != "monthly" {
		t.Errorf("数值默认值不对: %+v", p)
	}
	if p.GateMode != "grid-only" || p.WeightMethod != "target-weight" || p.GridAggregation != "highest" {
		t.Errorf("枚举默认值不对: %+v", p)
	}
	if p.GateThreshold != 0 {
		t.Errorf("闸门阈值不参与默认值填充（0 是合法值），实际 %v", p.GateThreshold)
	}

	t.Run("已有值不被覆盖", func(t *testing.T) {
		p := &domain.GridPlan{Market: "us", Currency: "$", Unit: 1, GateThreshold: 20}
		applyPlanDefaults(p)
		if p.Market != "us" || p.Currency != "$" || p.Unit != 1 {
			t.Errorf("覆盖了已有值: %+v", p)
		}
	})
}

func TestFilterMergeRemoveKeys(t *testing.T) {
	t.Run("filterKeys 去空去重", func(t *testing.T) {
		got := filterKeys([]string{"g1", " g2 ", "", "g1", "  "})
		want := []string{"g1", "g2"}
		if len(got) != len(want) {
			t.Fatalf("got %v", got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
	})

	t.Run("mergeKeys 保留原顺序并追加新项", func(t *testing.T) {
		got := mergeKeys([]string{"g1"}, []string{"g2", "g1", "g3"})
		want := []string{"g1", "g2", "g3"}
		if len(got) != len(want) {
			t.Fatalf("got %v", got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
	})

	t.Run("removeKeys 只删指定的", func(t *testing.T) {
		got := removeKeys([]string{"g1", "g2", "g3"}, []string{"g2"})
		if len(got) != 2 || got[0] != "g1" || got[1] != "g3" {
			t.Fatalf("got %v", got)
		}
	})
}

func TestRound2(t *testing.T) {
	cases := map[float64]float64{
		0.005:     0.01,
		1.004:     1.0,
		1234.5678: 1234.57,
		-0.004:    -0,
	}
	for in, want := range cases {
		if got := round2(in); got != want {
			t.Errorf("round2(%v) = %v, want %v", in, got, want)
		}
	}
}

func TestAdvanceByFrequency(t *testing.T) {
	t.Run("过去的日期往前推进到未来", func(t *testing.T) {
		base := domain.ParseDateOnly("2020-01-01")
		next := advanceByFrequency(base, "monthly")
		if !next.Valid {
			t.Fatal("返回了无效日期")
		}
		// 30 天为一步，最多推进 60 步，一定能越过 2020 年
		if next.String() <= "2020-01-01" {
			t.Errorf("没有推进: %s", next.String())
		}
	})

	t.Run("非法频率按 30 天处理", func(t *testing.T) {
		base := domain.ParseDateOnly("2020-01-01")
		next := advanceByFrequency(base, "unknown")
		if !next.Valid {
			t.Fatal("返回了无效日期")
		}
	})

	t.Run("零值按今天起算", func(t *testing.T) {
		next := advanceByFrequency(domain.DateOnly{}, "weekly")
		if !next.Valid {
			t.Fatal("零值输入应回退到今天")
		}
	})
}

func TestNormalizeSide(t *testing.T) {
	cases := map[string]string{
		"":       domain.ExecutionSideBuy,
		"buy":    domain.ExecutionSideBuy,
		"sell":   domain.ExecutionSideSell,
		" sell ": domain.ExecutionSideSell,
		"short":  "",
	}
	for input, want := range cases {
		if got := normalizeSide(input); got != want {
			t.Errorf("normalizeSide(%q) = %q, 期望 %q", input, got, want)
		}
	}
}

func TestNormalizePoint(t *testing.T) {
	if got := normalizePoint(nil); got != nil {
		t.Errorf("nil 应返回 nil，得到 %v", *got)
	}
	zero := 0.0
	if got := normalizePoint(&zero); got != nil {
		t.Errorf("非正数应返回 nil，得到 %v", *got)
	}
	raw := 2248.314159
	got := normalizePoint(&raw)
	if got == nil || *got != 2248.3142 {
		t.Fatalf("点位应保留 4 位小数，得到 %v", got)
	}
	if raw != 2248.314159 {
		t.Error("不应改动调用方传进来的值")
	}
}
