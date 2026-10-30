import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { bundledDecoderInventory, bundledDecoderProfiles } from "../../shared/decoderLibraryModel";
import { cvValueFixture } from "../../test/fixtures/vehicles";
import { translate } from "../../shared/i18n";
import { decoderAnalysis } from "./decoderAnalysis";
import { cvValueInfo } from "./cvValueInfo";
import { VehicleCVDetails } from "./VehicleCVDetails";
import { VehicleCVValuesTable } from "./VehicleCVValuesTable";
import { DecoderLibraryNotice } from "./DecoderLibraryNotice";

const type = "LokPilot 5 micro DCC";
const provenance = "ECoS @ 192.168.1.150 · #1002 · Cache/Speicher · CV31/CV32: ?";
const reading = (cvNumber: number, value: number) => cvValueFixture({
  id: `ecos-${cvNumber}`, cvNumber, value, decoderProfile: "ECoS", protocol: "DCC28", description: provenance
});
const values = [reading(8, 151), reading(1, 3), reading(29, 14), reading(27, 3), reading(112, 42),
  reading(257, 32), reading(291, 31)];
const analyze = (rows = values, decoderType = type, catalog = bundledDecoderProfiles) =>
  decoderAnalysis(rows, "ECoS", catalog, bundledDecoderInventory, decoderType);
const t = (key: string, args?: Record<string, string | number>) => translate("de", key, args);

