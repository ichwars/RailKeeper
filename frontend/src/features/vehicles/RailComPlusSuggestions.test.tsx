import { act, fireEvent, render, renderHook, screen } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import type { MasterDataEntry, VehicleFunctionInput } from "../../shared/api";
import { exactECoSSymbol } from "../../shared/ecosSymbols";
import { RailComPlusSuggestions } from "./RailComPlusSuggestions";
import { RailComPlusPreview } from "./RailComPlusPreview";
import { decoderAnalysis } from "./decoderAnalysis";
import { CVImportRowsTable } from "./CVImportRowsTable";
import { buildCVImportPreview } from "./cvImport";
import { railComPlusFlagPatch } from "./railComPlusFunctionFlags";
import { useVehicleFunctionsController } from "./useVehicleFunctionsController";
import { emptyVehicle } from "./vehicleViewModel";
import { translate } from "../../shared/i18n";

const label = "LokPilot 5 micro DCC";
const values = [{ cvNumber: 8, value: 151, decoderProfile: label },
  { cvNumber: 257, value: 96, indexCV31: 1, indexCV32: 1, decoderProfile: label },
  { cvNumber: 258, value: 9, indexCV31: 1, indexCV32: 1, decoderProfile: label }];
const data = decoderAnalysis(values, label).railComPlus;
const symbol: MasterDataEntry = { id: "symbols:shunting", type: "symbols", key: "shunting", label: "Rangiergang",
  active: true, sortOrder: 1, metadata: { ecosCode: 9 }, createdAt: "", updatedAt: "" };

it("fills a missing symbol while preserving names, mode, direction and notes", () => {
  const update = vi.fn(), current: VehicleFunctionInput = { name: "Mein Rangiergang", notes: "Eigene Notiz", mode: "moment", directionDependent: true };
  render(<RailComPlusSuggestions data={data} symbols={[symbol]} readonly={false} edit={() => current} update={update} />);
  fireEvent.click(screen.getByRole("button", { name: "Symbol für F0 ergänzen" }));
  expect(update).toHaveBeenCalledOnce();
  const patch = update.mock.calls[0][1];
  expect(patch.symbolKey).toBe("shunting");
  expect(patch.notes).toContain("Eigene Notiz\nRailComPlus");
  for (const key of ["name", "mode", "directionDependent"]) expect(patch).not.toHaveProperty(key);
});

it("protects existing symbols and read-only users", () => {
  const update = vi.fn();
  const { rerender } = render(<RailComPlusSuggestions data={data} symbols={[symbol]} readonly={false}
    edit={() => ({ symbolKey: "my-custom-symbol" })} update={update} />);
  expect(screen.getByRole("button", { name: "Symbol für F0 ergänzen" })).toBeDisabled();
  rerender(<RailComPlusSuggestions data={data} symbols={[symbol]} readonly edit={() => ({})} update={update} />);
  expect(screen.getByRole("button", { name: "Symbol für F0 ergänzen" })).toBeDisabled();
  expect(update).not.toHaveBeenCalled();
});

it("requires unambiguous numeric symbol metadata", () => {
  expect(exactECoSSymbol(9, [{ ...symbol, key: "symbol-9", metadata: {} }])).toBeUndefined();
  expect(exactECoSSymbol(9, [symbol, { ...symbol, key: "duplicate" }])).toBeUndefined();
  expect(exactECoSSymbol(9, [{ ...symbol, metadata: { ecosCode: 9, esuCode: 3 } }])).toBeUndefined();
  expect(exactECoSSymbol(9, [{ ...symbol, active: false }])).toBeUndefined();
  expect(exactECoSSymbol(9, [{ ...symbol, metadata: { esuCode: "9" } }])?.key).toBe("shunting");
  render(<RailComPlusSuggestions data={data} symbols={[]} readonly={false} edit={() => ({})} update={vi.fn()} />);
  expect(screen.getByRole("button", { name: "Symbol für F0 ergänzen" })).toBeDisabled();
  expect(screen.getByText(/empirischen Vergleichsdefinition/)).toBeVisible();
});

