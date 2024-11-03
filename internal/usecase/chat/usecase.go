package chat

import (
	"badbuddy/internal/delivery/dto/requests"
	"badbuddy/internal/delivery/dto/responses"
	"badbuddy/internal/domain/models"
	"badbuddy/internal/repositories/interfaces"
	"context"
	"errors"
	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("unauthorized")

	ErrValidation = errors.New("validation error")

	ErrChatNotFound = errors.New("chat not found")
)

type useCase struct {
	chatRepo interfaces.ChatRepository
	userRepo interfaces.UserRepository
}

func NewChatUseCase(chatRepo interfaces.ChatRepository, userRepo interfaces.UserRepository) UseCase {
	return &useCase{
		chatRepo: chatRepo,
		userRepo: userRepo,
	}
}

func (uc *useCase) GetChatMessageByID(ctx context.Context, chatID uuid.UUID, limit int, offset int, userID uuid.UUID) (*responses.ChatResponse, error) {
	isPartOfChat, err := uc.chatRepo.IsUserPartOfChat(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}
	if !isPartOfChat {
		return nil, ErrUnauthorized
	}

	chat, err := uc.chatRepo.GetChatMessageByID(ctx, chatID, limit, offset)

	if err != nil {
		return nil, err
	}

	err = uc.chatRepo.UpdateChatMessageReadStatus(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}

	chatMassage := []responses.ChatMassageResponse{}

	for _, m := range *chat {
		chatMassage = append(chatMassage, responses.ChatMassageResponse{
			ID:     m.ID.String(),
			ChatID: m.ChatID.String(),
			Autor: responses.UserResponse{
				ID:           m.SenderID.String(),
				Email:        m.Email,
				FirstName:    m.FirstName,
				LastName:     m.LastName,
				Phone:        m.Phone,
				PlayLevel:    string(m.PlayLevel),
				Location:     *m.Location,
				Bio:          *m.Bio,
				AvatarURL:    *m.AvatarURL,
				LastActiveAt: m.LastActiveAt,
			},
			Message:       m.Content,
			Timestamp:     m.CreatedAt,
			EditTimeStamp: m.UpdatedAt,
		})

	}

	return &responses.ChatResponse{
		ChatMassage: chatMassage,
	}, nil

}

func (uc *useCase) SendMessage(ctx context.Context, userID, chatID uuid.UUID, req requests.SendAndUpdateMessageRequest) error {
	if req.Message == "" {
		return ErrValidation
	}

	isPartOfChat, err := uc.chatRepo.IsUserPartOfChat(ctx, userID, chatID)
	if err != nil {
		return err
	}
	if !isPartOfChat {
		return ErrUnauthorized
	}

	_, err = uc.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return ErrChatNotFound
	}

	message := models.Message{
		ID:       uuid.New(),
		ChatID:   chatID,
		SenderID: userID,
		Type:     models.MessageTypeText,
		Content:  req.Message,
		Status:   models.MessageStatusSent,
	}

	err = uc.chatRepo.SaveMessage(ctx, &message)
	if err != nil {
		return err
	}

	return nil
}

func (uc *useCase) DeleteMessage(ctx context.Context, chatID, messageID, userID uuid.UUID) error {
	isPartOfChat, err := uc.chatRepo.IsUserPartOfChat(ctx, userID, chatID)
	if err != nil {
		return err
	}
	if !isPartOfChat {
		return ErrUnauthorized
	}

	message, err := uc.chatRepo.GetChatMessageByID(ctx, chatID, 1, 0)
	if err != nil {
		return err
	}

	if len(*message) == 0 {
		return ErrChatNotFound
	}

	if (*message)[0].SenderID != userID {
		return ErrUnauthorized
	}

	err = uc.chatRepo.DeleteChatMessage(ctx, messageID)
	if err != nil {
		return err
	}

	return nil
}

func (uc *useCase) UpdateMessage(ctx context.Context, chatID, messageID, userID uuid.UUID, req requests.SendAndUpdateMessageRequest) error {
	isPartOfChat, err := uc.chatRepo.IsUserPartOfChat(ctx, userID, chatID)
	if err != nil {
		return err
	}

	if !isPartOfChat {
		return ErrUnauthorized
	}

	message, err := uc.chatRepo.GetChatMessageByID(ctx, chatID, 1, 0)
	if err != nil {
		return err
	}

	if len(*message) == 0 {
		return ErrChatNotFound
	}

	if (*message)[0].SenderID != userID {
		return ErrUnauthorized
	}

	messageToUpdate := models.Message{
		ID:      messageID,
		Content: req.Message,
	}

	err = uc.chatRepo.UpdateChatMessage(ctx, &messageToUpdate)
	if err != nil {
		return err
	}

	return nil
}
