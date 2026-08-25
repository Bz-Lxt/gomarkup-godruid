import { useMemo } from "react";
import { stateLabel, stateMark } from "../lib/format";
import type { ConnectionView } from "../types";

function tone(state: string): string {
  switch (state) {
    case "IDLE":
      return "bg-phosphor/20 border-phosphor text-phosphor";
    case "IN_USE":
      return "bg-sodium/20 border-sodium text-sodium";
    case "PROBING":
      return "bg-alarm/20 border-alarm text-alarm";
    case "RECONNECTING":
      return "bg-violet/20 border-violet text-violet";
    case "CONNECTING":
      return "bg-cyan/15 border-cyan text-cyan";
    default:
      return "bg-ash/10 border-ash text-ash";
  }
}

export function ConnectionWall({
  connections,
  onSelect,
}: {
  connections: ConnectionView[];
  onSelect: (c: ConnectionView) => void;
}) {
  const cells = useMemo(() => connections.slice(0, 2000), [connections]);
  if (cells.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center border border-dashed border-rail text-fog" role="status">
        股道空闲 · 尚无逻辑连接
      </div>
    );
  }
  return (
    <div
      className="grid w-full gap-1.5"
      style={{ gridTemplateColumns: "repeat(auto-fill, minmax(72px, 1fr))" }}
      role="list"
      aria-label="连接状态墙"
    >
      {cells.map((c) => (
        <button
          key={c.connection_id}
          type="button"
          role="listitem"
          onClick={() => onSelect(c)}
          className={`min-h-[72px] border px-1.5 py-1 text-left transition hover:brightness-125 ${tone(c.state)} ${c.state === "IN_USE" ? "bg-[repeating-linear-gradient(135deg,transparent,transparent_4px,rgba(245,193,74,0.12)_4px,rgba(245,193,74,0.12)_8px)]" : ""}`}
          aria-label={`${c.connection_id} ${stateLabel(c.state)} 代数${c.generation}`}
        >
          <div className="flex items-center justify-between text-[10px]">
            <span aria-hidden>{stateMark(c.state)}</span>
            <span>g{c.generation}</span>
          </div>
          <div className="font-display text-lg leading-none">{c.connection_id.replace("c-", "")}</div>
          <div className="text-[10px] uppercase tracking-wider">{stateLabel(c.state)}</div>
        </button>
      ))}
    </div>
  );
}