it("uses the same interpretation in import preview without offering writes", () => {
  render(<RailComPlusPreview values={values} />);
  fireEvent.click(screen.getByText("RailComPlus-Daten der Importdatei prüfen"));
  expect(screen.getByText("Rangiergang", { exact: true })).toBeVisible();
  expect(screen.getByText(/CV257=96/)).toBeVisible();
  expect(screen.queryByRole("button")).not.toBeInTheDocument();
});

it("does not hide a duplicate byte when interpreting an import preview", () => {
  const duplicate = [...values, values[1]];
  const preview = buildCVImportPreview("duplicate.json", duplicate, []);
  expect(preview.rows.at(-1)?.status).toBe("invalid");
  const { container } = render(<CVImportRowsTable preview={preview} saving={false} readonly={false} toggle={vi.fn()} />);
  fireEvent.click(screen.getByText("RailComPlus-Daten der Importdatei prüfen"));
  expect(container.querySelectorAll(".railcom-plus article")).toHaveLength(0);
  expect(screen.getByText(/32 unvollständige oder widersprüchliche Bytepaare/)).toBeVisible();
});

it("shows a known export date even without a recorded source line", () => {
  const dated = decoderAnalysis(values.map(row => row.cvNumber === 258
    ? { ...row, sourceExportedAt: "2026-10-17T17:41:59" } : row), label).railComPlus;
  render(<RailComPlusSuggestions data={dated} />);
  expect(screen.getByText(/Exportdatum: 17\.10\.2026 17:41:59/)).toBeVisible();
  expect(screen.queryByText(/Originalzeile/)).not.toBeInTheDocument();
});

const verified = (value = 176, functionKey: "F0" | "F1" | "F3" = "F1", exportedAt = "2026-10-20T01:41:03") => {
  const { firstCV, firstLine, icon } = {
    F0: { firstCV: 257, firstLine: 175, icon: 3 },
    F1: { firstCV: 259, firstLine: 177, icon: 2 },
    F3: { firstCV: 263, firstLine: 181, icon: 9 }
  }[functionKey];
  return decoderAnalysis([
  { cvNumber: 8, value: 151, decoderProfile: label },
  { cvNumber: 7, value: 255, decoderProfile: label,
    description: "LokProgrammer Decoder: LokPilot 5 micro DCC (5.14.183)" },
  { cvNumber: firstCV, value, indexCV31: 1, indexCV32: 1, decoderProfile: label,
    sourceLine: firstLine, sourceExportedAt: exportedAt },
  { cvNumber: firstCV + 1, value: icon, indexCV31: 1, indexCV32: 1, decoderProfile: label,
    sourceLine: firstLine + 1, sourceExportedAt: exportedAt }
  ], label).railComPlus!;
};

it.each(["de", "en"] as const)("retains field-specific provenance in %s without changing unrelated fields", language => {
  const result = verified();
  const patch = railComPlusFlagPatch(result, result.functions[0],
    { notes: "Eigene Notiz", name: "Meine Funktion", symbolKey: "custom", directionDependent: true },
    (key, args) => translate(language, key, args));
  expect(patch).toMatchObject({ mode: "moment", inverted: true, flagsVersion: 2 });
  expect(patch?.notes).toContain("Eigene Notiz\nRailComPlus");
  for (const text of ["CV31=1 / CV32=1", "CV259=176; CV260=2", "177 / 178", "2026-10-20T01:41:03",
    "LokPilot 5 micro DCC", "5.14.183", "5.2.18", "F0", "F1", "F3",
    language === "de" ? "Licht/Physikalisch/Logisch" : "Light/Physical/Logical"]) expect(patch?.notes).toContain(text);
  for (const key of ["name", "symbolKey", "directionDependent", "legacyDirectionDependent"]) {
    expect(patch).not.toHaveProperty(key);
  }
});

