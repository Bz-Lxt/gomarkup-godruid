import { useEffect, useState } from "react";
import { ConnectionDrawer } from "./components/ConnectionDrawer";
import { ConnectionWall } from "./components/ConnectionWall";
import { DemoPanel } from "./components/DemoPanel";
import { MetricCards } from "./components/MetricCards";
import { QueueAlert } from "./components/QueueAlert";
import { StreamStatus } from "./components/StreamStatus";
import { ThroughputChart } from "./components/ThroughputChart";
import { usePoolStream } from "./hooks/usePoolStream";
import { displayTime } from "./lib/format";
import type { ConnectionView } from "./types";

export default function App() {
  const { snap, series, chartWindow, setChartWindow, mode, error } = usePoolStream();
  const [sel, setSel] = useState<ConnectionView | null>(null);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setSel(null);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  return (
    <div className="relative min-h-screen w-full px-4 py-5 md:px-6">
      <header className="mb-5 flex w-full flex-wrap items-end justify-between gap-4 border-b border-rail pb-4">
        <div>
          <p className="text-xs tracking-[0.35em] text-cyan">YARD / SIGNAL BOX</p>
          <h1 className="font-display text-6xl leading-none md:text-7xl">GODRUID</h1>
          <p className="mt-1 text-sm text-fog">手写连接池观测 · 空闲磷光 / 借用钠黄 / 探测警报</p>
        </div>
        <div className="text-right">
          <StreamStatus mode={mode} time={snap?.server_time} error={error} />
          <p className="mt-1 text-xs text-fog">池 {snap?.pool_id ?? "—"} · seq {snap?.seq ?? 0}</p>
          <p className="text-xs text-fog">北京时间 {displayTime(snap?.server_time)}</p>
        </div>
      </header>

      <MetricCards snap={snap} />

      <div className="mt-4 grid w-full gap-4 lg:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)]">
        <div className="space-y-4">
          <QueueAlert waiting={snap?.counts.waiting ?? 0} avgMs={snap?.wait.avg_ms ?? 0} />
          <ConnectionWall connections={snap?.connections ?? []} onSelect={setSel} />
        </div>
        <div className="space-y-4">
          <ThroughputChart points={series} chartWindow={chartWindow} onWindow={setChartWindow} />
          <DemoPanel />
          <section className="border border-rail bg-panel/70 p-4 text-xs text-fog" aria-label="状态图例">
            <p>
              <span className="text-phosphor">● 空闲</span>
              <span className="ml-3 text-sodium">▨ 借用</span>
              <span className="ml-3 text-alarm">✚ 探测</span>
              <span className="ml-3 text-violet">↻ 重连</span>
            </p>
          </section>
        </div>
      </div>
      <ConnectionDrawer conn={sel} onClose={() => setSel(null)} />
    </div>
  );
}
