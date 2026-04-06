package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/resoul/studio.go.api/internal/domain"
	"gorm.io/gorm"
)

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) domain.ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateChannel(ctx context.Context, channel *domain.Channel) error {
	return r.db.WithContext(ctx).Create(channel).Error
}

func (r *chatRepository) GetChannel(ctx context.Context, id uuid.UUID) (*domain.Channel, error) {
	var channel domain.Channel
	if err := r.db.WithContext(ctx).First(&channel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *chatRepository) ListChannels(ctx context.Context, workspaceID uuid.UUID, userID string) ([]domain.Channel, error) {
	var channels []domain.Channel
	// List all public channels in workspace OR private channels where user is a member
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND (type = ? OR id IN (SELECT channel_id FROM channel_members WHERE user_id = ?))",
			workspaceID, domain.ChannelTypePublic, userID).
		Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}

func (r *chatRepository) DeleteChannel(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Channel{}, "id = ?", id).Error
}

func (r *chatRepository) CreateConversation(ctx context.Context, conv *domain.DirectMessageConversation) error {
	return r.db.WithContext(ctx).Create(conv).Error
}

func (r *chatRepository) GetConversation(ctx context.Context, id uuid.UUID) (*domain.DirectMessageConversation, error) {
	var conv domain.DirectMessageConversation
	if err := r.db.WithContext(ctx).First(&conv, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *chatRepository) FindConversation(ctx context.Context, workspaceID uuid.UUID, user1ID, user2ID string) (*domain.DirectMessageConversation, error) {
	var conv domain.DirectMessageConversation
	// Search for conversation where both users are participants
	err := r.db.WithContext(ctx).Where("workspace_id = ? AND ((user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?))",
		workspaceID, user1ID, user2ID, user2ID, user1ID).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *chatRepository) ListConversations(ctx context.Context, workspaceID uuid.UUID, userID string) ([]domain.DirectMessageConversation, error) {
	var convs []domain.DirectMessageConversation
	if err := r.db.WithContext(ctx).Where("workspace_id = ? AND (user1_id = ? OR user2_id = ?)", workspaceID, userID, userID).Find(&convs).Error; err != nil {
		return nil, err
	}
	return convs, nil
}

func (r *chatRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *chatRepository) ListMessages(ctx context.Context, targetID uuid.UUID, limit, offset int) ([]domain.Message, error) {
	var messages []domain.Message
	// targetID can be either a ChannelID or a ConversationID
	// We check both fields.
	err := r.db.WithContext(ctx).
		Where("channel_id = ? OR conversation_id = ?", targetID, targetID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *chatRepository) AddChannelMember(ctx context.Context, member *domain.ChannelMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *chatRepository) IsMember(ctx context.Context, channelID uuid.UUID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.ChannelMember{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *chatRepository) ListChannelMembers(ctx context.Context, channelID uuid.UUID) ([]domain.ChannelMember, error) {
	var members []domain.ChannelMember
	err := r.db.WithContext(ctx).Where("channel_id = ?", channelID).Find(&members).Error
	return members, err
}
