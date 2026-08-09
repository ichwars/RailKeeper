import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const layoutsCss = readFileSync(resolve(process.cwd(), "src/styles/layouts.css"), "utf8");

describe("layout workspace responsive presentation", () => {
  it("keeps panel actions compact and status pills content-sized on narrow screens", () => {
    expect(layoutsCss).toMatch(
      /@media\s*\(max-width:\s*600px\)[\s\S]*?\.layout-panel-head\s*\{[^}]*align-items:\s*center[^}]*flex-wrap:\s*wrap/s
    );
    expect(layoutsCss).toMatch(
      /@media\s*\(max-width:\s*600px\)[\s\S]*?\.layout-workspace-head \.status-pill\s*\{[^}]*align-self:\s*flex-start/s
    );
  });
});
