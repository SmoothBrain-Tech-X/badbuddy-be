// internal/domain/models/venue.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type VenueStatus string
type CourtStatus string

const (
	VenueStatusActive      VenueStatus = "active"
	VenueStatusInactive    VenueStatus = "inactive"
	VenueStatusMaintenance VenueStatus = "maintenance"

	CourtStatusAvailable   CourtStatus = "available"
	CourtStatusOccupied    CourtStatus = "occupied"
	CourtStatusMaintenance CourtStatus = "maintenance"
)

type Venue struct {
	ID           uuid.UUID   `db:"id"`
	Name         string      `db:"name"`
	Description  string      `db:"description"`
	Address      string      `db:"address"`
	Location     string      `db:"location"`
	Phone        string      `db:"phone"`
	Email        string      `db:"email"`
	OpenTime     string      `db:"open_time"`
	CloseTime    string      `db:"close_time"`
	ImageURLs    string      `db:"image_urls"`
	Status       VenueStatus `db:"status"`
	Rating       float64     `db:"rating"`
	TotalReviews int         `db:"total_reviews"`
	OwnerID      uuid.UUID   `db:"owner_id"`
	CreatedAt    time.Time   `db:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at"`
	DeletedAt    *time.Time  `db:"deleted_at"`
}

type Court struct {
	ID           uuid.UUID   `db:"id"`
	VenueID      uuid.UUID   `db:"venue_id"`
	Name         string      `db:"name"`
	Description  string      `db:"description"`
	PricePerHour float64     `db:"price_per_hour"`
	Status       CourtStatus `db:"status"`
	CreatedAt    time.Time   `db:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at"`
	DeletedAt    *time.Time  `db:"deleted_at"`
}

type VenueWithCourts struct {
	Venue
	Courts []Court `db:"courts"`
}

type VenueReview struct {
	ID        uuid.UUID `db:"id"`
	VenueID   uuid.UUID `db:"venue_id"`
	UserID    uuid.UUID `db:"user_id"`
	Rating    int       `db:"rating"`
	Comment   string    `db:"comment"`
	CreatedAt time.Time `db:"created_at"`
	UpdateAt  time.Time `db:"updated_at"`
}
