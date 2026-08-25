package fake_test

import (
	"context"
	"testing"

	"godruid/internal/adapter"
	"godruid/internal/adapter/fake"
)

func TestFakeContract(t *testing.T) {
	adapter.CheckConnector(t, fake.New(fake.Options{}))
}

func TestFakeFaults(t *testing.T) {
	c := fake.New(fake.Options{})
	c.SetFailDial(true)
	if _, err := c.Connect(context.Background()); err == nil {
		t.Fatal("expected dial fail")
	}
	c.SetFailDial(false)
	c.SetFailPing(true)
	conn, err := c.Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Ping(context.Background()); err == nil {
		t.Fatal("expected ping fail")
	}
	_ = conn.Close()
}
