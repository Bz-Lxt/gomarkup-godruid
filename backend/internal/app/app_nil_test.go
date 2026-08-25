package app_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"godruid/internal/app"
	"godruid/internal/pool"
)

func TestFaultEndpointWithExternalConnector(t *testing.T) {
	settings := app.Settings{
		Listen:    "127.0.0.1:0",
		Demo:      true,
		Connector: "tcp",
		Target:    "127.0.0.1:1",
		Pool:      pool.DefaultConfig(),
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	instance, err := app.New(settings, logger)
	if err != nil {
		t.Fatalf("create app: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/demo/faults", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("fault endpoint panicked: %v", recovered)
			}
		}()
		instance.Handler().ServeHTTP(rec, req)
	}()

	if rec.Code != http.StatusOK {
		t.Fatalf("fault endpoint status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var state struct {
		FailPing bool `json:"fail_ping"`
		FailDial bool `json:"fail_dial"`
		DropNext int  `json:"drop_next"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&state); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if state.FailPing || state.FailDial || state.DropNext != 0 {
		t.Fatalf("unsupported fault injection returned non-empty state: %+v", state)
	}
}
