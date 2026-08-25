#!/usr/bin/env python3
"""Offline API smoke against the Compose frontend proxy. Cost ¥0."""
import json
import os
import urllib.request

BASE = os.environ.get("GODRUID_BASE", "http://127.0.0.1:18080")


def get(path: str):
    with urllib.request.urlopen(BASE + path, timeout=5) as r:
        return r.status, json.loads(r.read().decode())


def post(path: str, body: dict):
    req = urllib.request.Request(
        BASE + path,
        data=json.dumps(body).encode(),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=5) as r:
        return r.status, json.loads(r.read().decode())


def main():
    st, health = get("/healthz")
    assert st == 200 and health.get("status") == "ok", health
    st, ready = get("/readyz")
    assert st == 200 and ready.get("status") == "ready", ready
    st, pools = get("/api/v1/pools")
    assert st == 200 and pools["pools"], pools
    st, snap = get("/api/v1/pools/default/snapshot")
    assert st == 200 and "counts" in snap and "seq" in snap, snap
    st, metrics = get("/api/v1/pools/default/metrics?window=1m")
    assert st == 200 and metrics.get("window") == "1m", metrics
    st, wl = post("/api/v1/demo/workload", {"running": True, "concurrency": 8, "hold_ms": 20, "think_ms": 10})
    assert st == 200 and wl.get("concurrency") == 8, wl
    print("API smoke PASS")


if __name__ == "__main__":
    main()
