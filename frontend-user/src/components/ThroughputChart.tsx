import { useMemo } from "react";
import type { SeriesPoint } from "../types";

export function ThroughputChart({
  points,
  chartWindow,
  onWindow,
}: {
  points: SeriesPoint[];
  chartWindow: string;
  onWindow: (w: string) => void;
}) {
  const path = useMemo(() => {
    if (points.length < 2) return "";
    const max = Math.max(...points.map((p) => Math.max(p.borrow_rps, p.return_rps)), 1);
    return points
      .map((p, i) => {
        const x = (i / (points.length - 1)) * 100;
        const y = 36 - (p.borrow_rps / max) * 32;
        return `${i === 0 ? "M" : "L"}${x},${y}`;
      })
      .join(" ");
  }, [points]);
  const ret = useMemo(() => {
    if (points.length < 2) return "";
    const max = Math.max(...points.map((p) => Math.max(p.borrow_rps, p.return_rps)), 1);
    return points
      .map((p, i) => {
        const x = (i / (points.length - 1)) * 100;
        const y = 36 - (p.return_rps / max) * 32;
        return `${i === 0 ? "M" : "L"}${x},${y}`;
      })
      .join(" ");
  }, [points]);

  return (
    <section className="border border-rail bg-panel/90 p-4" aria-label="吞吐时序">
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <h2 className="font-display text-2xl tracking-wide">THROUGHPUT RAIL</h2>
        <div className="flex gap-2" role="group" aria-label="时间窗口">
          {["1m", "5m", "15m"].map((w) => (
            <button
              key={w}
              type="button"
              onClick={() => onWindow(w)}
              className={`border px-3 py-1 text-sm ${chartWindow === w ? "border-cyan text-cyan" : "border-rail text-fog"}`}
              aria-pressed={chartWindow === w}
            >
              {w}
            </button>
          ))}
        </div>
      </div>
      {points.length < 2 ? (
        <p className="text-fog">采集不足，等待滚动窗口…</p>
      ) : (
        <svg viewBox="0 0 100 40" className="h-40 w-full" role="img" aria-label="借还 RPS 时序图">
          <path d={path} fill="none" stroke="#3fd4e8" strokeWidth="0.6" />
          <path d={ret} fill="none" stroke="#3ee88a" strokeWidth="0.6" />
        </svg>
      )}
      <p className="mt-2 text-xs text-fog">
        <span className="text-cyan">■ 借 RPS</span>
        <span className="ml-4 text-phosphor">■ 还 RPS</span>
      </p>
    </section>
  );
}
