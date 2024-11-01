package venue

import (
	"context"
	"fmt"
	"time"

	"badbuddy/internal/delivery/dto/requests"
	"badbuddy/internal/delivery/dto/responses"
	"badbuddy/internal/domain/models"
	"badbuddy/internal/repositories/interfaces"

	"github.com/google/uuid"
)

type useCase struct {
	venueRepo interfaces.VenueRepository
	userRepo  interfaces.UserRepository
}

func NewVenueUseCase(venueRepo interfaces.VenueRepository, userRepo interfaces.UserRepository) UseCase {
	return &useCase{
		venueRepo: venueRepo,
		userRepo:  userRepo,
	}
}

func (uc *useCase) CreateVenue(ctx context.Context, ownerID uuid.UUID, req requests.CreateVenueRequest) (*responses.VenueResponse, error) {
	venue := &models.Venue{
		OpenTime:    req.OpenTime,
		CloseTime:   req.CloseTime,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Location:    req.Location,
		Phone:       req.Phone,
		Email:       req.Email,
		ImageURLs:   req.ImageURLs,
		Status:      models.VenueStatusActive,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := uc.venueRepo.Create(ctx, venue); err != nil {
		return nil, fmt.Errorf("failed to create venue: %w", err)
	}

	return &responses.VenueResponse{
		ID:           venue.ID.String(),
		Name:         venue.Name,
		Description:  venue.Description,
		Address:      venue.Address,
		Location:     venue.Location,
		Phone:        venue.Phone,
		Email:        venue.Email,
		OpenTime:     venue.OpenTime,
		CloseTime:    venue.CloseTime,
		ImageURLs:    venue.ImageURLs,
		Status:       string(venue.Status),
		Rating:       venue.Rating,
		TotalReviews: venue.TotalReviews,
	}, nil
}

func (uc *useCase) GetVenue(ctx context.Context, id uuid.UUID) (*responses.VenueResponse, error) {
	venueWithCourts, err := uc.venueRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get venue: %w", err)
	}

	courts := make([]responses.CourtResponse, len(venueWithCourts.Courts))
	for i, court := range venueWithCourts.Courts {
		courts[i] = responses.CourtResponse{
			ID:           court.ID.String(),
			Name:         court.Name,
			Description:  court.Description,
			PricePerHour: court.PricePerHour,
			Status:       string(court.Status),
		}
	}

	return &responses.VenueResponse{
		ID:           venueWithCourts.ID.String(),
		Name:         venueWithCourts.Name,
		Description:  venueWithCourts.Description,
		Address:      venueWithCourts.Address,
		Location:     venueWithCourts.Location,
		Phone:        venueWithCourts.Phone,
		Email:        venueWithCourts.Email,
		OpenTime:     venueWithCourts.OpenTime,
		CloseTime:    venueWithCourts.CloseTime,
		ImageURLs:    venueWithCourts.ImageURLs,
		Status:       string(venueWithCourts.Status),
		Rating:       venueWithCourts.Rating,
		TotalReviews: venueWithCourts.TotalReviews,
		Courts:       courts,
	}, nil
}

func (uc *useCase) UpdateVenue(ctx context.Context, id uuid.UUID, req requests.UpdateVenueRequest) error {
	venue, err := uc.venueRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get venue: %w", err)
	}

	// Update fields if provided
	if req.Name != "" {
		venue.Name = req.Name
	}
	if req.Description != "" {
		venue.Description = req.Description
	}
	if req.Address != "" {
		venue.Address = req.Address
	}
	if req.Location != "" {
		venue.Location = req.Location
	}
	if req.Phone != "" {
		venue.Phone = req.Phone
	}
	if req.Email != "" {
		venue.Email = req.Email
	}
	if !req.OpenTime.IsZero() {
		openTime, err := time.Parse(time.RFC3339, req.OpenTime.Format(time.RFC3339))
		if err != nil {
			return fmt.Errorf("invalid open time format: %w", err)
		}
		venue.OpenTime = openTime
	}
	if !req.CloseTime.IsZero() {
		closeTime, err := time.Parse(time.RFC3339, req.CloseTime.Format(time.RFC3339))
		if err != nil {
			return fmt.Errorf("invalid close time format: %w", err)
		}
		venue.CloseTime = closeTime
	}
	if req.ImageURLs != "" {
		venue.ImageURLs = req.ImageURLs
	}
	if req.Status != "" {
		venue.Status = models.VenueStatus(req.Status)
	}

	venue.UpdatedAt = time.Now()

	if err := uc.venueRepo.Update(ctx, &venue.Venue); err != nil {
		return fmt.Errorf("failed to update venue: %w", err)
	}

	return nil
}

func (uc *useCase) ListVenues(ctx context.Context, location string, limit, offset int) ([]responses.VenueResponse, error) {
	venues, err := uc.venueRepo.List(ctx, location, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list venues: %w", err)
	}

	venueResponses := make([]responses.VenueResponse, len(venues))
	for i, venue := range venues {
		venueResponses[i] = responses.VenueResponse{
			ID:           venue.ID.String(),
			Name:         venue.Name,
			Description:  venue.Description,
			Address:      venue.Address,
			Location:     venue.Location,
			Phone:        venue.Phone,
			Email:        venue.Email,
			OpenTime:     venue.OpenTime,
			CloseTime:    venue.CloseTime,
			ImageURLs:    venue.ImageURLs,
			Status:       string(venue.Status),
			Rating:       venue.Rating,
			TotalReviews: venue.TotalReviews,
		}
	}

	return venueResponses, nil
}

