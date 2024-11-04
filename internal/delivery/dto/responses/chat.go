package responses

import "time"

type ChatMassageListResponse struct {
	ChatMassage []ChatMassageResponse `json:"chat_massage"`
}

type ChatMassageResponse struct {
	ID            string           `json:"id"`
	ChatID        string           `json:"chat_id"`
	Autor         UserChatResponse `json:"autor"`
	Message       string           `json:"message"`
	Timestamp     time.Time        `json:"timestamp"`
	EditTimeStamp time.Time        `json:"edit_timestamp"`
}

type BoardCastMessageResponse struct {
	MessageaType string      `json:"message_type"`
	Data         interface{} `json:"data,omitempty"`
}

type ChatListResponse struct {
	Chats []ChatResponse `json:"chats"`
}

type ChatResponse struct {
	ID          string               `json:"id"`
	Type        string               `json:"type"`
	LastMessage *ChatMassageResponse `json:"last_message,omitempty"`
	Users       []UserChatResponse   `json:"users"`
}

type UserChatResponse struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Phone        string    `json:"phone"`
	PlayLevel    string    `json:"play_level"`
	Location     string    `json:"location"`
	Bio          string    `json:"bio"`
	AvatarURL    string    `json:"avatar_url"`
	LastActiveAt time.Time `json:"last_active_at"`
}
