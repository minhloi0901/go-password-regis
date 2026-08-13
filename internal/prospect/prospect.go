package prospect

type Prospect struct {
	ID       string
	Username string
	Email    string
	Status   string
}

const (
	StatusPending   = "pending"
	StatusActive    = "active"
	StatusSuspended = "suspended"
)
