package booking

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
	bookingRepo interfaces.BookingRepository
	courtRepo   interfaces.CourtRepository
	venueRepo   interfaces.VenueRepository
}

func NewBookingUseCase(
	bookingRepo interfaces.BookingRepository,
	courtRepo interfaces.CourtRepository,
	venueRepo interfaces.VenueRepository,
) UseCase {
	return &useCase{
		bookingRepo: bookingRepo,
		courtRepo:   courtRepo,
		venueRepo:   venueRepo,
	}
}

func (uc *useCase) CreateBooking(ctx context.Context, userID uuid.UUID, req requests.CreateBookingRequest) (*responses.BookingResponse, error) {
	// Parse and validate court ID
	courtID, err := uuid.Parse(req.CourtID)
	if err != nil {
		return nil, fmt.Errorf("invalid court ID: %w", err)
	}

	// Get court details
	court, err := uc.courtRepo.GetByID(ctx, courtID)
	if err != nil {
		return nil, fmt.Errorf("court not found: %w", err)
	}

	// Validate venue is active
	venue, err := uc.venueRepo.GetByID(ctx, court.VenueID)
	if err != nil {
		return nil, fmt.Errorf("venue not found: %w", err)
	}

	if venue.Status != models.VenueStatusActive {
		return nil, fmt.Errorf("venue is not active")
	}

	// Parse dates and times
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	// Check venue operating hours
	venueOpen, _ := time.Parse("15:04", venue.OpenTime.Format("15:04"))
	venueClose, _ := time.Parse("15:04", venue.CloseTime.Format("15:04"))

	if startTime.Before(venueOpen) || endTime.After(venueClose) {
		return nil, fmt.Errorf("booking time must be within venue operating hours (%s - %s)",
			venue.OpenTime, venue.CloseTime)
	}

	// Check availability
	available, err := uc.bookingRepo.CheckCourtAvailability(ctx, courtID, date, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}
	if !available {
		return nil, fmt.Errorf("court is not available for the selected time slot")
	}

	// Calculate duration and total amount
	duration := endTime.Sub(startTime)
	hours := duration.Hours()
	totalAmount := hours * court.PricePerHour

	// Create booking
	booking := &models.CourtBooking{
		ID:          uuid.New(),
		CourtID:     courtID,
		UserID:      userID,
		Date:        date,
		StartTime:   startTime,
		EndTime:     endTime,
		TotalAmount: totalAmount,
		Status:      models.BookingStatusPending,
		Notes:       req.Notes,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := booking.Validate(); err != nil {
		return nil, fmt.Errorf("invalid booking: %w", err)
	}

	if err := uc.bookingRepo.Create(ctx, booking); err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// Get complete booking details
	bookingDetail, err := uc.bookingRepo.GetByID(ctx, booking.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking details: %w", err)
	}

	return bookingDetail.ToResponse(), nil
}

func (uc *useCase) GetBooking(ctx context.Context, id uuid.UUID) (*responses.BookingResponse, error) {
	booking, err := uc.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	return booking.ToResponse(), nil
}
func (uc *useCase) ListBookings(ctx context.Context, req requests.ListBookingsRequest) (*responses.BookingListResponse, error) {
	filters := make(map[string]interface{})

	if req.CourtID != "" {
		courtID, err := uuid.Parse(req.CourtID)
		if err != nil {
			return nil, fmt.Errorf("invalid court ID: %w", err)
		}
		filters["court_id"] = courtID
	}

	if req.VenueID != "" {
		venueID, err := uuid.Parse(req.VenueID)
		if err != nil {
			return nil, fmt.Errorf("invalid venue ID: %w", err)
		}
		filters["venue_id"] = venueID
	}

	if req.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", req.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format: %w", err)
		}
		filters["date_from"] = dateFrom
	}

	if req.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", req.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format: %w", err)
		}
		filters["date_to"] = dateTo
	}

	if req.Status != "" {
		filters["status"] = models.BookingStatus(req.Status)
	}

	// Set default limit and offset
	limit := 10
	if req.Limit > 0 && req.Limit <= 100 {
		limit = req.Limit
	}

	offset := 0
	if req.Offset > 0 {
		offset = req.Offset
	}

	// Get total count
	total, err := uc.bookingRepo.Count(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get bookings
	bookings, err := uc.bookingRepo.List(ctx, filters, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings: %w", err)
	}

	// Convert to response
	bookingResponses := make([]responses.BookingResponse, len(bookings))
	for i, booking := range bookings {
		bookingResponses[i] = *booking.ToResponse()
	}

	return &responses.BookingListResponse{
		Bookings: bookingResponses,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func (uc *useCase) UpdateBooking(ctx context.Context, id uuid.UUID, req requests.UpdateBookingRequest) (*responses.BookingResponse, error) {
	booking, err := uc.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	if booking.Status == models.BookingStatusCancelled {
		return nil, fmt.Errorf("cannot update cancelled booking")
	}

	if req.Status != "" {
		booking.Status = models.BookingStatus(req.Status)
	}

	if req.Notes != nil {
		booking.Notes = req.Notes
	}

	booking.UpdatedAt = time.Now()

	if err := uc.bookingRepo.Update(ctx, booking); err != nil {
		return nil, fmt.Errorf("failed to update booking: %w", err)
	}

	return booking.ToResponse(), nil
}

func (uc *useCase) CancelBooking(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	booking, err := uc.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}

	if booking.UserID != userID {
		return fmt.Errorf("unauthorized to cancel this booking")
	}

	if !booking.CanBeCancelled() {
		return fmt.Errorf("booking cannot be cancelled")
	}

	if err := uc.bookingRepo.CancelBooking(ctx, id); err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	// Handle payment refund if needed
	if booking.Payment != nil && booking.Payment.Status == models.PaymentStatusCompleted {
		payment := booking.Payment
		payment.Status = models.PaymentStatusRefunded
		payment.UpdatedAt = time.Now()

		if err := uc.bookingRepo.UpdatePayment(ctx, payment); err != nil {
			return fmt.Errorf("failed to update payment status: %w", err)
		}
	}

	return nil
}

func (uc *useCase) GetUserBookings(ctx context.Context, userID uuid.UUID, includeHistory bool) ([]responses.BookingResponse, error) {
	bookings, err := uc.bookingRepo.GetUserBookings(ctx, userID, includeHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bookings: %w", err)
	}

	responses := make([]responses.BookingResponse, len(bookings))
	for i, booking := range bookings {
		responses[i] = *booking.ToResponse()
	}

	return responses, nil
}

func (uc *useCase) CheckAvailability(ctx context.Context, req requests.CheckAvailabilityRequest) (*responses.CourtAvailabilityResponse, error) {
	courtID, err := uuid.Parse(req.CourtID)
	if err != nil {
		return nil, fmt.Errorf("invalid court ID: %w", err)
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	// Get court details
	court, err := uc.courtRepo.GetByID(ctx, courtID)
	if err != nil {
		return nil, fmt.Errorf("court not found: %w", err)
	}

	// Check availability
	available, err := uc.bookingRepo.CheckCourtAvailability(ctx, courtID, date, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}

	// Get existing bookings for the day
	bookings, err := uc.bookingRepo.GetCourtBookings(ctx, courtID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get court bookings: %w", err)
	}

	conflicts := make([]responses.BookingSlot, 0)
	for _, booking := range bookings {
		if booking.Status != models.BookingStatusCancelled {
			conflicts = append(conflicts, responses.BookingSlot{
				StartTime: booking.StartTime.Format("15:04"),
				EndTime:   booking.EndTime.Format("15:04"),
				Status:    string(booking.Status),
			})
		}
	}

	return &responses.CourtAvailabilityResponse{
		CourtID:   courtID.String(),
		CourtName: court.Name,
		Date:      date.Format("2006-01-02"),
		Available: available,
		Conflicts: conflicts,
	}, nil
}

func (uc *useCase) CreatePayment(ctx context.Context, bookingID uuid.UUID, req requests.CreatePaymentRequest) (*responses.PaymentResponse, error) {
	booking, err := uc.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	if booking.Status != models.BookingStatusPending {
		return nil, fmt.Errorf("booking is not in pending state")
	}

	if booking.Payment != nil {
		return nil, fmt.Errorf("payment already exists for this booking")
	}

	if req.Amount != booking.TotalAmount {
		return nil, fmt.Errorf("payment amount does not match booking amount")
	}

	payment := &models.Payment{
		ID:            uuid.New(),
		BookingID:     bookingID,
		Amount:        req.Amount,
		Status:        models.PaymentStatusPending,
		PaymentMethod: models.PaymentMethod(req.PaymentMethod),
		TransactionID: req.TransactionID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := uc.bookingRepo.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Update booking status
	booking.Status = models.BookingStatusConfirmed
	booking.UpdatedAt = time.Now()

	if err := uc.bookingRepo.Update(ctx, booking); err != nil {
		return nil, fmt.Errorf("failed to update booking status: %w", err)
	}

	return &responses.PaymentResponse{
		ID:            payment.ID.String(),
		Amount:        payment.Amount,
		Status:        string(payment.Status),
		PaymentMethod: string(payment.PaymentMethod),
		TransactionID: *payment.TransactionID,
		CreatedAt:     payment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     payment.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// Helper methods

// validateBookingTime validates if the booking time is valid
func (uc *useCase) validateBookingTime(date time.Time, startTime, endTime time.Time, venue *models.Venue) error {
	now := time.Now()

	// Check if date is in the future
	if date.Before(now.Truncate(24 * time.Hour)) {
		return fmt.Errorf("booking date must be in the future")
	}

	// Check if date is not too far in advance (e.g., 3 months)
	if date.After(now.AddDate(0, 3, 0)) {
		return fmt.Errorf("cannot book more than 3 months in advance")
	}

	// Create full datetime for comparison
	bookingStart := time.Date(
		date.Year(), date.Month(), date.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.Local)
	bookingEnd := time.Date(
		date.Year(), date.Month(), date.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, time.Local)

	// Check minimum booking duration (30 minutes)
	if bookingEnd.Sub(bookingStart) < 30*time.Minute {
		return fmt.Errorf("booking duration must be at least 30 minutes")
	}

	// Check maximum booking duration (4 hours)
	if bookingEnd.Sub(bookingStart) > 4*time.Hour {
		return fmt.Errorf("booking duration cannot exceed 4 hours")
	}

	// Check venue operating hours
	venueOpen, _ := time.Parse("15:04", venue.OpenTime.Format("15:04"))
	venueClose, _ := time.Parse("15:04", venue.CloseTime.Format("15:04"))

	if startTime.Before(venueOpen) || endTime.After(venueClose) {
		return fmt.Errorf("booking must be within venue operating hours (%s - %s)",
			venue.OpenTime, venue.CloseTime)
	}

	return nil
}

// calculateBookingAmount calculates the total amount for a booking
func (uc *useCase) calculateBookingAmount(startTime, endTime time.Time, pricePerHour float64) float64 {
	duration := endTime.Sub(startTime)
	hours := duration.Hours()
	return hours * pricePerHour
}

// generateTimeSlots generates available time slots for a given date
func (uc *useCase) generateTimeSlots(ctx context.Context, courtID uuid.UUID, date time.Time, venue *models.Venue) ([]responses.TimeSlot, error) {
	// Parse venue operating hours
	venueOpen, _ := time.Parse("15:04", venue.OpenTime.Format("15:04"))
	venueClose, _ := time.Parse("15:04", venue.CloseTime.Format("15:04"))

	// Get existing bookings for the day
	bookings, err := uc.bookingRepo.GetCourtBookings(ctx, courtID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get court bookings: %w", err)
	}

	// Create map of booked times
	bookedTimes := make(map[string]bool)
	for _, booking := range bookings {
		if booking.Status != models.BookingStatusCancelled {
			for t := booking.StartTime; t.Before(booking.EndTime); t = t.Add(30 * time.Minute) {
				bookedTimes[t.Format("15:04")] = true
			}
		}
	}

	// Generate available time slots
	var slots []responses.TimeSlot
	for t := venueOpen; t.Before(venueClose); t = t.Add(30 * time.Minute) {
		if !bookedTimes[t.Format("15:04")] {
			endTime := t.Add(30 * time.Minute)
			if !endTime.After(venueClose) {
				slots = append(slots, responses.TimeSlot{
					StartTime: t.Format("15:04"),
					EndTime:   endTime.Format("15:04"),
				})
			}
		}
	}

	return slots, nil
}

// handlePaymentStatus updates booking status based on payment status
func (uc *useCase) handlePaymentStatus(ctx context.Context, bookingID uuid.UUID, paymentStatus models.PaymentStatus) error {
	booking, err := uc.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}

	switch paymentStatus {
	case models.PaymentStatusCompleted:
		booking.Status = models.BookingStatusConfirmed
	case models.PaymentStatusFailed:
		booking.Status = models.BookingStatusPending
	case models.PaymentStatusRefunded:
		booking.Status = models.BookingStatusCancelled
		booking.CancelledAt = toPtr(time.Now())
	}

	booking.UpdatedAt = time.Now()
	if err := uc.bookingRepo.Update(ctx, booking); err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}

	return nil
}

// validateRefundEligibility checks if a booking is eligible for refund
func (uc *useCase) validateRefundEligibility(booking *models.CourtBooking) error {
	if booking.Status != models.BookingStatusConfirmed {
		return fmt.Errorf("booking must be confirmed to be eligible for refund")
	}

	if booking.Payment == nil || booking.Payment.Status != models.PaymentStatusCompleted {
		return fmt.Errorf("no completed payment found for booking")
	}

	// Check cancellation deadline (24 hours before start time)
	bookingStart := time.Date(
		booking.Date.Year(), booking.Date.Month(), booking.Date.Day(),
		booking.StartTime.Hour(), booking.StartTime.Minute(), 0, 0, time.Local)

	if time.Now().After(bookingStart.Add(-24 * time.Hour)) {
		return fmt.Errorf("cancellation deadline has passed (24 hours before start time)")
	}

	return nil
}

// processRefund handles the refund process for a cancelled booking
func (uc *useCase) processRefund(ctx context.Context, booking *models.CourtBooking) error {
	if err := uc.validateRefundEligibility(booking); err != nil {
		return fmt.Errorf("refund not eligible: %w", err)
	}

	// Update payment status
	payment := booking.Payment
	payment.Status = models.PaymentStatusRefunded
	payment.UpdatedAt = time.Now()

	if err := uc.bookingRepo.UpdatePayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Update booking status
	booking.Status = models.BookingStatusCancelled
	booking.CancelledAt = toPtr(time.Now())
	booking.UpdatedAt = time.Now()

	if err := uc.bookingRepo.Update(ctx, booking); err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}

	return nil
}

// Helper function to create pointer to time
func toPtr(t time.Time) *time.Time {
	return &t
}

// Additional helper methods

// isBookingConflict checks if two bookings conflict in time
func (uc *useCase) isBookingConflict(booking1, booking2 *models.CourtBooking) bool {
	if booking1.CourtID != booking2.CourtID || !booking1.Date.Equal(booking2.Date) {
		return false
	}

	return booking1.StartTime.Before(booking2.EndTime) && booking2.StartTime.Before(booking1.EndTime)
}

// validateBookingUpdate checks if a booking can be updated
func (uc *useCase) validateBookingUpdate(booking *models.CourtBooking) error {
	if booking.Status == models.BookingStatusCancelled {
		return fmt.Errorf("cannot update cancelled booking")
	}

	if booking.Status == models.BookingStatusCompleted {
		return fmt.Errorf("cannot update completed booking")
	}

	bookingStart := time.Date(
		booking.Date.Year(), booking.Date.Month(), booking.Date.Day(),
		booking.StartTime.Hour(), booking.StartTime.Minute(), 0, 0, time.Local)

	if time.Now().After(bookingStart) {
		return fmt.Errorf("cannot update past or ongoing bookings")
	}

	return nil
}

// validatePaymentCreate validates payment creation request
func (uc *useCase) validatePaymentCreate(booking *models.CourtBooking, amount float64) error {
	if booking.Status != models.BookingStatusPending {
		return fmt.Errorf("booking must be in pending status to add payment")
	}

	if booking.Payment != nil {
		return fmt.Errorf("payment already exists for this booking")
	}

	if amount != booking.TotalAmount {
		return fmt.Errorf("payment amount (%f) does not match booking amount (%f)",
			amount, booking.TotalAmount)
	}

	return nil
}
