package rest

import (
	"badbuddy/internal/delivery/http/middleware"
	"badbuddy/internal/usecase/session"

	"badbuddy/internal/delivery/dto/requests"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SessionHandler struct {
	sessionUseCase session.UseCase
}

func NewSessionHandler(sessionUseCase session.UseCase) *SessionHandler {
	return &SessionHandler{
		sessionUseCase: sessionUseCase,
	}
}
func (h *SessionHandler) SetupSessionRoutes(app *fiber.App) {
	sessions := app.Group("/api/sessions")

	// Public routes
	sessions.Get("/", h.ListSessions)
	sessions.Get("/:id", h.GetSession)

	// Protected routes
	sessions.Use(middleware.AuthRequired())
	sessions.Post("/", h.CreateSession)
	sessions.Put("/:id", h.UpdateSession)
	sessions.Post("/:id/join", h.JoinSession)
	sessions.Post("/:id/leave", h.LeaveSession)
	sessions.Get("/user/me", h.GetUserSessions)
}
func (h *SessionHandler) CreateSession(c *fiber.Ctx) error {
	var req requests.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	hostID := c.Locals("userID").(uuid.UUID)

	session, err := h.sessionUseCase.CreateSession(c.Context(), hostID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(session)
}

func (h *SessionHandler) GetSession(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid session ID",
		})
	}

	session, err := h.sessionUseCase.GetSession(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(session)
}

func (h *SessionHandler) ListSessions(c *fiber.Ctx) error {
	filters := map[string]interface{}{
		"date":         c.Query("date"),
		"location":     c.Query("location"),
		"player_level": c.Query("player_level"),
		"status":       c.Query("status"),
	}

	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	sessions, err := h.sessionUseCase.ListSessions(c.Context(), filters, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(sessions)
}

func (h *SessionHandler) JoinSession(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid session ID",
		})
	}

	userID := c.Locals("userID").(uuid.UUID)

	if err := h.sessionUseCase.JoinSession(c.Context(), sessionID, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully joined session",
	})
}

// internal/delivery/http/handlers/session.go (continued)

func (h *SessionHandler) LeaveSession(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid session ID",
		})
	}

	userID := c.Locals("userID").(uuid.UUID)

	if err := h.sessionUseCase.LeaveSession(c.Context(), sessionID, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully left session",
	})
}

func (h *SessionHandler) UpdateSession(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid session ID",
		})
	}

	var req requests.UpdateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.sessionUseCase.UpdateSession(c.Context(), sessionID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Session updated successfully",
	})
}

func (h *SessionHandler) GetUserSessions(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	includeHistory := c.QueryBool("include_history", false)

	sessions, err := h.sessionUseCase.GetUserSessions(c.Context(), userID, includeHistory)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"sessions": sessions,
	})
}
