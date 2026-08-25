export function displayTime(iso: string | undefined): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

export function fmtNum(n: number, digits = 1): string {
  if (!Number.isFinite(n)) return "0";
  return n.toFixed(digits);
}

export function stateLabel(state: string): string {
  switch (state) {
    case "IDLE":
      return "空闲";
    case "IN_USE":
      return "借用";
    case "PROBING":
      return "探测";
    case "RECONNECTING":
      return "重连";
    case "CONNECTING":
      return "连接";
    case "CLOSING":
      return "关闭中";
    case "CLOSED":
      return "已关闭";
    default:
      return state;
  }
}

export function stateMark(state: string): string {
  switch (state) {
    case "IDLE":
      return "●";
    case "IN_USE":
      return "▨";
    case "PROBING":
      return "✚";
    case "RECONNECTING":
      return "↻";
    case "CONNECTING":
      return "◌";
    default:
      return "■";
  }
}
