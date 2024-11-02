package requests

import "time"

type CreateVenueRequest struct {
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Address     string    `json:"address" validate:"required"`
	Location    string    `json:"location" validate:"required"`
	Phone       string    `json:"phone" validate:"required"`
	Email       string    `json:"email" validate:"required,email"`
	OpenTime    time.Time `json:"open_time" validate:"required"`
	CloseTime   time.Time `json:"close_time" validate:"required"`
	ImageURLs   string    `json:"image_urls"`
}

type UpdateVenueRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Address     string    `json:"address"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	OpenTime    time.Time `json:"open_time"`
	CloseTime   time.Time `json:"close_time"`
	ImageURLs   string    `json:"image_urls"`
	Status      string    `json:"status"`
}

// type CreateCourtRequest struct {
// 	Name         string  `json:"name" validate:"required"`
// 	Description  string  `json:"description"`
// 	PricePerHour float64 `json:"price_per_hour" validate:"required,gt=0"`
// }

// type UpdateCourtRequest struct {
// 	CourtID      string  `json:"court_id" validate:"required,uuid"`
// 	Name         string  `json:"name"`
// 	Description  string  `json:"description"`
// 	PricePerHour float64 `json:"price_per_hour" validate:"gt=0"`
// 	Status       string  `json:"status"`
// }

type AddReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment"`
}
