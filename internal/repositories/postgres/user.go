// postgres/user_repository.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"badbuddy/internal/domain/models"
	"badbuddy/internal/repositories/interfaces"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrDuplicateEmail = errors.New("email already exists")
	ErrInvalidInput   = errors.New("invalid input")
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) interfaces.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (
            id, email, password, first_name, last_name,
            phone, play_level, location, bio, 
          avatar_url, status, 
            created_at,last_active_at
        ) VALUES (
            :id, :email, :password, :first_name, :last_name,
            :phone, :play_level, :location, :bio,
           :avatar_url, :status,
            :created_at, :last_active_at
        )`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	now := time.Now()
	user.CreatedAt = now
	user.LastActiveAt = now

	if user.Status == "" {
		user.Status = models.UserStatusActive
	}

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return ErrDuplicateEmail
			case "23502": // not_null_violation
				return fmt.Errorf("%w: missing required field", ErrInvalidInput)
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.GetContext(ctx, &user, `
        SELECT * FROM users 
        WHERE id = $1 AND status != $2`,
		id, models.UserStatusInactive)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.GetContext(ctx, &user, `
        SELECT * FROM users 
        WHERE email = $1 AND status != $2`,
		email, models.UserStatusInactive)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET
			first_name = :first_name,
			last_name = :last_name,
			phone = :phone,
			play_level = :play_level,
			location = :location,
			bio = :bio,
			avatar_url = :avatar_url
		WHERE id = :id AND status != 'inactive'`

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error) {
	query := `
        -- First, create a view or use a CTE to calculate regular partners
WITH session_counts AS (
    -- Count sessions played together between pairs of players
    SELECT 
        CASE 
            WHEN sp1.user_id < sp2.user_id THEN sp1.user_id 
            ELSE sp2.user_id 
        END as player1_id,
        CASE 
            WHEN sp1.user_id < sp2.user_id THEN sp2.user_id 
            ELSE sp1.user_id 
        END as player2_id,
        COUNT(DISTINCT sp1.session_id) as sessions_together
    FROM session_participants sp1
    JOIN session_participants sp2 ON sp1.session_id = sp2.session_id 
        AND sp1.user_id != sp2.user_id
    JOIN play_sessions ps ON ps.id = sp1.session_id 
        AND ps.status != 'cancelled'
    GROUP BY 
        CASE 
            WHEN sp1.user_id < sp2.user_id THEN sp1.user_id 
            ELSE sp2.user_id 
        END,
        CASE 
            WHEN sp1.user_id < sp2.user_id THEN sp2.user_id 
            ELSE sp1.user_id 
        END
    HAVING COUNT(DISTINCT sp1.session_id) >= 3
),
user_stats AS (
    SELECT 
        u.*,
        -- Hosted sessions (excluding cancelled)
        COUNT(DISTINCT ps.id) FILTER (
            WHERE ps.host_id = u.id 
            AND ps.status != 'cancelled'
        ) as hosted_sessions,
        
        -- Joined sessions (excluding cancelled)
        COUNT(DISTINCT sp.session_id) FILTER (
            WHERE ps.status != 'cancelled'
            AND ps.host_id != u.id
        ) as joined_sessions,
        
        -- Average rating
        COALESCE(AVG(pr.rating), 0) as avg_rating,
        
        -- Total reviews received
        COUNT(DISTINCT pr.id) as total_reviews,
        
        -- Regular partners (played 3 or more sessions together)
        COALESCE((
            SELECT COUNT(DISTINCT 
                CASE 
                    WHEN sc.player1_id = u.id THEN sc.player2_id 
                    ELSE sc.player1_id 
                END
            )
            FROM session_counts sc
            WHERE sc.player1_id = u.id 
            OR sc.player2_id = u.id
        ), 0) as regular_partners
    FROM users u
    LEFT JOIN play_sessions ps ON ps.host_id = u.id
    LEFT JOIN session_participants sp ON sp.user_id = u.id
    LEFT JOIN player_reviews pr ON pr.reviewed_id = u.id
    WHERE u.id = $1 AND u.status != $2
    GROUP BY u.id
)
SELECT * FROM user_stats;`

	var profile models.UserProfile
	err := r.db.GetContext(ctx, &profile, query, userID, models.UserStatusInactive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return &profile, nil
}

func (r *userRepository) UpdateLastActive(ctx context.Context, userID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
        UPDATE users 
        SET last_active_at = CURRENT_TIMESTAMP 
        WHERE id = $1 AND status != $2`,
		userID, models.UserStatusInactive)

	if err != nil {
		return fmt.Errorf("failed to update last active: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepository) SearchUsers(ctx context.Context, query string, filters interfaces.UserSearchFilters) ([]models.User, error) {
	queryBuilder := `
        SELECT * FROM users
        WHERE status != $1
        AND search_vector @@ plainto_tsquery($2)`

	args := []interface{}{models.UserStatusInactive, query}
	argCount := 3

	if filters.PlayLevel != "" {
		queryBuilder += fmt.Sprintf(" AND play_level = $%d", argCount)
		args = append(args, filters.PlayLevel)
		argCount++
	}

	if filters.Location != "" {
		queryBuilder += fmt.Sprintf(" AND location = $%d", argCount)
		args = append(args, filters.Location)
		argCount++
	}

	// Add ordering
	queryBuilder += `
        ORDER BY 
            CASE WHEN last_active_at > NOW() - INTERVAL '7 days' THEN 1 ELSE 0 END DESC,
            ts_rank(search_vector, plainto_tsquery($2)) DESC,
            created_at DESC
        LIMIT $%d OFFSET $%d`

	args = append(args, filters.Limit, filters.Offset)

	var users []models.User
	err := r.db.SelectContext(ctx, &users, queryBuilder, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}
