import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

describe("data transfer dashboard styles", () => {
  it("keeps the dashboard in its focused stylesheet", () => {
    const appStyles = readFileSync(resolve(process.cwd(), "src/app/styles.css"), "utf8");
    const importExportStyles = readFileSync(resolve(process.cwd(), "src/styles/import-export.css"), "utf8");
    const dashboardStyles = readFileSync(resolve(process.cwd(), "src/styles/data-transfer-dashboard.css"), "utf8");

    expect(appStyles).toContain('@import "../styles/data-transfer-dashboard.css";');
    expect(importExportStyles).not.toContain(".data-transfer-workspace");
    expect(dashboardStyles).toContain(".data-transfer-workspace");
    expect(dashboardStyles).toContain("grid-template-columns: minmax(280px, 0.72fr)");
  });
});
