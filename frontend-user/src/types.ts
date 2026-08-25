export type ConnState =
  | "CONNECTING"
  | "IDLE"
  | "IN_USE"
  | "PROBING"
  | "RECONNECTING"
  | "CLOSING"
  | "CLOSED";

export interface ConnectionView {
  connection_id: string;
  generation: number;
  state: ConnState | string;
  created_at: string;
  last_borrow_at?: string;
  last_return_at?: string;
  last_probe_at?: string;
  borrow_count: number;
  last_error: string;
}

export interface Snapshot {
  seq: number;
  server_time: string;
  pool_id: string;
  counts: {
    idle: number;
    in_use: number;
    probing: number;
    reconnecting: number;
    dialing: number;
    connecting: number;
    closing: number;
    live: number;
    waiting: number;
  };
  rates: {
    borrow_rps: number;
    return_rps: number;
    hit_rate: number;
    hit_rate_sample: boolean;
  };
  wait: {
    avg_ms: number;
    p50_ms: number;
    p95_ms: number;
    p99_ms: number;
    samples: number;
  };
  connections: ConnectionView[];
}

export interface SeriesPoint {
  t: string;
  borrow_rps: number;
  return_rps: number;
  hit_rate: number;
  live: number;
  waiting: number;
}

export type StreamMode = "live" | "reconnect" | "poll" | "offline";
