import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "tests",
  timeout: 30_000,
  use: { baseURL: process.env.GODRUID_BASE ?? "http://127.0.0.1:18080" },
});
