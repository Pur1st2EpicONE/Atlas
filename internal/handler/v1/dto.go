package v1

import "time"

type RegisterDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CreateEventDTO struct {
	ID          string `json:"event_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Seats       int    `json:"seats"`
	BookingTTL  string `json:"booking_ttl"`
}

type InfoResponseDTO struct {
	ID          string    `json:"event_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Seats       int       `json:"seats"`
}
