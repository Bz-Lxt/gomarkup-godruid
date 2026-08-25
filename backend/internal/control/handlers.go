package control

import (
	"encoding/json"
	"net/http"

	"godruid/internal/errorsx"
	"godruid/internal/metrics"
	"godruid/internal/observe"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /readyz", s.ready)
	mux.HandleFunc("GET /api/v1/pools", s.listPools)
	mux.HandleFunc("GET /api/v1/pools/{id}", s.getPool)
	mux.HandleFunc("GET /api/v1/pools/{id}/snapshot", s.snapshot)
	mux.HandleFunc("GET /api/v1/pools/{id}/metrics", s.series)
	mux.HandleFunc("GET /api/v1/pools/{id}/events", s.events)
	mux.HandleFunc("POST /api/v1/demo/workload", s.workload)
	mux.HandleFunc("POST /api/v1/demo/faults", s.faults)
	return withJSON(mux)
}

func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, _ *http.Request) {
	if s.app == nil || s.app.Pool() == nil || s.app.Pool().Closed() {
		errorsx.Write(w, errorsx.New(errorsx.CodeNotReady, "pool not ready", http.StatusServiceUnavailable))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready", "pool_id": s.app.Pool().ID()})
}

func (s *Server) listPools(w http.ResponseWriter, _ *http.Request) {
	p := s.app.Pool()
	writeJSON(w, http.StatusOK, PoolList{Pools: []PoolMeta{s.meta(p.ID())}})
}

func (s *Server) getPool(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.known(id) {
		errorsx.Write(w, errorsx.New(errorsx.CodeNotFound, "pool not found", http.StatusNotFound))
		return
	}
	writeJSON(w, http.StatusOK, s.meta(id))
}

func (s *Server) snapshot(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.known(id) {
		errorsx.Write(w, errorsx.New(errorsx.CodeNotFound, "pool not found", http.StatusNotFound))
		return
	}
	snap := s.app.CurrentSnapshot()
	writeJSON(w, http.StatusOK, SnapshotFrom(snap))
}

func (s *Server) series(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.known(id) {
		errorsx.Write(w, errorsx.New(errorsx.CodeNotFound, "pool not found", http.StatusNotFound))
		return
	}
	win, ok := metrics.ParseWindow(r.URL.Query().Get("window"))
	if !ok {
		errorsx.Write(w, errorsx.Invalid("window", "window must be 1m, 5m or 15m"))
		return
	}
	snap := s.app.CurrentSnapshot()
	points := s.app.Ring().Window(snap.ServerTime, win)
	dto := make([]SeriesPointDTO, 0, len(points))
	for _, p := range points {
		dto = append(dto, SeriesPointDTO{
			T: iso(p.T), BorrowRPS: p.BorrowRPS, ReturnRPS: p.ReturnRPS,
			HitRate: p.HitRate, Live: p.Live, Waiting: p.Waiting,
		})
	}
	label := r.URL.Query().Get("window")
	if label == "" {
		label = "1m"
	}
	writeJSON(w, http.StatusOK, MetricsDTO{
		Window:     label,
		Seq:        snap.Seq,
		ServerTime: iso(snap.ServerTime),
		Points:     dto,
	})
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.known(id) {
		errorsx.Write(w, errorsx.New(errorsx.CodeNotFound, "pool not found", http.StatusNotFound))
		return
	}
	observe.PrepareSSE(w)
	ch, cancel := s.app.Hub().Subscribe()
	defer cancel()
	_ = observe.WriteSnapshot(w, s.app.CurrentSnapshot())
	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}
	flusher.Flush()
	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case snap, ok := <-ch:
			if !ok {
				return
			}
			if err := observe.WriteSnapshot(w, snap); err != nil {
				return
			}
		}
	}
}

func (s *Server) workload(w http.ResponseWriter, r *http.Request) {
	if !s.app.DemoEnabled() {
		errorsx.Write(w, errorsx.New(errorsx.CodeDemoDisabled, "demo endpoints disabled", http.StatusForbidden))
		return
	}
	var req WorkloadReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorsx.Write(w, errorsx.Invalid("body", "invalid json"))
		return
	}
	if err := req.Validate(); err != nil {
		errorsx.Write(w, err)
		return
	}
	running := true
	if req.Running != nil {
		running = *req.Running
	}
	out := s.app.Workload().Apply(running, req.Concurrency, req.HoldMS, req.ThinkMS)
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) faults(w http.ResponseWriter, r *http.Request) {
	if !s.app.DemoEnabled() {
		errorsx.Write(w, errorsx.New(errorsx.CodeDemoDisabled, "demo endpoints disabled", http.StatusForbidden))
		return
	}
	var req FaultReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorsx.Write(w, errorsx.Invalid("body", "invalid json"))
		return
	}
	if err := req.Validate(); err != nil {
		errorsx.Write(w, err)
		return
	}
	out := s.app.ApplyFaults(req)
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) known(id string) bool {
	return s.app != nil && s.app.Pool() != nil && (id == s.app.Pool().ID() || id == "default")
}

func (s *Server) meta(id string) PoolMeta {
	p := s.app.Pool()
	return PoolMeta{
		ID:        p.ID(),
		Name:      p.Name(),
		Connector: p.ConnectorKind(),
		MaxIdle:   p.Config().MaxIdle,
		MaxActive: p.Config().MaxActive,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
