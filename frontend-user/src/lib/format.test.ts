import { displayTime, stateLabel } from "./format";

test("formats beijing-ish civil time", () => {
  const s = displayTime("2026-08-25T11:40:00+08:00");
  expect(s).toMatch(/2026-08-25 11:40:00/);
});

test("state labels", () => {
  expect(stateLabel("IDLE")).toBe("空闲");
  expect(stateLabel("PROBING")).toBe("探测");
});
