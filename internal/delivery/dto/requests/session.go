package requests

type CreateSessionRequest struct {
	VenueID                   string   `json:"venue_id" validate:"required,uuid"`
	CourtIDs                  []string `json:"court_ids" validate:"required,min=1,dive,uuid"`
	Title                     string   `json:"title" validate:"required"`
	Description               string   `json:"description"`
	SessionDate               string   `json:"session_date" validate:"required,datetime"`
	StartTime                 string   `json:"start_time" validate:"required,datetime"`
	EndTime                   string   `json:"end_time" validate:"required,datetime"`
	PlayerLevel               string   `json:"player_level" validate:"required,oneof=beginner intermediate advanced"`
	MaxParticipants           int      `json:"max_participants" validate:"required,min=2"`
	CostPerPerson             float64  `json:"cost_per_person" validate:"required,min=0"`
	AllowCancellation         bool     `json:"allow_cancellation"`
	CancellationDeadlineHours int      `json:"cancellation_deadline_hours" validate:"required_if=AllowCancellation true,min=0"`
	Rules                     []string `json:"rules" validate:"omitempty,dive,min=1"`
}

type UpdateSessionRequest struct {
	Title                     string   `json:"title"`
	Description               string   `json:"description"`
	CourtIDs                  []string `json:"court_ids" validate:"omitempty,min=1,dive,uuid"`
	PlayerLevel               string   `json:"player_level" validate:"omitempty,oneof=beginner intermediate advanced"`
	MaxParticipants           int      `json:"max_participants" validate:"omitempty,min=2"`
	CostPerPerson             float64  `json:"cost_per_person" validate:"omitempty,min=0"`
	Status                    string   `json:"status" validate:"omitempty,oneof=open full cancelled completed"`
	AllowCancellation         bool     `json:"allow_cancellation"`
	CancellationDeadlineHours int      `json:"cancellation_deadline_hours" validate:"omitempty,min=0"`
	Rules                     []string `json:"rules" validate:"omitempty,dive,min=1"`
}

type JoinSessionRequest struct {
	Message string `json:"message"` // Optional message for the host
}

type AddSessionRuleRequest struct {
	RuleText string `json:"rule_text" validate:"required,min=1"`
}
