import { useState } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { ECoSStoredCVPanel } from "./ECoSStoredCVPanel";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";

const item: DigitalCenterWorkItem = {
  id: "item-1", sessionId: "session-1", centerObjectId: "1002", vehicleId: "vehicle-1",
  name: "Test", decoderAddress: 3, protocol: "DCC", compareStatus: "ok", stationStatus: "read",
  center: {}, railkeeper: {}, proposed: {}, conflicts: [], createdAt: "", updatedAt: ""
};

function database(supplied = true): DigitalCenterWorkItem {
  return { ...item, center: {
    profile: "LokPilot 5 micro DCC",
    cvs: supplied ? Array.from({ length: 1457 }, (_, i) => i < 83
      ? { number: i + 1, value: 0 }
      : { number: 257 + (i - 83) % 256, value: i % 256, indexCV31: 16, indexCV32: Math.floor((i - 83) / 256) }) : [],
    cvRead: { source: "database", nextCV: 1025, unavailable: 0,
      readAt: "2026-10-27T07:00:00Z", sourceHost: "ecos.local" }
  } };
}

const legacy: DigitalCenterWorkItem = { ...item, center: {
  cvs: [{ number: 8, value: 151 }], cvRead: { nextCV: 33, unavailable: 31,
    readAt: "2026-10-27T06:00:00Z", sourceHost: "ecos.local" }
} };

function Harness({ initial = item }: { initial?: DigitalCenterWorkItem }) {
  const [snapshot, setSnapshot] = useState(initial);
  return <ECoSStoredCVPanel item={snapshot} onChange={setSnapshot} onBusy={() => undefined} />;
}

