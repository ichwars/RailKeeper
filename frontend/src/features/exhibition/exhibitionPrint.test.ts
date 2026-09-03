import { afterEach, describe, expect, it, vi } from "vitest";

import { ExhibitionList, ExhibitionWorkspaceEntry } from "../../shared/api";
import { buildExhibitionPrintHTML, printFunctionChips, printList } from "./exhibitionPrint";

const list: ExhibitionList = {
  id: "event", designation: "Modellbahntage Köln", date: "2026-08-31", endDate: "2026-09-05",
  location: "Vereinshalle", description: "Betrieb mit Gästen", organizationNotes: "Aufbau ab 08:00",
  status: "locked", locked: true, lockReason: "Planung abgeschlossen", entryCount: 1, revision: 3,
  createdAt: "2026-08-01", updatedAt: "2026-08-30"
};
const entry: ExhibitionWorkspaceEntry = {
  id: "entry", listId: list.id, locomotiveName: "BR 103 113-7", owner: "Fahrgemeinschaft Köln",
  manufacturer: "Roco", series: "103", gattung: "Elektrolokomotive", epoch: "IV", railwayCompany: "DB",
  dtDecoder: true, decoderType: "LokSound 5", decoderNumber: "103", sxAddress: "42",
  adapter: "PluX22", interfaceName: "ECoS", analog: false, availability: "unavailable",
  dayScope: "day1,day5", notes: "Kupplung prüfen\nNur im äußeren Gleis", sortOrder: 2,
  functionKeys: JSON.stringify([{ key: "F31", name: "Bahnhofsansage", type: "sound", symbolKey: "speaker" }]),
  imageUrl: "/vehicle.png", revision: 1, status: "unavailable", conflictIds: [],
  createdAt: "2026-08-01", updatedAt: "2026-08-30"
};
const t = (key: string, values?: Record<string, string | number>) => `${key}${values ? ` ${values.count}` : ""}`;

describe("exhibition print report", () => {
  afterEach(() => {
    document.querySelectorAll("iframe").forEach((frame) => frame.remove());
    vi.useRealTimers();
  });

  it("includes all operational fields, event metadata and saved functions", () => {
    const html = buildExhibitionPrintHTML(list, [entry], [], "de", t);
    const doc = new DOMParser().parseFromString(html, "text/html");
    for (const value of [
      list.designation, "31.8.2026", "5.9.2026", list.location, list.description, list.organizationNotes,
      list.lockReason, entry.locomotiveName, entry.owner, entry.manufacturer, entry.series, entry.gattung,
      entry.epoch, entry.railwayCompany, entry.decoderType, entry.decoderNumber, entry.sxAddress,
      entry.adapter, entry.interfaceName, "Nicht verfügbar", "Tag 1, Tag 5", entry.notes,
      "F31", "Bahnhofsansage", "sound", "speaker"
    ]) expect(doc.body.textContent).toContain(value);
    expect(doc.querySelector("img")?.getAttribute("src")).toBe(entry.imageUrl);
    expect(doc.querySelectorAll("tbody.entry")).toHaveLength(1);
    expect(doc.querySelectorAll("button, nav, aside")).toHaveLength(0);
    expect(html).toContain("@page { size: A4 landscape");
    expect(html).toContain("table-header-group");
    expect(html).toContain("break-inside: avoid");
    expect(html).toContain("white-space: pre-wrap");
    expect(html).not.toContain("white-space: nowrap");
  });

  it("keeps imported text inert and rejects executable image URLs", () => {
    const payload = '<img src=x onerror="alert(1)">';
    const html = buildExhibitionPrintHTML({ ...list, designation: payload }, [{
      ...entry, owner: payload, notes: payload, imageUrl: "javascript:alert(1)",
      functionKeys: JSON.stringify([{ key: "F1", name: payload, type: "sound" }])
    }], [], "de", t);
    const doc = new DOMParser().parseFromString(html, "text/html");
    expect(doc.querySelectorAll("img, script, [onerror]")).toHaveLength(0);
    expect(doc.body.textContent).toContain(payload);
    expect(doc.querySelector('meta[http-equiv="Content-Security-Policy"]')?.getAttribute("content"))
      .toContain("default-src 'none'");
  });

  it("preserves legacy or malformed function text without adding unsaved F0 defaults", () => {
    expect(printFunctionChips("F2: Horn\nF31: Ansage", [])).toContain("F2: Horn\nF31: Ansage");
    expect(printFunctionChips("F2: Horn", [])).not.toContain("F0");
    expect(printFunctionChips("", [])).toBe("–");
    expect(printFunctionChips("[]", [])).toBe("–");
    expect(printFunctionChips('[null,{"name":"Unvollständig"}]', [])).toContain("Unvollständig");
  });

  it("supports empty lists, English and the existing image-free option", () => {
    expect(buildExhibitionPrintHTML(list, [], [], "en", t)).toContain("exhibition.printEmpty");
    const html = buildExhibitionPrintHTML(list, [entry], [], "en", t, { includeImages: false });
    expect(html).toContain("Day 1, Day 5");
    expect(html).toContain("Unavailable");
    expect(html).toContain("Organisation notes");
    expect(html).not.toContain("<img");
  });

  it("prints only its loaded document and retains it until the print dialog closes", async () => {
    vi.useFakeTimers();
    const pagePrint = vi.spyOn(window, "print").mockImplementation(() => {});
    const pending = printList(list, [entry], [], "de", t);
    const frame = document.querySelector("iframe");
    expect(frame?.srcdoc).toContain(list.designation);
    const printWindow = frame!.contentWindow!;
    vi.spyOn(printWindow, "focus").mockImplementation(() => {});
    const print = vi.spyOn(printWindow, "print").mockImplementation(() => {});
    expect(print).not.toHaveBeenCalled();
    frame!.dispatchEvent(new Event("load"));
    await pending;
    expect(print).toHaveBeenCalledOnce();
    expect(pagePrint).not.toHaveBeenCalled();
    vi.advanceTimersByTime(2_000);
    expect(frame?.isConnected).toBe(true);
    printWindow.dispatchEvent(new Event("afterprint"));
    expect(frame?.isConnected).toBe(false);
  });

  it("cleans up and rejects when the print document fails to load", async () => {
    vi.useFakeTimers();
    const pending = printList(list, [entry], [], "de", t);
    const rejected = expect(pending).rejects.toThrow("Druckansicht konnte nicht geladen werden");
    vi.advanceTimersByTime(15_000);
    await rejected;
    expect(document.querySelector("iframe")).toBeNull();
  });
});
