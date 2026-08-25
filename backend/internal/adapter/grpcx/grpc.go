package grpcx

import (
	"context"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"godruid/internal/connx"
)

// Connector opens a single grpc.ClientConn. Callers must not wrap it with
// a client-side load-balancing pool.
type Connector struct {
	Target string
	Dial   func(ctx context.Context, target string) (*grpc.ClientConn, error)
}

func New(target string) *Connector {
	return &Connector{Target: target}
}

func (c *Connector) Kind() string { return "grpc" }

func (c *Connector) Connect(ctx context.Context) (connx.Connection, error) {
	dial := c.Dial
	if dial == nil {
		dial = func(ctx context.Context, target string) (*grpc.ClientConn, error) {
			return grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		}
	}
	cc, err := dial(ctx, c.Target)
	if err != nil {
		return nil, err
	}
	return &Conn{cc: cc, target: c.Target}, nil
}

type Conn struct {
	cc     *grpc.ClientConn
	target string
	closed atomic.Bool
}

func (c *Conn) Ping(ctx context.Context) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
	}
	_, err := healthpb.NewHealthClient(c.cc).Check(ctx, &healthpb.HealthCheckRequest{})
	return err
}

func (c *Conn) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	return c.cc.Close()
}

func (c *Conn) Metadata() connx.Metadata {
	return connx.Metadata{
		Kind:     "grpc",
		Remote:   connx.SanitizeRemote(c.target),
		Isolated: true,
		Note:     "single ClientConn; no extra grpc pool",
	}
}
