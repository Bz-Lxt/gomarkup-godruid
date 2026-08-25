import { fmtNum } from "../lib/format";

export function QueueAlert({ waiting, avgMs }: { waiting: number; avgMs: number }) {
  const hot = waiting > 0;
  return (
    <section
      className={`border px-4 py-3 ${hot ? "alert-flash border-alarm bg-alarm/10" : "border-rail bg-panel/80"}`}
      aria-live="polite"
      aria-label="借还排队告警"
    >
      <p className="font-display tracking-[0.2em] text-fog">QUEUE / HOLDING SIDING</p>
      <div className="mt-1 flex flex-wrap items-end gap-8">
        <div>
          <p className="text-xs text-fog">排队 Goroutine</p>
          <p className={`font-display text-5xl leading-none ${hot ? "text-alarm" : "text-paper"}`}>{waiting}</p>
        </div>
        <div>
          <p className="text-xs text-fog">平均等待</p>
          <p className={`font-display text-5xl leading-none ${hot ? "text-sodium" : "text-paper"}`}>
            {fmtNum(avgMs, 0)}
            <span className="ml-1 text-lg text-fog">ms</span>
          </p>
        </div>
      </div>
    </section>
  );
}