describe("ECoS decoder-library interpretation", () => {
  it("explains supplied index pages through the cached model profile while keeping missing CV8 absent", () => {
    const rows = [{ ...reading(257, 20), indexCV31: 16, indexCV32: 3 },
      { ...reading(257, 1), indexCV31: 16, indexCV32: 8 }];
    const a = decoderAnalysis(rows, "ECoS", bundledDecoderProfiles, bundledDecoderInventory, "", {
      name: type, objectId: "1002", host: "ecos.test", readAt: "2026-10-27T07:00:00Z"
    });
    expect(a.definition(rows[0])?.definition.name.de).toContain("Mapping-Bedingung A");
    expect(a.definition(rows[1])?.definition.name.de).toContain("Mapping-Aktion K");
    expect(a.get(8)).toBeUndefined();
    render(<DecoderLibraryNotice values={rows} profile="ECoS" match={a.libraryMatch} />);
    expect(screen.getByText(/Zuordnung über das gespeicherte ECoS-Profil/)).toBeVisible();
    expect(screen.queryByText(/CV8=151/)).not.toBeInTheDocument();
  });

  it("uses supplied database pages exactly and keeps missing decoder identity unresolved", () => {
    const rows = [
      { ...reading(257, 20), indexCV31: 16, indexCV32: 3 },
      { ...reading(257, 1), indexCV31: 16, indexCV32: 8 }
    ];
    const identified = analyze([reading(8, 151), ...rows]);
    expect(identified.definition(rows[0])?.definition.name.de).toContain("Mapping-Bedingung A");
    expect(identified.definition(rows[1])?.definition.name.de).toContain("Mapping-Aktion K");
    expect(identified.indexPages).toBe(2);
    const missing = analyze(rows);
    expect(missing.modelRules).toBe(false);
    expect(missing.definition(rows[0])).toBeUndefined();
    expect(missing.indexUnknown(rows[0])).toBe(false);
  });
  it("separates cache source from a vehicle model confirmed by CV8", () => {
    const a = analyze();
    expect(a.libraryMatch).toMatchObject({ status: "matched", matchedBy: "name",
      profileId: "esu-lokpilot-5-micro-dcc", totalValues: values.length, knownValues: 5 });
    expect(a.abc).toBe(true);
    expect(a.address).toBe(3);
    expect(a.definition(values[4])?.definition.factoryDefault).toBe(20);
    expect(a.get(112)).toBe(42);
    expect(a.get(254)).toBeUndefined();
    render(<DecoderLibraryNotice values={values} profile="ECoS" match={a.libraryMatch}
      vehicleDecoderType={a.vehicleDecoderType} />);
    expect(screen.getByText(/hinterlegten Decodertyp und CV8=151/)).toBeVisible();
    expect(screen.getByText(/keine neue Decoderlesung/)).toBeVisible();
  });

  it("does not borrow identity or readings from other profiles, or trust a name without CV8", () => {
    expect(analyze(values, "").libraryMatch.status).toBe("manufacturer");
    expect(analyze(values.filter(row => row.cvNumber !== 8)).libraryMatch.status).toBe("suggested");
    const different = values.map(row => ({ ...row, id: `other-${row.id}`, decoderProfile: type }));
    const a = analyze([...values.filter(row => row.cvNumber !== 8), ...different]);
    expect(a.libraryMatch.status).toBe("suggested");
    expect(a.abc).toBeUndefined();
    expect(analyze(values.map(row => row.cvNumber === 8 ? { ...row, value: 99 } : row))
      .libraryMatch.status).toBe("conflict");
    expect(analyze([...values, { ...values[0], id: "conflict", value: 99 }]).libraryMatch.status).toBe("conflict");
    expect(analyze(values.map(row => ({ ...row, protocol: "MM" }))).libraryMatch.status).toBe("unsupported");
    const disabled = bundledDecoderProfiles.map(p => ({ ...p, active: false }));
    expect(analyze(values, type, disabled).libraryMatch.status).toBe("unavailable");
  });

  it("does not infer any missing index page from the decoder, current selector bytes or a sole documented page", () => {
    const a = analyze([...values, reading(31, 16), reading(32, 3)]);
    expect(a.definition(values[5])).toBeUndefined();
    expect(a.indexUnknown(values[5])).toBe(true);
    expect(a.get(257)).toBeUndefined();
    expect(a.indexPages).toBe(0);
    expect(a.definition({ ...values[5], indexCV31: 16, indexCV32: 3 })?.definition.name.de)
      .toContain("Mapping-Bedingung A");
    expect(a.definition({ ...values[5], indexCV31: 16, indexCV32: 8 })?.definition.name.de)
      .toContain("Mapping-Aktion K");
    expect(a.definition({ ...values[5], indexCV31: 0, indexCV32: 0 })).toBeUndefined();
    expect(a.indexUnknown({ ...values[5], indexCV32: 3 })).toBe(true);
    const catalog = structuredClone(bundledDecoderProfiles);
    const model = catalog.find(p => p.id === a.libraryMatch.profileId)!;
    const definition = model.definitions.find(d => d.cv === 112)!;
    model.definitions.push({ ...definition, cv: 900, indexCV31: 16, indexCV32: 99 });
    expect(analyze(values, type, catalog).definition(reading(900, 20))).toBeUndefined();
  });

  it("accepts explicitly unindexed higher registers and rejects a number shared with indexed registers", () => {
    const catalog = structuredClone(bundledDecoderProfiles);
    const model = catalog.find(p => p.id === "esu-lokpilot-5-micro-dcc")!;
    const definition = model.definitions.find(d => d.cv === 112)!;
    model.definitions.push({ ...definition, cv: 900 }, { ...definition, cv: 901 },
      { ...definition, cv: 901, indexCV31: 16, indexCV32: 0 });
    const a = analyze(values, type, catalog);
    expect(a.definition(reading(900, 0))).toBeDefined();
    expect(a.indexUnknown(reading(900, 0))).toBe(false);
    expect(a.definition(reading(901, 0))).toBeUndefined();
    expect(a.indexUnknown(reading(901, 0))).toBe(true);
  });

  it("preserves all raw and indexed records and does not synthesize missing values or provenance", () => {
    const indexed = Array.from({ length: 1367 }, (_, i) => cvValueFixture({ id: `txt-${i}`, decoderProfile: type,
      cvNumber: 257 + i % 256, value: i % 256, indexCV31: 16, indexCV32: Math.floor(i / 256),
      sourceLine: i + 4, sourceExportedAt: "2026-10-17T17:41:59" }));
    const rows = [...values, ...indexed], before = structuredClone(rows);
    const a = analyze(rows);
    for (const value of values) cvValueInfo(value, a, t);
    expect(a.rows).toHaveLength(values.length);
    expect(a.railComPlus?.functions || []).toHaveLength(0);
    expect(rows).toEqual(before);
    expect(values.every(row => row.sourceLine == null && row.sourceExportedAt == null)).toBe(true);
  });

  it("uses stored CV-based model recognition without requiring a vehicle label", () => {
    const a = analyze([reading(8, 97), reading(261, 180)], "");
    expect(a.libraryMatch).toMatchObject({ status: "matched", matchedBy: "cv", profileId: "dh-dh18a" });
    expect(a.definition(reading(261, 180))).toBeDefined();
    expect(a.indexUnknown(reading(261, 180))).toBe(false);
  });
});

