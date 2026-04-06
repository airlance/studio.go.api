package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ory "github.com/ory/client-go"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/resoul/studio.go.api/internal/transport/http/utils"
)

type ChatHandler struct {
	service domain.ChatService
}

func NewChatHandler(service domain.ChatService) *ChatHandler {
	return &ChatHandler{service: service}
}

func (h *ChatHandler) CreateChannel(c *gin.Context) {
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_ID", "Invalid workspace ID")
		return
	}

	var input struct {
		Name         string   `json:"name" binding:"required"`
		Description  string   `json:"description"`
		IsPrivate    bool     `json:"is_private"`
		Participants []string `json:"participants"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_INPUT", err.Error())
		return
	}

	user := c.MustGet("user").(*ory.Identity)
	channel, err := h.service.CreateChannel(c.Request.Context(), workspaceID, input.Name, input.Description, input.IsPrivate, user.Id, input.Participants)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}

	utils.RespondCreated(c, channel)
}

func (h *ChatHandler) ListChannels(c *gin.Context) {
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_ID", "Invalid workspace ID")
		return
	}

	user := c.MustGet("user").(*ory.Identity)
	channels, err := h.service.ListChannels(c.Request.Context(), workspaceID, user.Id)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}

	utils.RespondOK(c, channels)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("chat_id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_ID", "Invalid ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// We'll use a single method for both channels and DMs since the service handles it
	messages, err := h.service.GetChannelMessages(c.Request.Context(), targetID, limit, offset)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}

	utils.RespondOK(c, messages)
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("chat_id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_ID", "Invalid ID")
		return
	}

	var input struct {
		Content   string `json:"content" binding:"required"`
		IsChannel bool   `json:"is_channel"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_INPUT", err.Error())
		return
	}

	user := c.MustGet("user").(*ory.Identity)
	msg, err := h.service.SendMessage(c.Request.Context(), user.Id, targetID, input.Content, input.IsChannel)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}

	utils.RespondCreated(c, msg)
}

func (h *ChatHandler) GetOrCreateConversation(c *gin.Context) {
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_ID", "Invalid workspace ID")
		return
	}

	var input struct {
		TargetUserID string `json:"target_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_INPUT", err.Error())
		return
	}

	user := c.MustGet("user").(*ory.Identity)
	conv, err := h.service.GetOrCreateConversation(c.Request.Context(), workspaceID, user.Id, input.TargetUserID)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}

	utils.RespondOK(c, conv)
}

func (h *ChatHandler) ListConversations(c *gin.Context) {
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "SNAKE_CASE_INVALID_ID", "Invalid workspace ID")
		return
	}

	user := c.MustGet("user").(*ory.Identity)
	convs, err := h.service.ListConversations(c.Request.Context(), workspaceID, user.Id)
	if err != nil {
		utils.RespondMapped(c, err)
		return
	}

	utils.RespondOK(c, convs)
}
