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
    expect(dashboardStyles).toMatch(/\.data-transfer-workspace\s*\{[^}]*overflow-x:\s*hidden/s);
    expect(dashboardStyles).toMatch(/\.data-transfer-table-wrap\s*\{[^}]*overflow:\s*auto/s);
    expect(dashboardStyles).toContain("@media (max-width: 1180px)");
    expect(dashboardStyles).toContain("@media (max-width: 900px)");
  });

  it("keeps truncation inside profile and history table cells", () => {
    const profiles = readFileSync(
      resolve(process.cwd(), "src/features/importExport/TransferProfilesTable.tsx"),
      "utf8"
    );
    const history = readFileSync(
      resolve(process.cwd(), "src/features/importExport/TransferHistoryTable.tsx"),
      "utf8"
    );

    expect(profiles).not.toMatch(/<td[^>]*className="data-transfer-truncate"/);
    expect(history).not.toMatch(/<td[^>]*className="data-transfer-truncate"/);
    expect(profiles).toMatch(/<td[^>]*>\s*<span className="data-transfer-truncate">/);
    expect(history).toMatch(/<td[^>]*>\s*<span className="data-transfer-truncate">/);
  });

  it("keeps the transfer dialog body vertically scrollable", () => {
    const dialogStyles = readFileSync(resolve(process.cwd(), "src/styles/data-transfer-dialogs.css"), "utf8");

    expect(dialogStyles).toMatch(
      /\.transfer-import-dialog\s*\{[^}]*grid-template-rows:\s*auto auto minmax\(0, 1fr\) auto/s
    );
    expect(dialogStyles).toMatch(/\.data-transfer-dialog-body\s*\{[^}]*overflow-y:\s*auto/s);
    expect(dialogStyles).toMatch(/\.data-transfer-dialog-body\s*\{[^}]*scrollbar-gutter:\s*stable/s);
    expect(dialogStyles).toMatch(/\.transfer-review-wrap\s*\{[^}]*overflow:\s*auto/s);
  });

  it("keeps the review table cell in the table formatting context", () => {
    const dialogStyles = readFileSync(resolve(process.cwd(), "src/styles/data-transfer-dialogs.css"), "utf8");

    expect(dialogStyles).not.toMatch(/\.transfer-review-table td:nth-child\(2\)\s*\{[^}]*display:\s*grid/s);
    expect(dialogStyles).toMatch(/\.transfer-review-record\s*\{[^}]*display:\s*grid/s);
  });

  it("keeps editable profile rows visible on focus and dialog actions aligned", () => {
    const dashboardStyles = readFileSync(resolve(process.cwd(), "src/styles/data-transfer-dashboard.css"), "utf8");
    const dialogStyles = readFileSync(resolve(process.cwd(), "src/styles/data-transfer-dialogs.css"), "utf8");

    expect(dashboardStyles).toMatch(/\.transfer-profile-row:focus-visible\s*\{[^}]*outline:/s);
    expect(dialogStyles).toMatch(/\.data-transfer-dialog-actions-main\s*\{[^}]*flex-wrap:\s*nowrap/s);
    expect(dialogStyles).toMatch(
      /@media \(max-width:\s*680px\)[\s\S]*\.data-transfer-dialog-actions-main[^}]*width:\s*100%/
    );
    expect(dialogStyles).toMatch(
      /@media \(max-width:\s*680px\)[\s\S]*\.confirm-layer\.data-transfer-dialog-layer\s*\{[^}]*padding:\s*8px/
    );
  });
});
