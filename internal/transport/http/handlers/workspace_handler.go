package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ory "github.com/ory/client-go"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/resoul/studio.go.api/internal/transport/http/utils"
)

type WorkspaceHandler struct {
	service domain.WorkspaceService
}

func NewWorkspaceHandler(service domain.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{service: service}
}

func (h *WorkspaceHandler) Create(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	name := c.PostForm("name")
	if name == "" {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Name is required")
		return
	}

	input := domain.CreateWorkspaceInput{
		Name:        name,
		Description: c.PostForm("description"),
		OwnerID:     oryIdentity.Id,
	}

	if file, header, err := c.Request.FormFile("logo"); err == nil {
		defer file.Close()
		input.Logo = file
		input.LogoSize = header.Size
		input.LogoType = header.Header.Get("Content-Type")
	}

	ws, err := h.service.CreateWorkspace(c.Request.Context(), input)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.RespondOK(c, ws)
}

func (h *WorkspaceHandler) List(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	workspaces, err := h.service.ListForUser(c.Request.Context(), oryIdentity.Id)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.RespondOK(c, workspaces)
}

type InvitePreviewResponse struct {
	ID           uuid.UUID `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	LogoURL      string    `json:"logo_url"`
	MembersCount int64     `json:"members_count"`
}

func (h *WorkspaceHandler) GetInvitePreview(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Token is required")
		return
	}

	ws, membersCount, err := h.service.PreviewInvite(c.Request.Context(), token)
	if err != nil {
		utils.RespondError(c, http.StatusNotFound, "NOT_FOUND", "Invite not found or expired")
		return
	}

	utils.RespondOK(c, InvitePreviewResponse{
		ID:           ws.ID,
		Slug:         ws.Slug,
		Name:         ws.Name,
		LogoURL:      ws.LogoURL,
		MembersCount: membersCount,
	})
}

func (h *WorkspaceHandler) AcceptInvite(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	token := c.Param("token")
	if token == "" {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Token is required")
		return
	}

	if err := h.service.AcceptInvite(c.Request.Context(), token, oryIdentity.Id); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// createInviteRequest is the typed JSON body for POST /workspaces/:id/invites.
type createInviteRequest struct {
	Email     string               `json:"email"      binding:"required,email"`
	Role      domain.WorkspaceRole `json:"role"       binding:"required"`
	SendEmail bool                 `json:"send_email"`
}

func (h *WorkspaceHandler) CreateInvite(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	var req createInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	// Build the base URL for the invite link from the incoming request so it
	// works correctly across environments (local, staging, production) without
	// needing an extra env var.
	inviteBaseURL := fmt.Sprintf("%s://%s", scheme(c), c.Request.Host)

	input := domain.CreateInviteInput{
		WorkspaceID:   wsID,
		Email:         req.Email,
		Role:          req.Role,
		SendEmail:     req.SendEmail,
		InviteBaseURL: inviteBaseURL,
	}

	invite, err := h.service.InviteUser(c.Request.Context(), input)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.RespondOK(c, invite)
}

func (h *WorkspaceHandler) GetCurrent(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	ws, err := h.service.GetCurrentWorkspace(c.Request.Context(), oryIdentity.Id)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.RespondOK(c, ws)
}

func (h *WorkspaceHandler) SetCurrent(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	if err = h.service.SetCurrentWorkspace(c.Request.Context(), oryIdentity.Id, wsID); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *WorkspaceHandler) Update(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	input := domain.UpdateWorkspaceInput{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
	}

	if file, header, err := c.Request.FormFile("logo"); err == nil {
		defer file.Close()
		input.Logo = file
		input.LogoSize = header.Size
		input.LogoType = header.Header.Get("Content-Type")
	}

	ws, err := h.service.UpdateWorkspace(c.Request.Context(), wsID, input)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.RespondOK(c, ws)
}

func (h *WorkspaceHandler) GetConfig(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	config, err := h.service.GetCurrentConfig(c.Request.Context(), oryIdentity.Id)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.RespondOK(c, config)
}

func (h *WorkspaceHandler) UpdateConfig(c *gin.Context) {
	identity, exists := c.Get("user")
	if !exists {
		utils.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found in context")
		return
	}
	oryIdentity := identity.(*ory.Identity)

	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid workspace id")
		return
	}

	var req struct {
		Language string `json:"language"`
		Theme    string `json:"theme"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err = h.service.UpdateConfig(c.Request.Context(), oryIdentity.Id, wsID, req.Language, req.Theme); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// scheme returns "https" when the request came over TLS or via a proxy header,
// falling back to "http".
func scheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	return "http"
}
