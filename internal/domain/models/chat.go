// internal/domain/models/chat.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type ChatType string
type MessageType string
type MessageStatus string

const (
	ChatTypeDirect  ChatType = "direct"
	ChatTypeGroup   ChatType = "group"
	ChatTypeSession ChatType = "session"

	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeSystem MessageType = "system"

	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

// Chat represents a conversation between users
type Chat struct {
	ID        uuid.UUID  `db:"id"`
	Type      ChatType   `db:"type"`
	Name      string     `db:"name"`       // For group chats
	SessionID *uuid.UUID `db:"session_id"` // For session chats
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`

	// Populated fields
	Participants []ChatParticipant `db:"participants,omitempty"`
	LastMessage  *Message          `db:"last_message,omitempty"`
	UnreadCount  int               `db:"unread_count,omitempty"`
}

// ChatParticipant represents a user in a chat
type ChatParticipant struct {
	ID         uuid.UUID  `db:"id"`
	ChatID     uuid.UUID  `db:"chat_id"`
	UserID     uuid.UUID  `db:"user_id"`
	IsAdmin    bool       `db:"is_admin"`
	LastReadAt time.Time  `db:"last_read_at"`
	JoinedAt   time.Time  `db:"joined_at"`
	LeftAt     *time.Time `db:"left_at"`

	// Populated fields
	User *User `db:"user,omitempty"`
}

// Message represents a single message in a chat
type Message struct {
	ID        uuid.UUID     `db:"id"`
	ChatID    uuid.UUID     `db:"chat_id"`
	SenderID  uuid.UUID     `db:"sender_id"`
	Type      MessageType   `db:"type"`
	Content   string        `db:"content"`
	Status    MessageStatus `db:"status"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
	DeletedAt *time.Time    `db:"deleted_at"`

	// Populated fields
	Sender *User       `db:"sender,omitempty"`
	ReadBy []uuid.UUID `db:"read_by,omitempty"`
}

// MessageReceipt tracks message delivery and read status
type MessageReceipt struct {
	ID        uuid.UUID     `db:"id"`
	MessageID uuid.UUID     `db:"message_id"`
	UserID    uuid.UUID     `db:"user_id"`
	Status    MessageStatus `db:"status"`
	ReadAt    *time.Time    `db:"read_at"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
}
