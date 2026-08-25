package metrics

import (
	"sync/atomic"
	"time"
)

type Registry struct {
	BorrowReq      atomic.Int64
	BorrowOK       atomic.Int64
	BorrowTimeout  atomic.Int64
	BorrowCancel   atomic.Int64
	ReturnOK       atomic.Int64
	ReturnReject   atomic.Int64
	IdleHit        atomic.Int64
	CreateOK       atomic.Int64
	CreateFail     atomic.Int64
	CloseOK        atomic.Int64
	ProbeOK        atomic.Int64
	ProbeFail      atomic.Int64
	ReconnectOK    atomic.Int64
	ReconnectFail  atomic.Int64
	lastBorrowOK   atomic.Int64
	lastReturnOK   atomic.Int64
	lastIdleHit    atomic.Int64
}

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) SnapshotRates(elapsed time.Duration) (borrowRPS, returnRPS, hitRate float64, sampled bool) {
	if r == nil || elapsed <= 0 {
		return 0, 0, 0, false
	}
	sec := elapsed.Seconds()
	b := r.BorrowOK.Load()
	ret := r.ReturnOK.Load()
	hit := r.IdleHit.Load()
	db := b - r.lastBorrowOK.Swap(b)
	dr := ret - r.lastReturnOK.Swap(ret)
	dh := hit - r.lastIdleHit.Swap(hit)
	borrowRPS = float64(db) / sec
	returnRPS = float64(dr) / sec
	if db > 0 {
		hitRate = float64(dh) / float64(db)
		sampled = true
	}
	return borrowRPS, returnRPS, hitRate, sampled
}
