import { createRef, useState, type ComponentProps } from "react";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { ExhibitionEntryInput, ExhibitionWorkspace, MasterDataEntry } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { ExhibitionEntryDialog } from "./ExhibitionEntryDialog";

const at = "2026-11-08T12:00:00Z";
const workspace: ExhibitionWorkspace = {
  list: { id: "event", designation: "Event", date: "2026-11-08", endDate: "2026-11-10", status: "open", revision: 4, locked: false, entryCount: 0, createdAt: at, updatedAt: at },
  entries: [], conflicts: [], summary: { entryCount: 0, ownerCount: 0, conflictCount: 0, readyCount: 0 },
  readiness: { total: 0, addressesChecked: 0, functionsDocumented: 0, imagesPresent: 0, problems: 0 }, dayScopes: ["day1", "day2", "day3"]
};
const initial: ExhibitionEntryInput = { owner: "Stored owner", locomotiveName: "Guest locomotive", dayScope: "all", dtDecoder: false, analog: false, availability: "available", notes: "Keep this note" };
function entry(label: string): MasterDataEntry { return { id: label, type: "manufacturer", key: label, label, active: true, metadata: {}, sortOrder: 1, createdAt: at, updatedAt: at }; }

function renderDialog(options: Partial<ComponentProps<typeof ExhibitionEntryDialog>> = {}) {
  const saved = vi.fn(), changed = vi.fn(), selected = vi.fn(), close = vi.fn(), removed = vi.fn();
  function Harness() {
    const [form, setForm] = useState(options.form || initial);
    return <ExhibitionEntryDialog admin closeRef={createRef()} de editingEntry={null}
      masterData={{ manufacturers: [entry("Piko")], gattungen: [entry("Diesellok")], epochs: [entry("IV")], railwayCompanies: [entry("DR")] }}
      saving={false} symbols={[]} vehicles={[vehicleFixture()]} workspace={workspace} canReadInventory
      onClose={close} onDelete={removed} onSelectVehicle={selected} onChange={patch => { changed(patch); setForm(current => ({ ...current, ...patch })); }}
      onSubmit={event => { event.preventDefault(); saved(form); }} {...options} form={form} />;
  }
  return { ...render(<Harness />), saved, changed, selected, close, removed };
}
async function choose(user: ReturnType<typeof userEvent.setup>, label: string, option: string) {
  await user.click(screen.getByRole("button", { name: label }));
  await user.click(screen.getByRole("option", { name: option }));
}

