import { act, fireEvent, render, renderHook, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { api } from "../../shared/api";
import { functionFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { functionDraftChanges, functionNoteParts, withFunctionOwnNote } from "./functionDraft";
import { useVehicleFunctionsController } from "./useVehicleFunctionsController";
import { emptyVehicle } from "./vehicleViewModel";

const original = "Eigene Notiz\nRailComPlus · Betriebsart: Dauer; Invertiert: nein · CV31=1 / CV32=1 · CV257=0; CV258=3 · Originalzeile 175 / 176 · Exportdatum: 2026-10-17T17:41:59";

describe("one function draft", () => {
  it("retains large opaque ECoS descriptors as source notes when the own note changes", () => {
    const source = "ECoS · #1004 · F3 · funcexists=2147483647 · dauer · 2026-10-27T08:00:00Z";
    expect(functionNoteParts("Eigene Notiz\n" + source, "F3")).toEqual({ own: "Eigene Notiz", sources: [source] });
    expect(withFunctionOwnNote("Eigene Notiz\n" + source, "F3", "Geändert")).toBe("Geändert\n" + source);
  });
  it("separates source presentation without rewriting historical notes", () => {
    const ecos = "ECoS · #1002 · F0 · funcexists=3 · dauer · 2026-10-26T08:00:00Z";
    expect(functionNoteParts("Own note\n" + ecos, "F0")).toEqual({ own: "Own note", sources: [ecos] });
    expect(functionNoteParts(ecos, "F1").own).toBe(ecos);
    expect(functionNoteParts(original, "F0")).toEqual({ own: "Eigene Notiz", sources: [original.split("\n")[1]] });
    expect(functionNoteParts(original, "F1").own).toBe(original);
    const changed = withFunctionOwnNote(original, "F0", "Neue Notiz\nZweite Zeile");
    expect(changed).toBe("Neue Notiz\nZweite Zeile\n" + original.split("\n")[1]);
    expect(functionNoteParts("RailComPlus ist meine eigene Beschreibung", "F0").own).toBe("RailComPlus ist meine eigene Beschreibung");
    expect(functionDraftChanges({ F0: { ...functionFixture(), persisted: true, notes: original } })).toEqual([]);
  });

  it("stages edits and deletions, preserves revisions across refresh, and resets on cancel", () => {
    const { result } = renderHook(() => useVehicleFunctionsController({ selected: vehicleFixture(), form: emptyVehicle,
      setSaving: vi.fn(), onMessage: vi.fn(), refreshSelectedVehicle: vi.fn() }));
    const saved = functionFixture({ revision: "a".repeat(64), notes: original });
    expect(result.current.state.showConfiguredOnly).toBe(true);
    expect(result.current.state.configuredKeys).toEqual([]);
    act(() => result.current.commands.loadDetail([saved], "vehicle-1"));
    act(() => result.current.commands.updateFunctionEdit("F0", { mode: "moment" }));
    act(() => result.current.commands.loadDetail([{ ...saved, revision: "b".repeat(64), name: "Changed elsewhere" }], "vehicle-1"));
    const changes = result.current.commands.changes();
    expect(changes).toMatchObject([{ functionKey: "F0", expectedRevision: "a".repeat(64), value: { mode: "moment", notes: original } }]);
    expect(changes[0].value).not.toHaveProperty("dirty");
    expect(changes[0].value).not.toHaveProperty("modeAssigned");
    act(() => result.current.commands.remove("F0"));
    expect(result.current.commands.changes()).toEqual([{ functionKey: "F0", expectedRevision: "a".repeat(64), delete: true }]);
    act(() => result.current.commands.remove("F0"));
    expect(result.current.commands.changes()[0].value?.mode).toBe("moment");
    act(() => result.current.commands.reset());
    act(() => result.current.commands.loadDetail([saved], "vehicle-1"));
    expect(result.current.commands.changes()).toEqual([]);
    expect(result.current.commands.functionEdit("F0").notes).toBe(original);
  });

  it("imports into the draft without requests and preserves confirmed flags from older files", async () => {
    const upsert = vi.spyOn(api, "updateVehicleFunction"), remove = vi.spyOn(api, "deleteVehicleFunction");
    const onMessage = vi.fn();
    const { result } = renderHook(() => useVehicleFunctionsController({ selected: vehicleFixture(), form: emptyVehicle,
      setSaving: vi.fn(), onMessage, refreshSelectedVehicle: vi.fn() }));
    act(() => result.current.commands.loadDetail([functionFixture({ inverted: false, directionDependent: true, revision: "a".repeat(64) })], "vehicle-1"));
    render(<input type="file" aria-label="Import file" onChange={event => result.current.commands.importValues(event.target.files)} />);
    const file = { text: async () => JSON.stringify({ functions: [{ functionKey: "F0", name: "Imported", mode: "moment", directionDependent: false }] }) };
    fireEvent.change(screen.getByLabelText("Import file"), { target: { files: [file] } });
    await waitFor(() => expect(result.current.commands.functionEdit("F0").name).toBe("Imported"));
    expect(result.current.commands.changes()[0]).toMatchObject({ expectedRevision: "a".repeat(64),
      value: { flagsVersion: 2, inverted: false, directionDependent: true, legacyDirectionDependent: false } });
    expect(upsert).not.toHaveBeenCalled(); expect(remove).not.toHaveBeenCalled();
    const before = result.current.commands.changes();
    fireEvent.change(screen.getByLabelText("Import file"), { target: { files: [{ text: async () => JSON.stringify([
      { functionKey: "F1", name: "Valid" }, { functionKey: "F32", name: "Invalid" }
    ]) }] } });
    await waitFor(() => expect(onMessage).toHaveBeenLastCalledWith(expect.stringContaining("ungültige")));
    expect(result.current.commands.changes()).toEqual(before);
    upsert.mockRestore(); remove.mockRestore();
  });
});
