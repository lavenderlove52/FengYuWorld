package repository_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/repository"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 仓储层测试连的是真实 PostgreSQL（本地 docker 的 myblog 库），
// 所以每个用例都在事务里跑，结束时回滚，不污染开发数据。
// 没有设置 GRIDPLAN_TEST_DSN 时整体跳过。

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("GRIDPLAN_TEST_DSN")
	if dsn == "" {
		dsn = "postgres://myblog:myblog@localhost:5432/myblog?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Skipf("skip: 连不上测试库 %v", err)
	}
	return db
}

// inTx 在事务里执行 fn，结束后无条件回滚。
func inTx(t *testing.T, db *gorm.DB, fn func(tx *gorm.DB)) {
	t.Helper()
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin tx: %v", tx.Error)
	}
	defer tx.Rollback()
	fn(tx)
}

// seedUser 建两个用户，返回 id。用户名带时间戳避免和真实数据冲突。
func seedUser(t *testing.T, tx *gorm.DB, name string) int64 {
	t.Helper()
	suffix := time.Now().UnixNano()
	u := &domain.User{
		Username:     name + "-" + itoa(suffix),
		Email:        name + "-" + itoa(suffix) + "@test.local",
		PasswordHash: "x",
		Nickname:     name,
	}
	if err := tx.Create(u).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u.ID
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func newPlan(userID int64, name string) *domain.GridPlan {
	return &domain.GridPlan{
		UserID:          userID,
		Name:            name,
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
				Name: "上证综指", Code: "000001.SH", Enabled: true, Weight: 50,
				Invested: 10000, MarketValue: 12000, BaseAmount: 3000,
				Grid:           []domain.GridLayer{{Key: "g1", Threshold: 8, Allocation: 20}},
				ExecutedLayers: []string{},
			},
			{
				Name: "创业板指", Code: "399006.SZ", Enabled: true, Weight: 50,
				Invested: 8000, MarketValue: 7000, BaseAmount: 2000,
				Grid:           []domain.GridLayer{{Key: "g1", Threshold: 12, Allocation: 25}},
				ExecutedLayers: []string{},
			},
		},
	}
}

func TestGridPlan_UserIsolation(t *testing.T) {
	db := testDB(t)
	inTx(t, db, func(tx *gorm.DB) {
		repo := repository.NewGridPlanRepository(tx)
		alice := seedUser(t, tx, "alice")
		bob := seedUser(t, tx, "bob")

		plan := newPlan(alice, "Alice 的计划")
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create: %v", err)
		}

		got, err := repo.GetByID(alice, plan.ID)
		if err != nil {
			t.Fatalf("alice get: %v", err)
		}
		if got == nil {
			t.Fatal("alice 读不到自己的计划")
		}
		if len(got.Assets) != 2 {
			t.Fatalf("期望 2 个标的，得到 %d", len(got.Assets))
		}

		stolen, err := repo.GetByID(bob, plan.ID)
		if err != nil {
			t.Fatalf("bob get: %v", err)
		}
		if stolen != nil {
			t.Fatal("bob 读到了 alice 的计划")
		}

		// bob 的列表里不应该出现 alice 的计划
		list, err := repo.Summaries(bob, &bob)
		if err != nil {
			t.Fatalf("summaries: %v", err)
		}
		for _, s := range list {
			if s.ID == plan.ID {
				t.Fatal("bob 的列表里出现了 alice 的计划")
			}
		}
	})
}

func TestGridPlan_SummaryAggregation(t *testing.T) {
	db := testDB(t)
	inTx(t, db, func(tx *gorm.DB) {
		repo := repository.NewGridPlanRepository(tx)
		uid := seedUser(t, tx, "summary")

		plan := newPlan(uid, "汇总测试")
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create: %v", err)
		}

		list, err := repo.Summaries(uid, &uid)
		if err != nil {
			t.Fatalf("summaries: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("期望 1 个计划，得到 %d", len(list))
		}
		s := list[0]
		if s.AssetCount != 2 {
			t.Errorf("asset_count = %d, 期望 2", s.AssetCount)
		}
		if s.Invested != 18000 {
			t.Errorf("invested = %v, 期望 18000", s.Invested)
		}
		if s.MarketValue != 19000 {
			t.Errorf("market_value = %v, 期望 19000", s.MarketValue)
		}
		if s.Pnl != 1000 {
			t.Errorf("pnl = %v, 期望 1000", s.Pnl)
		}
	})
}

