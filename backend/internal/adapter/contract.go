package adapter

import (
	"context"
	"testing"
	"time"

	"godruid/internal/connx"
)

// CheckConnector is a shared contract used by fake/tcp/redis/mysql/grpc tests.
func CheckConnector(t *testing.T, c connx.Connector) {
	t.Helper()
	if c.Kind() == "" {
		t.Fatal("kind must not be empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := c.Connect(ctx)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := conn.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	md := conn.Metadata()
	if md.Kind == "" {
		t.Fatal("metadata.kind required")
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close must be idempotent: %v", err)
	}
}
