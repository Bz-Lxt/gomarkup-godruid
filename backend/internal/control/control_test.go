package control_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"godruid/internal/adapter/fake"
	"godruid/internal/app"
	"godruid/internal/logx"
	"godruid/internal/pool"
)

func TestHealthAndSnapshot(t *testing.T) {
	cfg := pool.DefaultConfig()
	settings := app.Settings{
		Listen:    "127.0.0.1:0",
		Demo:      true,
		Connector: "fake",
		LogLevel:  "error",
		Pool:      cfg,
	}
	inst, err := app.New(settings, logx.Init("error", nil))
	if err != nil {
		t.Fatal(err)
	}
	_ = inst.Start()
	t.Cleanup(func() {
		_ = inst.Shutdown(t.Context())
	})
	h := inst.Handler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != 200 {
		t.Fatalf("health %d", rr.Code)
	}
	time.Sleep(20 * time.Millisecond)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/pools/default/snapshot", nil))
	if rr.Code != 200 {
		t.Fatalf("snapshot %d %s", rr.Code, rr.Body.String())
	}
	body := map[string]any{"running": false, "concurrency": 2, "hold_ms": 5, "think_ms": 5}
	raw, _ := json.Marshal(body)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/demo/workload", bytes.NewReader(raw)))
	if rr.Code != 200 {
		t.Fatalf("workload %d %s", rr.Code, rr.Body.String())
	}
	_ = fake.New(fake.Options{})
}