it.each([
  [96, "dauer", false, "2026-10-20T02:52:38"], [224, "moment", false, "2026-10-20T02:53:04"],
  [240, "moment", true, "2026-10-20T02:54:19"], [112, "dauer", true, "2026-10-20T02:54:43"]
])("applies F3 Logical export %i with all three comparison references", (value, mode, inverted, exportedAt) => {
  const update = vi.fn(), result = verified(value, "F3", exportedAt);
  render(<RailComPlusSuggestions data={result} readonly={false}
    edit={() => ({ notes: "Eigene Notiz", symbolKey: "custom" })} update={update} />);
  expect(screen.getByText(/Je vier Zustände von F0 \(Licht\), F1 \(Physikalisch\) und F3 \(Logisch\)/)).toBeVisible();
  fireEvent.click(screen.getByRole("button", { name: "Schaltart / Invertierung für F3 ergänzen" }));
  expect(update).toHaveBeenCalledOnce();
  expect(update).toHaveBeenCalledWith("F3", expect.objectContaining({ mode, inverted, flagsVersion: 2 }));
  const patch = update.mock.calls[0][1];
  for (const text of ["Eigene Notiz\nRailComPlus", `CV263=${value}; CV264=9`, "181 / 182", exportedAt,
    "F0-/F1-/F3-Vergleich (Licht/Physikalisch/Logisch)"]) expect(patch.notes).toContain(text);
  for (const field of ["symbolKey", "name", "directionDependent", "legacyDirectionDependent"]) {
    expect(patch).not.toHaveProperty(field);
  }
});

it.each(["de", "en"] as const)("keeps saved F3 settings and historical notes intact in %s", language => {
  const result = verified(240, "F3"), translateKey = (key: string, args?: Record<string, string | number>) =>
    translate(language, key, args);
  const note = "Eigene Notiz\nRailComPlus · F1-Vergleich";
  const patch = railComPlusFlagPatch(result, result.functions[0],
    { mode: "dauer", inverted: null, persisted: true, notes: note }, translateKey);
  expect(patch).toMatchObject({ inverted: true });
  expect(patch).not.toHaveProperty("mode");
  expect(patch?.notes).toContain(note + "\nRailComPlus");
  expect(patch?.notes).toContain(language === "de" ? "F0-/F1-/F3-Vergleich" : "F0/F1/F3 comparisons");
  expect(railComPlusFlagPatch(result, result.functions[0],
    { mode: "dauer", inverted: false, persisted: true, notes: note }, translateKey)).toBeUndefined();
});

it.each([
  [0, "dauer", false, "2026-10-20T03:14:07"], [128, "moment", false, "2026-10-20T03:14:39"],
  [144, "moment", true, "2026-10-20T03:15:02"], [16, "dauer", true, "2026-10-20T03:15:37"]
])("applies F0 Light export %i with source evidence without changing its name or symbol",
  (value, mode, inverted, exportedAt) => {
    const update = vi.fn(), result = verified(value, "F0", exportedAt);
    render(<RailComPlusSuggestions data={result} readonly={false}
      edit={() => ({ name: "Mein Fahrlicht", symbolKey: "light", mode: "dauer", inverted: null })} update={update} />);
    expect(screen.getByText(/F0 \(Licht\), F1 \(Physikalisch\) und F3 \(Logisch\)/)).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "Schaltart / Invertierung für F0 ergänzen" }));
    expect(update).toHaveBeenCalledOnce();
    expect(update).toHaveBeenCalledWith("F0", expect.objectContaining({ mode, inverted, flagsVersion: 2 }));
    const patch = update.mock.calls[0][1];
    for (const text of [`CV257=${value}; CV258=3`, "175 / 176", exportedAt,
      "F0-/F1-/F3-Vergleich (Licht/Physikalisch/Logisch)"]) expect(patch.notes).toContain(text);
    for (const field of ["symbolKey", "name", "directionDependent", "legacyDirectionDependent"]) {
      expect(patch).not.toHaveProperty(field);
    }
  });

