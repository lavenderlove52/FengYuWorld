package handler

import (
	"net/http"
	"strconv"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PinHandler struct {
	pins *service.PinService
}

func NewPinHandler(pins *service.PinService) *PinHandler {
	return &PinHandler{pins: pins}
}

func (h *PinHandler) GetSitePins(c *gin.Context) {
	pins, err := h.pins.GetSitePins()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(pins))
}

func (h *PinHandler) SitePinPost(c *gin.Context) {
	userID := middleware.MustUserID(c)
	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	if err := h.pins.SitePinPost(userID, postID); err != nil {
		switch err {
		case service.ErrPostNotFound:
			c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "文章不存在"))
		case service.ErrPinForbidden:
			c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "无权限执行首页置顶"))
		case service.ErrAlreadyPinned:
			c.JSON(http.StatusConflict, dto.Err(dto.CodeConflict, "该文章已置顶"))
		case service.ErrPinLimitExceeded:
			c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "首页置顶数量已达上限"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		}
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *PinHandler) SiteUnpinPost(c *gin.Context) {
	userID := middleware.MustUserID(c)
	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	if err := h.pins.SiteUnpinPost(userID, postID); err != nil {
		switch err {
		case service.ErrPinNotFound:
			c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "置顶记录不存在"))
		case service.ErrPinForbidden:
			c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "无权限取消首页置顶"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		}
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *PinHandler) GetMyPins(c *gin.Context) {
	userID := middleware.MustUserID(c)
	pins, err := h.pins.GetUserPins(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(pins))
}

func (h *PinHandler) UserPinPost(c *gin.Context) {
	userID := middleware.MustUserID(c)
	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	if err := h.pins.UserPinPost(userID, postID); err != nil {
		switch err {
		case service.ErrPostNotFound:
			c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "文章不存在"))
		case service.ErrPinNotOwner:
			c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "只能置顶自己的文章"))
		case service.ErrAlreadyPinned:
			c.JSON(http.StatusConflict, dto.Err(dto.CodeConflict, "该文章已置顶"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		}
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *PinHandler) UserUnpinPost(c *gin.Context) {
	userID := middleware.MustUserID(c)
	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	if err := h.pins.UserUnpinPost(userID, postID); err != nil {
		switch err {
		case service.ErrPinNotFound:
			c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "置顶记录不存在"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		}
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *PinHandler) GrantSitePinPermission(c *gin.Context) {
	adminID := middleware.MustUserID(c)
	targetUserID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	if err := h.pins.GrantSitePinPermission(adminID, targetUserID); err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "用户不存在"))
		case service.ErrPinForbidden:
			c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "无权限"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		}
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *PinHandler) RevokeSitePinPermission(c *gin.Context) {
	adminID := middleware.MustUserID(c)
	targetUserID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	if err := h.pins.RevokeSitePinPermission(adminID, targetUserID); err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "用户不存在"))
		case service.ErrPinForbidden:
			c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "无权限"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		}
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}
