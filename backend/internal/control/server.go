package control

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"godruid/internal/demo"
	"godruid/internal/metrics"
	"godruid/internal/observe"
	"godruid/internal/pool"
)

type Runtime interface {
	Pool() *pool.Pool
	CurrentSnapshot() metrics.Snapshot
	Ring() *metrics.Ring
	Hub() *observe.Hub
	DemoEnabled() bool
	Workload() *demo.Engine
	ApplyFaults(FaultReq) any
}

type Server struct {
	app  Runtime
	log  *slog.Logger
	http *http.Server
}

func New(addr string, rt Runtime, log *slog.Logger) *Server {
	s := &Server{app: rt, log: log}
	s.http = &http.Server{
		Addr:              addr,
		Handler:           s.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.http.Addr)
	if err != nil {
		return err
	}
	go func() {
		if err := s.http.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.log.Error("http serve", "err", err)
		}
	}()
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func (s *Server) Handler() http.Handler { return s.http.Handler }
func (s *Server) Addr() string          { return s.http.Addr }
