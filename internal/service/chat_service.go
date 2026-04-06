package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/resoul/studio.go.api/internal/domain"
	"github.com/sirupsen/logrus"
)

type chatService struct {
	repo domain.ChatRepository
	hub  domain.PresenceHub
}

func NewChatService(repo domain.ChatRepository, hub domain.PresenceHub) domain.ChatService {
	return &chatService{
		repo: repo,
		hub:  hub,
	}
}

func (s *chatService) CreateChannel(ctx context.Context, workspaceID uuid.UUID, name, description string, isPrivate bool, creatorID string, participants []string) (*domain.Channel, error) {
	logrus.WithFields(logrus.Fields{
		"workspace_id": workspaceID,
		"name":         name,
		"creator_id":   creatorID,
	}).Info("Creating new channel")

	channel := &domain.Channel{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
		Type:        domain.ChannelTypePublic,
		CreatedBy:   creatorID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if isPrivate {
		channel.Type = domain.ChannelTypePrivate
	}

	if err := s.repo.CreateChannel(ctx, channel); err != nil {
		return nil, err
	}

	// Add creator as member
	_ = s.repo.AddChannelMember(ctx, &domain.ChannelMember{
		ChannelID: channel.ID,
		UserID:    creatorID,
		JoinedAt:  time.Now(),
	})

	// Add other participants
	for _, pID := range participants {
		if pID != creatorID && pID != "" {
			_ = s.repo.AddChannelMember(ctx, &domain.ChannelMember{
				ChannelID: channel.ID,
				UserID:    pID,
				JoinedAt:  time.Now(),
			})
		}
	}

	return channel, nil
}

func (s *chatService) ListChannels(ctx context.Context, workspaceID uuid.UUID, userID string) ([]domain.Channel, error) {
	channels, err := s.repo.ListChannels(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}

	if len(channels) == 0 {
		// If no channels exist for this user, check if any public channels exist in the workspace
		publicChannels, _ := s.repo.ListChannels(ctx, workspaceID, "") // Empty userID only gets public channels
		if len(publicChannels) == 0 {
			// Create default #general channel
			general, err := s.CreateChannel(ctx, workspaceID, "general", "Default channel for everyone", false, "system", nil)
			if err == nil {
				channels = append(channels, *general)
			}
		} else {
			// User is not a member of any private channel and there are public ones they should see?
			// Actually my repo logic handles public + joined private.
			// So if they see nothing, they just see nothing (or we just return the public ones).
			return publicChannels, nil
		}
	}

	return channels, nil
}

func (s *chatService) GetChannelMessages(ctx context.Context, channelID uuid.UUID, limit, offset int) ([]domain.Message, error) {
	return s.repo.ListMessages(ctx, channelID, limit, offset)
}

func (s *chatService) GetOrCreateConversation(ctx context.Context, workspaceID uuid.UUID, user1ID, user2ID string) (*domain.DirectMessageConversation, error) {
	conv, err := s.repo.FindConversation(ctx, workspaceID, user1ID, user2ID)
	if err == nil {
		return conv, nil
	}

	conv = &domain.DirectMessageConversation{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		User1ID:     user1ID,
		User2ID:     user2ID,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}

	return conv, nil
}

func (s *chatService) ListConversations(ctx context.Context, workspaceID uuid.UUID, userID string) ([]domain.DirectMessageConversation, error) {
	return s.repo.ListConversations(ctx, workspaceID, userID)
}

func (s *chatService) GetConversationMessages(ctx context.Context, convID uuid.UUID, limit, offset int) ([]domain.Message, error) {
	return s.repo.ListMessages(ctx, convID, limit, offset)
}

func (s *chatService) SendMessage(ctx context.Context, senderID string, targetID uuid.UUID, content string, isChannel bool) (*domain.Message, error) {
	msg := &domain.Message{
		ID:        uuid.New(),
		SenderID:  senderID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if isChannel {
		msg.ChannelID = &targetID
	} else {
		msg.ConversationID = &targetID
	}

	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	// Broadcast message via Hub
	logrus.WithFields(logrus.Fields{
		"message_id": msg.ID,
		"sender_id":  senderID,
		"target_id":  targetID,
		"is_channel": isChannel,
	}).Info("Broadcasting chat message via Hub")
	s.hub.Broadcast(ctx, msg)

	return msg, nil
}
