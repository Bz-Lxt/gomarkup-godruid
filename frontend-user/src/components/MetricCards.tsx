import { fmtNum } from "../lib/format";
import type { Snapshot } from "../types";

export function MetricCards({ snap }: { snap: Snapshot | null }) {
  const items = [
    { k: "BORROW RPS", v: fmtNum(snap?.rates.borrow_rps ?? 0), c: "text-cyan" },
    { k: "RETURN RPS", v: fmtNum(snap?.rates.return_rps ?? 0), c: "text-phosphor" },
    {
      k: "HIT RATE",
      v: snap?.rates.hit_rate_sample ? `${fmtNum((snap?.rates.hit_rate ?? 0) * 100, 0)}%` : "0 · 无样本",
      c: "text-sodium",
    },
    { k: "LIVE", v: String(snap?.counts.live ?? 0), c: "text-paper" },
    { k: "WAITING", v: String(snap?.counts.waiting ?? 0), c: snap && snap.counts.waiting > 0 ? "text-alarm" : "text-fog" },
  ];
  return (
    <section className="grid w-full grid-cols-2 gap-3 md:grid-cols-5" aria-label="关键指标">
      {items.map((it) => (
        <article key={it.k} className="border border-rail bg-panel/90 px-4 py-3">
          <p className="font-display text-sm tracking-[0.2em] text-fog">{it.k}</p>
          <p className={`font-display text-4xl leading-none ${it.c}`}>{it.v}</p>
        </article>
      ))}
    </section>
  );
}
