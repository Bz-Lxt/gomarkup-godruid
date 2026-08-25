package observe

import (
	"encoding/json"
	"fmt"
	"net/http"

	"godruid/internal/metrics"
)

func WriteSnapshot(w http.ResponseWriter, s metrics.Snapshot) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported")
	}
	body, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: snapshot\ndata: %s\n\n", body); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func PrepareSSE(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
}