func TestGridPlan_EmptyPlanSummaryHasZero(t *testing.T) {
	db := testDB(t)
	inTx(t, db, func(tx *gorm.DB) {
		repo := repository.NewGridPlanRepository(tx)
		uid := seedUser(t, tx, "empty")

		plan := newPlan(uid, "空计划")
		plan.Assets = nil
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create: %v", err)
		}

		list, err := repo.Summaries(uid, &uid)
		if err != nil {
			t.Fatalf("summaries: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("LEFT JOIN 丢掉了没有标的的计划，得到 %d 条", len(list))
		}
		if list[0].Invested != 0 || list[0].MarketValue != 0 || list[0].Pnl != 0 {
			t.Errorf("空计划汇总应全为 0，得到 %+v", list[0])
		}
	})
}

func TestGridPlan_UpdateReplacesAssets(t *testing.T) {
	db := testDB(t)
	inTx(t, db, func(tx *gorm.DB) {
		repo := repository.NewGridPlanRepository(tx)
		uid := seedUser(t, tx, "update")

		plan := newPlan(uid, "更新测试")
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create: %v", err)
		}

		loaded, err := repo.GetByID(uid, plan.ID)
		if err != nil || loaded == nil {
			t.Fatalf("get: %v", err)
		}
		loaded.Name = "改过的名字"
		loaded.TotalBudget = 200000
		loaded.Assets = []domain.GridAsset{
			{Name: "只剩一个", Code: "X", Enabled: true, Weight: 100, Invested: 500, MarketValue: 600},
		}
		if err := repo.Update(loaded, 0); err != nil {
			t.Fatalf("update: %v", err)
		}

		again, err := repo.GetByID(uid, plan.ID)
		if err != nil {
			t.Fatalf("get after update: %v", err)
		}
		if again.Name != "改过的名字" || again.TotalBudget != 200000 {
			t.Errorf("计划字段没更新: %+v", again)
		}
		if len(again.Assets) != 1 || again.Assets[0].Name != "只剩一个" {
			t.Errorf("标的列表未整体替换: %+v", again.Assets)
		}
	})
}

func TestGridPlan_DeleteCascades(t *testing.T) {
	db := testDB(t)
	inTx(t, db, func(tx *gorm.DB) {
		repo := repository.NewGridPlanRepository(tx)
		uid := seedUser(t, tx, "cascade")

		plan := newPlan(uid, "级联删除")
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := repo.UpsertSnapshot(&domain.GridPlanSnapshot{
			PlanID: plan.ID, SnapshotDate: domain.ParseDateOnly("2026-11-19"),
			Invested: 18000, MarketValue: 19000, Pnl: 1000,
		}); err != nil {
			t.Fatalf("snapshot: %v", err)
		}
		if err := repo.AddExecution(&domain.GridPlanExecution{
			PlanID: plan.ID, AssetName: "上证综指", Amount: 3000,
		}); err != nil {
			t.Fatalf("execution: %v", err)
		}

		ok, err := repo.Delete(uid, plan.ID)
		if err != nil || !ok {
			t.Fatalf("delete: ok=%v err=%v", ok, err)
		}

		var assetCount, snapCount, execCount int64
		tx.Model(&domain.GridAsset{}).Where("plan_id = ?", plan.ID).Count(&assetCount)
		tx.Model(&domain.GridPlanSnapshot{}).Where("plan_id = ?", plan.ID).Count(&snapCount)
		tx.Model(&domain.GridPlanExecution{}).Where("plan_id = ?", plan.ID).Count(&execCount)
		if assetCount != 0 || snapCount != 0 || execCount != 0 {
			t.Errorf("级联未清干净: assets=%d snaps=%d execs=%d", assetCount, snapCount, execCount)
		}
	})
}

func TestGridPlan_DeleteRejectsOtherUser(t *testing.T) {
	db := testDB(t)
	inTx(t, db, func(tx *gorm.DB) {
		repo := repository.NewGridPlanRepository(tx)
		alice := seedUser(t, tx, "del-alice")
		bob := seedUser(t, tx, "del-bob")

		plan := newPlan(alice, "alice 的")
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create: %v", err)
		}

		ok, err := repo.Delete(bob, plan.ID)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}
		if ok {
			t.Fatal("bob 删掉了 alice 的计划")
		}

		still, _ := repo.GetByID(alice, plan.ID)
		if still == nil {
			t.Fatal("计划被误删")
		}
	})
}