describe("exhibition entry inputs retain event-only ownership", () => {
  it("submits edited event fields without mutating the inventory source or unrelated notes", async () => {
    const user = userEvent.setup();
    const { saved, selected } = renderDialog();
    for (const [label, value] of [["Besitzer", "Updated owner"], ["Lok-Bezeichnung", "Updated locomotive"], ["Baureihe", "218"], ["Decoder-Typ", "LokSound"], ["Adresse DCC", "21"], ["Adresse SX", "12"], ["Zentrale / Schnittstelle", "ECoS"]]) {
      await user.clear(screen.getByLabelText(label)); await user.type(screen.getByLabelText(label), value);
    }
    await choose(user, "Hersteller", "Piko"); await choose(user, "Gattung", "Diesellok");
    await choose(user, "Epoche", "IV"); await choose(user, "Bahnverwaltung", "DR");
    await choose(user, "Verfügbarkeit", "Nicht verfügbar");
    await user.click(screen.getByLabelText("Analog"));
    await user.click(screen.getByRole("button", { name: "Tag 1" }));
    await user.click(screen.getByRole("button", { name: "Tag 3" }));
    await user.click(screen.getByRole("button", { name: "Speichern" }));
    expect(saved).toHaveBeenLastCalledWith(expect.objectContaining({ owner: "Updated owner", locomotiveName: "Updated locomotive", series: "218", manufacturer: "Piko", gattung: "Diesellok", epoch: "IV", railwayCompany: "DR", availability: "unavailable", analog: true, decoderType: "LokSound", decoderNumber: "21", sxAddress: "12", interfaceName: "ECoS", dayScope: "day1,day3", notes: initial.notes }));
    expect(selected).not.toHaveBeenCalled();
  });

  it("restores all days when the final explicit day is removed and preserves raw legacy manufacturer labels", async () => {
    const user = userEvent.setup();
    const { saved } = renderDialog({ form: { ...initial, dayScope: "day2", manufacturer: " Legacy Maker " } });
    expect(screen.getByRole("button", { name: "Hersteller" })).toHaveTextContent("Legacy Maker");
    await user.click(screen.getByRole("button", { name: "Tag 2" }));
    await user.click(screen.getByRole("button", { name: "Speichern" }));
    expect(saved).toHaveBeenLastCalledWith(expect.objectContaining({ dayScope: "all", manufacturer: " Legacy Maker " }));
    await user.click(screen.getByRole("button", { name: "Tag 1" }));
    await user.click(screen.getByRole("button", { name: "Alle Tage" }));
    await user.click(screen.getByRole("button", { name: "Speichern" }));
    expect(saved).toHaveBeenLastCalledWith(expect.objectContaining({ dayScope: "all" }));
  });

  it("keeps old function-key text editable and serializes only explicitly configured functions", async () => {
    const user = userEvent.setup();
    const { saved } = renderDialog({ form: { ...initial, functionKeys: "F0: Old light;F1: Horn" } });
    await user.click(screen.getByRole("tab", { name: /Funktionstasten/ }));
    expect(screen.getByLabelText("F0 Beschreibung")).toHaveValue("Old light");
    expect(screen.getByLabelText("F1 Beschreibung")).toHaveValue("Horn");
    await user.click(screen.getByRole("button", { name: "Funktion hinzufügen" }));
    await user.type(screen.getByLabelText("F2 Beschreibung"), "Coupling");
    await choose(user, "F2 Typ", "Kupplung");
    await user.click(screen.getByRole("button", { name: "F1 entfernen" }));
    await user.click(screen.getByRole("button", { name: "Speichern" }));
    expect(JSON.parse(saved.mock.lastCall![0].functionKeys)).toEqual([
      { key: "F0", name: "Old light", type: "licht", symbolKey: "light" },
      { key: "F2", name: "Coupling", type: "kupplung", symbolKey: "" }
    ]);
    expect(screen.queryByLabelText("F1 Beschreibung")).not.toBeInTheDocument();
  });

  it("retains existing JSON symbols, allows function removal, and keeps unrelated event input", async () => {
    const user = userEvent.setup();
    const { saved } = renderDialog({ form: { ...initial, functionKeys: JSON.stringify([{ key: "F3", name: "Stored sound", type: "sound", symbolKey: "sound" }]) } });
    await user.click(screen.getByRole("tab", { name: /Funktionstasten/ }));
    expect(screen.getByLabelText("F3 Beschreibung")).toHaveValue("Stored sound");
    await user.click(screen.getByRole("button", { name: "F3 entfernen" }));
    await user.click(screen.getByRole("tab", { name: "Allgemein" }));
    expect(screen.getByLabelText("Besitzer")).toHaveValue(initial.owner);
    await user.click(screen.getByRole("button", { name: "Speichern" }));
    expect(JSON.parse(saved.mock.lastCall![0].functionKeys).some((item: { key: string }) => item.key === "F3")).toBe(false);
  });

  it("adopts image bytes into this event and removes the image only on explicit removal", async () => {
    const user = userEvent.setup();
    const { changed, saved } = renderDialog();
    await user.click(screen.getByRole("tab", { name: /Fahrzeugbild/ }));
    await user.upload(screen.getByLabelText<HTMLInputElement>("Bilddatei", { selector: "input" }), new File(["synthetic image"], "event.png", { type: "image/png" }));
    await waitFor(() => expect(changed).toHaveBeenCalledWith(expect.objectContaining({ imageUrl: expect.stringMatching(/^data:image\/png;base64,/) })));
    await user.click(screen.getByRole("button", { name: "Bild entfernen" }));
    expect(screen.getByText("Kein Fahrzeugbild")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Speichern" }));
    expect(saved).toHaveBeenLastCalledWith(expect.objectContaining({ imageUrl: "", notes: initial.notes }));
  });

  it("never offers inventory lookup or deletion to a Messe entry editor", async () => {
    const { close } = renderDialog({ admin: false, canReadInventory: false, de: false, saving: true });
    expect(screen.queryByRole("button", { name: "Inventory vehicle (optional)" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Delete" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();
    await userEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(close).toHaveBeenCalledOnce();
  });
});
