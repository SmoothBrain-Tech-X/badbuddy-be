package rest

import (
	"badbuddy/internal/delivery/dto/requests"
	"badbuddy/internal/delivery/dto/responses"
	"badbuddy/internal/delivery/http/middleware"
	"badbuddy/internal/usecase/chat"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"errors"
)

type ChatHandler struct {
	chatUseCase chat.UseCase
}

func NewChatHandler(chatUseCase chat.UseCase) *ChatHandler {
	return &ChatHandler{
		chatUseCase: chatUseCase,
	}
}

func (h *ChatHandler) SetupChatRoutes(app *fiber.App) {
	chat := app.Group("/api/chats")

	// Public routes

	// Protected routes
	chat.Use(middleware.AuthRequired())
	chat.Get("/:chatID/messages", h.GetChatMessage)
	chat.Post("/:chatID/messages", h.SendMessage)
	chat.Delete("/:chatID/messages/:messageID", h.DeleteMessage)
	chat.Put("/:chatID/messages/:messageID", h.UpdateMessage)
}

func (h *ChatHandler) GetChatMessage(c *fiber.Ctx) error {
	chatID := c.Params("chatID")
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return h.handleError(c, errors.New("invalid limit format"))
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return h.handleError(c, errors.New("invalid offset format"))
	}

	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return h.handleError(c, errors.New("invalid chat ID format"))
	}

	userID := c.Locals("userID").(uuid.UUID)

	chat, err := h.chatUseCase.GetChatMessageByID(c.Context(), chatUUID, limit, offset, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(chat)
}

func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	var req requests.SendAndUpdateMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, errors.New("invalid request body"))
	}

	userID := c.Locals("userID").(uuid.UUID)

	chatID := c.Params("chatID")
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return h.handleError(c, errors.New("invalid chat ID format"))
	}

	err = h.chatUseCase.SendMessage(c.Context(), userID, chatUUID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(responses.SuccessResponse{
		Message: "Message sent successfully",
	})
}

func (h *ChatHandler) handleError(c *fiber.Ctx, err error) error {
	var status int
	var errorResponse responses.ErrorResponse

	// Add specific error type checks here if needed
	switch {
	case errors.Is(err, chat.ErrChatNotFound):
		status = fiber.StatusNotFound
		errorResponse = responses.ErrorResponse{
			Error: "Chat not found",
			Code:  "CHAT_NOT_FOUND",
		}
	case errors.Is(err, chat.ErrUnauthorized):
		status = fiber.StatusUnauthorized
		errorResponse = responses.ErrorResponse{
			Error: "Unauthorized",
			Code:  "UNAUTHORIZED",
		}
	case errors.Is(err, chat.ErrValidation):
		status = fiber.StatusBadRequest
		errorResponse = responses.ErrorResponse{
			Error: "Validation error",
			Code:  "VALIDATION_ERROR",
		}
	default:
		status = fiber.StatusInternalServerError
		errorResponse = responses.ErrorResponse{
			Error: "Internal server error",
			Code:  "INTERNAL_ERROR",
		}
	}

	errorResponse.Description = err.Error()
	return c.Status(status).JSON(errorResponse)
}

func (h *ChatHandler) DeleteMessage(c *fiber.Ctx) error {
	chatID := c.Params("chatID")
	messageID := c.Params("messageID")

	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return h.handleError(c, errors.New("invalid chat ID format"))
	}

	messageUUID, err := uuid.Parse(messageID)
	if err != nil {
		return h.handleError(c, errors.New("invalid message ID format"))
	}

	userID := c.Locals("userID").(uuid.UUID)

	err = h.chatUseCase.DeleteMessage(c.Context(), chatUUID, messageUUID, userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(responses.SuccessResponse{
		Message: "Message deleted successfully",
	})
}

func (h *ChatHandler) UpdateMessage(c *fiber.Ctx) error {
	var req requests.SendAndUpdateMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, errors.New("invalid request body"))
	}

	chatID := c.Params("chatID")
	messageID := c.Params("messageID")

	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return h.handleError(c, errors.New("invalid chat ID format"))
	}

	messageUUID, err := uuid.Parse(messageID)
	if err != nil {
		return h.handleError(c, errors.New("invalid message ID format"))
	}

	userID := c.Locals("userID").(uuid.UUID)

	err = h.chatUseCase.UpdateMessage(c.Context(), chatUUID, messageUUID, userID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(responses.SuccessResponse{
		Message: "Message updated successfully",
	})
}
