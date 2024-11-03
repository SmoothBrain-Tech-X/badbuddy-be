package postgres

import (
	"badbuddy/internal/domain/models"
	"badbuddy/internal/repositories/interfaces"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type chatRepository struct {
	db *sqlx.DB
}

func NewChatRepository(db *sqlx.DB) interfaces.ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) GetChatMessageByID(ctx context.Context, chatID uuid.UUID, limit int, offset int) (*[]models.Message, error) {
	// Get chat
	chat := models.Chat{}

	query := `SELECT * FROM chats WHERE id = $1`

	err := r.db.GetContext(ctx, &chat, query, chatID)
	if err != nil {
		return nil, err
	}

	query = `
		SELECT 
			m.id AS m_id,
			m.chat_id,
			m.sender_id,
			m.type,
			m.content,
			m.created_at,
			m.updated_at,
			u.email,
			u.first_name,
			u.last_name,
			u.phone,
			u.play_level,
			u.avatar_url,
			u.play_level,
			u.gender,
			u.location,
			u.bio,
			u.last_active_at
		FROM 
			chat_messages m
		JOIN 
			users u ON m.sender_id = u.id
		WHERE 
			m.chat_id = $1
			AND m.delete_at IS NULL
		ORDER BY 
			m.created_at DESC
		LIMIT $2
		OFFSET $3`

	// Get messages
	messages := []models.Message{}
	err = r.db.SelectContext(ctx, &messages, query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &messages, nil
}

func (r *chatRepository) GetChatByID(ctx context.Context, chatID uuid.UUID) (*models.Chat, error) {
	chat := models.Chat{}

	query := `SELECT * FROM chats WHERE id = $1`

	err := r.db.GetContext(ctx, &chat, query, chatID)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *chatRepository) IsUserPartOfChat(ctx context.Context, userID, chatID uuid.UUID) (bool, error) {
	var count int

	query := `SELECT COUNT(*) FROM chat_participants WHERE user_id = $1 AND chat_id = $2`

	err := r.db.GetContext(ctx, &count, query, userID, chatID)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *chatRepository) SaveMessage(ctx context.Context, message *models.Message) error {

	query := `INSERT INTO chat_messages (id, chat_id, sender_id, type, content, created_at, updated_at, status) VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), $6)`

	_, err := r.db.ExecContext(ctx, query, message.ID, message.ChatID, message.SenderID, message.Type, message.Content, message.Status)
	if err != nil {
		return err
	}

	return nil
}

func (r *chatRepository) CreateChat(ctx context.Context, chat *models.Chat) error {

	query := `INSERT INTO chats (id, type, session_id) VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, query, chat.ID, chat.Type, chat.SessionID)
	if err != nil {
		return err
	}

	return nil
}

func (r *chatRepository) AddUserToChat(ctx context.Context, userID, chatID uuid.UUID) error {

	query := `INSERT INTO chat_participants (id, chat_id, user_id) VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, query, uuid.New(), chatID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *chatRepository) RemoveUserFromChat(ctx context.Context, userID, chatID uuid.UUID) error {

	query := `DELETE FROM chat_participants WHERE chat_id = $1 AND user_id = $2`

	_, err := r.db.ExecContext(ctx, query, chatID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *chatRepository) UpdateChatMessage(ctx context.Context, message *models.Message) error {

	query := `UPDATE chat_messages SET content = $1, updated_at = NOW() WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, message.Content, message.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *chatRepository) DeleteChatMessage(ctx context.Context, messageID uuid.UUID) error {

	query := `UPDATE chat_messages SET delete_at = NOW(), updated_at = NOW() WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return err
	}

	return nil
}

func (r *chatRepository) UpdateChatMessageReadStatus(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) error {

	query := `UPDATE chat_messages SET status = 'read' WHERE chat_id = $1 AND sender_id != $2 AND status = 'sent'`

	_, err := r.db.ExecContext(ctx, query, chatID, userID)
	if err != nil {
		return err
	}

	return nil
}
