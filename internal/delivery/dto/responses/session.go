package responses

type ParticipantResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Status      string `json:"status"`
	JoinedAt    string `json:"joined_at"`
	CancelledAt string `json:"cancelled_at,omitempty"`
}

type SessionResponse struct {
	ID             string                `json:"id"`
	Title          string                `json:"title"`
	Description    string                `json:"description"`
	VenueName      string                `json:"venue_name"`
	VenueLocation  string                `json:"venue_location"`
	CourtName      string                `json:"court_name"`
	HostName       string                `json:"host_name"`
	HostLevel      string                `json:"host_level"`
	SessionDate    string                `json:"session_date"`
	StartTime      string                `json:"start_time"`
	EndTime        string                `json:"end_time"`
	PlayerLevel    string                `json:"player_level"`
	MinPlayers     int                   `json:"min_players"`
	MaxPlayers     int                   `json:"max_players"`
	CostPerPerson  float64               `json:"cost_per_person"`
	Status         string                `json:"status"`
	CurrentPlayers int                   `json:"current_players"`
	WaitlistCount  int                   `json:"waitlist_count"`
	Participants   []ParticipantResponse `json:"participants,omitempty"`
	CreatedAt      string                `json:"created_at"`
	UpdatedAt      string                `json:"updated_at"`
}

type SessionListResponse struct {
	Sessions []SessionResponse `json:"sessions"`
	Total    int               `json:"total"`
}
