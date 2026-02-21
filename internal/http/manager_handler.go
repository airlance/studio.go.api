package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appconfig "github.com/football.manager.api/internal/config"
	"github.com/football.manager.api/internal/domain"
	platformauth "github.com/football.manager.api/internal/platform/auth"
	"github.com/football.manager.api/internal/platform/httpx"
	"github.com/football.manager.api/internal/platform/storage"
	"github.com/football.manager.api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ManagerHandler struct {
	managerUC      usecase.ManagerUseCase
	storage        storage.PublicObjectStorage
	maxAvatarBytes int64
}

func NewManagerHandler(managerUC usecase.ManagerUseCase, storage storage.PublicObjectStorage, storageCfg appconfig.StorageConfig) *ManagerHandler {
	maxAvatarBytes := storageCfg.MaxAvatarMB * 1024 * 1024
	if maxAvatarBytes <= 0 {
		maxAvatarBytes = 5 * 1024 * 1024
	}

	return &ManagerHandler{
		managerUC:      managerUC,
		storage:        storage,
		maxAvatarBytes: maxAvatarBytes,
	}
}

func (h *ManagerHandler) CreateMe(c *gin.Context) {
	userID, ok := platformauth.GetUserIDFromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	req, err := h.parseCreateManagerRequest(c, userID)
	if err != nil {
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "too large") {
			httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		logrus.WithError(err).Error("Failed to parse manager create request")
		httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to process avatar upload")
		return
	}

	manager, err := h.managerUC.Create(c.Request.Context(), userID, toCreateManagerDTO(req))
	if err != nil {
		switch err {
		case domain.ErrManagerExists:
			httpx.RespondError(c, http.StatusConflict, "manager_exists", "Manager already exists")
		default:
			if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "status must be") {
				httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
				return
			}
			logrus.WithError(err).Error("Failed to create manager")
			httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to create manager")
		}
		return
	}

	httpx.RespondCreated(c, gin.H{
		"message": "Manager created",
		"manager": toManagerResponse(manager),
	})
}

func (h *ManagerHandler) GetMe(c *gin.Context) {
	userID, ok := platformauth.GetUserIDFromContext(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	manager, err := h.managerUC.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		if err == domain.ErrManagerNotFound {
			httpx.RespondError(c, http.StatusNotFound, "not_found", "Manager not found")
			return
		}
		logrus.WithError(err).Error("Failed to get current manager")
		httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to get manager")
		return
	}

	httpx.RespondOK(c, gin.H{
		"manager": toManagerResponse(manager),
	})
}

func (h *ManagerHandler) parseCreateManagerRequest(c *gin.Context, userID uint) (CreateManagerRequest, error) {
	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data") {
		return h.parseMultipartCreateManagerRequest(c, userID)
	}

	var req CreateManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return CreateManagerRequest{}, err
	}

	return req, nil
}

func (h *ManagerHandler) parseMultipartCreateManagerRequest(c *gin.Context, userID uint) (CreateManagerRequest, error) {
	var req CreateManagerRequest

	req.Name = strings.TrimSpace(c.PostForm("name"))
	req.Status = strings.TrimSpace(c.PostForm("status"))

	if rawCountryID := strings.TrimSpace(c.PostForm("country_id")); rawCountryID != "" {
		countryID, err := strconv.ParseUint(rawCountryID, 10, 64)
		if err != nil || countryID == 0 {
			return CreateManagerRequest{}, fmt.Errorf("invalid country_id")
		}
		parsedCountryID := uint(countryID)
		req.CountryID = &parsedCountryID
	}

	if req.Name == "" {
		return CreateManagerRequest{}, fmt.Errorf("name is required")
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		return CreateManagerRequest{}, fmt.Errorf("avatar is required")
	}
	defer file.Close()

	avatarURL, err := h.uploadAvatar(c.Request.Context(), userID, file, header)
	if err != nil {
		return CreateManagerRequest{}, err
	}
	req.Avatar = avatarURL

	return req, nil
}

func (h *ManagerHandler) uploadAvatar(ctx context.Context, userID uint, file multipart.File, header *multipart.FileHeader) (string, error) {
	if h.storage == nil {
		return "", fmt.Errorf("avatar storage is not configured")
	}

	if header.Size > h.maxAvatarBytes {
		return "", fmt.Errorf("avatar file is too large (max %d MB)", h.maxAvatarBytes/(1024*1024))
	}

	probe := make([]byte, 512)
	n, err := file.Read(probe)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read avatar image: %w", err)
	}
	contentType := http.DetectContentType(probe[:n])
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("avatar must be an image")
	}

	stream := io.MultiReader(bytes.NewReader(probe[:n]), file)

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = extensionFromContentType(contentType)
	}
	if ext == "" {
		ext = ".bin"
	}

	key := fmt.Sprintf("avatars/managers/%d/%d-%s%s", userID, time.Now().UTC().Unix(), uuid.NewString(), ext)
	url, err := h.storage.PutPublicObject(ctx, key, stream, header.Size, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload avatar: %w", err)
	}

	return url, nil
}

func extensionFromContentType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "image/svg+xml":
		return ".svg"
	default:
		return ""
	}
}
