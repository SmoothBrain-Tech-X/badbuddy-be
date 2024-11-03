package interfaces

import (
	"context"
	"badbuddy/internal/domain/models"

	"github.com/google/uuid"
)

type ChatRepository interface {
	GetChatMessageByID(ctx context.Context, chatID uuid.UUID, limit int, offset int) (*[]models.Message, error) // Get messages of a chat
	GetChatByID(ctx context.Context, chatID uuid.UUID) (*models.Chat, error)
	IsUserPartOfChat(ctx context.Context, userID, chatID uuid.UUID) (bool, error)
	SaveMessage(ctx context.Context, message *models.Message) (*models.Message, error)
	CreateChat(ctx context.Context, chat *models.Chat) error
	AddUserToChat(ctx context.Context, userID, chatID uuid.UUID) error
	RemoveUserFromChat(ctx context.Context, userID, chatID uuid.UUID) error
	UpdateChatMessage(ctx context.Context, message *models.Message) error
	DeleteChatMessage(ctx context.Context, messageID uuid.UUID) error
	UpdateChatMessageReadStatus(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) error
	GetMessageByID(ctx context.Context, messageID uuid.UUID) (*models.Message, error) // Get a message by ID
}