func TestGridPlan_SnapshotSameDayOverwrites(t *testing.T) {
	db := testDB(t)
	inTx(t, db, func(tx *gorm.DB) {
		repo := repository.NewGridPlanRepository(tx)
		uid := seedUser(t, tx, "snap")

		plan := newPlan(uid, "快照测试")
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create: %v", err)
		}

		for _, v := range []struct {
			mv, pnl float64
		}{{19000, 1000}, {20000, 2000}, {21000, 3000}} {
			if err := repo.UpsertSnapshot(&domain.GridPlanSnapshot{
				PlanID: plan.ID, SnapshotDate: domain.ParseDateOnly("2026-11-19"),
				Invested: 18000, MarketValue: v.mv, Pnl: v.pnl,
			}); err != nil {
				t.Fatalf("upsert: %v", err)
			}
		}

		snaps, err := repo.ListSnapshots(plan.ID, 0)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(snaps) != 1 {
			t.Fatalf("同一天应只有 1 条，得到 %d", len(snaps))
		}
		if snaps[0].MarketValue != 21000 || snaps[0].Pnl != 3000 {
			t.Errorf("应保留最后一次写入，得到 %+v", snaps[0])
		}
	})
}

func TestGridPlan_SnapshotOrderAndWindow(t *testing.T) {
	db := testDB(t)
	inTx(t, db, func(tx *gorm.DB) {
		repo := repository.NewGridPlanRepository(tx)
		uid := seedUser(t, tx, "snapwin")

		plan := newPlan(uid, "窗口测试")
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create: %v", err)
		}
		for _, d := range []string{"2026-11-15", "2026-11-17", "2026-11-16"} {
			if err := repo.UpsertSnapshot(&domain.GridPlanSnapshot{
				PlanID: plan.ID, SnapshotDate: domain.ParseDateOnly(d), Invested: 1, MarketValue: 2, Pnl: 1,
			}); err != nil {
				t.Fatalf("upsert %s: %v", d, err)
			}
		}

		snaps, err := repo.ListSnapshots(plan.ID, 0)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(snaps) != 3 {
			t.Fatalf("期望 3 条，得到 %d", len(snaps))
		}
		for i := 1; i < len(snaps); i++ {
			if snaps[i-1].SnapshotDate.String() > snaps[i].SnapshotDate.String() {
				t.Fatalf("快照未按日期升序: %v", snaps)
			}
		}
	})
}

func TestGridPlan_ExecutionLifecycle(t *testing.T) {
	db := testDB(t)
	inTx(t, db, func(tx *gorm.DB) {
		repo := repository.NewGridPlanRepository(tx)
		uid := seedUser(t, tx, "exec")

		plan := newPlan(uid, "执行记录")
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create: %v", err)
		}
		loaded, _ := repo.GetByID(uid, plan.ID)
		assetID := loaded.Assets[0].ID

		if err := repo.AddExecution(&domain.GridPlanExecution{
			PlanID: plan.ID, AssetID: &assetID, AssetName: "上证综指",
			Amount: 3000, LayerKeys: []string{"g1"}, WasRegular: true,
			PrevNextDate: domain.ParseDateOnly("2026-12-15"),
		}); err != nil {
			t.Fatalf("add execution: %v", err)
		}

		last, err := repo.LastExecution(plan.ID)
		if err != nil || last == nil {
			t.Fatalf("last execution: %v", err)
		}
		if last.Amount != 3000 || len(last.LayerKeys) != 1 || last.LayerKeys[0] != "g1" {
			t.Errorf("执行记录字段不对: %+v", last)
		}
		if last.PrevNextDate.String() != "2026-12-15" {
			t.Errorf("prev_next_date 不对: %v", last.PrevNextDate.String())
		}

		if err := repo.DeleteExecution(plan.ID, last.ID); err != nil {
			t.Fatalf("delete execution: %v", err)
		}
		after, _ := repo.LastExecution(plan.ID)
		if after != nil {
			t.Error("删除后仍有执行记录")
		}
	})
}

