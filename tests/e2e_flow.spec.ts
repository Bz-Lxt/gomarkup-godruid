import { expect, test } from "@playwright/test";

const base = process.env.GODRUID_BASE ?? "http://127.0.0.1:18080";

test("signal box shows wall, queue, and stream", async ({ page }) => {
  await page.goto(base);
  await expect(page.getByRole("heading", { name: "GODRUID" })).toBeVisible();
  await expect(page.getByLabel("关键指标")).toBeVisible();
  await expect(page.getByLabel("借还排队告警")).toBeVisible();
  await expect(page.getByLabel("连接状态墙").or(page.getByText("尚无逻辑连接"))).toBeVisible();
  await page.getByRole("button", { name: "启动负载" }).click();
  await expect(page.getByText(/负载已启动|YARD CONTROL/)).toBeVisible();
});

test("responsive 768 and 480", async ({ page }) => {
  await page.setViewportSize({ width: 768, height: 900 });
  await page.goto(base);
  await expect(page.getByRole("heading", { name: "GODRUID" })).toBeVisible();
  await page.setViewportSize({ width: 480, height: 800 });
  await expect(page.getByLabel("关键指标")).toBeVisible();
});
