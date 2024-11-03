package responses

import "time"

type ChatResponse struct {
	ChatMassage []ChatMassageResponse `json:"chat_massage"`
}

type ChatMassageResponse struct {
	ID            string       `json:"id"`
	ChatID        string       `json:"chat_id"`
	Autor         UserResponse `json:"autor"`
	Message       string       `json:"message"`
	Timestamp     time.Time    `json:"timestamp"`
	EditTimeStamp time.Time       `json:"edit_timestamp"`
}

type BoardCastMessageResponse struct {
	MessageaType string `json:"message_type"`
	Data    interface{} `json:"data,omitempty"`
}
