package handler

import (
	"net/http"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	uploads *service.UploadService
}

func NewUploadHandler(uploads *service.UploadService) *UploadHandler {
	return &UploadHandler{uploads: uploads}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	userID := middleware.MustUserID(c)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	defer f.Close()

	resp, err := h.uploads.SaveImage(userID, file.Filename, file.Size, f, 0)
	if err == service.ErrInvalidUpload {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "文件无效"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(resp))
}

func (h *UploadHandler) UploadBanner(c *gin.Context) {
	userID := middleware.MustUserID(c)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	defer f.Close()

	resp, err := h.uploads.SaveBannerImage(userID, file.Filename, file.Size, f)
	if err == service.ErrInvalidUpload {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "文件无效"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(resp))
}

func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	userID := middleware.MustUserID(c)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	defer f.Close()

	resp, err := h.uploads.SaveAvatar(userID, file.Filename, file.Size, f)
	if err == service.ErrInvalidUpload {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "文件无效"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(resp))
}
