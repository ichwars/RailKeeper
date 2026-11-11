import { act, renderHook, waitFor } from "@testing-library/react";
import type { SetStateAction } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError, api, type ExhibitionEntry, type ExhibitionList, type Vehicle } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { useVehicleOutputController } from "./useVehicleOutputController";
import * as qr from "./vehicleQr";
import * as reports from "./vehicleReports";
import { vehicleToForm } from "./vehicleTransforms";

const list: ExhibitionList = { id: "event", designation: "Clubabend", date: "2026-11-07", endDate: "2026-11-07", status: "open", revision: 1, locked: false, entryCount: 0, createdAt: "", updatedAt: "" };
function entry(overrides: Partial<ExhibitionEntry> = {}): ExhibitionEntry {
  return { id: "entry", listId: list.id, owner: "operator", locomotiveName: "Other vehicle", decoderNumber: "7", dayScope: "all", availability: "available", revision: 1, dtDecoder: true, analog: false, sortOrder: 1, createdAt: "", updatedAt: "", ...overrides };
}
function workspace(vehicles: Vehicle[] = [vehicleFixture()]) {
  let collection = vehicles;
  const onMessage = vi.fn();
  const setSelectedDetail = vi.fn();
  const setVehicles = vi.fn((next: SetStateAction<Vehicle[]>) => { collection = typeof next === "function" ? next(collection) : next; });
  const selected = vehicles[0] || null;
  const hook = renderHook(() => useVehicleOutputController({
    username: "operator", form: vehicleToForm(selected || vehicleFixture({ name: "", inventoryNumber: "" })), selected,
    query: "BR", sort: { key: "name", direction: "asc" }, sortedVehicles: vehicles, selectedVisibleVehicles: vehicles.slice(0, 1),
    setVehicles, setSelectedDetail, onMessage, t: (key) => key
  }));
  return { ...hook, onMessage, setSelectedDetail, collection: () => collection };
}

