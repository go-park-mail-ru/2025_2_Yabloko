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

type TicketWithMessages struct {
	Ticket   *Ticket
	Messages []*Message
	Rating   *Rating
}

type TicketFilter struct {
	Status   *string
	Category *string
	Limit    int
	Offset   int
}

type Statistics struct {
	TotalTickets      int
	OpenTickets       int
	InProgressTickets int
	ClosedTickets     int
	TicketsByCategory map[string]int
	AverageRating     float64
}

type UserInfo struct {
	ID    string
	Name  string
	Email string
}
