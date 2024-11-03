package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"badbuddy/internal/delivery/dto/requests"
	"badbuddy/internal/delivery/dto/responses"
	"badbuddy/internal/domain/models"
	"badbuddy/internal/repositories/interfaces"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("unauthorized")

	ErrValidation = errors.New("validation error")

	ErrSessionNotFound = errors.New("session not found")
)

type useCase struct {
	sessionRepo interfaces.SessionRepository
	venueRepo   interfaces.VenueRepository
}

func NewSessionUseCase(sessionRepo interfaces.SessionRepository, venueRepo interfaces.VenueRepository) UseCase {
	return &useCase{
		sessionRepo: sessionRepo,
		venueRepo:   venueRepo,
	}
}

func (uc *useCase) CreateSession(ctx context.Context, hostID uuid.UUID, req requests.CreateSessionRequest) (*responses.SessionResponse, error) {
	// Validate venue exists and is active
	venue, err := uc.venueRepo.GetByID(ctx, uuid.MustParse(req.VenueID))
	if err != nil {
		return nil, fmt.Errorf("invalid venue: %w", err)
	}

	if venue.Status != models.VenueStatusActive {
		return nil, fmt.Errorf("venue is not active")
	}

	// Parse times
	sessionDate, err := time.Parse("2006-01-02", req.SessionDate)
	if err != nil {
		return nil, fmt.Errorf("invalid session date: %w", err)
	}

	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %w", err)
	}

	// Parse and validate court IDs
	openRanges := []responses.OpenRangeResponse{}

	if json.Unmarshal(json.RawMessage(venue.OpenRange.RawMessage), &openRanges) != nil {
		return nil, fmt.Errorf("error decoding enroll response: %v", err)
	}

	// Validate session time including venue operating hours
	for _, openRange := range openRanges {
		venueOpenTime, _ := time.Parse("15:04", openRange.OpenTime.String())
		venueCloseTime, _ := time.Parse("15:04", openRange.CloseTime.String())
		if err := uc.validateSessionTime(sessionDate, startTime, endTime, venueOpenTime, venueCloseTime); err != nil {
			return nil, err
		}
	}

	courtIDs := make([]uuid.UUID, len(req.CourtIDs))
	for i, courtIDStr := range req.CourtIDs {
		courtID, err := uuid.Parse(courtIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid court ID format: %s", courtIDStr)
		}
		courtIDs[i] = courtID

		// Check court conflicts
		if err := uc.checkSessionConflict(ctx, sessionDate, startTime, endTime, courtID); err != nil {
			return nil, err
		}
	}

	// Create session
	session := &models.Session{
		ID:                        uuid.New(),
		HostID:                    hostID,
		VenueID:                   uuid.MustParse(req.VenueID),
		Title:                     req.Title,
		Description:               &req.Description,
		SessionDate:               sessionDate,
		StartTime:                 startTime,
		EndTime:                   endTime,
		PlayerLevel:               models.PlayerLevel(req.PlayerLevel),
		MaxParticipants:           req.MaxParticipants,
		CostPerPerson:             req.CostPerPerson,
		AllowCancellation:         req.AllowCancellation,
		CancellationDeadlineHours: &req.CancellationDeadlineHours,
		Status:                    models.SessionStatusOpen,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
		CourtIDs:                  courtIDs,
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Add host as confirmed participant
	participant := &models.SessionParticipant{
		ID:        uuid.New(),
		SessionID: session.ID,
		UserID:    hostID,
		Status:    models.ParticipantStatusConfirmed,
		JoinedAt:  time.Now(),
	}

	if err := uc.sessionRepo.AddParticipant(ctx, participant); err != nil {
		return nil, fmt.Errorf("failed to add host as participant: %w", err)
	}

	// Get complete session details
	sessionDetail, err := uc.sessionRepo.GetByID(ctx, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session details: %w", err)
	}

	return uc.toSessionResponse(sessionDetail), nil
}

