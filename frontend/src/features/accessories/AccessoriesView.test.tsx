import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  api,
  type AccessoryInstallation,
  type AccessoryProduct,
  type AccessoryReservation,
  type Layout,
  type LayoutUnit
} from "../../shared/api";
import { AccessoriesView } from "./AccessoriesView";

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((next) => { resolve = next; });
  return { promise, resolve };
}

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

const layout: Layout = {
  id: "layout-1", name: "Clubanlage Bahnhof", kind: "club", gauge: "TT", scale: "1:120", version: 1,
  archived: false, createdAt: "2026-08-07T10:00:00Z", updatedAt: "2026-08-07T10:00:00Z"
};

const layoutUnit: LayoutUnit = {
  id: "unit-1", layoutId: "layout-1", name: "Bahnhofsmodul", kind: "module", widthMm: 1200,
  heightMm: 500, version: 1, archived: false, createdAt: "2026-08-07T10:00:00Z",
  updatedAt: "2026-08-07T10:00:00Z"
};

const reservation: AccessoryReservation = {
  id: "reservation-1", productId: "product-1", locationId: "location-1", quantity: 2, layoutId: "layout-1",
  status: "active", createdBy: "planner", createdAt: "2026-08-07T10:00:00Z",
  updatedAt: "2026-08-07T10:00:00Z"
};

