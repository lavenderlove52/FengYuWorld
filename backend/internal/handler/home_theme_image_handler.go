package handler

import (
	"errors"
	"net/http"
	"strconv"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type HomeThemeImageHandler struct {
	images *service.HomeThemeImageService
}

func NewHomeThemeImageHandler(images *service.HomeThemeImageService) *HomeThemeImageHandler {
	return &HomeThemeImageHandler{images: images}
}

func (h *HomeThemeImageHandler) List(c *gin.Context) {
	images, err := h.images.List(middleware.MustUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(images))
}

func (h *HomeThemeImageHandler) Upload(c *gin.Context) {
	userID := middleware.MustUserID(c)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid file"))
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "open upload"))
		return
	}
	defer f.Close()

	image, err := h.images.Upload(userID, file.Filename, file.Size, f)
	switch {
	case errors.Is(err, service.ErrInvalidUpload):
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid image"))
	case errors.Is(err, service.ErrHomeThemeImageLimit):
		c.JSON(http.StatusConflict, dto.Err(dto.CodeConflict, "home theme image limit reached; delete an image before uploading"))
	case err != nil:
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, err.Error()))
	default:
		c.JSON(http.StatusCreated, dto.OK(image))
	}
}

func (h *HomeThemeImageHandler) Delete(c *gin.Context) {
	imageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || imageID <= 0 {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid image id"))
		return
	}
	err = h.images.Delete(middleware.MustUserID(c), imageID)
	switch {
	case errors.Is(err, service.ErrHomeThemeImageNotFound):
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "home theme image not found"))
	case err != nil:
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, err.Error()))
	default:
		c.JSON(http.StatusOK, dto.OK(nil))
	}
}
