package grpcx_test

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"

	"godruid/internal/adapter"
	"godruid/internal/adapter/grpcx"
)

type hs struct {
	healthpb.UnimplementedHealthServer
}

func (hs) Check(context.Context, *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func TestGRPCContract(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	healthpb.RegisterHealthServer(srv, hs{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	c := grpcx.New("buf")
	c.Dial = func(ctx context.Context, _ string) (*grpc.ClientConn, error) {
		return grpc.DialContext(ctx, "buf",
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	}
	adapter.CheckConnector(t, c)
}
