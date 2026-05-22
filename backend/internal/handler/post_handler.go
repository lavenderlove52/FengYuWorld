package handler

import (
	"net/http"
	"strconv"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	posts *service.PostService
}

func NewPostHandler(posts *service.PostService) *PostHandler {
	return &PostHandler{posts: posts}
}

func (h *PostHandler) ListPublic(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))
	result, err := h.posts.ListPublic(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(result))
}

func (h *PostHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	var viewerID *int64
	if uid, ok := middleware.GetUserID(c); ok {
		viewerID = uid
	}
	post, err := h.posts.GetDetail(viewerID, id)
	if err == service.ErrPostNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "资源不存在"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(post))
}

func (h *PostHandler) Create(c *gin.Context) {
	userID := middleware.MustUserID(c)
	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	post, err := h.posts.Create(userID, req)
	if err == service.ErrInvalidPostData {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "文章数据无效"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(post))
}

func (h *PostHandler) Update(c *gin.Context) {
	userID := middleware.MustUserID(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	var req dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	post, err := h.posts.Update(userID, id, req)
	if err == service.ErrPostNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "资源不存在"))
		return
	}
	if err == service.ErrPostForbidden {
		c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "无权限"))
		return
	}
	if err == service.ErrInvalidPostData {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "文章数据无效"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(post))
}

func (h *PostHandler) Delete(c *gin.Context) {
	userID := middleware.MustUserID(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	if err := h.posts.Delete(userID, id); err == service.ErrPostNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "资源不存在"))
		return
	} else if err == service.ErrPostForbidden {
		c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "无权限"))
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *PostHandler) ListMine(c *gin.Context) {
	userID := middleware.MustUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))
	status := c.Query("status")
	visibility := c.Query("visibility")
	result, err := h.posts.ListMine(userID, status, visibility, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(result))
}
