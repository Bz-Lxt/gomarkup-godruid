package metrics

import "time"

type Counts struct {
	Idle         int `json:"idle"`
	InUse        int `json:"in_use"`
	Probing      int `json:"probing"`
	Reconnecting int `json:"reconnecting"`
	Dialing      int `json:"dialing"`
	Connecting   int `json:"connecting"`
	Closing      int `json:"closing"`
	Live         int `json:"live"`
	Waiting      int `json:"waiting"`
}

type Rates struct {
	BorrowRPS     float64 `json:"borrow_rps"`
	ReturnRPS     float64 `json:"return_rps"`
	HitRate       float64 `json:"hit_rate"`
	HitRateSample bool    `json:"hit_rate_sample"`
}

type WaitStats struct {
	AvgMS   float64 `json:"avg_ms"`
	P50MS   float64 `json:"p50_ms"`
	P95MS   float64 `json:"p95_ms"`
	P99MS   float64 `json:"p99_ms"`
	Samples int     `json:"samples"`
}

type ConnView struct {
	ConnectionID string    `json:"connection_id"`
	Generation   uint64    `json:"generation"`
	State        string    `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
	LastBorrowAt time.Time `json:"last_borrow_at,omitempty"`
	LastReturnAt time.Time `json:"last_return_at,omitempty"`
	LastProbeAt  time.Time `json:"last_probe_at,omitempty"`
	BorrowCount  uint64    `json:"borrow_count"`
	LastError    string    `json:"last_error"`
}

type Snapshot struct {
	Seq         uint64     `json:"seq"`
	ServerTime  time.Time  `json:"server_time"`
	PoolID      string     `json:"pool_id"`
	Counts      Counts     `json:"counts"`
	Rates       Rates      `json:"rates"`
	Wait        WaitStats  `json:"wait"`
	Connections []ConnView `json:"connections"`
}

type SeriesPoint struct {
	T         time.Time `json:"t"`
	BorrowRPS float64   `json:"borrow_rps"`
	ReturnRPS float64   `json:"return_rps"`
	HitRate   float64   `json:"hit_rate"`
	Live      int       `json:"live"`
	Waiting   int       `json:"waiting"`
}

func DurationMS(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}
