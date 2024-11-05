package responses

type ParticipantResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Status      string `json:"status"`
	JoinedAt    string `json:"joined_at"`
	CancelledAt string `json:"cancelled_at,omitempty"`
}

type SessionRuleResponse struct {
	ID        string `json:"id"`
	RuleText  string `json:"rule_text"`
	CreatedAt string `json:"created_at"`
}

type SessionResponse struct {
	ID                        string                `json:"id"`
	Title                     string                `json:"title"`
	Description               string                `json:"description"`
	VenueName                 string                `json:"venue_name"`
	VenueLocation             string                `json:"venue_location"`
	HostID                    string                `json:"host_id"`
	HostName                  string                `json:"host_name"`
	HostLevel                 string                `json:"host_level"`
	HostGender                string                `json:"host_gender"`
	SessionDate               string                `json:"session_date"`
	StartTime                 string                `json:"start_time"`
	EndTime                   string                `json:"end_time"`
	PlayerLevel               string                `json:"player_level"`
	MaxParticipants           int                   `json:"max_participants"`
	CostPerPerson             float64               `json:"cost_per_person"`
	Status                    string                `json:"status"`
	AllowCancellation         bool                  `json:"allow_cancellation"`
	CancellationDeadlineHours *int                  `json:"cancellation_deadline_hours,omitempty"`
	IsPublic                  bool                  `json:"is_public"`
	ConfirmedPlayers          int                   `json:"confirmed_players"`
	PendingPlayers            int                   `json:"pending_players"`
	Participants              []ParticipantResponse `json:"participants,omitempty"`
	Rules                     []SessionRuleResponse `json:"rules,omitempty"`
	CreatedAt                 string                `json:"created_at"`
	UpdatedAt                 string                `json:"updated_at"`
}

type SessionListResponse struct {
	Sessions []SessionResponse `json:"sessions"`
	Total    int               `json:"total"`
}

// Error responses
type ErrorResponse struct {
	Error       string `json:"error"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
}

// Success responses
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
