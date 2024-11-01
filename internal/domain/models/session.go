// internal/domain/models/session.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type SessionStatus string
type ParticipantStatus string

const (
	SessionStatusOpen      SessionStatus = "open"
	SessionStatusFull      SessionStatus = "full"
	SessionStatusCancelled SessionStatus = "cancelled"
	SessionStatusCompleted SessionStatus = "completed"

	ParticipantStatusJoined    ParticipantStatus = "joined"
	ParticipantStatusWaitlist  ParticipantStatus = "waitlist"
	ParticipantStatusCancelled ParticipantStatus = "cancelled"
)

type Session struct {
	ID            uuid.UUID     `db:"id"`
	HostID        uuid.UUID     `db:"host_id"`
	VenueID       uuid.UUID     `db:"venue_id"`
	CourtID       uuid.UUID     `db:"court_id"`
	Title         string        `db:"title"`
	Description   string        `db:"description"`
	SessionDate   time.Time     `db:"session_date"`
	StartTime     time.Time     `db:"start_time"`
	EndTime       time.Time     `db:"end_time"`
	PlayerLevel   PlayerLevel   `db:"player_level"`
	MinPlayers    int           `db:"min_players"`
	MaxPlayers    int           `db:"max_players"`
	CostPerPerson float64       `db:"cost_per_person"`
	Status        SessionStatus `db:"status"`
	CreatedAt     time.Time     `db:"created_at"`
	UpdatedAt     time.Time     `db:"updated_at"`
}

type SessionParticipant struct {
	ID          uuid.UUID         `db:"id"`
	SessionID   uuid.UUID         `db:"session_id"`
	UserID      uuid.UUID         `db:"user_id"`
	Status      ParticipantStatus `db:"status"`
	JoinedAt    time.Time         `db:"joined_at"`
	CancelledAt *time.Time        `db:"cancelled_at"`
}

type SessionDetail struct {
	Session
	VenueName      string               `db:"venue_name"`
	VenueLocation  string               `db:"venue_location"`
	CourtName      string               `db:"court_name"`
	HostName       string               `db:"host_name"`
	HostLevel      string               `db:"host_level"`
	CurrentPlayers int                  `db:"current_players"`
	WaitlistCount  int                  `db:"waitlist_count"`
	Participants   []SessionParticipant `db:"participants,omitempty"`
}
