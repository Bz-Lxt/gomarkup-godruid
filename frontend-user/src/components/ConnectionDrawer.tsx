import { displayTime, stateLabel } from "../lib/format";
import type { ConnectionView } from "../types";

export function ConnectionDrawer({
  conn,
  onClose,
}: {
  conn: ConnectionView | null;
  onClose: () => void;
}) {
  if (!conn) return null;
  return (
    <div className="fixed inset-0 z-20 flex items-end justify-end bg-black/50 p-4 md:items-center" role="dialog" aria-modal="true" aria-labelledby="conn-title">
      <div className="w-full max-w-md border border-rail bg-panel p-5">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-xs tracking-[0.2em] text-fog">CAR / CONNECTION</p>
            <h2 id="conn-title" className="font-display text-4xl">
              {conn.connection_id}
            </h2>
          </div>
          <button type="button" onClick={onClose} className="border border-rail px-3 py-1 text-fog" aria-label="关闭连接详情">
            关闭
          </button>
        </div>
        <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
          <div>
            <dt className="text-fog">状态</dt>
            <dd>{stateLabel(conn.state)}</dd>
          </div>
          <div>
            <dt className="text-fog">generation</dt>
            <dd>{conn.generation}</dd>
          </div>
          <div>
            <dt className="text-fog">创建</dt>
            <dd>{displayTime(conn.created_at)}</dd>
          </div>
          <div>
            <dt className="text-fog">最近借用</dt>
            <dd>{displayTime(conn.last_borrow_at)}</dd>
          </div>
          <div>
            <dt className="text-fog">最近归还</dt>
            <dd>{displayTime(conn.last_return_at)}</dd>
          </div>
          <div>
            <dt className="text-fog">最近探测</dt>
            <dd>{displayTime(conn.last_probe_at)}</dd>
          </div>
          <div>
            <dt className="text-fog">借用次数</dt>
            <dd>{conn.borrow_count}</dd>
          </div>
          <div className="col-span-2">
            <dt className="text-fog">错误摘要</dt>
            <dd>{conn.last_error || "—"}</dd>
          </div>
        </dl>
      </div>
    </div>
  );
}
