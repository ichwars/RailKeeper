import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const readStyle = (name: string) =>
  readFileSync(resolve(process.cwd(), `src/styles/${name}.css`), "utf8");

describe("CSS ownership after the design audit", () => {
  it("owns the shared switch control in forms instead of settings", () => {
    const formsCss = readStyle("forms-controls");
    const settingsCss = readStyle("settings");

    expect(formsCss).toMatch(/\.switch-field\s*\{[^}]*display:\s*inline-flex/s);
    expect(formsCss.indexOf(".switch-field {")).toBeLessThan(
      formsCss.indexOf(".switch-field input:disabled + span"),
    );
    expect(settingsCss).not.toMatch(/(?:^|\n)\.switch-field(?:\s|>|\{|input|span)/);
    expect(settingsCss).toContain(".smtp-settings-card .smtp-toggle-row .switch-field");
  });

  it("owns the inventory switch alongside the inventory feature", () => {
    const formsCss = readStyle("forms-controls");
    const inventoryCss = readStyle("vehicle-inventory");

    expect(inventoryCss).toMatch(/\.inventory-inline-switch\s*\{/);
    expect(formsCss).not.toContain(".inventory-inline-switch");
  });
});
