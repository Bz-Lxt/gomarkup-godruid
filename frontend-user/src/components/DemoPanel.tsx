import { useState } from "react";
import { setFaults, setWorkload } from "../api";

export function DemoPanel() {
  const [msg, setMsg] = useState("");
  const [conc, setConc] = useState(32);
  const apply = async (running: boolean) => {
    try {
      await setWorkload({ running, concurrency: conc, hold_ms: 40, think_ms: 12 });
      setMsg(running ? "负载已启动" : "负载已暂停");
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "操作失败");
    }
  };
  const fault = async (fail_ping: boolean) => {
    try {
      await setFaults({ fail_ping, drop_next: fail_ping ? 3 : 0 });
      setMsg(fail_ping ? "已注入探测失败" : "故障已清除");
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "操作失败");
    }
  };
  return (
    <section className="border border-rail bg-panel/80 p-4" aria-label="演示控制">
      <h2 className="font-display text-2xl">YARD CONTROL</h2>
      <label className="mt-3 block text-xs text-fog">
        并发
        <input
          className="ml-2 w-24 border border-rail bg-void px-2 py-1 text-paper"
          type="number"
          min={1}
          max={1000}
          value={conc}
          onChange={(e) => setConc(Number(e.target.value))}
        />
      </label>
      <div className="mt-3 flex flex-wrap gap-2">
        <button type="button" className="border border-cyan px-3 py-1 text-cyan" onClick={() => apply(true)}>
          启动负载
        </button>
        <button type="button" className="border border-rail px-3 py-1 text-fog" onClick={() => apply(false)}>
          暂停流量
        </button>
        <button type="button" className="border border-alarm px-3 py-1 text-alarm" onClick={() => fault(true)}>
          注入断连
        </button>
        <button type="button" className="border border-phosphor px-3 py-1 text-phosphor" onClick={() => fault(false)}>
          恢复探测
        </button>
      </div>
      {msg ? (
        <p className="mt-2 text-xs text-sodium" role="status">
          {msg}
        </p>
      ) : null}
    </section>
  );
}
