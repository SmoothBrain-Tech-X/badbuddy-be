package models

import "github.com/google/uuid"

type Facility struct {
	ID   uuid.UUID `json:"id"`
	Name string `json:"name"`
}