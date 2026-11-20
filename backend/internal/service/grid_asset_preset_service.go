package service

import (
	"strings"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/repository"
)

// GridAssetPresetService 自建标的联想项的读写。
type GridAssetPresetService struct {
	repo *repository.GridAssetPresetRepository
}

func NewGridAssetPresetService(repo *repository.GridAssetPresetRepository) *GridAssetPresetService {
	return &GridAssetPresetService{repo: repo}
}

func (s *GridAssetPresetService) List(userID int64, market string) ([]domain.GridAssetPreset, error) {
	if market != "us" {
		market = "cn"
	}
	return s.repo.ListByMarket(userID, market)
}

// Delete 返回 false 表示不存在或不属于该用户。
func (s *GridAssetPresetService) Delete(userID, id int64) (bool, error) {
	return s.repo.Delete(userID, id)
}

// CollectFromAssets 把某个计划里非预置的标的记进自建列表。
// 由 GridPlanService.Update 在保存后调用，保证「点保存才记住」。
func (s *GridAssetPresetService) CollectFromAssets(userID int64, market string, assets []domain.GridAsset) error {
	// 只有这两个市场有预置名单，其它值会让 CHECK 约束拒绝整批写入
	if market != "us" {
		market = "cn"
	}
	items := make([]domain.GridAssetPreset, 0, len(assets))
	seen := make(map[string]struct{}, len(assets))
	for _, asset := range assets {
		name := strings.TrimSpace(asset.Name)
		if name == "" || domain.IsPresetIndex(market, name) {
			continue
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		items = append(items, domain.GridAssetPreset{
			Name:      name,
			Code:      strings.TrimSpace(asset.Code),
			Leveraged: asset.Leveraged,
		})
	}
	return s.repo.UpsertMany(userID, market, items)
}
