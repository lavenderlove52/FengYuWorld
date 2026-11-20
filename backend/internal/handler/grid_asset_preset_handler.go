package handler

import (
	"net/http"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type GridAssetPresetHandler struct {
	presets *service.GridAssetPresetService
}

func NewGridAssetPresetHandler(presets *service.GridAssetPresetService) *GridAssetPresetHandler {
	return &GridAssetPresetHandler{presets: presets}
}

// List GET /api/grid-presets?market=cn
func (h *GridAssetPresetHandler) List(c *gin.Context) {
	userID := middleware.MustUserID(c)
	items, err := h.presets.List(userID, c.DefaultQuery("market", "cn"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	out := make([]dto.GridAssetPresetDTO, 0, len(items))
	for _, item := range items {
		out = append(out, dto.GridAssetPresetDTO{
			ID:        item.ID,
			Market:    item.Market,
			Name:      item.Name,
			Code:      item.Code,
			Leveraged: item.Leveraged,
		})
	}
	c.JSON(http.StatusOK, dto.OK(out))
}

// Delete DELETE /api/grid-presets/:id
func (h *GridAssetPresetHandler) Delete(c *gin.Context) {
	userID := middleware.MustUserID(c)
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	deleted, err := h.presets.Delete(userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "标的不存在"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}
