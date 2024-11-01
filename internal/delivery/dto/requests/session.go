package requests

type CreateSessionRequest struct {
	VenueID       string  `json:"venue_id" validate:"required,uuid"`
	CourtID       string  `json:"court_id" validate:"required,uuid"`
	Title         string  `json:"title" validate:"required"`
	Description   string  `json:"description"`
	SessionDate   string  `json:"session_date" validate:"required,datetime"`
	StartTime     string  `json:"start_time" validate:"required,datetime"`
	EndTime       string  `json:"end_time" validate:"required,datetime"`
	PlayerLevel   string  `json:"player_level" validate:"required,oneof=beginner intermediate advanced all"`
	MinPlayers    int     `json:"min_players" validate:"required,min=2"`
	MaxPlayers    int     `json:"max_players" validate:"required,gtefield=MinPlayers"`
	CostPerPerson float64 `json:"cost_per_person" validate:"required,min=0"`
}
type UpdateSessionRequest struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	PlayerLevel   string  `json:"player_level" validate:"omitempty,oneof=beginner intermediate advanced all"`
	MinPlayers    int     `json:"min_players" validate:"omitempty,min=2"`
	MaxPlayers    int     `json:"max_players" validate:"omitempty,gtefield=MinPlayers"`
	CostPerPerson float64 `json:"cost_per_person" validate:"omitempty,min=0"`
	Status        string  `json:"status" validate:"omitempty,oneof=open full cancelled completed"`
}
type JoinSessionRequest struct {
	Message string `json:"message"` // Optional message for the host
}
