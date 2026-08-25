import { render, screen } from "@testing-library/react";
import { ConnectionWall } from "./ConnectionWall";
import type { ConnectionView } from "../types";

const sample = (state: string, id: string): ConnectionView => ({
  connection_id: id,
  generation: 1,
  state,
  created_at: "2026-08-25T11:00:00+08:00",
  borrow_count: 1,
  last_error: "",
});

test("renders three color-coded states with accessible names", () => {
  render(
    <ConnectionWall
      connections={[sample("IDLE", "c-0001"), sample("IN_USE", "c-0002"), sample("PROBING", "c-0003")]}
      onSelect={() => undefined}
    />,
  );
  expect(screen.getByLabelText(/c-0001 空闲/)).toBeInTheDocument();
  expect(screen.getByLabelText(/c-0002 借用/)).toBeInTheDocument();
  expect(screen.getByLabelText(/c-0003 探测/)).toBeInTheDocument();
});

test("empty yard", () => {
  render(<ConnectionWall connections={[]} onSelect={() => undefined} />);
  expect(screen.getByRole("status")).toHaveTextContent("尚无逻辑连接");
});