describe("ECoSStoredCVPanel", () => {
  beforeEach(() => { vi.restoreAllMocks(); setLanguage("de"); });

  it("reviews and explicitly imports backup values without any device refresh", async () => {
    const snapshot = database();
    snapshot.center.cvs = [{ number: 257, value: 0, indexCV31: 0, indexCV32: 0 }];
    snapshot.center.cvRead!.source = "backup";
    const read = vi.spyOn(api, "readDigitalCenterCVBatch");
    const save = vi.spyOn(api, "importDigitalCenterCVs").mockResolvedValue({ imported: 0, preserved: 1 });
    render(<Harness initial={snapshot} />);
    expect(screen.getByText(/Quelle: ECoS-Sicherung/)).toBeVisible();
    expect(screen.queryByRole("button", { name: "Alle gespeicherten CVs lesen" })).not.toBeInTheDocument();
    expect(screen.queryByRole("progressbar")).not.toBeInTheDocument();
    expect(save).not.toHaveBeenCalled();
    fireEvent.click(screen.getByText("Gelieferte Werte und Herkunft ansehen"));
    expect(screen.getByText("0 / 0")).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "CV-Werte ergänzen" }));
    await screen.findByText("0 CV-Werte ergänzt · 1 vorhandene Angaben geschützt.");
    expect(save).toHaveBeenCalledWith("session-1", "item-1", "vehicle-1");
    expect(read).not.toHaveBeenCalled();
  });

  it("compares saved source profiles with the active library and flags a conflicting manual decoder", () => {
    const snapshot = database();
    snapshot.center.cvs = [{ number: 257, value: 20, indexCV31: 16, indexCV32: 3 }];
    const { container, rerender } = render(<ECoSStoredCVPanel item={snapshot} onChange={vi.fn()} onBusy={vi.fn()} />);
    expect(container.querySelector(".decoder-library-notice")).toHaveAttribute("data-status", "matched");
    expect(screen.getByText(/Zuordnung über das gespeicherte ECoS-Profil/)).toBeVisible();
    rerender(<ECoSStoredCVPanel item={{ ...snapshot, railkeeper: { decoderType: "DH18A" } }}
      onChange={vi.fn()} onBusy={vi.fn()} />);
    expect(container.querySelector(".decoder-library-notice")).toHaveAttribute("data-status", "conflict");
    expect(snapshot.railkeeper.decoderType).toBeUndefined();
  });

  it("previews a database larger than 1024 with exact pages before explicit import", async () => {
    const read = vi.spyOn(api, "readDigitalCenterCVBatch").mockResolvedValue(database());
    const save = vi.spyOn(api, "importDigitalCenterCVs").mockResolvedValue({ imported: 1456, preserved: 1 });
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "Alle gespeicherten CVs lesen" }));
    await screen.findByText(/1457 gespeicherte Werte · 1374 mit Indexseite/);
    expect(read).toHaveBeenCalledOnce();
    expect(read).toHaveBeenCalledWith("session-1", "item-1", 1, { source: "database", restart: false });
    expect(save).not.toHaveBeenCalled();
    fireEvent.click(screen.getByText("Gelieferte Werte und Herkunft ansehen"));
    expect(screen.getAllByRole("rowheader", { name: "257" })).toHaveLength(6);
    expect(screen.getByText("ECoS-Decoderprofil: LokPilot 5 micro DCC")).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "CV-Werte ergänzen" }));
    await screen.findByText("1456 CV-Werte ergänzt · 1 vorhandene Angaben geschützt.");
    expect(save).toHaveBeenCalledWith("session-1", "item-1", "vehicle-1");
  });

  it("reports an empty database without inventing zero values", async () => {
    vi.spyOn(api, "readDigitalCenterCVBatch").mockResolvedValue(database(false));
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "Alle gespeicherten CVs lesen" }));
    await screen.findByText(/Die ECoS hat für diese Lok keine gespeicherten CV-Werte bereitgestellt/);
    expect(screen.queryByRole("button", { name: "CV-Werte ergänzen" })).not.toBeInTheDocument();
  });

  it("replaces a partial legacy scan with the complete database before import", async () => {
    const read = vi.spyOn(api, "readDigitalCenterCVBatch").mockResolvedValue(database());
    const save = vi.spyOn(api, "importDigitalCenterCVs").mockResolvedValue({ imported: 1457, preserved: 0 });
    render(<Harness initial={legacy} />);
    fireEvent.click(screen.getByRole("button", { name: "CV-Werte ergänzen" }));
    await screen.findByText("1457 CV-Werte ergänzt · 0 vorhandene Angaben geschützt.");
    expect(read).toHaveBeenCalledOnce();
    expect(save).toHaveBeenCalledOnce();
  });

  it("retains the previous snapshot on failure and retries the database", async () => {
    const read = vi.spyOn(api, "readDigitalCenterCVBatch")
      .mockRejectedValueOnce(new Error("offline")).mockResolvedValue(database());
    render(<Harness initial={legacy} />);
    fireEvent.click(screen.getByRole("button", { name: "Lesen fortsetzen" }));
    await screen.findByRole("alert");
    expect(screen.getByText(/1 Werte verfügbar · 32 von 1024/)).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "Lesen fortsetzen" }));
    await screen.findByText(/1457 gespeicherte Werte/);
    expect(read).toHaveBeenCalledTimes(2);
  });

  it("adopts a complete normal comparison without another read workflow", async () => {
    const read = vi.spyOn(api, "readDigitalCenterCVBatch");
    const save = vi.spyOn(api, "importDigitalCenterCVs").mockResolvedValue({ imported: 0, preserved: 1457 });
    render(<Harness initial={database()} />);
    expect(screen.queryByRole("button", { name: "Alle gespeicherten CVs lesen" })).not.toBeInTheDocument();
    expect(screen.getByText(/Die gespeicherten CV-Werte wurden beim ECoS-Abgleich mit erfasst/)).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "CV-Werte ergänzen" }));
    await screen.findByText("0 CV-Werte ergänzt · 1457 vorhandene Angaben geschützt.");
    expect(read).not.toHaveBeenCalled();
    expect(save).toHaveBeenCalledOnce();
  });

  it("shows conflicting full addresses separately without making those values importable", () => {
    const snapshot = database(false);
    snapshot.center.cvSources = { sources: ["database", "backup"], backupStatus: "checked", backupAdded: 0,
      conflicts: [{ number: 257, indexCV31: 16, indexCV32: 3, databaseValue: 0, backupValue: 22 }] };
    render(<Harness initial={snapshot} />);
    fireEvent.click(screen.getByText("1 abweichende CV-Angaben prüfen"));
    expect(screen.getByText("16 / 3")).toBeVisible();
    expect(screen.getByText("22")).toBeVisible();
    expect(screen.getByText(/Diese Werte werden nicht übernommen/)).toBeVisible();
    expect(screen.queryByRole("button", { name: "CV-Werte ergänzen" })).not.toBeInTheDocument();
  });

  it("keeps the original profile visible without qualifying a conflicted decoder identity", () => {
    const snapshot = database();
    snapshot.center.profile = "";
    snapshot.center.sourceProfile = "LokPilot 5 micro DCC";
    snapshot.center.cvs = [{ number: 257, value: 10, indexCV31: 16, indexCV32: 3 }];
    snapshot.center.cvSources = { sources: ["database", "backup"], backupStatus: "checked", backupAdded: 0,
      conflicts: [{ number: 8, databaseValue: 0, backupValue: 151 }] };
    const { container } = render(<Harness initial={snapshot} />);
    expect(screen.getByText("ECoS-Decoderprofil: LokPilot 5 micro DCC")).toBeVisible();
    expect(screen.getByText(/wird er nicht zur Decoderzuordnung verwendet/)).toBeVisible();
    expect(container.querySelector(".decoder-library-notice")).not.toHaveAttribute("data-status", "matched");
  });

  it("shows three-source conflicts and unverified raw bytes only as read-only evidence", () => {
    const snapshot = database(false);
    snapshot.center.profile = "";
    snapshot.center.sourceProfile = "LokPilot 5 micro DCC";
    snapshot.center.cvSources = { sources: ["database", "backup"], backupStatus: "checked", backupAdded: 0,
      storedCacheStatus: "available", unverifiedValues: [{ number: 8, value: 151 }, { number: 42, value: 77 }],
      conflicts: [{ number: 8, databaseValue: 0, backupValue: 5, rawCacheValue: 151 }] };
    const before = structuredClone(snapshot);
    const { rerender } = render(<Harness initial={snapshot} />);
    fireEvent.click(screen.getByText("1 abweichende CV-Angaben prüfen"));
    expect(screen.getByText(/Zusätzlicher CV-Rohwert: 151/)).toBeVisible();
    fireEvent.click(screen.getByText("2 weitere gespeicherte CV-Rohwerte ansehen"));
    expect(screen.getByText(/Ihr Gültigkeits- und Messstatus ist nicht belegt/)).toBeVisible();
    expect(screen.getByText("77")).toBeVisible();
    expect(screen.queryByRole("button", { name: "CV-Werte ergänzen" })).not.toBeInTheDocument();
    expect(snapshot).toEqual(before);
    const twoSource = structuredClone(snapshot);
    twoSource.center.cvSources!.conflicts = [{ number: 8, databaseValue: 0, backupValue: 151, backupSource: "stored_cache" }];
    rerender(<ECoSStoredCVPanel item={twoSource} onChange={vi.fn()} onBusy={vi.fn()} />);
    fireEvent.click(screen.getByText("1 abweichende CV-Angaben prüfen"));
    expect(screen.getByText("CV-Rohspeicher · Gültigkeit und Messstatus unbekannt")).toBeVisible();
  });

  it("keeps an incomplete normal comparison visible but prevents partial CV adoption or another separate read", () => {
    const snapshot = structuredClone(legacy);
    snapshot.center.cvRead = undefined;
    snapshot.center.cvSources = { sources: [], backupStatus: "failed", backupAdded: 0,
      conflicts: [], warning: "Unavailable" };
    const read = vi.spyOn(api, "readDigitalCenterCVBatch"), save = vi.spyOn(api, "importDigitalCenterCVs");
    render(<Harness initial={snapshot} />);
    expect(screen.getByText(/Die CV-Erfassung ist unvollständig/)).toBeVisible();
    expect(screen.getByText("1 verfügbare CV-Angaben · Erfassung unvollständig")).toBeVisible();
    expect(screen.queryByRole("button", { name: "Alle gespeicherten CVs lesen" })).not.toBeInTheDocument();
    const adopt = screen.getByRole("button", { name: "CV-Werte ergänzen" });
    expect(adopt).toBeDisabled();
    fireEvent.click(adopt);
    expect(read).not.toHaveBeenCalled();
    expect(save).not.toHaveBeenCalled();
  });

  it("marks unavailable binary cache details without blocking a complete database snapshot", async () => {
    const snapshot = database();
    snapshot.center.cvSources = { sources: ["database", "backup"], backupStatus: "checked", backupAdded: 0,
      storedCacheStatus: "unavailable", conflicts: [] };
    const read = vi.spyOn(api, "readDigitalCenterCVBatch");
    const save = vi.spyOn(api, "importDigitalCenterCVs").mockResolvedValue({ imported: 0, preserved: 1457 });
    render(<Harness initial={snapshot} />);
    expect(screen.getByText(/Der zusätzliche gespeicherte Datenbestand konnte nicht vollständig geprüft werden/)).toBeVisible();
    const adopt = screen.getByRole("button", { name: "CV-Werte ergänzen" });
    expect(adopt).toBeEnabled();
    fireEvent.click(adopt);
    await screen.findByText("0 CV-Werte ergänzt · 1457 vorhandene Angaben geschützt.");
    expect(read).not.toHaveBeenCalled();
    expect(save).toHaveBeenCalledOnce();
  });

  it("does not adopt a stopped preparation even if the export finishes", async () => {
    let resolve: ((value: DigitalCenterWorkItem) => void) | undefined;
    const read = vi.spyOn(api, "readDigitalCenterCVBatch").mockImplementation(() =>
      new Promise(done => { resolve = done; }));
    const save = vi.spyOn(api, "importDigitalCenterCVs");
    render(<Harness initial={legacy} />);
    fireEvent.click(screen.getByRole("button", { name: "CV-Werte ergänzen" }));
    fireEvent.click(screen.getByRole("button", { name: "Abfrage anhalten" }));
    await act(async () => { resolve?.(database()); });
    expect(read).toHaveBeenCalledOnce();
    expect(save).not.toHaveBeenCalled();
    expect(screen.getByText(/1 Werte verfügbar · 32 von 1024/)).toBeVisible();
  });
});