const installation: AccessoryInstallation = {
  id: "installation-1", productId: "product-1", sourceLocationId: "location-1", quantity: 2,
  layoutId: "layout-1", condition: "ready", installedBy: "editor", installedAt: "2026-08-07T10:00:00Z"
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
    vi.spyOn(api, "vehicles").mockResolvedValue([]);
    vi.spyOn(api, "layouts").mockResolvedValue([layout]);
    vi.spyOn(api, "layoutUnits").mockResolvedValue([layoutUnit]);
    vi.spyOn(api, "accessoryReservations").mockResolvedValue([]);
    vi.spyOn(api, "accessoryInstallations").mockResolvedValue([]);
    vi.spyOn(api, "accessoryAllocationSummary").mockResolvedValue({ productId: "product-1", owned: 5,
      stored: 5, reserved: 0, installed: 0, available: 5, missing: 0 });
  });

  it("loads the workspace in parallel and renders long German product names", async () => {
    render(<AccessoriesView roles={["Viewer"]} />);

    expect(screen.getByText("Zubehördaten werden geladen...")).toBeInTheDocument();
    expect(await screen.findByText(quantityProduct.name)).toBeInTheDocument();
    expect(api.accessoryProducts).toHaveBeenCalledWith("");
    expect(api.storageLocations).toHaveBeenCalledOnce();
    await waitFor(() => expect(api.accessoryStock).toHaveBeenCalledWith("product-1"));
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

  it("preserves the product search when switching tabs", async () => {
    const user = userEvent.setup();
    render(<AccessoriesView roles={["Viewer"]} />);
    await screen.findByText(quantityProduct.name);

    await user.type(screen.getByLabelText("Produkte suchen"), "Tillig TT");
    await user.click(screen.getByRole("tab", { name: "Bestand" }));
    await user.click(screen.getByRole("tab", { name: "Produkte" }));

    expect(screen.getByLabelText("Produkte suchen")).toHaveValue("Tillig TT");
  });

  it("keeps storage-location administration out of the accessory feature", async () => {
    render(<AccessoriesView roles={["Editor"]} />);
    await screen.findByText(quantityProduct.name);

    expect(screen.queryByRole("tab", { name: "Lagerorte" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Lagerort speichern" })).not.toBeInTheDocument();
  });

  it("ignores stale product detail responses", async () => {
    const user = userEvent.setup();
    const staleStock = deferred<Awaited<ReturnType<typeof api.accessoryStock>>>();
    const staleSummary = deferred<Awaited<ReturnType<typeof api.accessoryAllocationSummary>>>();
    vi.mocked(api.accessoryProducts).mockResolvedValue([quantityProduct, individualProduct]);
    vi.mocked(api.accessoryStock).mockImplementation((id) => id === "product-1" ? staleStock.promise
      : Promise.resolve({ productId: "product-2", trackingMode: "individual", totalQuantity: 1, locations: [] }));
    vi.mocked(api.accessoryAllocationSummary).mockImplementation((id) => id === "product-1" ? staleSummary.promise
      : Promise.resolve({ productId: "product-2", owned: 1, stored: 1, reserved: 0, installed: 0,
        available: 1, missing: 0 }));
    render(<AccessoriesView roles={["Viewer"]} />);
    await screen.findByText(individualProduct.name);

    await user.click(screen.getByText(individualProduct.name).closest("button")!);
    await waitFor(() => expect(api.accessoryStock).toHaveBeenCalledWith("product-2"));
    staleStock.resolve({ productId: "product-1", trackingMode: "quantity", totalQuantity: 99, locations: [] });
    staleSummary.resolve({ productId: "product-1", owned: 99, stored: 99, reserved: 0, installed: 0,
      available: 99, missing: 0 });

    await waitFor(() => expect(screen.getByRole("region", { name: "Bestandszusammenfassung" }))
      .not.toHaveTextContent("99"));
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

  it("lets planners create and cancel reservations but not installations", async () => {
    const user = userEvent.setup();
    vi.mocked(api.accessoryReservations).mockResolvedValue([reservation]);
    vi.spyOn(api, "createAccessoryReservation").mockResolvedValue(reservation);
    vi.spyOn(api, "cancelAccessoryReservation").mockResolvedValue({ ...reservation, status: "cancelled" });
    render(<AccessoriesView roles={["Planner"]} />);
    await screen.findByText(quantityProduct.name);

    await user.click(screen.getByRole("tab", { name: "Reservierungen" }));
    await user.click(screen.getByRole("button", { name: "Reservierung anlegen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    await waitFor(() => expect(api.createAccessoryReservation).toHaveBeenCalledWith(expect.objectContaining({
      productId: "product-1", layoutId: "layout-1", locationId: "location-1", quantity: 1
    })));

    await user.click(screen.getByRole("button", { name: "Stornieren" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    await waitFor(() => expect(api.cancelAccessoryReservation).toHaveBeenCalledWith("reservation-1"));

    await user.click(screen.getByRole("tab", { name: "Einbauhistorie" }));
    expect(screen.queryByText("Zubehör einbauen")).not.toBeInTheDocument();
  });

  it("lets editors record and remove installations with confirmation", async () => {
    const user = userEvent.setup();
    vi.mocked(api.accessoryInstallations).mockResolvedValue([installation]);
    vi.spyOn(api, "createAccessoryInstallation").mockResolvedValue(installation);
    vi.spyOn(api, "updateAccessoryInstallationCondition").mockResolvedValue({ ...installation,
      condition: "maintenance_due" });
    vi.spyOn(api, "removeAccessoryInstallation").mockResolvedValue({ ...installation,
      removedAt: "2026-08-07T11:00:00Z", removedBy: "editor", removalDisposition: "stored" });
    render(<AccessoriesView roles={["Editor"]} />);
    await screen.findByText(quantityProduct.name);

    await user.click(screen.getByRole("tab", { name: "Einbauhistorie" }));
    await user.click(screen.getByRole("button", { name: "Einbau erfassen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    await waitFor(() => expect(api.createAccessoryInstallation).toHaveBeenCalledWith(expect.objectContaining({
      productId: "product-1", layoutId: "layout-1", sourceLocationId: "location-1", quantity: 1
    })));

    await user.click(screen.getByRole("button", { name: "Zustand speichern" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    await waitFor(() => expect(api.updateAccessoryInstallationCondition)
      .toHaveBeenCalledWith("installation-1", { condition: "ready" }));

    await user.click(screen.getByRole("button", { name: "Ausbauen" }));
    await user.click(screen.getAllByRole("button", { name: "Ausbauen" })[1]);
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    await waitFor(() => expect(api.removeAccessoryInstallation).toHaveBeenCalledWith("installation-1", {
      disposition: "stored", storageLocationId: "location-1", notes: undefined
    }));
  });

  it("shows loading failures", async () => {
    vi.mocked(api.accessoryProducts).mockRejectedValue(new Error("Zubehör nicht erreichbar"));
    render(<AccessoriesView roles={["Viewer"]} />);
    expect(await screen.findByText("Zubehör nicht erreichbar")).toBeInTheDocument();
  });
});
