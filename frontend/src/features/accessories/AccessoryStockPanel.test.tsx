import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { api, type AccessoryArticle, type AccessoryAsset, type StorageLocation } from "../../shared/api";
import { AccessoryStockPanel } from "./AccessoryStockPanel";

const article: AccessoryArticle = {
  id: "article-1", inventoryNumber: "RK-ART-000001", manufacturer: "Tillig", name: "Signal",
  category: "signal", trackingMode: "individual",
  manufacturerStatus: "available", articleType: "signal", subtype: "main_signal", gauges: ["TT"], packageQuantity: 1,
  stockUnit: "piece", minimumStock: 0, inventoryStrategy: "individual", alternativeNumbers: [], keywords: [],
  archived: false, attributes: [], createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T09:00:00Z"
};
const asset: AccessoryAsset = { id: "asset-1", productId: article.id, inventoryNumber: "RK-1", serialNumber: "ALT",
  condition: "ready", lifecycle: "stored", storageLocationId: "location-1", purchaseDate: "2026-01-01",
  purchasePrice: "12.50", warrantyUntil: "2028-01-01", notes: "Alt", createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T09:00:00Z" };
const reservedAsset: AccessoryAsset = {
  ...asset, id: "asset-2", inventoryNumber: "RK-2", lifecycle: "reserved"
};
const location: StorageLocation = { id: "location-1", name: "Schublade",
  archived: false, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T09:00:00Z" };

describe("AccessoryStockPanel", () => {
  it("aligns stock adjustment and transfer as peer forms", () => {
    const quantityArticle = {
      ...article,
      trackingMode: "quantity" as const,
      inventoryStrategy: "quantity" as const
    };
    const view = render(<AccessoryStockPanel article={quantityArticle} stock={null} movements={[]} assets={[]}
      locations={[location]} canEdit onChanged={vi.fn()} onDirtyChange={vi.fn()} />);

    const commands = view.container.querySelector(".article-stock-commands");
    const forms = commands?.querySelectorAll(":scope > .article-stock-form");
    expect(forms).toHaveLength(2);
    expect(forms?.[0]?.querySelector(".primary-button")).toHaveTextContent("Bestand buchen");
    expect(forms?.[1]).toHaveClass("article-transfer-form");
    expect(forms?.[1]?.querySelector(".primary-button")).toHaveTextContent("Umbuchen");

    const transferFields = forms?.[1]?.querySelector(".article-transfer-fields");
    expect(transferFields).toBeInTheDocument();
    expect(transferFields?.children).toHaveLength(4);
  });

  it("edits every supported individual item field through typed app controls", async () => {
    const user = userEvent.setup();
    const update = vi.spyOn(api, "updateAccessoryAsset").mockResolvedValue(asset);
    render(<AccessoryStockPanel article={article} stock={null} movements={[]} assets={[asset]} locations={[location]}
      canEdit onChanged={vi.fn()} onDirtyChange={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Einzelstück bearbeiten: RK-1" }));
    const serial = screen.getByRole("textbox", { name: "Seriennummer" });
    await user.clear(serial); await user.type(serial, "NEU");
    const notes = screen.getByRole("textbox", { name: "Notizen" });
    await user.clear(notes); await user.type(notes, "Geprüft");
    expect(screen.getByRole("textbox", { name: "Kaufdatum" })).toHaveValue("01.01.2026");
    expect(screen.getByRole("spinbutton", { name: "Kaufpreis" })).toHaveValue(12.5);
    expect(screen.getByRole("button", { name: "Zustand" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Einzelstück speichern" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    expect(update).toHaveBeenCalledWith("asset-1", expect.objectContaining({ inventoryNumber: "RK-1",
      serialNumber: "NEU", condition: "ready", lifecycle: "stored", purchaseDate: "2026-01-01", purchasePrice: "12.50",
      warrantyUntil: "2028-01-01", notes: "Geprüft" }));
  });

  it("does not offer asset editing for reserved or installed lifecycles", () => {
    render(<AccessoryStockPanel article={article} stock={null} movements={[]}
      assets={[reservedAsset, { ...reservedAsset, id: "asset-3", inventoryNumber: "RK-3", lifecycle: "installed" }]}
      locations={[location]} canEdit onChanged={vi.fn()} onDirtyChange={vi.fn()} />);

    expect(screen.queryByRole("button", { name: "Einzelstück bearbeiten: RK-2" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Einzelstück bearbeiten: RK-3" })).not.toBeInTheDocument();
  });
});
