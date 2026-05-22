package handler

import (
	"net/http"
	"strconv"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categories *service.CategoryService
}

func NewCategoryHandler(categories *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categories: categories}
}

func (h *CategoryHandler) List(c *gin.Context) {
	userID := middleware.MustUserID(c)
	list, err := h.categories.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(list))
}

func (h *CategoryHandler) Create(c *gin.Context) {
	userID := middleware.MustUserID(c)
	var req dto.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	item, err := h.categories.Create(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *CategoryHandler) Update(c *gin.Context) {
	userID := middleware.MustUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req dto.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	item, err := h.categories.Update(userID, id, req)
	if err == service.ErrCategoryNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "分类不存在"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	userID := middleware.MustUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.categories.Delete(userID, id); err == service.ErrCategoryNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "分类不存在"))
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

type TagHandler struct {
	tags *service.TagService
}

func NewTagHandler(tags *service.TagService) *TagHandler {
	return &TagHandler{tags: tags}
}

func (h *TagHandler) List(c *gin.Context) {
	userID := middleware.MustUserID(c)
	list, err := h.tags.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(list))
}

func (h *TagHandler) Create(c *gin.Context) {
	userID := middleware.MustUserID(c)
	var req dto.TagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	item, err := h.tags.Create(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *TagHandler) Delete(c *gin.Context) {
	userID := middleware.MustUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.tags.Delete(userID, id); err == service.ErrTagNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "标签不存在"))
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}
