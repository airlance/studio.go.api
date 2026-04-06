package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ChannelType string

const (
	ChannelTypePublic  ChannelType = "public"
	ChannelTypePrivate ChannelType = "private"
)

type Channel struct {
	ID          uuid.UUID   `gorm:"primaryKey;type:uuid" json:"id"`
	WorkspaceID uuid.UUID   `gorm:"type:uuid;index" json:"workspace_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        ChannelType `gorm:"default:'public'" json:"type"`
	CreatedBy   string      `json:"created_by"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type DirectMessageConversation struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index" json:"workspace_id"`
	User1ID     string    `gorm:"index" json:"user1_id"`
	User2ID     string    `gorm:"index" json:"user2_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Message struct {
	ID             uuid.UUID  `gorm:"primaryKey;type:uuid" json:"id"`
	ChannelID      *uuid.UUID `gorm:"type:uuid;index" json:"channel_id,omitempty"`
	ConversationID *uuid.UUID `gorm:"type:uuid;index" json:"conversation_id,omitempty"`
	SenderID       string     `gorm:"index" json:"sender_id"`
	Content        string     `json:"content"`
	CreatedAt      time.Time  `json:"created_at"`
}

type ChannelMember struct {
	ChannelID uuid.UUID `gorm:"primaryKey;type:uuid" json:"channel_id"`
	UserID    string    `gorm:"primaryKey" json:"user_id"`
	JoinedAt  time.Time `json:"joined_at"`
}

type ChatRepository interface {
	CreateChannel(ctx context.Context, channel *Channel) error
	GetChannel(ctx context.Context, id uuid.UUID) (*Channel, error)
	ListChannels(ctx context.Context, workspaceID uuid.UUID, userID string) ([]Channel, error)
	DeleteChannel(ctx context.Context, id uuid.UUID) error

	CreateConversation(ctx context.Context, conv *DirectMessageConversation) error
	GetConversation(ctx context.Context, id uuid.UUID) (*DirectMessageConversation, error)
	FindConversation(ctx context.Context, workspaceID uuid.UUID, user1ID, user2ID string) (*DirectMessageConversation, error)
	ListConversations(ctx context.Context, workspaceID uuid.UUID, userID string) ([]DirectMessageConversation, error)

	SaveMessage(ctx context.Context, msg *Message) error
	ListMessages(ctx context.Context, targetID uuid.UUID, limit, offset int) ([]Message, error)

	AddChannelMember(ctx context.Context, member *ChannelMember) error
	IsMember(ctx context.Context, channelID uuid.UUID, userID string) (bool, error)
	ListChannelMembers(ctx context.Context, channelID uuid.UUID) ([]ChannelMember, error)
}

type ChatService interface {
	CreateChannel(ctx context.Context, workspaceID uuid.UUID, name, description string, isPrivate bool, creatorID string, participants []string) (*Channel, error)
	ListChannels(ctx context.Context, workspaceID uuid.UUID, userID string) ([]Channel, error)
	GetChannelMessages(ctx context.Context, channelID uuid.UUID, limit, offset int) ([]Message, error)

	GetOrCreateConversation(ctx context.Context, workspaceID uuid.UUID, user1ID, user2ID string) (*DirectMessageConversation, error)
	ListConversations(ctx context.Context, workspaceID uuid.UUID, userID string) ([]DirectMessageConversation, error)
	GetConversationMessages(ctx context.Context, convID uuid.UUID, limit, offset int) ([]Message, error)

	SendMessage(ctx context.Context, senderID string, targetID uuid.UUID, content string, isChannel bool) (*Message, error)
}
