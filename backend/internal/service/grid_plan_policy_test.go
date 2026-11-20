package service

import (
	"testing"

	"my-blog/backend/internal/domain"
)

// 可见性策略是纯判定逻辑，用替身覆盖各档位，不连数据库。

type fakePlanAccess struct {
	visible map[int64][]int64
}

func (f *fakePlanAccess) IsVisibleToUser(planID, userID int64) (bool, error) {
	for _, id := range f.visible[planID] {
		if id == userID {
			return true, nil
		}
	}
	return false, nil
}

type fakeFriends struct {
	pairs map[[2]int64]bool
}

func (f *fakeFriends) AreFriends(userA, userB int64) (bool, error) {
	return f.pairs[[2]int64{userA, userB}] || f.pairs[[2]int64{userB, userA}], nil
}

func int64p(v int64) *int64 { return &v }

func TestGridPlanPolicy_Visibility(t *testing.T) {
	owner := int64(10)
	viewer := int64(20)
	stranger := int64(30)

	policy := NewGridPlanPolicy(
		&fakePlanAccess{visible: map[int64][]int64{7: {viewer}}},
		&fakeFriends{pairs: map[[2]int64]bool{{owner, viewer}: true}},
	)

	cases := []struct {
		name       string
		visibility string
		planID     int64
		viewer     *int64
		want       bool
	}{
		{"作者看私密", domain.PlanVisibilityPrivate, 1, int64p(owner), true},
		{"访客看私密", domain.PlanVisibilityPrivate, 1, int64p(stranger), false},
		{"未登录看私密", domain.PlanVisibilityPrivate, 1, nil, false},
		{"未登录看公开", domain.PlanVisibilityPublic, 2, nil, true},
		{"访客看公开", domain.PlanVisibilityPublic, 2, int64p(stranger), true},
		{"未登录看登录可见", domain.PlanVisibilityLogin, 3, nil, false},
		{"登录用户看登录可见", domain.PlanVisibilityLogin, 3, int64p(stranger), true},
		{"未登录看好友可见", domain.PlanVisibilityFriends, 4, nil, false},
		{"好友看好友可见", domain.PlanVisibilityFriends, 4, int64p(viewer), true},
		{"非好友看好友可见", domain.PlanVisibilityFriends, 4, int64p(stranger), false},
		{"指定用户看指定可见", domain.PlanVisibilitySpecified, 7, int64p(viewer), true},
		{"其他人看指定可见", domain.PlanVisibilitySpecified, 7, int64p(stranger), false},
		{"未登录看指定可见", domain.PlanVisibilitySpecified, 7, nil, false},
		{"未知档位一律拒绝", "weird", 8, int64p(viewer), false},
	}

	for _, tc := range cases {
		plan := &domain.GridPlan{ID: tc.planID, UserID: owner, Visibility: tc.visibility}
		got, err := policy.CanReadPlan(tc.viewer, plan)
		if err != nil {
			t.Fatalf("%s: 意外错误 %v", tc.name, err)
		}
		if got != tc.want {
			t.Errorf("%s: 期望 %v，得到 %v", tc.name, tc.want, got)
		}
	}

	if _, err := policy.CanReadPlan(nil, nil); err != nil {
		t.Errorf("nil 计划不应报错: %v", err)
	}
}

func TestGridPlanVisibility_Labels(t *testing.T) {
	// 常量必须与迁移里的 CHECK 约束一致，否则保存会被数据库拒绝
	want := map[string]bool{
		"private": true, "login": true, "friends": true, "specified": true, "public": true,
	}
	for key := range want {
		if _, ok := allowedPlanVisibility[key]; !ok {
			t.Errorf("allowedPlanVisibility 缺少 %q", key)
		}
	}
	if len(allowedPlanVisibility) != len(want) {
		t.Errorf("allowedPlanVisibility 有 %d 项，期望 %d 项", len(allowedPlanVisibility), len(want))
	}
}