func (uc *useCase) SearchVenues(ctx context.Context, query string, limit, offset int) ([]responses.VenueResponse, error) {
	venues, err := uc.venueRepo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search venues: %w", err)
	}

	venueResponses := make([]responses.VenueResponse, len(venues))
	for i, venue := range venues {
		venueResponses[i] = responses.VenueResponse{
			ID:           venue.ID.String(),
			Name:         venue.Name,
			Description:  venue.Description,
			Address:      venue.Address,
			Location:     venue.Location,
			Phone:        venue.Phone,
			Email:        venue.Email,
			OpenTime:     venue.OpenTime,
			CloseTime:    venue.CloseTime,
			ImageURLs:    venue.ImageURLs,
			Status:       string(venue.Status),
			Rating:       venue.Rating,
			TotalReviews: venue.TotalReviews,
		}
	}

	return venueResponses, nil
}

func (uc *useCase) AddCourt(ctx context.Context, venueID uuid.UUID, req requests.CreateCourtRequest) (*responses.CourtResponse, error) {

	courts, err := uc.venueRepo.GetCourts(ctx, venueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get courts: %w", err)
	}

	for _, court := range courts {
		if court.Name == req.Name {
			return nil, fmt.Errorf("court name already exists")
		}
	}

	court := &models.Court{
		ID:           uuid.New(),
		VenueID:      venueID,
		Name:         req.Name,
		Description:  req.Description,
		PricePerHour: req.PricePerHour,
		Status:       models.CourtStatusAvailable,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := uc.venueRepo.AddCourt(ctx, court); err != nil {
		return nil, fmt.Errorf("failed to add court: %w", err)
	}

	return &responses.CourtResponse{
		ID:           court.ID.String(),
		Name:         court.Name,
		Description:  court.Description,
		PricePerHour: court.PricePerHour,
		Status:       string(court.Status),
	}, nil
}

func (uc *useCase) UpdateCourt(ctx context.Context, venueID uuid.UUID, req requests.UpdateCourtRequest) error {

	courts, err := uc.venueRepo.GetCourts(ctx, venueID)
	if err != nil {
		return fmt.Errorf("failed to get court: %w", err)
	}
	courtUUID, err := uuid.Parse(req.CourtID)
	if err != nil {
		return fmt.Errorf("invalid court ID: %w", err)
	}

	var court *models.Court
	for i := range courts {

		if courts[i].ID == courtUUID {
			court = &courts[i]
			break
		}
	}

	if court == nil {
		return fmt.Errorf("court not found")
	}

	if req.Name != "" {
		court.Name = req.Name
	}
	if req.Description != "" {
		court.Description = req.Description
	}
	if req.PricePerHour > 0 {
		court.PricePerHour = req.PricePerHour
	}
	if req.Status != "" {
		court.Status = models.CourtStatus(req.Status)
	}

	court.UpdatedAt = time.Now()

	if err := uc.venueRepo.UpdateCourt(ctx, court); err != nil {
		return fmt.Errorf("failed to update court: %w", err)
	}

	return nil
}

func (uc *useCase) DeleteCourt(ctx context.Context, venueID uuid.UUID, courtID uuid.UUID) error {

	courts, err := uc.venueRepo.GetCourts(ctx, venueID)
	if err != nil {
		return fmt.Errorf("failed to get court: %w", err)
	}

	var court *models.Court
	for i := range courts {
		if courts[i].ID == courtID {
			court = &courts[i]
			break
		}
	}

	if court == nil {
		return fmt.Errorf("court not found")
	}

	if err := uc.venueRepo.DeleteCourt(ctx, courtID); err != nil {
		return fmt.Errorf("failed to delete court: %w", err)
	}

	return nil

}

func (uc *useCase) AddReview(ctx context.Context, venueID uuid.UUID, userID uuid.UUID, req requests.AddReviewRequest) error {
	review := &models.VenueReview{
		ID:        uuid.New(),
		VenueID:   venueID,
		UserID:    userID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		CreatedAt: time.Now(),
	}

	fmt.Println("review added before")

	if err := uc.venueRepo.AddReview(ctx, review); err != nil {
		return fmt.Errorf("failed to add review: %w", err)
	}

	fmt.Println("review added")

	return nil
}

func (uc *useCase) GetReviews(ctx context.Context, venueID uuid.UUID, limit, offset int) ([]responses.ReviewResponse, error) {
	reviews, err := uc.venueRepo.GetReviews(ctx, venueID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}

	user, err := uc.userRepo.GetByID(ctx, reviews[0].UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewer: %w", err)
	}

	reviewResponses := make([]responses.ReviewResponse, len(reviews))
	for i, review := range reviews {
		reviewResponses[i] = responses.ReviewResponse{
			ID:        review.ID.String(),
			Rating:    review.Rating,
			Comment:   review.Comment,
			CreatedAt: review.CreatedAt.Format(time.RFC3339),
			Reviewer: responses.ReviewerResponse{
				FirstName: user.FirstName,
				LastName:  user.LastName,
				AvatarURL: user.AvatarURL,
			},
		}
	}

	return reviewResponses, nil
}
