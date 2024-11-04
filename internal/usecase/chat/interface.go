package chat

import (
	"badbuddy/internal/delivery/dto/requests"
	"badbuddy/internal/delivery/dto/responses"
	"context"

	"github.com/google/uuid"
)

type UseCase interface {
	GetChatMessageByID(ctx context.Context, chatID uuid.UUID, limit int, offset int, userID uuid.UUID) (*responses.ChatMassageListResponse, error)

	SendMessage(ctx context.Context, userID uuid.UUID, chatID uuid.UUID, req requests.SendAndUpdateMessageRequest) (*responses.ChatMassageResponse, error)

	DeleteMessage(ctx context.Context, chatID uuid.UUID, messageID uuid.UUID, userID uuid.UUID) error

	UpdateMessage(ctx context.Context, chatID uuid.UUID, messageID uuid.UUID, userID uuid.UUID, req requests.SendAndUpdateMessageRequest) error

	GetChats(ctx context.Context, userID uuid.UUID) (*responses.ChatListResponse, error)
}