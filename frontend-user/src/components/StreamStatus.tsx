import { displayTime } from "../lib/format";
import type { StreamMode } from "../types";

const label: Record<StreamMode, string> = {
  live: "实时 SSE",
  reconnect: "重连中",
  poll: "快照轮询",
  offline: "离线",
};

export function StreamStatus({ mode, time, error }: { mode: StreamMode; time?: string; error?: string }) {
  const color =
    mode === "live" ? "text-phosphor" : mode === "offline" ? "text-alarm" : mode === "poll" ? "text-sodium" : "text-cyan";
  return (
    <p className="text-xs text-fog" role="status" aria-live="polite">
      <span className={color}>● {label[mode]}</span>
      <span className="ml-3">最后更新 {displayTime(time)}</span>
      {error ? <span className="ml-3 text-alarm">{error}</span> : null}
    </p>
  );
}
