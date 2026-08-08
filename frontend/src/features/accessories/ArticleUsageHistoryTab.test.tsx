import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { setLanguage, type Language } from "../../shared/i18n";
import type { AccessoryArticle, Layout } from "../../shared/api";
import { ArticleUsageHistoryTab } from "./ArticleUsageHistoryTab";

const article: AccessoryArticle = {
  id: "article-1", manufacturer: "Tillig", name: "Signal", category: "signal", trackingMode: "quantity",
  manufacturerStatus: "available", articleType: "signal", subtype: "main_signal", gauges: ["TT"], packageQuantity: 1,
  stockUnit: "piece", minimumStock: 0, inventoryStrategy: "quantity", alternativeNumbers: [], keywords: [],
  archived: false, attributes: [], createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T09:00:00Z"
};
const layout: Layout = { id: "layout-1", name: "Clubanlage", kind: "club", gauge: "TT", scale: "1:120",
  version: 1, archived: false, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T09:00:00Z" };

function renderHistory(language: Language) {
  setLanguage(language);
  return render(<ArticleUsageHistoryTab article={article} resources={{
    locations: [], stock: null, movements: [], assets: [], purchases: [], documents: [], reservations: [],
    installations: [], vehicles: [], layouts: [layout], units: [], usageHistory: { productId: article.id, events: [
      { id: "install", productId: article.id, type: "installation", layoutId: layout.id, quantity: 1,
        condition: "defective", occurredAt: "2026-08-08T12:30:00" },
      { id: "remove", productId: article.id, type: "removal", layoutId: layout.id, quantity: 1,
        removalDisposition: "retired", occurredAt: "2026-08-09T12:30:00" }
    ] }
  }} />);
}

describe("ArticleUsageHistoryTab", () => {
  it("shows target names, condition, removal disposition, and German date formatting", () => {
    renderHistory("de");
    expect(screen.getAllByText("Clubanlage")).toHaveLength(2);
    expect(screen.getByText("Defekt")).toBeInTheDocument();
    expect(screen.getByText("Ausgemustert")).toBeInTheDocument();
    expect(screen.getByText(/08\.08\.26/)).toBeInTheDocument();
    expect(screen.queryByText("layout-1")).not.toBeInTheDocument();
  });

  it("formats dates from the active English language instead of the browser locale", () => {
    renderHistory("en");
    expect(screen.getByText(/08\/08\/2026/)).toBeInTheDocument();
  });
});
