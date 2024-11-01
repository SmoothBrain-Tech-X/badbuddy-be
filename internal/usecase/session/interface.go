package session

import (
	"context"

	"badbuddy/internal/delivery/dto/requests"
	"badbuddy/internal/delivery/dto/responses"

	"github.com/google/uuid"
)

type UseCase interface {
	CreateSession(ctx context.Context, hostID uuid.UUID, req requests.CreateSessionRequest) (*responses.SessionResponse, error)
	GetSession(ctx context.Context, id uuid.UUID) (*responses.SessionResponse, error)
	UpdateSession(ctx context.Context, id uuid.UUID, req requests.UpdateSessionRequest) error
	ListSessions(ctx context.Context, filters map[string]interface{}, limit, offset int) (*responses.SessionListResponse, error)
	JoinSession(ctx context.Context, sessionID, userID uuid.UUID) error
	LeaveSession(ctx context.Context, sessionID, userID uuid.UUID) error
	GetUserSessions(ctx context.Context, userID uuid.UUID, includeHistory bool) ([]responses.SessionResponse, error)
}
