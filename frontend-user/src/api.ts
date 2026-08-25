import type { SeriesPoint, Snapshot } from "./types";

const base = "";

export async function fetchSnapshot(id = "default"): Promise<Snapshot> {
  const res = await fetch(`${base}/api/v1/pools/${id}/snapshot`);
  if (!res.ok) throw new Error(`snapshot ${res.status}`);
  return res.json();
}

export async function fetchMetrics(window: string, id = "default"): Promise<SeriesPoint[]> {
  const res = await fetch(`${base}/api/v1/pools/${id}/metrics?window=${window}`);
  if (!res.ok) throw new Error(`metrics ${res.status}`);
  const body = await res.json();
  return body.points ?? [];
}

export async function setWorkload(body: {
  running: boolean;
  concurrency: number;
  hold_ms: number;
  think_ms: number;
}) {
  const res = await fetch(`${base}/api/v1/demo/workload`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`workload ${res.status}`);
  return res.json();
}

export async function setFaults(body: { fail_ping?: boolean; fail_dial?: boolean; drop_next?: number }) {
  const res = await fetch(`${base}/api/v1/demo/faults`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`faults ${res.status}`);
  return res.json();
}

export function openEvents(id: string, onSnapshot: (s: Snapshot) => void, onError: () => void): EventSource {
  const es = new EventSource(`${base}/api/v1/pools/${id}/events`);
  es.addEventListener("snapshot", (ev) => {
    try {
      onSnapshot(JSON.parse((ev as MessageEvent).data));
    } catch {
      onError();
    }
  });
  es.onerror = () => onError();
  return es;
}
