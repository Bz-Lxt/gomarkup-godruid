package app

import (
	"context"
	"log/slog"
	"time"

	"net/http"

	"godruid/internal/adapter/fake"
	grpcx "godruid/internal/adapter/grpcx"
	"godruid/internal/adapter/mysql"
	"godruid/internal/adapter/redis"
	"godruid/internal/adapter/tcp"
	"godruid/internal/connx"
	"godruid/internal/control"
	"godruid/internal/demo"
	"godruid/internal/lifecycle"
	"godruid/internal/metrics"
	"godruid/internal/observe"
	"godruid/internal/pool"
	"godruid/internal/timeutil"
)

type App struct {
	settings Settings
	log      *slog.Logger
	pool     *pool.Pool
	super    *lifecycle.Supervisor
	ring     *metrics.Ring
	hub      *observe.Hub
	work     *demo.Engine
	fake     *fake.Connector
	server   *control.Server
	stop     context.CancelFunc
}

func New(settings Settings, log *slog.Logger) (*App, error) {
	connector, fakeConn, err := buildConnector(settings)
	if err != nil {
		return nil, err
	}
	p, err := pool.New(connector, settings.Pool, pool.WithLogger(log), pool.WithID("default"), pool.WithName("demo-"+connector.Kind()))
	if err != nil {
		return nil, err
	}
	a := &App{
		settings: settings,
		log:      log,
		pool:     p,
		super:    lifecycle.NewSupervisor(p, log),
		ring:     metrics.NewRing(settings.Pool.MetricsRetention, settings.Pool.SnapshotInterval),
		hub:      observe.NewHub(),
		work:     demo.NewEngine(p),
		fake:     fakeConn,
	}
	a.server = control.New(settings.Listen, a, log)
	return a, nil
}

func (a *App) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.stop = cancel
	a.super.Start()
	go a.sampleLoop(ctx)
	if err := a.server.Start(); err != nil {
		cancel()
		return err
	}
	if a.settings.Demo {
		a.work.Start()
	}
	a.log.Info("godruid started", "listen", a.settings.Listen, "connector", a.settings.Connector, "demo", a.settings.Demo)
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.stop != nil {
		a.stop()
	}
	a.work.Stop()
	a.super.Stop()
	_ = a.pool.Close()
	return a.server.Shutdown(ctx)
}

func (a *App) Pool() *pool.Pool               { return a.pool }
func (a *App) Ring() *metrics.Ring            { return a.ring }
func (a *App) Hub() *observe.Hub              { return a.hub }
func (a *App) DemoEnabled() bool              { return a.settings.Demo }
func (a *App) Workload() *demo.Engine         { return a.work }
func (a *App) Handler() http.Handler { return a.server.Handler() }

func (a *App) ApplyFaults(req control.FaultReq) any {
	if a.fake == nil {
		return demo.FaultState{}
	}
	return demo.ApplyFaults(a.fake, req.FailPing, req.FailDial, req.DropNext)
}

func (a *App) CurrentSnapshot() metrics.Snapshot {
	if s, ok := a.ring.Latest(); ok {
		return s
	}
	return a.capture(a.ring.NextSeq())
}

func (a *App) sampleLoop(ctx context.Context) {
	t := time.NewTicker(a.settings.Pool.SnapshotInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s := a.capture(a.ring.NextSeq())
			a.ring.Append(s)
			a.hub.Publish(s)
		}
	}
}

func (a *App) capture(seq uint64) metrics.Snapshot {
	now := timeutil.Now()
	borrow, ret, hit, sampled := a.pool.Metrics().SnapshotRates(a.settings.Pool.SnapshotInterval)
	avg, p50, p95, p99, n := a.pool.Waits().Stats(a.pool.Clock().Now())
	return metrics.Snapshot{
		Seq:        seq,
		ServerTime: now,
		PoolID:     a.pool.ID(),
		Counts:     a.pool.Counts(),
		Rates: metrics.Rates{
			BorrowRPS:     borrow,
			ReturnRPS:     ret,
			HitRate:       hit,
			HitRateSample: sampled,
		},
		Wait: metrics.WaitStats{
			AvgMS:   metrics.DurationMS(avg),
			P50MS:   metrics.DurationMS(p50),
			P95MS:   metrics.DurationMS(p95),
			P99MS:   metrics.DurationMS(p99),
			Samples: n,
		},
		Connections: a.pool.Views(),
	}
}

func buildConnector(s Settings) (connx.Connector, *fake.Connector, error) {
	switch s.Connector {
	case "tcp":
		return tcp.New(s.Target), nil, nil
	case "redis":
		return redis.New(s.Target), nil, nil
	case "mysql":
		return mysql.New(s.Target), nil, nil
	case "grpc":
		return grpcx.New(s.Target), nil, nil
	default:
		f := fake.New(fake.Options{})
		return f, f, nil
	}
}