func TestGridPlan_DeleteAssetKeepsExecution(t *testing.T) {
	db := testDB(t)
	inTx(t, db, func(tx *gorm.DB) {
		repo := repository.NewGridPlanRepository(tx)
		uid := seedUser(t, tx, "assetdel")

		plan := newPlan(uid, "删标的")
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create: %v", err)
		}
		loaded, _ := repo.GetByID(uid, plan.ID)
		assetID := loaded.Assets[0].ID

		if err := repo.AddExecution(&domain.GridPlanExecution{
			PlanID: plan.ID, AssetID: &assetID, AssetName: "上证综指", Amount: 3000,
		}); err != nil {
			t.Fatalf("add execution: %v", err)
		}

		// 整体替换标的列表，旧的 asset 行被删
		loaded.Assets = []domain.GridAsset{{Name: "新标的", Enabled: true, Weight: 100}}
		if err := repo.Update(loaded, 0); err != nil {
			t.Fatalf("update: %v", err)
		}

		last, err := repo.LastExecution(plan.ID)
		if err != nil || last == nil {
			t.Fatalf("执行记录应保留: %v", err)
		}
		if last.AssetID != nil {
			t.Errorf("asset_id 应被置空，得到 %v", *last.AssetID)
		}
		if last.AssetName != "上证综指" {
			t.Errorf("asset_name 冗余应保留，得到 %q", last.AssetName)
		}
	})
}

func TestGridPlan_UpdateAssetsInvested(t *testing.T) {
	db := testDB(t)
	inTx(t, db, func(tx *gorm.DB) {
		repo := repository.NewGridPlanRepository(tx)
		uid := seedUser(t, tx, "invested")

		plan := newPlan(uid, "回写")
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create: %v", err)
		}
		loaded, _ := repo.GetByID(uid, plan.ID)
		assetID := loaded.Assets[0].ID

		if err := repo.UpdateAssetsInvested(plan.ID, assetID, 33333, []string{"g1", "g2"}); err != nil {
			t.Fatalf("update invested: %v", err)
		}

		again, _ := repo.GetByID(uid, plan.ID)
		if again.Assets[0].Invested != 33333 {
			t.Errorf("invested = %v", again.Assets[0].Invested)
		}
		if len(again.Assets[0].ExecutedLayers) != 2 {
			t.Errorf("executed_layers = %v", again.Assets[0].ExecutedLayers)
		}
	})
}

func TestGridPlan_UpdateRejectsStaleVersion(t *testing.T) {
	db := testDB(t)
	inTx(t, db, func(tx *gorm.DB) {
		repo := repository.NewGridPlanRepository(tx)
		uid := seedUser(t, tx, "version")

		plan := newPlan(uid, "乐观锁")
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create: %v", err)
		}
		v1, _ := repo.GetByID(uid, plan.ID)
		if v1.Version != 1 {
			t.Fatalf("新建计划 version 应为 1，得到 %d", v1.Version)
		}

		// 两台设备同时读到 version=1，先到的保存成功
		first, _ := repo.GetByID(uid, plan.ID)
		first.Name = "先保存的"
		if err := repo.Update(first, v1.Version); err != nil {
			t.Fatalf("首次保存应成功: %v", err)
		}
		after, _ := repo.GetByID(uid, plan.ID)
		if after.Version != 2 {
			t.Errorf("保存后 version 应递增到 2，得到 %d", after.Version)
		}

		// 后到的还拿着 version=1，必须被拒绝，且不能覆盖已写入的数据
		stale, _ := repo.GetByID(uid, plan.ID)
		stale.Name = "过期的保存"
		err := repo.Update(stale, v1.Version)
		if !errors.Is(err, repository.ErrVersionConflict) {
			t.Fatalf("期望 ErrVersionConflict，得到 %v", err)
		}
		final, _ := repo.GetByID(uid, plan.ID)
		if final.Name != "先保存的" {
			t.Errorf("过期保存不应覆盖数据，得到 %q", final.Name)
		}

		// BumpPlanVersion 也要让手里的版本失效（执行/撤销走这条路）
		if err := repo.BumpPlanVersion(plan.ID); err != nil {
			t.Fatalf("bump version: %v", err)
		}
		bumped, _ := repo.GetByID(uid, plan.ID)
		if bumped.Version != 3 {
			t.Errorf("bump 后 version 应为 3，得到 %d", bumped.Version)
		}
	})
}
