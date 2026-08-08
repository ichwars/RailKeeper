import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api, type AccessoryArticle } from "../../shared/api";
import { ArticlePurchaseDocumentsTab } from "./ArticlePurchaseDocumentsTab";

const article: AccessoryArticle = {
  id: "article-1", manufacturer: "Tillig", name: "Gleis", category: "straight", trackingMode: "quantity",
  manufacturerStatus: "available", articleType: "track", subtype: "straight", gauges: ["TT"], packageQuantity: 1,
  stockUnit: "piece", minimumStock: 0, inventoryStrategy: "quantity", alternativeNumbers: [], keywords: [],
  archived: false, attributes: [], createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T09:00:00Z"
};

describe("ArticlePurchaseDocumentsTab", () => {
  afterEach(() => vi.useRealTimers());

  it("books one purchase with one purchase command and no client-side stock adjustment", async () => {
    const user = userEvent.setup();
    const createPurchase = vi.spyOn(api, "createAccessoryPurchase").mockResolvedValue({} as never);
    const adjustStock = vi.spyOn(api, "adjustAccessoryStock");
    const onChanged = vi.fn().mockResolvedValue(undefined);
    render(<ArticlePurchaseDocumentsTab article={article} disabled={false} onChanged={onChanged} resources={{
      locations: [], stock: null, movements: [], assets: [], purchases: [], documents: [], reservations: [],
      installations: [], usageHistory: null, vehicles: [], layouts: [], units: []
    }} onDirtyChange={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Kauf buchen" }));

    expect(createPurchase).toHaveBeenCalledOnce();
    expect(adjustStock).not.toHaveBeenCalled();
    expect(onChanged).toHaveBeenCalledOnce();
  });

  it("resets a committed purchase before a failed refresh so the original draft cannot be submitted twice", async () => {
    const user = userEvent.setup();
    const createPurchase = vi.spyOn(api, "createAccessoryPurchase").mockResolvedValue({} as never);
    const onChanged = vi.fn().mockRejectedValue(new Error("Aktualisierung fehlgeschlagen"));
    render(<ArticlePurchaseDocumentsTab article={article} disabled={false} onChanged={onChanged} resources={{
      locations: [], stock: null, movements: [], assets: [], purchases: [], documents: [], reservations: [],
      installations: [], usageHistory: null, vehicles: [], layouts: [], units: []
    }} onDirtyChange={vi.fn()} />);
    await user.type(screen.getByRole("textbox", { name: "Bezugsquelle" }), "Modellbahnshop");

    await user.click(screen.getByRole("button", { name: "Kauf buchen" }));

    expect(createPurchase).toHaveBeenCalledOnce();
    expect(onChanged).toHaveBeenCalledOnce();
    expect(screen.getByRole("textbox", { name: "Bezugsquelle" })).toHaveValue("");
  });

  it("uses the local calendar date for a new purchase", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 7, 8, 0, 30));
    render(<ArticlePurchaseDocumentsTab article={article} disabled={false} onChanged={vi.fn()}
      onDirtyChange={vi.fn()} resources={{ locations: [], stock: null, movements: [], assets: [], purchases: [],
        documents: [], reservations: [], installations: [], usageHistory: null, vehicles: [], layouts: [], units: [] }} />);

    expect(screen.getByRole("textbox", { name: "Kaufdatum" })).toHaveValue("08.08.2026");
  });

  it("requires delete confirmation and keeps the document visible after a failed delete", async () => {
    const user = userEvent.setup();
    const remove = vi.spyOn(api, "deleteAccessoryDocument").mockRejectedValue(new Error("Löschen fehlgeschlagen"));
    render(<ArticlePurchaseDocumentsTab article={article} disabled={false} onChanged={vi.fn()}
      onDirtyChange={vi.fn()} resources={{ locations: [], stock: null, movements: [], assets: [], purchases: [],
        documents: [{ id: "doc-1", productId: article.id, originalName: "Rechnung.pdf", fileName: "doc.pdf",
          category: "invoice", mimeType: "application/pdf", sizeBytes: 100, isPrimary: false, createdBy: "admin",
          createdAt: "2026-08-08T09:00:00Z", updatedAt: "2026-08-08T09:00:00Z" }],
        reservations: [], installations: [], usageHistory: null, vehicles: [], layouts: [], units: [] }} />);

    const deleteButton = screen.getByRole("button", { name: "Dokument löschen: Rechnung.pdf" });
    deleteButton.focus();
    await user.click(deleteButton);
    expect(remove).not.toHaveBeenCalled();
    const confirm = screen.getByRole("dialog", { name: "Dokument löschen" });
    expect(confirm).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Abbrechen" })).toHaveFocus();
    await user.click(screen.getByRole("button", { name: "Löschen" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Löschen fehlgeschlagen");
    expect(screen.getByText("Rechnung.pdf")).toBeInTheDocument();
  });
});
