package v1

type RegisterDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string `json:"role"`
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
