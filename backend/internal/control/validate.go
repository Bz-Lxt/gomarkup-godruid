package control

import "godruid/internal/errorsx"

func (r WorkloadReq) Validate() error {
	if r.Concurrency < 1 || r.Concurrency > 1000 {
		return errorsx.Invalid("concurrency", "concurrency must be between 1 and 1000")
	}
	if r.HoldMS < 0 || r.HoldMS > 60000 {
		return errorsx.Invalid("hold_ms", "hold_ms must be between 0 and 60000")
	}
	if r.ThinkMS < 0 || r.ThinkMS > 60000 {
		return errorsx.Invalid("think_ms", "think_ms must be between 0 and 60000")
	}
	return nil
}

func (r FaultReq) Validate() error {
	if r.DropNext != nil && (*r.DropNext < 0 || *r.DropNext > 10000) {
		return errorsx.Invalid("drop_next", "drop_next must be between 0 and 10000")
	}
	return nil
}
