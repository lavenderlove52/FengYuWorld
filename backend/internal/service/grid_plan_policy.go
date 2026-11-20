package service

import (
	"my-blog/backend/internal/domain"
)

// GridPlanPolicy 投资计划的对外可见性，口径与文章（PostPolicy）保持一致：
// 作者永远能读；其余人按 visibility 逐档判定，未登录访客只读 public。
type GridPlanPolicy struct {
	plans       gridPlanAccessRepo
	friendships friendChecker
}

// gridPlanAccessRepo 只取策略用得到的两个能力，方便测试替身。
type gridPlanAccessRepo interface {
	IsVisibleToUser(planID, userID int64) (bool, error)
}

type friendChecker interface {
	AreFriends(userA, userB int64) (bool, error)
}

func NewGridPlanPolicy(plans gridPlanAccessRepo, friendships friendChecker) *GridPlanPolicy {
	return &GridPlanPolicy{plans: plans, friendships: friendships}
}

// CanReadPlan viewerID 为 nil 表示未登录访客。
func (p *GridPlanPolicy) CanReadPlan(viewerID *int64, plan *domain.GridPlan) (bool, error) {
	if plan == nil {
		return false, nil
	}
	if viewerID != nil && *viewerID == plan.UserID {
		return true, nil
	}
	switch plan.Visibility {
	case domain.PlanVisibilityPublic:
		return true, nil
	case domain.PlanVisibilityPrivate:
		return false, nil
	case domain.PlanVisibilityLogin:
		return viewerID != nil, nil
	case domain.PlanVisibilityFriends:
		if viewerID == nil {
			return false, nil
		}
		return p.friendships.AreFriends(plan.UserID, *viewerID)
	case domain.PlanVisibilitySpecified:
		if viewerID == nil {
			return false, nil
		}
		return p.plans.IsVisibleToUser(plan.ID, *viewerID)
	default:
		return false, nil
	}
}
