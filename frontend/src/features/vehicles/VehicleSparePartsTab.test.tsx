import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { api, type VehicleSparePart } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleSparePartsTab } from "./VehicleSparePartsTab";
import { useVehicleSparePartsController } from "./useVehicleSparePartsController";

const part: VehicleSparePart = { id: "part", vehicleId: "vehicle-1", articleNumber: "ET-12", description: "Radsatz", price: "12,50 EUR", url: "https://example.invalid/part", createdAt: "", updatedAt: "" };
const originalScroll = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "scrollIntoView");
const selected = vehicleFixture({ manufacturer: "Piko", spareParts: [part] });

function Workspace({ readonly = false }: { readonly?: boolean }) {
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const { t } = useI18n();
  const { state, commands } = useVehicleSparePartsController({
    selected, active: false, attachmentEdits: {}, setSaving, onMessage: setMessage,
    onOpenSpareParts: () => { }, refreshSelectedVehicle: async () => { }, t
  });
  return <><p role="status">{message}</p><VehicleSparePartsTab
    selected={selected} readonly={readonly} saving={saving} sparePartForm={state.form} editingSparePartID={state.editingId}
    sparePartLookupLoadingID={state.lookupLoadingId} sparePartLookupErrors={state.lookupErrors} sparePartLookupResults={state.lookupResults}
    sparePartStatusLoading={state.statusLoading} sparePartStatuses={state.statuses} importAllSparePartsLoading={state.importAllLoading}
    canImportAllSpareParts={state.canImportAll} importAllSparePartsTitle={state.importAllTitle}
    onUpdateSparePartForm={commands.updateForm} onResetSparePartForm={commands.resetForm} onSaveSparePart={commands.save}
    onEditSparePart={commands.edit} onDeleteSparePart={commands.remove} onSearchSparePart={commands.searchSingle}
    onApplySparePartLookup={commands.applyLookup} onImportAllSpareParts={commands.importAll} sparePartSort={state.sort} onToggleSparePartSort={commands.toggleSort}
  /></>;
}

describe("Stored spare parts and explicit source adoption", () => {
  afterEach(() => {
    if (originalScroll) Object.defineProperty(HTMLElement.prototype, "scrollIntoView", originalScroll);
    else Reflect.deleteProperty(HTMLElement.prototype, "scrollIntoView");
  });
  beforeEach(() => {
    Object.defineProperty(HTMLElement.prototype, "scrollIntoView", { configurable: true, value: vi.fn() });
  });

  it("keeps rejected edits, trims only on save and preserves the stored article identity", async () => {
    const update = vi.spyOn(api, "updateVehicleSparePart").mockRejectedValueOnce(new Error("Speicherung abgelehnt")).mockResolvedValue(part);
    render(<Workspace />);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Ersatzteil bearbeiten" }));
    expect(screen.getByLabelText("Art. Nr.")).toHaveValue("ET-12");
    await user.clear(screen.getByLabelText("Preis"));
    await user.type(screen.getByLabelText("Preis"), " 13,40 EUR ");
    await user.clear(screen.getByLabelText("Beschreibung"));
    await user.type(screen.getByLabelText("Beschreibung"), " Neuer Radsatz ");
    expect(update).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Ersatzteil speichern" }));
    expect(await screen.findByRole("status")).toHaveTextContent("Speicherung abgelehnt");
    expect(screen.getByLabelText("Beschreibung")).toHaveValue(" Neuer Radsatz ");
    await user.click(screen.getByRole("button", { name: "Ersatzteil speichern" }));
    await waitFor(() => expect(screen.getByLabelText("Beschreibung")).toHaveValue(""));
    expect(update).toHaveBeenLastCalledWith("vehicle-1", "part", { articleNumber: "ET-12", description: "Neuer Radsatz", price: "13,40 EUR", url: part.url });
  });

  it("keeps a search suggestion separate until the user explicitly adopts it", async () => {
    const candidate = { articleNumber: "ET-12", description: "Radsatz", price: "14,50 EUR", url: "https://example.invalid/new", availability: "knapp lieferbar", source: "www.piko-shop.de" };
    vi.spyOn(api, "articleSearch").mockResolvedValue({ query: "ET-12", results: [{ title: "Hersteller", source: "Piko", url: "https://example.invalid", snippet: "", score: 100, fields: {}, spareParts: [candidate] }] });
    const update = vi.spyOn(api, "updateVehicleSparePart").mockResolvedValue(part);
    render(<Workspace />);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Dieses Ersatzteil suchen" }));
    expect(await screen.findByText("14,50 EUR")).toBeInTheDocument();
    expect(screen.getAllByRole("img", { name: "knapp lieferbar" })[0]).toHaveClass("limited");
    expect(update).not.toHaveBeenCalled();
    expect(screen.getByText("12,50 EUR")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Ersten Treffer übernehmen" }));
    await waitFor(() => expect(update).toHaveBeenCalledWith("vehicle-1", "part", expect.objectContaining({ articleNumber: "ET-12", price: "14,50 EUR", url: candidate.url })));
  });

  it("shows a lookup failure while preserving stored parts and permits a later search", async () => {
    vi.spyOn(api, "articleSearch").mockRejectedValueOnce(new Error("Hersteller nicht erreichbar")).mockResolvedValue({ query: "ET-12", results: [] });
    render(<Workspace />);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Dieses Ersatzteil suchen" }));
    expect(await screen.findByText("Hersteller nicht erreichbar")).toBeInTheDocument();
    expect(screen.getByText("Radsatz")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Dieses Ersatzteil suchen" }));
    await waitFor(() => expect(api.articleSearch).toHaveBeenCalledTimes(2));
    expect(screen.getByRole("button", { name: "Ersten Treffer übernehmen" })).toBeDisabled();
  });

  it("cancels an edited draft and supports independent sorting without changing stored data", async () => {
    const update = vi.spyOn(api, "updateVehicleSparePart");
    render(<Workspace />);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Ersatzteil bearbeiten" }));
    await user.type(screen.getByLabelText("Beschreibung"), " Verworfen");
    await user.click(screen.getByRole("button", { name: "Abbrechen" }));
    expect(screen.getByLabelText("Beschreibung")).toHaveValue("");
    await user.click(screen.getByRole("button", { name: "Nach Preis sortieren" }));
    await user.click(screen.getByRole("button", { name: "Nach Preis sortieren" }));
    await user.click(screen.getByRole("button", { name: "Nach Link sortieren" }));
    const row = within(screen.getByRole("row", { name: /ET-12/ }));
    expect(row.getByText("Radsatz")).toBeInTheDocument();
    expect(update).not.toHaveBeenCalled();
  });

  it("retains source links but blocks all changes for a viewer", () => {
    render(<Workspace readonly />);
    for (const name of ["Ersatzteil hinzufügen", "Ersatzteil bearbeiten", "Ersatzteil löschen", "Dieses Ersatzteil suchen", "Verfügbare Ersatzteile suchen"]) {
      expect(screen.getByRole("button", { name })).toBeDisabled();
    }
    expect(screen.getByRole("link", { name: "Öffnen" })).toHaveAttribute("href", part.url);
  });
});
