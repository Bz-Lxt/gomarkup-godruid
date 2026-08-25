package control

import (
	"time"

	"godruid/internal/metrics"
	"godruid/internal/timeutil"
)

type PoolMeta struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Connector string `json:"connector"`
	MaxIdle   int    `json:"max_idle"`
	MaxActive int    `json:"max_active"`
}

type PoolList struct {
	Pools []PoolMeta `json:"pools"`
}

type SnapshotDTO struct {
	Seq         uint64              `json:"seq"`
	ServerTime  string              `json:"server_time"`
	PoolID      string              `json:"pool_id"`
	Counts      metrics.Counts      `json:"counts"`
	Rates       metrics.Rates       `json:"rates"`
	Wait        metrics.WaitStats   `json:"wait"`
	Connections []ConnDTO           `json:"connections"`
}

type ConnDTO struct {
	ConnectionID string `json:"connection_id"`
	Generation   uint64 `json:"generation"`
	State        string `json:"state"`
	CreatedAt    string `json:"created_at"`
	LastBorrowAt string `json:"last_borrow_at,omitempty"`
	LastReturnAt string `json:"last_return_at,omitempty"`
	LastProbeAt  string `json:"last_probe_at,omitempty"`
	BorrowCount  uint64 `json:"borrow_count"`
	LastError    string `json:"last_error"`
}

type MetricsDTO struct {
	Window     string          `json:"window"`
	Seq        uint64          `json:"seq"`
	ServerTime string          `json:"server_time"`
	Points     []SeriesPointDTO `json:"points"`
}

type SeriesPointDTO struct {
	T         string  `json:"t"`
	BorrowRPS float64 `json:"borrow_rps"`
	ReturnRPS float64 `json:"return_rps"`
	HitRate   float64 `json:"hit_rate"`
	Live      int     `json:"live"`
	Waiting   int     `json:"waiting"`
}

type WorkloadReq struct {
	Running     *bool `json:"running"`
	Concurrency int   `json:"concurrency"`
	HoldMS      int   `json:"hold_ms"`
	ThinkMS     int   `json:"think_ms"`
}

type FaultReq struct {
	FailPing *bool `json:"fail_ping"`
	FailDial *bool `json:"fail_dial"`
	DropNext *int  `json:"drop_next"`
}

func iso(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return timeutil.FormatISO(t)
}

func SnapshotFrom(s metrics.Snapshot) SnapshotDTO {
	conns := make([]ConnDTO, 0, len(s.Connections))
	for _, c := range s.Connections {
		conns = append(conns, ConnDTO{
			ConnectionID: c.ConnectionID,
			Generation:   c.Generation,
			State:        c.State,
			CreatedAt:    iso(c.CreatedAt),
			LastBorrowAt: iso(c.LastBorrowAt),
			LastReturnAt: iso(c.LastReturnAt),
			LastProbeAt:  iso(c.LastProbeAt),
			BorrowCount:  c.BorrowCount,
			LastError:    c.LastError,
		})
	}
	return SnapshotDTO{
		Seq:         s.Seq,
		ServerTime:  iso(s.ServerTime),
		PoolID:      s.PoolID,
		Counts:      s.Counts,
		Rates:       s.Rates,
		Wait:        s.Wait,
		Connections: conns,
	}
}
