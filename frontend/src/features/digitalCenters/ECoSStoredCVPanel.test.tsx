import { useState } from "react";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
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

function batch(start: number, supplied = true): DigitalCenterWorkItem {
  return { ...item, center: {
    cvs: supplied ? Array.from({ length: start + 31 }, (_, i) => ({ number: i + 1, value: (i + 1) % 256 })) : [],
    cvRead: { nextCV: start + 32, unavailable: supplied ? 0 : start + 31,
      readAt: "2026-10-22T11:00:00Z", sourceHost: "ecos.local" }
  } };
}

function Harness({ initial = item }: { initial?: DigitalCenterWorkItem }) {
  const [snapshot, setSnapshot] = useState(initial);
  return <ECoSStoredCVPanel item={snapshot} onChange={setSnapshot} onBusy={() => undefined} />;
}

describe("ECoSStoredCVPanel", () => {
  beforeEach(() => { vi.restoreAllMocks(); setLanguage("de"); });

  it("reads every CV number, previews 1024 values and explicitly imports with protection", async () => {
    const read = vi.spyOn(api, "readDigitalCenterCVBatch").mockImplementation(async (_, __, start) => batch(start));
    const save = vi.spyOn(api, "importDigitalCenterCVs").mockResolvedValue({ imported: 1023, preserved: 1 });
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "Alle gespeicherten CVs lesen" }));
    await screen.findByText(/1024 Werte verfügbar · 1024 von 1024/);
    expect(read).toHaveBeenCalledTimes(32);
    expect(read.mock.calls.map((call) => call[2])).toEqual(Array.from({ length: 32 }, (_, i) => i * 32 + 1));
    expect(save).not.toHaveBeenCalled();
    fireEvent.click(screen.getByText("Gelieferte Werte und Herkunft ansehen"));
    expect(screen.getByRole("rowheader", { name: "1024" })).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "CV-Werte ergänzen" }));
    await screen.findByText("1023 CV-Werte ergänzt · 1 vorhandene Angaben geschützt.");
    expect(save).toHaveBeenCalledWith("session-1", "item-1", "vehicle-1");
  });

  it("shows device unavailability instead of invented zero values", async () => {
    vi.spyOn(api, "readDigitalCenterCVBatch").mockImplementation(async (_, __, start) => batch(start, false));
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "Alle gespeicherten CVs lesen" }));
    await screen.findByText(/Die ECoS hat für diese Lok keine gespeicherten CV-Werte bereitgestellt/);
    expect(screen.queryByRole("button", { name: "CV-Werte ergänzen" })).not.toBeInTheDocument();
  });

  it("finishes the full read before importing a partial snapshot", async () => {
    const read = vi.spyOn(api, "readDigitalCenterCVBatch").mockImplementation(async (_, __, start) => batch(start));
    const save = vi.spyOn(api, "importDigitalCenterCVs").mockResolvedValue({ imported: 1024, preserved: 0 });
    render(<Harness initial={batch(1)} />);
    fireEvent.click(screen.getByRole("button", { name: "CV-Werte ergänzen" }));
    await screen.findByText("1024 CV-Werte ergänzt · 0 vorhandene Angaben geschützt.");
    expect(read).toHaveBeenCalledTimes(32);
    expect(read.mock.calls[0][2]).toBe(1);
    expect(save).toHaveBeenCalledOnce();
  });

  it("retains confirmed data on failure and resumes at the next unconfirmed batch", async () => {
    const read = vi.spyOn(api, "readDigitalCenterCVBatch")
      .mockResolvedValueOnce(batch(1)).mockRejectedValueOnce(new Error("offline"))
      .mockImplementation(async (_, __, start) => batch(start));
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "Alle gespeicherten CVs lesen" }));
    await screen.findByRole("alert");
    expect(screen.getByText(/32 Werte verfügbar · 32 von 1024/)).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "Lesen fortsetzen" }));
    await screen.findByText(/1024 Werte verfügbar · 1024 von 1024/);
    expect(read.mock.calls.slice(0, 3).map((call) => call[2])).toEqual([1, 33, 33]);
  });

  it("resumes a lost first restart response without requesting another reset", async () => {
    const read = vi.spyOn(api, "readDigitalCenterCVBatch")
      .mockRejectedValueOnce(new Error("response lost after commit"))
      .mockImplementation(async (_, __, start) => batch(start));
    render(<Harness initial={batch(993)} />);
    fireEvent.click(screen.getByRole("button", { name: "Alle gespeicherten CVs lesen" }));
    await screen.findByRole("alert");
    fireEvent.click(screen.getByRole("button", { name: "Lesen fortsetzen" }));
    await waitFor(() => expect(screen.queryByRole("alert")).not.toBeInTheDocument());
    await screen.findByText(/1024 Werte verfügbar · 1024 von 1024/);
    expect(read.mock.calls.slice(0, 3).map((call) => [call[2], call[3]?.restart]))
      .toEqual([[1, true], [1, false], [33, false]]);
  });

  it("stops after the running batch without requesting more CVs", async () => {
    let resolve: ((value: DigitalCenterWorkItem) => void) | undefined;
    const read = vi.spyOn(api, "readDigitalCenterCVBatch").mockImplementation(() =>
      new Promise((done) => { resolve = done; }));
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "Alle gespeicherten CVs lesen" }));
    fireEvent.click(screen.getByRole("button", { name: "Nach diesem Paket anhalten" }));
    await act(async () => { resolve?.(batch(1)); });
    await waitFor(() => expect(screen.getByRole("button", { name: "Lesen fortsetzen" })).toBeEnabled());
    expect(read).toHaveBeenCalledTimes(1);
  });

  it("does not import when preparation is stopped", async () => {
    let resolve: ((value: DigitalCenterWorkItem) => void) | undefined;
    const read = vi.spyOn(api, "readDigitalCenterCVBatch").mockImplementation(() =>
      new Promise((done) => { resolve = done; }));
    const save = vi.spyOn(api, "importDigitalCenterCVs");
    render(<Harness initial={batch(1)} />);
    fireEvent.click(screen.getByRole("button", { name: "CV-Werte ergänzen" }));
    fireEvent.click(screen.getByRole("button", { name: "Nach diesem Paket anhalten" }));
    await act(async () => { resolve?.(batch(33)); });
    expect(read).toHaveBeenCalledTimes(1);
    expect(save).not.toHaveBeenCalled();
    expect(screen.getByText(/64 Werte verfügbar · 64 von 1024/)).toBeVisible();
  });
});
