import { expect, it } from "vitest";
import type { MasterDataEntry, VehicleCVValueInput } from "../../shared/api";
import { translate } from "../../shared/i18n";
import { decoderAnalysis } from "./decoderAnalysis";
import { ecosFunctionSuggestion } from "./ecosFunctionSuggestions";
import { emptyFunctionEdit } from "./vehicleViewModel";

const profile = "LokPilot 5 micro DCC", readAt = "2026-10-26T08:00:00Z";
const t = (key: string, args?: Record<string, string | number>) => translate("de", key, args);
const symbol = (key: string, code: number): MasterDataEntry => ({ id: key, key, type: "symbols", label: key,
  active: true, sortOrder: 0, createdAt: "", updatedAt: "", metadata: { ecosCode: code } });
const symbols = [symbol("light", 3), symbol("shunt", 9)];
const cv = (cvNumber: number, value: number): VehicleCVValueInput => ({ cvNumber, value, decoderProfile: profile });
const values = [cv(8, 151), { ...cv(7, 255), description: `LokProgrammer Decoder: ${profile} (5.14.183)` },
  { ...cv(257, 0), indexCV31: 1, indexCV32: 1 }, { ...cv(258, 3), indexCV31: 1, indexCV32: 1 }];
const value = { index: 0, description: 3, momentary: false };

it("uses the active decoder library and records both sources without inventing direction", () => {
  const analysis = decoderAnalysis(values, profile);
  expect(analysis.railComPlus?.flagsVerification).toBeDefined();
  const result = ecosFunctionSuggestion(value, analysis, {}, symbols, readAt, 1002, t);
  expect(result.status).toBe("ready");
  expect(result.patch).toMatchObject({ name: "light", symbolKey: "light", mode: "dauer", inverted: false });
  expect(result.patch?.directionDependent).toBeUndefined();
  expect(result.patch?.notes).toContain("RailComPlus");
  expect(result.patch?.notes).toContain("ECoS · #1002 · F0 · funcexists=3");
});

it.each([{ description: 9, momentary: false }, { description: 4095, momentary: false },
  { description: 3, momentary: true }])(
  "blocks contradictory library and ECoS symbols or modes", different => {
    const result = ecosFunctionSuggestion({ index: 0, ...different }, decoderAnalysis(values, profile),
      emptyFunctionEdit("F0"), symbols, readAt, 1002, t);
    expect(result.status).toBe("conflict"); expect(result.patch).toBeUndefined();
  });

it("preserves manual names, symbols, modes, flags, deleted rows and original notes", () => {
  const analysis = decoderAnalysis(values, profile);
  const current = { persisted: true, name: "Mein Licht", symbolKey: "manual", mode: "moment", inverted: true,
    directionDependent: false, notes: "Meine Notiz", revision: "a".repeat(64) };
  const original = structuredClone(current);
  const result = ecosFunctionSuggestion(value, analysis, current, symbols, readAt, 1002, t);
  expect(result.status).toBe("protected"); expect(result.patch).toBeUndefined(); expect(current).toEqual(original);
  const partial = ecosFunctionSuggestion(value, analysis, { ...current, symbolKey: "" }, symbols, readAt, 1002, t);
  expect(partial.patch).toMatchObject({ symbolKey: "light" });
  for (const field of ["name", "mode", "inverted", "directionDependent"]) expect(partial.patch).not.toHaveProperty(field);
  expect(partial.patch?.notes).toMatch(/^Meine Notiz\n/);
  expect(ecosFunctionSuggestion(value, analysis, { deleted: true }, symbols, readAt, 1002, t).patch).toBeUndefined();
});

it("keeps unknown symbols and unindexed cache entries unresolved without borrowing index pages", () => {
  const rows = values.map(row => ({ ...row, decoderProfile: "ECoS", indexCV31: null, indexCV32: null }));
  const result = ecosFunctionSuggestion({ ...value, description: 4095 }, decoderAnalysis(rows, "ECoS", undefined,
    undefined, profile), emptyFunctionEdit("F0"), symbols, readAt, 1002, t);
  expect(result.library?.symbolKey).toBeUndefined(); expect(result.stationSymbol).toBeUndefined();
  expect(result.patch?.symbolKey).toBeUndefined(); expect(result.patch?.inverted).toBeUndefined();
  expect(result.patch?.mode).toBe("dauer");
});

it("rejects ambiguous or inactive symbol metadata and overflowing provenance", () => {
  const analysis = decoderAnalysis([], "");
  for (const list of [[...symbols, symbol("duplicate", 3)], [{ ...symbols[0], active: false }]]) {
    expect(ecosFunctionSuggestion(value, analysis, emptyFunctionEdit("F0"), list, readAt, 1002, t)
      .patch?.symbolKey).toBeUndefined();
  }
  const result = ecosFunctionSuggestion(value, analysis, { notes: "ä".repeat(495) }, symbols, readAt, 1002, t);
  expect(result.status).toBe("notesFull"); expect(result.patch).toBeUndefined();
});

it.each([7431, 2147483647])("keeps opaque station descriptor %i without inventing a symbol or losing key mode", description => {
  const result = ecosFunctionSuggestion({ index: 3, description, momentary: true }, decoderAnalysis([], "ECoS"),
    {}, symbols, readAt, 1002, t);
  expect(result.status).toBe("ready");
  expect(result.stationSymbol).toBeUndefined();
  expect(result.patch).toMatchObject({ mode: "moment", flagsVersion: 2 });
  expect(result.patch).not.toHaveProperty("name");
  expect(result.patch).not.toHaveProperty("symbolKey");
  expect(result.patch?.notes).toContain(`funcexists=${description}`);
});