describe("ECoS CV presentation", () => {
  it("shows meanings, distinguishes base from unknown pages and edits the original record", () => {
    const edit = vi.fn(), explain = vi.fn();
    render(<VehicleCVValuesTable values={values} decoderType={type} readonly={false} saving={false}
      edit={edit} remove={vi.fn()} onExplain={explain} />);
    expect(screen.getByText("Bremsverfahren")).toBeVisible();
    const row = screen.getByText("CV 257").closest("tr")!;
    expect(within(row).getByText("Zuordnung ungeklärt")).toBeVisible();
    expect(within(row).getByText("Indexseite unbekannt")).toBeVisible();
    expect(within(row).queryByText("Basis (ohne Index)")).not.toBeInTheDocument();
    fireEvent.click(within(row).getByRole("button", { name: "CV bearbeiten" }));
    expect(edit).toHaveBeenCalledExactlyOnceWith(values[5]);
    fireEvent.click(within(row).getByRole("button", { name: "CV 257 erklären" }));
    expect(explain).toHaveBeenCalledExactlyOnceWith(values[5]);
    fireEvent.click(screen.getByRole("button", { name: "Indexseite" }));
    fireEvent.click(screen.getByRole("option", { name: "Indexseite unbekannt" }));
    expect(screen.getByRole("status")).toHaveTextContent("1–2 von 2 Werten");
    expect(screen.queryByText("CV 1")).not.toBeInTheDocument();
  });

  it("exposes actual cache provenance and manufacturer documentation as separate evidence", () => {
    render(<VehicleCVDetails value={values[4]} values={values} files={[]} decoderType={type}
      readonly={false} saving={false} onEdit={vi.fn()} onClose={vi.fn()} />);
    expect(screen.getAllByText(provenance).find(element => element.tagName === "DD")).toBeVisible();
    expect(screen.getByText("0x2A")).toBeVisible();
    expect(screen.getByText("S.85, CV112")).toBeVisible();
    expect(screen.getAllByText("Nicht erfasst")).toHaveLength(2);
    expect(screen.getByText("20")).toBeVisible();
    expect(screen.getByText(/kein ausgelesener Fahrzeugwert/)).toBeVisible();
    expect(screen.getByText(/keine neue Decoderlesung/)).toBeVisible();
  });

  it("keeps unknown page details free of guessed bit meanings and documentation hints", () => {
    render(<VehicleCVDetails value={values[5]} values={values} files={[]} decoderType={type}
      readonly={true} saving={false} onEdit={vi.fn()} onClose={vi.fn()} />);
    expect(screen.getByRole("heading", { name: "CV 257 · Zuordnung ungeklärt" })).toBeVisible();
    expect(screen.getByText("Indexseite unbekannt")).toBeVisible();
    expect(screen.getByText("0x20")).toBeVisible();
    expect(screen.getByText(/fehlt die vollständige Indexseite/)).toBeVisible();
    expect(screen.queryByText("Bit 0")).not.toBeInTheDocument();
    expect(screen.queryAllByRole("link")).toHaveLength(0);
  });
});
