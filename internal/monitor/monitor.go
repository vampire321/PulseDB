package monitor

import "time"

type Status string

const(
	StatusActive Status = "active"
	StatusPaused Status = "paused"
)

type Monitor struct{
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	URL           string     `json:"url"`
	IntervalS     int        `json:"interval_s"`
	Status        Status     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastCheckedAt *time.Time `json:"last_checked_at"`
}

type Check struct {
	ID         string    `json:"id"`
	MonitorID  string    `json:"monitor_id"`
	CheckedAt  time.Time `json:"checked_at"`
	StatusCode int       `json:"status_code"`
	ResponseMs int       `json:"response_ms"`
	Error      *string   `json:"error"`
}