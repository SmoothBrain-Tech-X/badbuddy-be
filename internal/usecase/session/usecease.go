package session

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
	// Validate venue and court exist
	venue, err := uc.venueRepo.GetByID(ctx, uuid.MustParse(req.VenueID))
	if err != nil {
		return nil, fmt.Errorf("invalid venue: %w", err)
	}

	courtFound := false
	for _, court := range venue.Courts {
		if court.ID.String() == req.CourtID {
			courtFound = true
			break
		}
	}
	if !courtFound {
		return nil, fmt.Errorf("invalid court for venue")
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

	if startTime.After(endTime) {
		return nil, fmt.Errorf("start time must be before end time")
	}

	session := &models.Session{
		ID:            uuid.New(),
		HostID:        hostID,
		VenueID:       uuid.MustParse(req.VenueID),
		CourtID:       uuid.MustParse(req.CourtID),
		Title:         req.Title,
		Description:   req.Description,
		SessionDate:   sessionDate,
		StartTime:     startTime,
		EndTime:       endTime,
		PlayerLevel:   models.PlayerLevel(req.PlayerLevel),
		MinPlayers:    req.MinPlayers,
		MaxPlayers:    req.MaxPlayers,
		CostPerPerson: req.CostPerPerson,
		Status:        models.SessionStatusOpen,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Auto-add host as participant
	participant := &models.SessionParticipant{
		ID:        uuid.New(),
		SessionID: session.ID,
		UserID:    hostID,
		Status:    models.ParticipantStatusJoined,
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

func (uc *useCase) JoinSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if session.Status != models.SessionStatusOpen {
		return fmt.Errorf("session is not open for joining")
	}

	// Check if user is already participating
	participants, err := uc.sessionRepo.GetParticipants(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	for _, p := range participants {
		if p.UserID == userID {
			return fmt.Errorf("user is already participating in this session")
		}
	}

	status := models.ParticipantStatusJoined
	if len(participants) >= session.MaxPlayers {
		status = models.ParticipantStatusWaitlist
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

	// Update session status if max players reached
	if status == models.ParticipantStatusJoined && len(participants)+1 >= session.MaxPlayers {
		session.Status = models.SessionStatusFull
		if err := uc.sessionRepo.Update(ctx, &session.Session); err != nil {
			return fmt.Errorf("failed to update session status: %w", err)
		}
	}

	return nil
}

// Helper method to convert model to response
func (uc *useCase) toSessionResponse(session *models.SessionDetail) *responses.SessionResponse {
	participants := make([]responses.ParticipantResponse, len(session.Participants))
	for i, p := range session.Participants {
		participants[i] = responses.ParticipantResponse{
			ID:     p.ID.String(),
			UserID: p.UserID.String(),
			// UserName: p.UserName,
			Status:   string(p.Status),
			JoinedAt: p.JoinedAt.Format(time.RFC3339),
		}
		if p.CancelledAt != nil {
			participants[i].CancelledAt = p.CancelledAt.Format(time.RFC3339)
		}
	}

	return &responses.SessionResponse{
		ID:             session.ID.String(),
		Title:          session.Title,
		Description:    session.Description,
		VenueName:      session.VenueName,
		VenueLocation:  session.VenueLocation,
		CourtName:      session.CourtName,
		HostName:       session.HostName,
		HostLevel:      session.HostLevel,
		SessionDate:    session.SessionDate.Format("2006-01-02"),
		StartTime:      session.StartTime.Format("15:04"),
		EndTime:        session.EndTime.Format("15:04"),
		PlayerLevel:    string(session.PlayerLevel),
		MinPlayers:     session.MinPlayers,
		MaxPlayers:     session.MaxPlayers,
		CostPerPerson:  session.CostPerPerson,
		Status:         string(session.Status),
		CurrentPlayers: session.CurrentPlayers,
		WaitlistCount:  session.WaitlistCount,
		Participants:   participants,
		CreatedAt:      session.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      session.UpdatedAt.Format(time.RFC3339),
	}
}
func (uc *useCase) GetSession(ctx context.Context, id uuid.UUID) (*responses.SessionResponse, error) {
	session, err := uc.sessionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return uc.toSessionResponse(session), nil
}

// UpdateSession implements updating session details
func (uc *useCase) UpdateSession(ctx context.Context, id uuid.UUID, req requests.UpdateSessionRequest) error {
	session, err := uc.sessionRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Update fields if provided
	if req.Title != "" {
		session.Title = req.Title
	}
	if req.Description != "" {
		session.Description = req.Description
	}
	if req.PlayerLevel != "" {
		session.PlayerLevel = models.PlayerLevel(req.PlayerLevel)
	}
	if req.MinPlayers > 0 {
		if req.MinPlayers > session.CurrentPlayers {
			return fmt.Errorf("min players cannot be greater than current participants")
		}
		session.MinPlayers = req.MinPlayers
	}
	if req.MaxPlayers > 0 {
		if req.MaxPlayers < session.CurrentPlayers {
			return fmt.Errorf("max players cannot be less than current participants")
		}
		session.MaxPlayers = req.MaxPlayers
	}
	if req.CostPerPerson >= 0 {
		session.CostPerPerson = req.CostPerPerson
	}
	if req.Status != "" {
		session.Status = models.SessionStatus(req.Status)
	}

	session.UpdatedAt = time.Now()

	if err := uc.sessionRepo.Update(ctx, &session.Session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// ListSessions implements listing sessions with filters
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

// LeaveSession implements participant leaving a session
func (uc *useCase) LeaveSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Check if user is participating
	participants, err := uc.sessionRepo.GetParticipants(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	var isParticipating bool
	var isHost = session.HostID == userID
	for _, p := range participants {
		if p.UserID == userID {
			isParticipating = true
			break
		}
	}

	if !isParticipating {
		return fmt.Errorf("user is not participating in this session")
	}

	if isHost {
		return fmt.Errorf("host cannot leave the session, consider cancelling instead")
	}

	// Update participant status to cancelled
	if err := uc.sessionRepo.UpdateParticipantStatus(ctx, sessionID, userID, models.ParticipantStatusCancelled); err != nil {
		return fmt.Errorf("failed to update participant status: %w", err)
	}

	// If session was full, check waitlist and promote first waitlisted participant
	if session.Status == models.SessionStatusFull {
		waitlistedParticipants := make([]models.SessionParticipant, 0)
		for _, p := range participants {
			if p.Status == models.ParticipantStatusWaitlist {
				waitlistedParticipants = append(waitlistedParticipants, p)
			}
		}

		if len(waitlistedParticipants) > 0 {
			// Promote first waitlisted participant
			firstWaitlisted := waitlistedParticipants[0]
			if err := uc.sessionRepo.UpdateParticipantStatus(ctx, sessionID, firstWaitlisted.UserID, models.ParticipantStatusJoined); err != nil {
				return fmt.Errorf("failed to promote waitlisted participant: %w", err)
			}
		} else {
			// No waitlisted participants, update session status to open
			session.Status = models.SessionStatusOpen
			if err := uc.sessionRepo.Update(ctx, &session.Session); err != nil {
				return fmt.Errorf("failed to update session status: %w", err)
			}
		}
	}

	return nil
}

// GetUserSessions implements getting user's sessions
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

// Additional helper methods that might be needed

// validateSessionTime validates if the session time is valid
func (uc *useCase) validateSessionTime(sessionDate time.Time, startTime, endTime time.Time) error {
	now := time.Now()

	// Session date must be in the future
	if sessionDate.Before(now.Truncate(24 * time.Hour)) {
		return fmt.Errorf("session date must be in the future")
	}

	// Session must be at least 30 minutes long
	if endTime.Sub(startTime) < 30*time.Minute {
		return fmt.Errorf("session must be at least 30 minutes long")
	}

	// Can't create sessions more than 3 months in advance
	if sessionDate.After(now.AddDate(0, 3, 0)) {
		return fmt.Errorf("cannot create sessions more than 3 months in advance")
	}

	return nil
}

// checkSessionConflict checks if there's any conflict with existing sessions
func (uc *useCase) checkSessionConflict(ctx context.Context, courtID uuid.UUID, sessionDate time.Time, startTime, endTime time.Time) error {
	filters := map[string]interface{}{
		"court_id": courtID,
		"date":     sessionDate,
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
		existingStart := time.Date(session.SessionDate.Year(), session.SessionDate.Month(), session.SessionDate.Day(),
			session.StartTime.Hour(), session.StartTime.Minute(), 0, 0, time.Local)
		existingEnd := time.Date(session.SessionDate.Year(), session.SessionDate.Month(), session.SessionDate.Day(),
			session.EndTime.Hour(), session.EndTime.Minute(), 0, 0, time.Local)

		// Check for overlap
		if proposedStart.Before(existingEnd) && existingStart.Before(proposedEnd) {
			return fmt.Errorf("session time conflicts with existing session")
		}
	}

	return nil
}
