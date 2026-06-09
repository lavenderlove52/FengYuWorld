package handler

import (
	"net/http"
	"strconv"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PostTreeHandler struct {
	tree *service.PostTreeService
}

func NewPostTreeHandler(tree *service.PostTreeService) *PostTreeHandler {
	return &PostTreeHandler{tree: tree}
}

func (h *PostTreeHandler) List(c *gin.Context) {
	userID := middleware.MustUserID(c)
	tree, err := h.tree.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(tree))
}

func (h *PostTreeHandler) CreateFolder(c *gin.Context) {
	userID := middleware.MustUserID(c)
	var req dto.CreatePostTreeFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	node, err := h.tree.CreateFolder(userID, req)
	h.respondNode(c, node, err)
}

func (h *PostTreeHandler) AddPost(c *gin.Context) {
	userID := middleware.MustUserID(c)
	var req dto.AddPostTreePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	node, err := h.tree.AddPost(userID, req)
	h.respondNode(c, node, err)
}

func (h *PostTreeHandler) Update(c *gin.Context) {
	userID := middleware.MustUserID(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	var req dto.UpdatePostTreeNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	node, err := h.tree.Update(userID, id, req)
	h.respondNode(c, node, err)
}

func (h *PostTreeHandler) Delete(c *gin.Context) {
	userID := middleware.MustUserID(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	if err := h.tree.Delete(userID, id); err == service.ErrPostTreeNodeNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "目录节点不存在"))
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *PostTreeHandler) respondNode(c *gin.Context, node *dto.PostTreeNodeResponse, err error) {
	switch err {
	case nil:
		c.JSON(http.StatusOK, dto.OK(node))
	case service.ErrPostTreeNodeNotFound:
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "目录节点不存在"))
	case service.ErrInvalidPostTreeNode:
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "目录节点无效"))
	case service.ErrPostAlreadyInTree:
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "文章已经在目录中"))
	default:
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
	}
}
