import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";
import { api, ApiError, type VehicleCatalogContext, type VehicleSet } from "../../shared/api";
import { catalogFieldsFrom } from "./vehicleCatalogModel";
import { VehicleSetEditorDialog } from "./VehicleSetEditorDialog";

const original: VehicleSet = { id: "set-1", inventoryNumber: "RK-SET-1", name: "Original set",
  manufacturer: "Roco", articleNumber: "45923", gauge: "H0", category: "Wagen", gattung: "Reisezugwagen",
  purchasePrice: "200", members: [], mainImageMode: "automatic",
  createdAt: "2026-08-01T00:00:00Z", updatedAt: "2026-08-01T00:00:00Z" };
const fields = catalogFieldsFrom(original);
const revision = { id: "revision", articleId: "article", revisionNumber: 1, fields, note: "", createdAt: original.createdAt };
const context: VehicleCatalogContext = { targetType: "set", targetId: original.id, authority: "catalog",
  currentFields: fields, overrides: {}, pinned: revision, revisions: [revision], revision: "1".repeat(64) };
const start = async () => {
  await screen.findByDisplayValue(original.name);
  await userEvent.click(screen.getByText("Artikelrevisionen"));
  await screen.findByText(/Festgelegter Artikelstand/);
  await userEvent.click(screen.getByRole("button", { name: "Neue Revision entwerfen" }));
};
afterEach(() => vi.restoreAllMocks());

it("retains set drafts across tabs and delayed catalog confirmation, then requires explicit conflict resolution", async () => {
  const changed = { ...original, name: "Confirmed article", manufacturer: "Corrected" };
  vi.spyOn(api, "vehicleSet").mockResolvedValueOnce(original).mockResolvedValue(changed);
  vi.spyOn(api, "catalogContext").mockResolvedValue(context);
  let finish!: (value: VehicleCatalogContext) => void;
  vi.spyOn(api, "createCatalogRevision").mockReturnValue(new Promise(resolve => { finish = resolve; }));
  const update = vi.spyOn(api, "updateVehicleSet");
  const onUpdated = vi.fn(), onClose = vi.fn(), onCatalogChanged = vi.fn().mockResolvedValue(undefined);
  render(<VehicleSetEditorDialog setId={original.id} onUpdated={onUpdated} onClose={onClose}
    onCatalogChanged={onCatalogChanged} />);
  await start();
  const name = screen.getByRole("textbox", { name: "Bezeichnung" });
  await userEvent.clear(name); await userEvent.type(name, "Unsent set name");
  const price = screen.getByRole("textbox", { name: "Preis" });
  await userEvent.clear(price); await userEvent.type(price, "321");
  await userEvent.type(screen.getByRole("textbox", { name: "Notiz zur neuen Revision" }), "Keep article draft");
  await userEvent.click(screen.getByRole("tab", { name: "Upload" }));
  await userEvent.click(screen.getByRole("tab", { name: "Allgemein" }));
  expect(screen.getByRole("textbox", { name: "Notiz zur neuen Revision" })).toHaveValue("Keep article draft");
  await userEvent.click(screen.getByRole("button", { name: "Revision erstellen und Auswahl speichern" }));
  expect(screen.getByRole("button", { name: "Speichern" })).toBeDisabled();
  await userEvent.click(screen.getByRole("tab", { name: "Upload" }));
  await act(async () => finish({ ...context, currentFields: catalogFieldsFrom(changed) }));
  await userEvent.click(screen.getByRole("tab", { name: "Allgemein" }));
  expect(screen.getByRole("textbox", { name: "Bezeichnung" })).toHaveValue("Unsent set name");
  expect(screen.getByRole("textbox", { name: "Preis" })).toHaveValue("321");
  expect(screen.getByRole("button", { name: "Speichern" })).toBeDisabled();
  expect(onCatalogChanged).toHaveBeenCalledOnce();
  await userEvent.click(screen.getByRole("button", { name: "Gespeicherte Werte für diese Felder übernehmen" }));
  expect(screen.getByRole("textbox", { name: "Bezeichnung" })).toHaveValue(changed.name);
  expect(screen.getByRole("textbox", { name: "Preis" })).toHaveValue("321");
  expect(screen.getByRole("button", { name: "Speichern" })).toBeEnabled();
  await userEvent.click(screen.getByRole("button", { name: "Abbrechen" }));
  expect(update).not.toHaveBeenCalled(); expect(onUpdated).not.toHaveBeenCalled(); expect(onClose).toHaveBeenCalledOnce();
});

it("blocks stale set saves after an unknown catalog result until a read-only refresh reconciles the draft", async () => {
  vi.spyOn(api, "vehicleSet").mockResolvedValue(original);
  vi.spyOn(api, "catalogContext").mockResolvedValue(context);
  const create = vi.spyOn(api, "createCatalogRevision").mockRejectedValue(new ApiError("Unknown result", "catalog_failed", 500));
  render(<VehicleSetEditorDialog setId={original.id} onUpdated={vi.fn()} onClose={vi.fn()} />);
  await start();
  await userEvent.click(screen.getByRole("button", { name: "Revision erstellen und Auswahl speichern" }));
  await screen.findByText(/Das Speicherergebnis ist noch nicht bestätigt/);
  expect(screen.getByRole("button", { name: "Speichern" })).toBeDisabled();
  await userEvent.click(screen.getByRole("button", { name: "Gespeicherte Ansicht erneut laden" }));
  await waitFor(() => expect(screen.getByRole("button", { name: "Speichern" })).toBeEnabled());
  expect(create).toHaveBeenCalledOnce();
  expect(screen.queryByText(/Bereits gespeichert:/)).not.toBeInTheDocument();
});

it("ignores late catalog acknowledgements after the set editor scope closes", async () => {
  const read = vi.spyOn(api, "vehicleSet").mockResolvedValue(original);
  vi.spyOn(api, "catalogContext").mockResolvedValue(context);
  let finish!: (value: VehicleCatalogContext) => void, current = true;
  vi.spyOn(api, "createCatalogRevision").mockReturnValue(new Promise(resolve => { finish = resolve; }));
  const onCatalogChanged = vi.fn();
  render(<VehicleSetEditorDialog setId={original.id} onUpdated={vi.fn()} onClose={vi.fn()}
    captureScope={() => () => current} onCatalogChanged={onCatalogChanged} />);
  await start();
  await userEvent.click(screen.getByRole("button", { name: "Revision erstellen und Auswahl speichern" }));
  current = false;
  await act(async () => finish(context));
  expect(read).toHaveBeenCalledOnce(); expect(onCatalogChanged).not.toHaveBeenCalled();
});