it.each(["de", "en"] as const)("protects saved and manual F0 settings and preserves older evidence in %s", language => {
  const result = verified(144, "F0", "2026-10-20T03:15:02");
  const t = (key: string, args?: Record<string, string | number>) => translate(language, key, args);
  const notes = "Eigene Notiz\nRailComPlus · F1-/F3-Vergleich (Physikalisch/Logisch)";
  const current = { mode: "dauer" as const, inverted: null, persisted: true, notes };
  const before = structuredClone(current);
  const patch = railComPlusFlagPatch(result, result.functions[0], current, t);
  expect(patch).toMatchObject({ inverted: true });
  expect(patch).not.toHaveProperty("mode");
  expect(patch?.notes).toContain(notes + "\nRailComPlus");
  expect(patch?.notes).toContain(language === "de"
    ? "F0-/F1-/F3-Vergleich (Licht/Physikalisch/Logisch)" : "F0/F1/F3 comparisons (Light/Physical/Logical)");
  expect(patch?.notes).not.toContain(language === "de" ? "Schaltart:" : "Mode:");
  expect(current).toEqual(before);
  for (const inverted of [false, true]) {
    expect(railComPlusFlagPatch(result, result.functions[0], { ...current, inverted }, t)).toBeUndefined();
    expect(railComPlusFlagPatch(result, result.functions[0],
      { mode: "dauer", modeAssigned: true, inverted, notes }, t)).toBeUndefined();
  }
});

it("offers independent flag adoption even when the symbol is already assigned or unavailable", () => {
  const update = vi.fn();
  const { rerender } = render(<RailComPlusSuggestions data={verified()} symbols={[]} readonly={false}
    edit={() => ({ symbolKey: "custom", mode: "dauer", inverted: null })} update={update} />);
  fireEvent.click(screen.getByRole("button", { name: "Schaltart / Invertierung für F1 ergänzen" }));
  expect(update).toHaveBeenCalledWith("F1", expect.objectContaining({ mode: "moment", inverted: true }));
  rerender(<RailComPlusSuggestions data={verified(32)} symbols={[]} readonly={false}
    edit={() => ({ symbolKey: "custom", mode: "moment", inverted: null, persisted: true })} update={update} />);
  fireEvent.click(screen.getByRole("button", { name: "Schaltart / Invertierung für F1 ergänzen" }));
  const patch = update.mock.calls[1][1];
  expect(patch).toMatchObject({ inverted: false });
  expect(patch).not.toHaveProperty("mode");
  expect(patch.notes).not.toContain("Schaltart:");
});

it("protects explicit false inversion and manually selected latching mode before saving", () => {
  const result = verified();
  expect(railComPlusFlagPatch(result, result.functions[0],
    { mode: "dauer", modeAssigned: true, inverted: false }, (key, args) => translate("de", key, args))).toBeUndefined();
  const { result: controller } = renderHook(() => useVehicleFunctionsController({
    selected: null, form: emptyVehicle, setSaving: vi.fn(), onMessage: vi.fn(), refreshSelectedVehicle: vi.fn()
  }));
  act(() => controller.current.commands.updateFunctionEdit("F1", { mode: "moment" }));
  act(() => controller.current.commands.updateFunctionEdit("F1", { mode: "dauer" }));
  const patch = railComPlusFlagPatch(result, result.functions[0], controller.current.commands.functionEdit("F1"),
    (key, args) => translate("de", key, args));
  expect(patch).not.toHaveProperty("mode");
  expect(patch?.inverted).toBe(true);
  act(() => controller.current.commands.updateFunctionEdit("F1", patch!));
  expect(railComPlusFlagPatch(result, result.functions[0], controller.current.commands.functionEdit("F1"),
    (key, args) => translate("de", key, args))).toBeUndefined();
});

it("disables writes for viewers and does not enable flag adoption without firmware proof", () => {
  const update = vi.fn();
  const { rerender } = render(<RailComPlusSuggestions data={verified()} readonly edit={() => ({})} update={update} />);
  expect(screen.getByRole("button", { name: "Schaltart / Invertierung für F1 ergänzen" })).toBeDisabled();
  rerender(<RailComPlusSuggestions data={data} readonly={false} edit={() => ({})} update={update} />);
  expect(screen.queryByRole("button", { name: /Schaltart \/ Invertierung/ })).not.toBeInTheDocument();
  expect(update).not.toHaveBeenCalled();
});

it("protects long manual notes and never produces an unsaveable partial adoption", () => {
  const update = vi.fn();
  render(<RailComPlusSuggestions data={verified()} readonly={false} edit={() => ({ notes: "ä".repeat(400) })}
    update={update} />);
  expect(screen.getByRole("button", { name: "Schaltart / Invertierung für F1 ergänzen" })).toBeDisabled();
  expect(screen.getByText(/es wird nichts abgeschnitten/)).toBeVisible();
  expect(update).not.toHaveBeenCalled();
});
