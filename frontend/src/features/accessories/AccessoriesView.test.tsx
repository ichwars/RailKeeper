import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { api, type AccessoryProduct } from "../../shared/api";
import { AccessoriesView } from "./AccessoriesView";

const quantityProduct: AccessoryProduct = {
  id: "product-1",
  manufacturer: "Tillig",
  articleNumber: "83101",
  name: "Gerades Modellgleis mit besonders langer Artikelbezeichnung",
  category: "Gleismaterial",
  trackingMode: "quantity",
  createdAt: "2026-08-07T10:00:00Z",
  updatedAt: "2026-08-07T10:00:00Z"
};

const individualProduct: AccessoryProduct = {
  ...quantityProduct,
  id: "product-2",
  articleNumber: "LS150",
  name: "Schaltdecoder",
  category: "Decoder",
  trackingMode: "individual"
};

describe("AccessoriesView", () => {
  afterEach(() => vi.restoreAllMocks());

  beforeEach(() => {
    vi.spyOn(api, "accessoryProducts").mockResolvedValue([quantityProduct]);
    vi.spyOn(api, "storageLocations").mockResolvedValue([{ id: "location-1", name: "Werkstatt", archived: false,
      createdAt: "2026-08-07T10:00:00Z", updatedAt: "2026-08-07T10:00:00Z" }]);
    vi.spyOn(api, "accessoryStock").mockResolvedValue({ productId: "product-1", trackingMode: "quantity",
      totalQuantity: 5, locations: [{ locationId: "location-1", locationName: "Werkstatt", quantity: 5,
        updatedAt: "2026-08-07T10:00:00Z" }] });
    vi.spyOn(api, "accessoryAssets").mockResolvedValue([]);
    vi.spyOn(api, "accessoryAllocationSummary").mockResolvedValue({ productId: "product-1", owned: 5,
      stored: 5, reserved: 0, installed: 0, available: 5, missing: 0 });
  });

  it("loads the workspace in parallel and renders long German product names", async () => {
    render(<AccessoriesView roles={["Viewer"]} />);

    expect(screen.getByText("Zubehördaten werden geladen...")).toBeInTheDocument();
    expect(await screen.findByText(quantityProduct.name)).toBeInTheDocument();
    expect(api.accessoryProducts).toHaveBeenCalledWith("");
    expect(api.storageLocations).toHaveBeenCalledOnce();
    expect(api.accessoryStock).toHaveBeenCalledWith("product-1");
    expect(screen.queryByText("Produkt anlegen")).not.toBeInTheDocument();
  });

  it("searches products and creates a product for editors", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "createAccessoryProduct").mockResolvedValue(quantityProduct);
    render(<AccessoriesView roles={["Editor"]} />);
    await screen.findByText(quantityProduct.name);

    await user.type(screen.getByLabelText("Produkte suchen"), "Tillig TT");
    await user.click(screen.getByRole("button", { name: "Suchen" }));
    expect(api.accessoryProducts).toHaveBeenLastCalledWith("Tillig TT");

    await user.type(screen.getByLabelText("Hersteller"), "Tillig");
    await user.type(screen.getByLabelText("Artikelnummer"), "83102");
    await user.type(screen.getByLabelText("Bezeichnung"), "Gebogenes Gleis");
    await user.type(screen.getByLabelText("Kategorie"), "Gleismaterial");
    await user.click(screen.getByRole("button", { name: "Produkt speichern" }));

    await waitFor(() => expect(api.createAccessoryProduct).toHaveBeenCalledWith(expect.objectContaining({
      manufacturer: "Tillig", articleNumber: "83102", name: "Gebogenes Gleis", category: "Gleismaterial"
    })));
  });

  it("updates the selected product for editors", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "updateAccessoryProduct").mockResolvedValue({ ...quantityProduct, name: "Gerades Modellgleis" });
    render(<AccessoriesView roles={["Editor"]} />);
    await screen.findByText(quantityProduct.name);

    await user.click(screen.getByRole("button", { name: "Ausgewähltes Produkt bearbeiten" }));
    const name = screen.getByLabelText("Bezeichnung");
    await user.clear(name);
    await user.type(name, "Gerades Modellgleis");
    await user.click(screen.getByRole("button", { name: "Produkt speichern" }));

    await waitFor(() => expect(api.updateAccessoryProduct).toHaveBeenCalledWith("product-1",
      expect.objectContaining({ name: "Gerades Modellgleis", trackingMode: "quantity" })));
  });

  it("confirms quantity stock adjustments and keeps failures visible", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "adjustAccessoryStock").mockRejectedValue(new Error("Bestand konnte nicht gebucht werden"));
    render(<AccessoriesView roles={["Editor"]} />);
    await screen.findByText(quantityProduct.name);

    await user.click(screen.getByRole("tab", { name: "Bestand" }));
    fireEvent.change(screen.getByLabelText("Mengenänderung"), { target: { value: "3" } });
    await user.click(screen.getByRole("button", { name: "Bestand buchen" }));
    expect(screen.getByRole("dialog", { name: "Bestandsbuchung bestätigen" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    expect(await screen.findByText("Bestand konnte nicht gebucht werden")).toBeInTheDocument();
    expect(screen.getByRole("dialog", { name: "Bestandsbuchung bestätigen" })).toBeInTheDocument();
  });

  it("creates individually tracked assets", async () => {
    const user = userEvent.setup();
    vi.mocked(api.accessoryProducts).mockResolvedValue([individualProduct]);
    vi.mocked(api.accessoryStock).mockResolvedValue({ productId: "product-2", trackingMode: "individual",
      totalQuantity: 0, locations: [] });
    vi.spyOn(api, "createAccessoryAsset").mockResolvedValue({ id: "asset-1", productId: "product-2",
      inventoryNumber: "RK-Z-0001", condition: "ready", lifecycle: "stored", storageLocationId: "location-1",
      createdAt: "2026-08-07T10:00:00Z", updatedAt: "2026-08-07T10:00:00Z" });
    render(<AccessoriesView roles={["Admin"]} />);
    await screen.findByText(individualProduct.name);

    await user.click(screen.getByRole("tab", { name: "Einzelobjekte" }));
    await user.type(screen.getByLabelText("Inventarnummer"), "RK-Z-0001");
    await user.type(screen.getByLabelText("Kaufpreis"), "24.90");
    await user.type(screen.getByLabelText("Notizen"), "Schaltdecoder für Bahnhof");
    await user.click(screen.getByRole("button", { name: "Einzelobjekt speichern" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    await waitFor(() => expect(api.createAccessoryAsset).toHaveBeenCalledWith("product-2", expect.objectContaining({
      inventoryNumber: "RK-Z-0001", lifecycle: "stored", condition: "ready", purchasePrice: "24.90",
      notes: "Schaltdecoder für Bahnhof"
    })));
  });

  it("shows loading failures", async () => {
    vi.mocked(api.accessoryProducts).mockRejectedValue(new Error("Zubehör nicht erreichbar"));
    render(<AccessoriesView roles={["Viewer"]} />);
    expect(await screen.findByText("Zubehör nicht erreichbar")).toBeInTheDocument();
  });
});
