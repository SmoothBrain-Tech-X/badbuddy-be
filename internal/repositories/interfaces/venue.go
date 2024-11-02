package interfaces

import (
	"badbuddy/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type VenueRepository interface {
	Create(ctx context.Context, venue *models.Venue) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.VenueWithCourts, error)
	Update(ctx context.Context, venue *models.Venue) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, location string, limit, offset int) ([]models.Venue, error)
	CountVenues(ctx context.Context) (int, error)
	Search(ctx context.Context, query string, limit, offset int) ([]models.Venue, error)
	AddCourt(ctx context.Context, court *models.Court) error
	UpdateCourt(ctx context.Context, court *models.Court) error
	DeleteCourt(ctx context.Context, id uuid.UUID) error
	GetCourts(ctx context.Context, venueID uuid.UUID) ([]models.Court, error)
	AddReview(ctx context.Context, review *models.VenueReview) error
	GetReviews(ctx context.Context, venueID uuid.UUID, limit, offset int) ([]models.VenueReview, error)
	UpdateVenueRating(ctx context.Context, venueID uuid.UUID) error
}
