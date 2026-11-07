import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    environmentOptions: {
      jsdom: {
        url: "http://localhost/"
      }
    },
    setupFiles: ["./src/test/setup.ts"],
    restoreMocks: true,
    clearMocks: true,
    unstubGlobals: true,
    pool: "threads",
    maxWorkers: 2,
    coverage: {
      provider: "v8",
      include: ["src/**/*.{ts,tsx}"],
      exclude: ["src/**/*.test.{ts,tsx}", "src/generated/**", "src/test/**", "src/**/*.d.ts"],
      reporter: ["text", "json", "html"],
      reportsDirectory: "./coverage",
      thresholds: {
        statements: 74,
        branches: 69,
        functions: 68,
        lines: 76
      }
    }
  }
});
