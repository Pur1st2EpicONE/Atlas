package models

import "time"

type User struct {
	ID       int64
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Event struct {
	DBID        int64
	ID          string
	UserID      int64
	Title       string
	Description string
	Date        time.Time
	Seats       int
	BookingTTL  time.Duration
}
