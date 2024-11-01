package postgres

import (
	"context"
	"fmt"
	"strings"

	"badbuddy/internal/domain/models"
	"badbuddy/internal/repositories/interfaces"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type sessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) interfaces.SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (
			id, host_id, venue_id, court_id, title, description,
			session_date, start_time, end_time, player_level,
			min_players, max_players, cost_per_person, status,
			created_at, updated_at
		) VALUES (
			:id, :host_id, :venue_id, :court_id, :title, :description,
			:session_date, :start_time, :end_time, :player_level,
			:min_players, :max_players, :cost_per_person, :status,
			:created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, session)
	return err
}

func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SessionDetail, error) {
	query := `
		SELECT 
			s.*,
			v.name as venue_name,
			v.location as venue_location,
			c.name as court_name,
			u.first_name || ' ' || u.last_name as host_name,
			u.play_level as host_level,
			COUNT(sp.id) FILTER (WHERE sp.status = 'joined') as current_players,
			COUNT(sp.id) FILTER (WHERE sp.status = 'waitlist') as waitlist_count
		FROM sessions s
		JOIN venues v ON v.id = s.venue_id
		JOIN courts c ON c.id = s.court_id
		JOIN users u ON u.id = s.host_id
		LEFT JOIN session_participants sp ON sp.session_id = s.id
		WHERE s.id = $1 AND s.deleted_at IS NULL
		GROUP BY s.id, v.name, v.location, c.name, u.first_name, u.last_name, u.play_level`

	session := &models.SessionDetail{}
	err := r.db.GetContext(ctx, session, query, id)
	if err != nil {
		return nil, err
	}

	// Get participants
	participantsQuery := `
		SELECT sp.*, u.first_name || ' ' || u.last_name as user_name
		FROM session_participants sp
		JOIN users u ON u.id = sp.user_id
		WHERE sp.session_id = $1
		ORDER BY sp.joined_at`

	err = r.db.SelectContext(ctx, &session.Participants, participantsQuery, id)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *sessionRepository) Update(ctx context.Context, session *models.Session) error {
	query := `
		UPDATE sessions SET
			title = :title,
			description = :description,
			session_date = :session_date,
			start_time = :start_time,
			end_time = :end_time,
			player_level = :player_level,
			min_players = :min_players,
			max_players = :max_players,
			cost_per_person = :cost_per_person,
			status = :status,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.db.NamedExecContext(ctx, query, session)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

func (r *sessionRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.SessionDetail, error) {
	conditions := []string{"s.deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	for key, value := range filters {
		switch key {
		case "date":
			conditions = append(conditions, fmt.Sprintf("s.session_date = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "location":
			conditions = append(conditions, fmt.Sprintf("v.location = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "player_level":
			conditions = append(conditions, fmt.Sprintf("s.player_level = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "status":
			conditions = append(conditions, fmt.Sprintf("s.status = $%d", argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	args = append(args, limit, offset)

	query := fmt.Sprintf(`
		SELECT 
			s.*,
			v.name as venue_name,
			v.location as venue_location,
			c.name as court_name,
			u.first_name || ' ' || u.last_name as host_name,
			u.play_level as host_level,
			COUNT(sp.id) FILTER (WHERE sp.status = 'joined') as current_players,
			COUNT(sp.id) FILTER (WHERE sp.status = 'waitlist') as waitlist_count
		FROM sessions s
		JOIN venues v ON v.id = s.venue_id
		JOIN courts c ON c.id = s.court_id
		JOIN users u ON u.id = s.host_id
		LEFT JOIN session_participants sp ON sp.session_id = s.id
		WHERE %s
		GROUP BY s.id, v.name, v.location, c.name, u.first_name, u.last_name, u.play_level
		ORDER BY s.session_date ASC, s.start_time ASC
		LIMIT $%d OFFSET $%d`,
		strings.Join(conditions, " AND "),
		argIndex,
		argIndex+1,
	)

	var sessions []models.SessionDetail
	err := r.db.SelectContext(ctx, &sessions, query, args...)
	return sessions, err
}

func (r *sessionRepository) AddParticipant(ctx context.Context, participant *models.SessionParticipant) error {
	query := `
		INSERT INTO session_participants (
			id, session_id, user_id, status, joined_at
		) VALUES (
			:id, :session_id, :user_id, :status, :joined_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, participant)
	return err
}

func (r *sessionRepository) UpdateParticipantStatus(ctx context.Context, sessionID, userID uuid.UUID, status models.ParticipantStatus) error {
	query := `
		UPDATE session_participants SET
			status = $1,
			cancelled_at = CASE WHEN $1 = 'cancelled' THEN NOW() ELSE cancelled_at END
		WHERE session_id = $2 AND user_id = $3`

	result, err := r.db.ExecContext(ctx, query, status, sessionID, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

func (r *sessionRepository) GetParticipants(ctx context.Context, sessionID uuid.UUID) ([]models.SessionParticipant, error) {
	query := `
		SELECT 
			sp.*,
			u.first_name || ' ' || u.last_name as user_name
		FROM session_participants sp
		JOIN users u ON u.id = sp.user_id
		WHERE sp.session_id = $1
		ORDER BY sp.joined_at`

	var participants []models.SessionParticipant
	err := r.db.SelectContext(ctx, &participants, query, sessionID)
	return participants, err
}

func (r *sessionRepository) GetUserSessions(ctx context.Context, userID uuid.UUID, includeHistory bool) ([]models.SessionDetail, error) {
	conditions := []string{
		"(s.host_id = $1 OR sp.user_id = $1)",
		"s.deleted_at IS NULL",
	}

	if !includeHistory {
		conditions = append(conditions, "s.session_date >= CURRENT_DATE")
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT
			s.*,
			v.name as venue_name,
			v.location as venue_location,
			c.name as court_name,
			u.first_name || ' ' || u.last_name as host_name,
			u.play_level as host_level,
			COUNT(sp2.id) FILTER (WHERE sp2.status = 'joined') as current_players,
			COUNT(sp2.id) FILTER (WHERE sp2.status = 'waitlist') as waitlist_count
		FROM sessions s
		JOIN venues v ON v.id = s.venue_id
		JOIN courts c ON c.id = s.court_id
		JOIN users u ON u.id = s.host_id
		LEFT JOIN session_participants sp ON sp.session_id = s.id
		LEFT JOIN session_participants sp2 ON sp2.session_id = s.id
		WHERE %s
		GROUP BY s.id, v.name, v.location, c.name, u.first_name, u.last_name, u.play_level
		ORDER BY s.session_date DESC, s.start_time DESC`,
		strings.Join(conditions, " AND "),
	)

	var sessions []models.SessionDetail
	err := r.db.SelectContext(ctx, &sessions, query, userID)
	return sessions, err
}
