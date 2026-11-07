import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { FormEvent } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { api, type DecoderCVSnapshot, type DecoderCVSnapshotSummary } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { DecoderArchiveDialog, DecoderCVArchive } from "./DecoderCVArchive";

function snapshot(id = "first", profile = " Legacy "): DecoderCVSnapshot {
  const now = "2026-11-07T03:00:00Z";
  return { id, workspaceId: "group", vehicleReferenceId: "deleted-vehicle", vehicleName: "Früheres Fahrzeug",
    vehicleInventoryNumber: "INV-OLD", profile, installationId: null, kind: "legacy", scope: "partial_working_state",
    note: `Original note ${id}`, recordedAt: now, valueCount: 2, hardware: null,
    payload: { values: [
      { id: "unknown", vehicleId: "deleted-vehicle", cvNumber: 257, value: -1, decoderProfile: profile,
        sourceFileId: "source", sourceLine: 41, sourceExportedAt: "2026-10-20T08:30:00", createdAt: now, updatedAt: now,
        history: [{ id: "history", vehicleId: "deleted-vehicle", cvValueId: "unknown", oldValue: 3, newValue: -1, changedAt: now }] },
      { id: "indexed", vehicleId: "deleted-vehicle", cvNumber: 257, value: 7, decoderProfile: profile,
        indexCV31: 0, indexCV32: 0, createdAt: now, updatedAt: now }
    ] }, sources: [
      { sourceFileId: "source", contentState: "stored", metadata: { id: "source", vehicleId: "deleted-vehicle",
        fileName: "original.csv", originalName: "original.csv", sizeBytes: 13, createdAt: now, updatedAt: now } },
      { sourceFileId: "missing", contentState: "unavailable", metadata: { id: "missing", vehicleId: "",
        fileName: "", originalName: "", sizeBytes: 0, createdAt: "", updatedAt: "" } }
    ] };
}

function summary(value: DecoderCVSnapshot): DecoderCVSnapshotSummary {
  return { id: value.id, workspaceId: value.workspaceId, vehicleReferenceId: value.vehicleReferenceId,
    vehicleName: value.vehicleName, vehicleInventoryNumber: value.vehicleInventoryNumber, profile: value.profile,
    installationId: value.installationId, kind: value.kind, scope: value.scope, note: value.note,
    recordedAt: value.recordedAt, valueCount: value.valueCount };
}

describe("private recorded CV archive", () => {
  beforeEach(() => { setLanguage("de"); });
  afterEach(() => { vi.restoreAllMocks(); });

  it("shows historical identity, unknown values, distinct pages and retained source provenance", async () => {
    const value = snapshot();
    const list = vi.spyOn(api, "decoderCVSnapshots").mockResolvedValue([summary(value)]);
    vi.spyOn(api, "decoderCVSnapshot").mockResolvedValue(value);
    const write = vi.spyOn(api, "captureDecoderCVCheckpoint");
    render(<DecoderCVArchive vehicleId="deleted-vehicle" />);
    expect(await screen.findByText("Unbekannt (-1)")).toBeVisible();
    expect(screen.getByText("Ohne Indexangabe")).toBeVisible();
    expect(screen.getByText("CV31=0 / CV32=0")).toBeVisible();
    expect(screen.getByText(/Originalzeile 41/)).toBeVisible();
    expect(screen.getByText(/Quellenzeit ohne Zeitzone: 2026-10-20 08:30:00/)).toBeVisible();
    expect(screen.getAllByText(/Früheres Fahrzeug/).length).toBeGreaterThan(0);
    expect(screen.getByText("Damals keinem Decoderexemplar zugeordnet.")).toBeVisible();
    const download = screen.getByRole("link", { name: "Originaldatei herunterladen" });
    expect(download).toHaveAttribute("href", api.decoderSnapshotSourceUrl(value.id, "source"));
    expect(screen.getAllByRole("link")).toHaveLength(1);
    expect(screen.getByText('" Legacy "', { selector: "h3" })).toBeVisible();
    await userEvent.click(screen.getByText("Gespeicherte Änderungen: 1"));
    expect(screen.getByText(/3 → Unbekannt \(-1\)/)).toBeVisible();
    expect(list).toHaveBeenCalledWith("deleted-vehicle");
    expect(write).not.toHaveBeenCalled();
  });

  it("ignores a late response after selecting another recorded state", async () => {
    const first = snapshot(), second = snapshot("second", "Other");
    let release: (value: DecoderCVSnapshot) => void = () => undefined;
    const pending = new Promise<DecoderCVSnapshot>(resolve => { release = resolve; });
    vi.spyOn(api, "decoderCVSnapshots").mockResolvedValue([summary(first), summary(second)]);
    vi.spyOn(api, "decoderCVSnapshot").mockImplementation(id => id === first.id ? pending : Promise.resolve(second));
    render(<DecoderCVArchive />);
    await userEvent.click(await screen.findByRole("button", { name: /Other.*Früheres Fahrzeug/ }));
    expect(await screen.findByText("Original note second")).toBeVisible();
    await act(async () => { release(first); await pending; });
    expect(screen.queryByText("Original note first")).not.toBeInTheDocument();
    expect(screen.getByText("Original note second")).toBeVisible();
  });

  it("keeps searches inside an existing vehicle form from submitting it", async () => {
    const value = snapshot(), submit = vi.fn((event: FormEvent<HTMLFormElement>) => event.preventDefault());
    vi.spyOn(api, "decoderCVSnapshots").mockResolvedValue([summary(value)]);
    vi.spyOn(api, "decoderCVSnapshot").mockResolvedValue(value);
    render(<form onSubmit={submit}><input aria-label="Offener Fahrzeugentwurf" defaultValue="Ungespeichert" />
      <DecoderCVArchive /><button type="submit">Fahrzeug speichern</button></form>);
    await screen.findByText("Original note first");
    await userEvent.type(screen.getByRole("searchbox", { name: "Teilstände durchsuchen" }), "Legacy{Enter}");
    await userEvent.type(screen.getByRole("searchbox", { name: "Suchen" }), "257{Enter}");
    expect(submit).not.toHaveBeenCalled();
    expect(screen.getByLabelText("Offener Fahrzeugentwurf")).toHaveValue("Ungespeichert");
  });

  it("provides loading, recoverable error and empty states in both languages", async () => {
    const list = vi.spyOn(api, "decoderCVSnapshots").mockRejectedValueOnce(new Error("Unavailable")).mockResolvedValue([]);
    render(<DecoderCVArchive />);
    expect(screen.getByText("Gespeicherte Teilstände werden geladen …")).toBeVisible();
    expect(await screen.findByRole("alert")).toHaveTextContent("konnten nicht geladen werden");
    await userEvent.click(screen.getByRole("button", { name: "Neu laden" }));
    expect(await screen.findByText("Noch keine CV-Teilstände vorhanden.")).toBeVisible();
    act(() => setLanguage("en"));
    expect(screen.getByText("No recorded CV states yet.")).toBeVisible();
    expect(list).toHaveBeenCalledTimes(2);
  });

  it("focuses the close action and closes on Escape even after focus leaves the dialog", async () => {
    vi.spyOn(api, "decoderCVSnapshots").mockResolvedValue([]);
    const close = vi.fn();
    render(<DecoderArchiveDialog onClose={close} />);
    await waitFor(() => expect(screen.getByRole("button", { name: "Schließen" })).toHaveFocus());
    screen.getByRole("button", { name: "Schließen" }).blur();
    fireEvent.keyDown(document.body, { key: "Escape" });
    expect(close).toHaveBeenCalledOnce();
  });
});
