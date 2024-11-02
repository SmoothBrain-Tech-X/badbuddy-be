package responses

type ChatResponse struct {
	ChatMassage []ChatMassageResponse `json:"chat_massage"`
}

type ChatMassageResponse struct {
	ID            string       `json:"id"`
	ChatID        string       `json:"chat_id"`
	Autor         UserResponse `json:"autor"`
	Message       string       `json:"message"`
	Timestamp     string       `json:"timestamp"`
	EditTimeStamp string       `json:"edit_timestamp"`
}