func (uc *useCase) SearchSessions(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) (*responses.SessionListResponse, error) {
	sessions, err := uc.sessionRepo.Search(ctx, query, filters, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search sessions: %w", err)
	}

	sessionResponses := make([]responses.SessionResponse, len(sessions))
	for i, session := range sessions {
		sessionResponses[i] = *uc.toSessionResponse(&session)
	}

	return &responses.SessionListResponse{
		Sessions: sessionResponses,
		Total:    len(sessionResponses),
	}, nil

}

func (uc *useCase) UpdateSession(ctx context.Context, sessionID uuid.UUID, hostID uuid.UUID, req requests.UpdateSessionRequest) error {
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Verify host
	if session.HostID != hostID {
		return fmt.Errorf("only host can update session")
	}

	// Check if session can be updated
	if err := uc.canUpdateSession(session); err != nil {
		return err
	}

	// Update fields if provided
	if req.Title != "" {
		session.Title = req.Title
	}
	if req.Description != "" {
		session.Description = &req.Description
	}
	if req.PlayerLevel != "" {
		if err := uc.validatePlayerLevel(req.PlayerLevel); err != nil {
			return err
		}
		session.PlayerLevel = models.PlayerLevel(req.PlayerLevel)
	}
	if req.MaxParticipants > 0 {
		confirmedCount, _ := uc.countParticipantsByStatus(session.Participants)
		if err := uc.validateParticipantLimit(confirmedCount, req.MaxParticipants); err != nil {
			return err
		}
		session.MaxParticipants = req.MaxParticipants
	}
	if req.CostPerPerson >= 0 {
		session.CostPerPerson = req.CostPerPerson
	}
	if req.Status != "" {
		session.Status = models.SessionStatus(req.Status)
	}

	// Update cancellation settings
	session.AllowCancellation = req.AllowCancellation
	if req.CancellationDeadlineHours > 0 {
		session.CancellationDeadlineHours = &req.CancellationDeadlineHours
	}

	// Update courts if provided
	if len(req.CourtIDs) > 0 {
		courtIDs := make([]uuid.UUID, len(req.CourtIDs))
		for i, courtIDStr := range req.CourtIDs {
			courtID, err := uuid.Parse(courtIDStr)
			if err != nil {
				return fmt.Errorf("invalid court ID format: %s", courtIDStr)
			}
			courtIDs[i] = courtID

			// Check court conflicts
			if err := uc.checkSessionConflict(ctx, session.SessionDate, session.StartTime, session.EndTime, courtID); err != nil {
				return err
			}
		}
		session.CourtIDs = courtIDs
	}

	session.UpdatedAt = time.Now()

	if err := uc.sessionRepo.Update(ctx, &session.Session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// validateParticipantLimit validates the participant limit
func (uc *useCase) validateParticipantLimit(confirmedCount, maxParticipants int) error {
	if confirmedCount > maxParticipants {
		return fmt.Errorf("confirmed participants exceed the maximum allowed")
	}
	return nil
}

func (uc *useCase) JoinSession(ctx context.Context, sessionID, userID uuid.UUID, req requests.JoinSessionRequest) error {
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if err := uc.canJoinSession(session, userID); err != nil {
		return err
	}

	// Check if user is already participating
	participants, err := uc.sessionRepo.GetParticipants(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	if isParticipating, status := uc.isParticipantInSession(participants, userID); isParticipating {
		if status == models.ParticipantStatusCancelled {
			return fmt.Errorf("you have previously cancelled participation in this session")
		}
		return fmt.Errorf("you are already participating in this session")
	}

	confirmedCount, _ := uc.countParticipantsByStatus(participants)
	status := models.ParticipantStatusConfirmed
	if confirmedCount >= session.MaxParticipants {
		status = models.ParticipantStatusPending
	}

	participant := &models.SessionParticipant{
		ID:        uuid.New(),
		SessionID: sessionID,
		UserID:    userID,
		Status:    status,
		JoinedAt:  time.Now(),
	}

	if err := uc.sessionRepo.AddParticipant(ctx, participant); err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	// Update session status if max participants reached
	if status == models.ParticipantStatusConfirmed && confirmedCount+1 >= session.MaxParticipants {
		session.Status = models.SessionStatusFull
		if err := uc.sessionRepo.Update(ctx, &session.Session); err != nil {
			return fmt.Errorf("failed to update session status: %w", err)
		}
	}

	return nil
}

func (uc *useCase) LeaveSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Check if user is host
	if session.HostID == userID {
		return fmt.Errorf("host cannot leave session, use cancel instead")
	}

	// Check cancellation policy
	if !session.AllowCancellation {
		return fmt.Errorf("cancellation is not allowed for this session")
	}

	if session.CancellationDeadlineHours != nil {
		deadline := session.SessionDate.Add(-time.Duration(*session.CancellationDeadlineHours) * time.Hour)
		if time.Now().After(deadline) {
			return fmt.Errorf("cancellation deadline has passed")
		}
	}

	// Check if user is participating
	participants, err := uc.sessionRepo.GetParticipants(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	isParticipating, currentStatus := uc.isParticipantInSession(participants, userID)
	if !isParticipating {
		return fmt.Errorf("user is not participating in this session")
	}

	// Update participant status to cancelled
	if err := uc.sessionRepo.UpdateParticipantStatus(ctx, sessionID, userID, models.ParticipantStatusCancelled); err != nil {
		return fmt.Errorf("failed to update participant status: %w", err)
	}

	// If user was confirmed, try to promote a pending participant
	if currentStatus == models.ParticipantStatusConfirmed {
		for _, p := range participants {
			if p.Status == models.ParticipantStatusPending {
				if err := uc.sessionRepo.UpdateParticipantStatus(ctx, sessionID, p.UserID, models.ParticipantStatusConfirmed); err != nil {
					return fmt.Errorf("failed to promote pending participant: %w", err)
				}
				return nil
			}
		}

		// No pending participants and session was full, update to open
		if session.Status == models.SessionStatusFull {
			session.Status = models.SessionStatusOpen
			if err := uc.sessionRepo.Update(ctx, &session.Session); err != nil {
				return fmt.Errorf("failed to update session status: %w", err)
			}
		}
	}

	return nil
}

func (uc *useCase) CancelSession(ctx context.Context, sessionID, hostID uuid.UUID) error {
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Verify host
	if session.HostID != hostID {
		return fmt.Errorf("only host can cancel session")
	}

	if session.Status == models.SessionStatusCancelled || session.Status == models.SessionStatusCompleted {
		return fmt.Errorf("session is already cancelled or completed")
	}

	// Update session status
	session.Status = models.SessionStatusCancelled
	session.UpdatedAt = time.Now()

	if err := uc.sessionRepo.Update(ctx, &session.Session); err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	// Update all active participants to cancelled
	participants, err := uc.sessionRepo.GetParticipants(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	for _, p := range participants {
		if p.Status != models.ParticipantStatusCancelled {
			if err := uc.sessionRepo.UpdateParticipantStatus(ctx, sessionID, p.UserID, models.ParticipantStatusCancelled); err != nil {
				return fmt.Errorf("failed to update participant status: %w", err)
			}
		}
	}

	return nil
}

func (uc *useCase) GetSession(ctx context.Context, id uuid.UUID) (*responses.SessionResponse, error) {
	session, err := uc.sessionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return uc.toSessionResponse(session), nil
}

func (uc *useCase) ListSessions(ctx context.Context, filters map[string]interface{}, limit, offset int) (*responses.SessionListResponse, error) {
	sessions, err := uc.sessionRepo.List(ctx, filters, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessionResponses := make([]responses.SessionResponse, len(sessions))
	for i, session := range sessions {
		sessionResponses[i] = *uc.toSessionResponse(&session)
	}

	return &responses.SessionListResponse{
		Sessions: sessionResponses,
		Total:    len(sessionResponses),
	}, nil
}

func (uc *useCase) GetUserSessions(ctx context.Context, userID uuid.UUID, includeHistory bool) ([]responses.SessionResponse, error) {
	sessions, err := uc.sessionRepo.GetUserSessions(ctx, userID, includeHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	sessionResponses := make([]responses.SessionResponse, len(sessions))
	for i, session := range sessions {
		sessionResponses[i] = *uc.toSessionResponse(&session)
	}

	return sessionResponses, nil
}

// Helper method to convert model to response
func (uc *useCase) toSessionResponse(session *models.SessionDetail) *responses.SessionResponse {
	participants := make([]responses.ParticipantResponse, len(session.Participants))
	for i, p := range session.Participants {
		participants[i] = responses.ParticipantResponse{
			ID:       p.ID.String(),
			UserID:   p.UserID.String(),
			UserName: p.UserName,
			Status:   string(p.Status),
			JoinedAt: p.JoinedAt.Format(time.RFC3339),
		}
		if p.CancelledAt != nil {
			participants[i].CancelledAt = p.CancelledAt.Format(time.RFC3339)
		}
	}

	confirmedPlayers, pendingPlayers := uc.countParticipantsByStatus(session.Participants)

	description := ""
	if session.Description != nil {
		description = *session.Description
	}

	var cancellationDeadlineHours *int
	if session.CancellationDeadlineHours != nil {
		cancellationDeadlineHours = session.CancellationDeadlineHours
	}

	return &responses.SessionResponse{
		ID:                        session.ID.String(),
		Title:                     session.Title,
		Description:               description,
		VenueName:                 session.VenueName,
		VenueLocation:             session.VenueLocation,
		HostName:                  session.HostName,
		HostLevel:                 string(session.HostLevel),
		SessionDate:               session.SessionDate.Format("2006-01-02"),
		StartTime:                 session.StartTime.Format("15:04"),
		EndTime:                   session.EndTime.Format("15:04"),
		PlayerLevel:               string(session.PlayerLevel),
		MaxParticipants:           session.MaxParticipants,
		CostPerPerson:             session.CostPerPerson,
		Status:                    string(session.Status),
		AllowCancellation:         session.AllowCancellation,
		CancellationDeadlineHours: cancellationDeadlineHours,
		ConfirmedPlayers:          confirmedPlayers,
		PendingPlayers:            pendingPlayers,
		Participants:              participants,
		CreatedAt:                 session.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                 session.UpdatedAt.Format(time.RFC3339),
	}
}

// validateSessionTime validates if the session time is valid including venue hours
func (uc *useCase) validateSessionTime(sessionDate time.Time, startTime, endTime, venueOpen, venueClose time.Time) error {
	now := time.Now()

	// Session date must be in the future
	if sessionDate.Before(now.Truncate(24 * time.Hour)) {
		return fmt.Errorf("session date must be in the future")
	}

	// Session must be at least 30 minutes long
	sessionStartTime := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.Local)
	sessionEndTime := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, time.Local)

	if sessionEndTime.Sub(sessionStartTime) < 30*time.Minute {
		return fmt.Errorf("session must be at least 30 minutes long")
	}

	// Can't create sessions more than 3 months in advance
	if sessionDate.After(now.AddDate(0, 3, 0)) {
		return fmt.Errorf("cannot create sessions more than 3 months in advance")
	}

	// Check if start time is before end time
	if startTime.After(endTime) {
		return fmt.Errorf("start time must be before end time")
	}

	// Check venue operating hours
	if startTime.Before(venueOpen) || endTime.After(venueClose) {
		return fmt.Errorf("session time must be within venue operating hours (%s - %s)",
			venueOpen.Format("15:04"), venueClose.Format("15:04"))
	}

	return nil
}

// checkSessionConflict checks if there's any conflict with existing sessions
func (uc *useCase) checkSessionConflict(ctx context.Context, sessionDate time.Time, startTime, endTime time.Time, courtID uuid.UUID) error {
	filters := map[string]interface{}{
		"date": sessionDate.Format("2006-01-02"),
	}

	existingSessions, err := uc.sessionRepo.List(ctx, filters, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to check session conflicts: %w", err)
	}

	proposedStart := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.Local)
	proposedEnd := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, time.Local)

	for _, session := range existingSessions {
		if session.Status != models.SessionStatusCancelled {
			existingStart := time.Date(session.SessionDate.Year(), session.SessionDate.Month(), session.SessionDate.Day(),
				session.StartTime.Hour(), session.StartTime.Minute(), 0, 0, time.Local)
			existingEnd := time.Date(session.SessionDate.Year(), session.SessionDate.Month(), session.SessionDate.Day(),
				session.EndTime.Hour(), session.EndTime.Minute(), 0, 0, time.Local)

			if proposedStart.Before(existingEnd) && existingStart.Before(proposedEnd) {
				return fmt.Errorf("court is already booked from %s to %s",
					existingStart.Format("15:04"),
					existingEnd.Format("15:04"))
			}
		}
	}

	return nil
}

// countParticipantsByStatus counts participants by their status
func (uc *useCase) countParticipantsByStatus(participants []models.SessionParticipant) (confirmed, pending int) {
	for _, p := range participants {
		switch p.Status {
		case models.ParticipantStatusConfirmed:
			confirmed++
		case models.ParticipantStatusPending:
			pending++
		}
	}
	return
}

// isParticipantInSession checks if a user is already participating
func (uc *useCase) isParticipantInSession(participants []models.SessionParticipant, userID uuid.UUID) (bool, models.ParticipantStatus) {
	for _, p := range participants {
		if p.UserID == userID {
			return true, p.Status
		}
	}
	return false, ""
}

// validatePlayerLevel validates the player level
func (uc *useCase) validatePlayerLevel(level string) error {
	validLevels := map[string]bool{
		string(models.PlayerLevelBeginner):     true,
		string(models.PlayerLevelIntermediate): true,
		string(models.PlayerLevelAdvanced):     true,
	}

	if !validLevels[level] {
		return fmt.Errorf("invalid player level: must be one of beginner, intermediate, or advanced")
	}
	return nil
}

// canUpdateSession checks if a session can be updated
func (uc *useCase) canUpdateSession(session *models.SessionDetail) error {
	if session.Status == models.SessionStatusCancelled {
		return fmt.Errorf("cannot update cancelled session")
	}
	if session.Status == models.SessionStatusCompleted {
		return fmt.Errorf("cannot update completed session")
	}

	sessionDateTime := time.Date(
		session.SessionDate.Year(),
		session.SessionDate.Month(),
		session.SessionDate.Day(),
		session.StartTime.Hour(),
		session.StartTime.Minute(),
		0, 0, time.Local)

	if time.Now().After(sessionDateTime) {
		return fmt.Errorf("cannot update session that has already started")
	}

	return nil
}

// canJoinSession validates if a user can join a session
func (uc *useCase) canJoinSession(session *models.SessionDetail, userID uuid.UUID) error {
	if session.Status != models.SessionStatusOpen && session.Status != models.SessionStatusFull {
		return fmt.Errorf("session is not open for joining")
	}

	sessionDateTime := time.Date(
		session.SessionDate.Year(),
		session.SessionDate.Month(),
		session.SessionDate.Day(),
		session.StartTime.Hour(),
		session.StartTime.Minute(),
		0, 0, time.Local)

	if time.Now().After(sessionDateTime) {
		return fmt.Errorf("cannot join session that has already started")
	}

	return nil
}