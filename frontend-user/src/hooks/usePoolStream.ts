import { useEffect, useRef, useState } from "react";
import { fetchMetrics, fetchSnapshot, openEvents } from "../api";
import type { SeriesPoint, Snapshot, StreamMode } from "../types";

export function usePoolStream(poolId = "default") {
  const [snap, setSnap] = useState<Snapshot | null>(null);
  const [series, setSeries] = useState<SeriesPoint[]>([]);
  const [chartWindow, setChartWindow] = useState("1m");
  const [mode, setMode] = useState<StreamMode>("reconnect");
  const [error, setError] = useState<string>("");
  const lastSeq = useRef(0);
  const failAt = useRef<number | null>(null);

  useEffect(() => {
    let es: EventSource | null = null;
    let poll: number | undefined;
    let cancelled = false;

    const apply = (s: Snapshot) => {
      if (s.seq < lastSeq.current) return;
      if (lastSeq.current && s.seq > lastSeq.current + 1) {
        fetchSnapshot(poolId).then(apply).catch(() => undefined);
      }
      lastSeq.current = s.seq;
      setSnap(s);
      setError("");
      failAt.current = null;
    };

    const startPoll = () => {
      setMode("poll");
      globalThis.clearInterval(poll);
      poll = globalThis.setInterval(() => {
        fetchSnapshot(poolId).then(apply).catch(() => {
          setMode("offline");
          setError("控制面不可达");
        });
      }, 1000);
    };

    const connect = () => {
      setMode("reconnect");
      es = openEvents(
        poolId,
        (s) => {
          setMode("live");
          apply(s);
        },
        () => {
          if (cancelled) return;
          if (!failAt.current) failAt.current = Date.now();
          if (Date.now() - failAt.current >= 3000) {
            es?.close();
            startPoll();
          }
        },
      );
    };

    fetchSnapshot(poolId).then(apply).catch(() => setError("等待首帧"));
    connect();
    return () => {
      cancelled = true;
      es?.close();
      globalThis.clearInterval(poll);
    };
  }, [poolId]);

  useEffect(() => {
    let alive = true;
    const load = () =>
      fetchMetrics(chartWindow)
        .then((pts) => {
          if (alive) setSeries(pts);
        })
        .catch(() => undefined);
    load();
    const id = globalThis.setInterval(load, 2000);
    return () => {
      alive = false;
      globalThis.clearInterval(id);
    };
  }, [chartWindow]);

  return { snap, series, chartWindow, setChartWindow, mode, error };
}
