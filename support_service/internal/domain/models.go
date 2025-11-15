package domain

import "time"

type Ticket struct {
	ID          string
	UserID      *string
	GuestID     *string
	UserName    string
	UserEmail   string
	Category    string
	Status      string
	Priority    string
	Title       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Message struct {
	ID        string
	TicketID  string
	UserID    *string
	GuestID   *string
	UserRole  string
	Content   string
	CreatedAt time.Time
}

type Rating struct {
	ID        string
	TicketID  string
	Rating    int
	Comment   string
	CreatedAt time.Time
}

type UserInfo struct {
	ID    string
	Name  string
	Email string
}
