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

  it("maps the Tillig fixture quantity and stock location into the booked purchase API input", async () => {
    const user = userEvent.setup();
    const createPurchase = vi.spyOn(api, "createAccessoryPurchase").mockResolvedValue({} as never);
    render(<ArticlePurchaseDocumentsTab article={{ ...article, articleNumber: "83101", attributes: [
      { key: "trackSystem", kind: "text", textValue: "Tillig TT Modellgleis" },
      { key: "lengthMm", kind: "number", numberValue: 166, unit: "mm" },
      { key: "connectionCount", kind: "number", numberValue: 2 }
    ] }} disabled={false} onChanged={vi.fn()} resources={{
      locations: [{ id: "location-track-shelf", name: "Gleislager", parentId: undefined, archived: false,
        createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z" }],
      stock: null, movements: [], assets: [], purchases: [], documents: [], reservations: [],
      installations: [], usageHistory: null, vehicles: [], layouts: [], units: []
    }} onDirtyChange={vi.fn()} />);

    const quantity = screen.getByRole("spinbutton", { name: "Menge" });
    await user.clear(quantity);
    await user.type(quantity, "12");
    await user.click(screen.getByRole("checkbox", { name: "Kauf bestandswirksam buchen" }));
    await user.click(screen.getByRole("button", { name: "Kauf buchen" }));

    expect(createPurchase).toHaveBeenCalledWith("article-1", expect.objectContaining({
      quantity: 12,
      storageLocationId: "location-track-shelf",
      bookToStock: true
    }));
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

  it("marks the first uploaded product image primary and preserves its draft after failure", async () => {
    const user = userEvent.setup();
    const upload = vi.spyOn(api, "uploadAccessoryDocument").mockRejectedValue(new Error("Upload fehlgeschlagen"));
    render(<ArticlePurchaseDocumentsTab article={article} disabled={false} onChanged={vi.fn()}
      onDirtyChange={vi.fn()} resources={{ locations: [], stock: null, movements: [], assets: [], purchases: [],
        documents: [], reservations: [], installations: [], usageHistory: null, vehicles: [], layouts: [], units: [] }} />);
    const image = new File(["image"], "produkt.png", { type: "image/png" });

    await user.upload(screen.getByLabelText("Datei", { selector: "input" }), image);
    await user.click(screen.getByRole("button", { name: "Dokumentart" }));
    await user.click(screen.getByRole("option", { name: "Produktbild" }));
    await user.type(screen.getByRole("textbox", { name: "Beschreibung" }), "Vorderansicht");
    await user.click(screen.getByRole("button", { name: "Dokument hochladen" }));

    expect(upload).toHaveBeenCalledWith(article.id, {
      file: image, category: "image", description: "Vorderansicht", isPrimary: true
    });
    expect(await screen.findByRole("alert")).toHaveTextContent("Upload fehlgeschlagen");
    expect(screen.getByText("produkt.png")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Dokumentart" })).toHaveTextContent("Produktbild");
    expect(screen.getByRole("textbox", { name: "Beschreibung" })).toHaveValue("Vorderansicht");
  });

  it("offers an explicit primary-image action for another uploaded image", async () => {
    const user = userEvent.setup();
    const update = vi.spyOn(api, "updateAccessoryDocument").mockResolvedValue({} as never);
    const onChanged = vi.fn().mockResolvedValue(undefined);
    render(<ArticlePurchaseDocumentsTab article={article} disabled={false} onChanged={onChanged}
      onDirtyChange={vi.fn()} resources={{ locations: [], stock: null, movements: [], assets: [], purchases: [],
        documents: [{ id: "image-2", productId: article.id, originalName: "Seite.png", fileName: "seite.png",
          category: "image", mimeType: "image/png", sizeBytes: 100, isPrimary: false, createdBy: "admin",
          createdAt: "2026-08-08T09:00:00Z", updatedAt: "2026-08-08T09:00:00Z" }],
        reservations: [], installations: [], usageHistory: null, vehicles: [], layouts: [], units: [] }} />);

    await user.click(screen.getByRole("button", { name: "Als Produktbild verwenden: Seite.png" }));

    expect(update).toHaveBeenCalledWith(article.id, "image-2", {
      category: "image", description: undefined, isPrimary: true
    });
    expect(onChanged).toHaveBeenCalledOnce();
  });
});
