package control_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"godruid/internal/app"
	"godruid/internal/logx"
	"godruid/internal/pool"
)

func TestMetricsSamplingContinuesAfterStart(t *testing.T) {
	cfg := pool.DefaultConfig()
	cfg.SnapshotInterval = 5 * time.Millisecond
	settings := app.Settings{
		Listen:    "127.0.0.1:0",
		Demo:      false,
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

	pointCount := func() int {
		t.Helper()
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/pools/default/metrics?window=1m", nil)
		inst.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("metrics status=%d body=%s", rr.Code, rr.Body.String())
		}
		var body struct {
			Points []json.RawMessage `json:"points"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode metrics: %v", err)
		}
		return len(body.Points)
	}

	time.Sleep(50 * time.Millisecond)
	before := pointCount()
	deadline := time.Now().Add(150 * time.Millisecond)
	for time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		if after := pointCount(); after > before {
			return
		}
	}
	t.Fatalf("metrics points stopped growing after Start returned: before=%d after=%d", before, pointCount())
}
