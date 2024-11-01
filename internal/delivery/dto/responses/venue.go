package responses

type CourtResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	PricePerHour float64 `json:"price_per_hour"`
	Status       string  `json:"status"`
}

type VenueResponse struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Address      string          `json:"address"`
	Location     string          `json:"location"`
	Phone        string          `json:"phone"`
	Email        string          `json:"email"`
	OpenTime     string          `json:"open_time"`
	CloseTime    string          `json:"close_time"`
	ImageURLs    string          `json:"image_urls"`
	Status       string          `json:"status"`
	Rating       float64         `json:"rating"`
	TotalReviews int             `json:"total_reviews"`
	Courts       []CourtResponse `json:"courts"`
}

type ReviewResponse struct {
	ID        string           `json:"id"`
	Rating    int              `json:"rating"`
	Comment   string           `json:"comment"`
	CreatedAt string           `json:"created_at"`
	Reviewer  ReviewerResponse `json:"reviewer"`
}

type ReviewerResponse struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url"`
}
