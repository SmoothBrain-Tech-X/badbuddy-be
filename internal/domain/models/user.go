package models

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string
type PlayerLevel string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"

	PlayerLevelBeginner     PlayerLevel = "beginner"
	PlayerLevelIntermediate PlayerLevel = "intermediate"
	PlayerLevelAdvanced     PlayerLevel = "advanced"
)

type User struct {
	ID           uuid.UUID   `db:"id"`
	Email        string      `db:"email"`
	Password     string      `db:"password"`
	FirstName    string      `db:"first_name"`
	LastName     string      `db:"last_name"`
	Phone        string      `db:"phone"`
	PlayLevel    PlayerLevel `db:"play_level"`
	Location     string      `db:"location"`
	Bio          string      `db:"bio"`
	AvatarURL    string      `db:"avatar_url"`
	Status       UserStatus  `db:"status"`
	CreatedAt    time.Time   `db:"created_at"`
	LastActiveAt time.Time   `db:"last_active_at"`
	Gender       *string     `db:"gender"`
	PlayHand     *string     `db:"play_hand"`
}

type UserProfile struct {
	User
	HostedSessions  int     `db:"hosted_sessions"`
	JoinedSessions  int     `db:"joined_sessions"`
	AverageRating   float64 `db:"avg_rating"`
	TotalReviews    int     `db:"total_reviews"`
	RegularPartners int     `db:"regular_partners"`
}
