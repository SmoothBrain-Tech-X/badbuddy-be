package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"badbuddy/internal/domain/models"
	"badbuddy/internal/repositories/interfaces"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type venueRepository struct {
	db *sqlx.DB
}

func NewVenueRepository(db *sqlx.DB) interfaces.VenueRepository {
	return &venueRepository{
		db: db,
	}
}

func (r *venueRepository) Create(ctx context.Context, venue *models.Venue) error {
	query := `
		INSERT INTO venues (
			id, name, description, address, location, phone, email,
			open_time, close_time, image_urls, status, rating,
			total_reviews, owner_id, created_at, updated_at
		) VALUES (
			:id, :name, :description, :address, :location, :phone, :email,
			:open_time, :close_time, :image_urls, :status, :rating,
			:total_reviews, :owner_id, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, venue)
	if err != nil {
		return fmt.Errorf("failed to create venue: %w", err)
	}
	return nil
}

func (r *venueRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.VenueWithCourts, error) {
	result := &models.VenueWithCourts{}

	// Get venue details
	query := `
		SELECT * FROM venues WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &result.Venue, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("venue not found")
		}
		return nil, fmt.Errorf("failed to get venue: %w", err)
	}

	// Get courts
	courtsQuery := `
		SELECT * FROM courts 
		WHERE venue_id = $1 AND deleted_at IS NULL 
		ORDER BY created_at`
	err = r.db.SelectContext(ctx, &result.Courts, courtsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get courts: %w", err)
	}

	return result, nil
}

func (r *venueRepository) Update(ctx context.Context, venue *models.Venue) error {
	query := `
		UPDATE venues SET
			name = :name,
			description = :description,
			address = :address,
			location = :location,
			phone = :phone,
			email = :email,
			open_time = :open_time,
			close_time = :close_time,
			image_urls = :image_urls,
			status = :status,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.db.NamedExecContext(ctx, query, venue)
	if err != nil {
		return fmt.Errorf("failed to update venue: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("venue not found")
	}

	return nil
}

func (r *venueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE venues 
		SET deleted_at = NOW(), updated_at = NOW() 
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete venue: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("venue not found")
	}

	return nil
}

func (r *venueRepository) List(ctx context.Context, location string, limit, offset int) ([]models.Venue, error) {
	query := `
		SELECT * FROM venues 
		WHERE deleted_at IS NULL
		AND ($1 = '' OR location = $1)
		ORDER BY rating DESC, total_reviews DESC, created_at DESC
		LIMIT $2 OFFSET $3`

	venues := []models.Venue{}
	err := r.db.SelectContext(ctx, &venues, query, location, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list venues: %w", err)
	}

	return venues, nil
}

func (r *venueRepository) Search(ctx context.Context, query string, limit, offset int) ([]models.Venue, error) {
	searchQuery := `
		SELECT * FROM venues 
		WHERE deleted_at IS NULL
		AND (
			search_vector @@ plainto_tsquery($1)
			OR name ILIKE '%' || $1 || '%'
			OR location ILIKE '%' || $1 || '%'
		)
		ORDER BY rating DESC, total_reviews DESC, created_at DESC
		LIMIT $2 OFFSET $3`

	venues := []models.Venue{}
	err := r.db.SelectContext(ctx, &venues, searchQuery, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search venues: %w", err)
	}

	return venues, nil
}

func (r *venueRepository) AddCourt(ctx context.Context, court *models.Court) error {
	query := `
		INSERT INTO courts (
			id, venue_id, name, description, price_per_hour,
			status, created_at, updated_at
		) VALUES (
			:id, :venue_id, :name, :description, :price_per_hour,
			:status, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, court)
	if err != nil {
		return fmt.Errorf("failed to add court: %w", err)
	}

	return nil
}

func (r *venueRepository) UpdateCourt(ctx context.Context, court *models.Court) error {
	query := `
		UPDATE courts SET
			name = :name,
			description = :description,
			price_per_hour = :price_per_hour,
			status = :status,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.db.NamedExecContext(ctx, query, court)
	if err != nil {
		return fmt.Errorf("failed to update court: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("court not found")
	}

	return nil
}

func (r *venueRepository) DeleteCourt(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE courts 
		SET deleted_at = NOW(), updated_at = NOW() 
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete court: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("court not found")
	}

	return nil
}

func (r *venueRepository) GetCourts(ctx context.Context, venueID uuid.UUID) ([]models.Court, error) {
	query := `
		SELECT * FROM courts 
		WHERE venue_id = $1 AND deleted_at IS NULL 
		ORDER BY created_at`

	courts := []models.Court{}
	err := r.db.SelectContext(ctx, &courts, query, venueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get courts: %w", err)
	}

	return courts, nil
}

func (r *venueRepository) AddReview(ctx context.Context, review *models.VenueReview) error {

	// Insert review
	query := `
		INSERT INTO venue_reviews (
			id, venue_id, user_id, rating, comment, created_at
		) VALUES (
			:id, :venue_id, :user_id, :rating, :comment, :created_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, review)
	if err != nil {
		return fmt.Errorf("failed to add review: %w", err)
	}

	fmt.Println(review)

	// Update venue rating
	err = r.UpdateVenueRating(ctx, review.VenueID)
	if err != nil {
		return fmt.Errorf("failed to update venue rating: %w", err)
	}

	return nil

}

func (r *venueRepository) GetReviews(ctx context.Context, venueID uuid.UUID, limit, offset int) ([]models.VenueReview, error) {
	query := `
		SELECT vr.*, 
			u.id as user_id
		FROM venue_reviews vr
		JOIN users u ON u.id = vr.user_id
		WHERE vr.venue_id = $1
		ORDER BY vr.created_at DESC
		LIMIT $2 OFFSET $3`

	reviews := []models.VenueReview{}
	err := r.db.SelectContext(ctx, &reviews, query, venueID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}

	return reviews, nil
}

func (r *venueRepository) UpdateVenueRating(ctx context.Context, venueID uuid.UUID) error {
	query := `
		UPDATE venues 
		SET 
			rating = (
				SELECT COALESCE(AVG(rating)::NUMERIC(3,2), 0)
				FROM venue_reviews
				WHERE venue_id = $1
			),
			total_reviews = (
				SELECT COUNT(*)
				FROM venue_reviews
				WHERE venue_id = $1
			),
			updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, venueID)
	if err != nil {
		return fmt.Errorf("failed to update venue rating: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("venue not found")
	}

	return nil
}
