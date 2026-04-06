package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID           int64      `json:"id"`
	TemplateID   *int64     `json:"template_id,omitempty"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Status       Status     `json:"status"`
	ScheduledFor *time.Time `json:"scheduled_for,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

type TaskTemplate struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Recurrence  Recurrence `json:"recurrence"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Active      bool       `json:"active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Recurrence struct {
	Type          RecurrenceType `json:"type"`
	EveryNDays    int            `json:"every_n_days,omitempty"`
	DayOfMonth    int            `json:"day_of_month,omitempty"`
	MonthParity   MonthParity    `json:"month_parity,omitempty"`
	SpecificDates []time.Time    `json:"specific_dates,omitempty"`
}

type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceMonthParity   RecurrenceType = "month_parity"
)

func (r RecurrenceType) Valid() bool {
	switch r {
	case RecurrenceDaily, RecurrenceMonthly, RecurrenceSpecificDates, RecurrenceMonthParity:
		return true
	default:
		return false
	}
}

type MonthParity string

const (
	MonthParityEven MonthParity = "even"
	MonthParityOdd  MonthParity = "odd"
)

func (p MonthParity) Valid() bool {
	switch p {
	case MonthParityEven, MonthParityOdd:
		return true
	default:
		return false
	}
}