describe("Vehicle report, QR and exhibition output", () => {
  beforeEach(() => {
    vi.spyOn(qr, "buildQrSvg").mockResolvedValue("<svg>synthetic</svg>");
    vi.spyOn(qr, "buildBrandedQrPngDataUrl").mockResolvedValue("data:image/png;base64,c3ludGhldGlj");
    vi.spyOn(reports, "reservePrintDocument").mockReturnValue(null);
    vi.spyOn(reports, "openPrintDocument").mockImplementation(() => { });
    vi.spyOn(api, "vehicle").mockResolvedValue(vehicleFixture());
    vi.spyOn(api, "vehicleReportData").mockResolvedValue([vehicleFixture()]);
    vi.spyOn(api, "exhibitionLists").mockResolvedValue([list]);
    vi.spyOn(api, "exhibitionEntries").mockResolvedValue([]);
  });

  it("generates QR from the current record and reports rendering or printing failures", async () => {
    vi.spyOn(qr, "downloadQrSvgFile").mockImplementation(() => { });
    vi.spyOn(qr, "downloadQrPngFile").mockResolvedValue(undefined);
    vi.spyOn(qr, "printQrSvgLabel").mockImplementation(() => { throw new Error("Druckfenster blockiert"); });
    const { result } = workspace();
    await act(() => result.current.qr.commands.generate());
    expect(result.current.qr.state).toEqual({ open: true, svg: "<svg>synthetic</svg>", error: "" });
    act(() => result.current.qr.commands.downloadSvg());
    await act(() => result.current.qr.commands.downloadPng());
    expect(qr.downloadQrSvgFile).toHaveBeenCalledWith("<svg>synthetic</svg>", "RK-LOK-000001");
    act(() => result.current.qr.commands.print());
    expect(result.current.qr.state.error).toBe("Druckfenster blockiert");
    vi.mocked(qr.buildQrSvg).mockRejectedValueOnce(new Error("QR nicht erzeugt"));
    await act(() => result.current.qr.commands.generate());
    expect(result.current.qr.state.svg).toBe("");
    expect(result.current.qr.state.error).toBe("QR nicht erzeugt");
    act(() => result.current.qr.commands.close());
    expect(result.current.qr.state.open).toBe(false);
  });

  it("refreshes a row before creating its QR and does not open a failed row load", async () => {
    const { result, setSelectedDetail } = workspace();
    act(() => result.current.qr.commands.openForVehicle(vehicleFixture()));
    await waitFor(() => expect(result.current.qr.state.svg).toContain("<svg>"));
    expect(setSelectedDetail).toHaveBeenCalledWith(vehicleFixture());
    act(() => result.current.qr.commands.close());
    vi.mocked(api.vehicle).mockRejectedValueOnce(new Error("Fahrzeug nicht lesbar"));
    act(() => result.current.qr.commands.openForVehicle(vehicleFixture()));
    await waitFor(() => expect(result.current.qr.state.error).toBe("Fahrzeug nicht lesbar"));
    expect(result.current.qr.state.open).toBe(false);
  });

  it("builds the report from complete selected records and preserves report choices", async () => {
    const a = vehicleFixture();
    const b = vehicleFixture({ id: "vehicle-2", name: "Other vehicle" });
    const { result } = workspace([a, b]);
    act(() => {
      result.current.report.setters.setOpen(true);
      result.current.report.setters.setSelection("selected");
      result.current.report.setters.setMode("details");
      result.current.report.setters.setTitle("  Clubbestand  ");
      result.current.report.setters.setIncludeQRCode(false);
      result.current.report.setters.setIncludeImages(false);
    });
    await act(() => result.current.report.commands.create());
    expect(api.vehicleReportData).toHaveBeenCalledExactlyOnceWith([a.id]);
    expect(api.vehicle).not.toHaveBeenCalled();
    expect(qr.buildBrandedQrPngDataUrl).not.toHaveBeenCalled();
    const html = vi.mocked(reports.openPrintDocument).mock.calls[0][0];
    expect(html).toContain("Clubbestand");
    expect(html).toContain(a.inventoryNumber!);
    expect(html).not.toContain("Other vehicle");
    expect(result.current.report.state.open).toBe(false);
  });

  it("keeps report options available after a failed data read and handles empty selection", async () => {
    const { result, onMessage } = workspace();
    act(() => result.current.report.setters.setOpen(true));
    vi.mocked(api.vehicleReportData).mockRejectedValueOnce(new Error("Reportdaten fehlen"));
    await act(() => result.current.report.commands.create());
    expect(onMessage).toHaveBeenLastCalledWith("Reportdaten fehlen");
    expect(result.current.report.state).toMatchObject({ open: true, creating: false });
    expect(reports.openPrintDocument).not.toHaveBeenCalled();
    const empty = workspace([]);
    await act(() => empty.result.current.report.commands.create());
    expect(empty.onMessage).toHaveBeenCalledWith("Es gibt keine Fahrzeuge für den PDF-Report.");
  });

  it("prints a full vehicle including its QR and reports a failed print read", async () => {
    const { result, onMessage } = workspace();
    await act(() => result.current.report.commands.printVehicle(vehicleFixture()));
    expect(qr.buildBrandedQrPngDataUrl).toHaveBeenCalledOnce();
    expect(vi.mocked(reports.openPrintDocument).mock.calls[0][0]).toContain("data:image/png;base64,c3ludGhldGlj");
    vi.mocked(api.vehicle).mockRejectedValueOnce(new Error("Druckdaten fehlen"));
    await act(() => result.current.report.commands.printVehicle(vehicleFixture()));
    expect(onMessage).toHaveBeenLastCalledWith("Druckdaten fehlen");
  });

  it("excludes locked events and avoids importing a duplicate decoder", async () => {
    vi.mocked(api.exhibitionLists).mockResolvedValue([{ ...list, id: "locked", locked: true }, list]);
    vi.mocked(api.exhibitionEntries).mockResolvedValue([entry({ decoderNumber: "1001" })]);
    const create = vi.spyOn(api, "createExhibitionEntry");
    const { result } = workspace();
    act(() => result.current.exhibition.commands.toggle(vehicleFixture(), true));
    await waitFor(() => expect(result.current.exhibition.state.assignment?.loadingEntries).toBe(false));
    expect(result.current.exhibition.state.assignment?.lists).toEqual([list]);
    expect(result.current.exhibition.state.duplicateDecoder?.id).toBe("entry");
    act(() => result.current.exhibition.commands.save());
    expect(create).not.toHaveBeenCalled();
    act(() => result.current.exhibition.commands.close());
    expect(result.current.exhibition.state.assignment).toBeNull();
  });

  it("assigns a fresh complete record and updates the flag with its current revision", async () => {
    const original = vehicleFixture();
    vi.spyOn(api, "createExhibitionEntry").mockResolvedValue(entry());
    vi.spyOn(api, "updateVehicle").mockResolvedValue(vehicleFixture({ exhibition: true, images: undefined, attachments: undefined }));
    const { result, collection, setSelectedDetail } = workspace();
    act(() => result.current.exhibition.commands.toggle(original, true));
    await waitFor(() => expect(result.current.exhibition.state.assignment?.loadingEntries).toBe(false));
    act(() => result.current.exhibition.commands.save());
    await waitFor(() => expect(result.current.exhibition.state.assignment).toBeNull());
    expect(api.createExhibitionEntry).toHaveBeenCalledWith("event", expect.objectContaining({ vehicleId: original.id, locomotiveName: original.name, owner: "operator" }));
    expect(api.updateVehicle).toHaveBeenCalledWith(original.id, expect.objectContaining({ exhibition: true, expectedRevision: original.revision, images: undefined }));
    expect(collection()[0].images).toEqual(original.images);
    expect(setSelectedDetail).toHaveBeenCalledWith(expect.objectContaining({ exhibition: true }));
  });

  it("preserves the assignment after rejection without changing the vehicle flag", async () => {
    vi.spyOn(api, "createExhibitionEntry").mockRejectedValue(new Error("Liste inzwischen gesperrt"));
    const update = vi.spyOn(api, "updateVehicle");
    const { result } = workspace();
    act(() => result.current.exhibition.commands.toggle(vehicleFixture(), true));
    await waitFor(() => expect(result.current.exhibition.state.assignment?.loadingEntries).toBe(false));
    act(() => result.current.exhibition.commands.save());
    await waitFor(() => expect(result.current.exhibition.state.assignment?.error).toBe("Liste inzwischen gesperrt"));
    expect(update).not.toHaveBeenCalled();
    expect(api.createExhibitionEntry).toHaveBeenCalledOnce();
  });

  it("retains the newer event selection when an older entry read completes late", async () => {
    let finish: (rows: ExhibitionEntry[]) => void = () => { };
    vi.mocked(api.exhibitionEntries).mockReturnValueOnce(new Promise(resolve => { finish = resolve; })).mockResolvedValueOnce([entry({ id: "new-entry" })]);
    const { result } = workspace();
    act(() => result.current.exhibition.commands.toggle(vehicleFixture(), true));
    await waitFor(() => expect(api.exhibitionEntries).toHaveBeenCalledWith("event"));
    act(() => result.current.exhibition.commands.loadEntries("next-event"));
    await waitFor(() => expect(result.current.exhibition.state.assignment?.entries[0]?.id).toBe("new-entry"));
    await act(async () => finish([entry({ id: "old-entry" })]));
    expect(result.current.exhibition.state.assignment?.selectedListID).toBe("next-event");
    expect(result.current.exhibition.state.assignment?.entries[0]?.id).toBe("new-entry");
    act(() => result.current.exhibition.commands.loadEntries(""));
    expect(result.current.exhibition.state.assignment?.entries).toEqual([]);
  });

  it("reports no available event, lookup errors and unsupported analog vehicles", async () => {
    vi.mocked(api.exhibitionLists).mockResolvedValueOnce([]).mockRejectedValueOnce(new Error("Listen fehlen"));
    const { result, onMessage } = workspace();
    act(() => result.current.exhibition.commands.toggle(vehicleFixture({ digital: false }), true));
    expect(onMessage).toHaveBeenLastCalledWith("vehicles.exhibition.requiresDecoder");
    act(() => result.current.exhibition.commands.toggle(vehicleFixture(), true));
    await waitFor(() => expect(result.current.exhibition.state.assignment?.error).toBe("vehicles.exhibition.noOpenLists"));
    act(() => result.current.exhibition.commands.toggle(vehicleFixture(), true));
    await waitFor(() => expect(result.current.exhibition.state.assignment?.error).toBe("Listen fehlen"));
  });

  it("keeps a concurrent flag change visible instead of overwriting it", async () => {
    vi.spyOn(api, "updateVehicle").mockRejectedValue(new ApiError("changed", "vehicle_conflict", 409));
    const { result, onMessage, collection } = workspace();
    act(() => result.current.exhibition.commands.toggle(vehicleFixture(), false));
    await waitFor(() => expect(onMessage).toHaveBeenLastCalledWith("vehicles.mediaConflict.changed"));
    expect(collection()[0].exhibition).toBe(false);
  });
});
