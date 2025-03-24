package constants

// SQL constants
const (
	SQLCase = "CASE"
	SQLEnd  = " END"
)

// Order direction
type OrderDirection string

func (t OrderDirection) String() string {
	return string(t)
}

const (
	Asc  OrderDirection = "ASC"
	Desc OrderDirection = "DESC"
)

// Granularity constants
const (
	Daily   = "DAILY"
	Weekly  = "WEEKLY"
	Monthly = "MONTHLY"
	Yearly  = "YEARLY"
)
