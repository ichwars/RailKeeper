import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, ApiError } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { ECoSBackupImport } from "./ECoSBackupImport";
import type { ECoSBackupImportResult } from "./digitalCenterModel";

const result: ECoSBackupImportResult = {
  session: { id: "backup-1", provider: "ecos", state: "ready", host: "ecos.local", port: 15471,
    capabilities: { testConnection: false, readLocomotives: false, liveMonitor: false,
      writeLocomotives: false, writeCVs: false, diagnose: false },
    readStartedAt: "2026-10-29T06:00:00Z", readCompletedAt: "2026-10-29T06:00:01Z",
    createdByUserId: "admin-1", createdAt: "2026-10-29T06:00:00Z", updatedAt: "2026-10-29T06:00:01Z" },
  summary: { stationSerial: 66, profiles: 2, cvValues: 12,
    importedAt: "2026-10-29T06:00:00Z", linkedVehicles: 0, identitySource: "saved_comparison" }
};

function choose(file: File) {
  fireEvent.change(screen.getByLabelText("ECoS-Sicherung (.eco)"), { target: { files: [file] } });
}

describe("ECoSBackupImport", () => {
  beforeEach(() => { vi.restoreAllMocks(); setLanguage("de"); });

  it("allows preview without confirming station ownership and opens only the returned saved session", async () => {
    const upload = vi.spyOn(api, "importECoSBackup").mockResolvedValue(result);
    const read = vi.spyOn(api, "startDigitalCenterReadSession");
    const importCVs = vi.spyOn(api, "importDigitalCenterCVs");
    const onImported = vi.fn(async () => undefined);
    render(<ECoSBackupImport stationHost="ecos.local" onImported={onImported} />);
    expect(screen.getByRole("button", { name: "Sicherung prüfen" })).toBeDisabled();
    expect(screen.getByRole("checkbox")).not.toBeChecked();
    const file = new File(["encrypted backup"], "lokbestand.eco");
    choose(file);
    fireEvent.click(screen.getByRole("button", { name: "Sicherung prüfen" }));
    await screen.findByText(/2 Lok-Konfigurationen mit 12 CV-Werten geprüft/);
    expect(upload).toHaveBeenCalledWith(file, { confirmStation: false, stationHost: "ecos.local" });
    expect(onImported).toHaveBeenCalledWith(result);
    expect(screen.getByText(/Gesicherte Serienkennung: 0x00000042/)).toBeVisible();
    expect(screen.getByText(/Prüfstand erstellt: 2026-10-29T06:00:00Z/)).toBeVisible();
    expect(read).not.toHaveBeenCalled();
    expect(importCVs).not.toHaveBeenCalled();
  });

  it("passes explicit station confirmation and resets it for another file", async () => {
    const upload = vi.spyOn(api, "importECoSBackup").mockResolvedValue(result);
    render(<ECoSBackupImport stationHost="ecos.local" onImported={vi.fn(async () => undefined)} />);
    const file = new File(["backup"], "bestand.ECO");
    choose(file);
    fireEvent.click(screen.getByRole("checkbox"));
    fireEvent.click(screen.getByRole("button", { name: "Sicherung prüfen" }));
    await waitFor(() => expect(upload).toHaveBeenCalledWith(file,
      { confirmStation: true, stationHost: "ecos.local" }));
    await waitFor(() => expect(screen.getByRole("button", { name: "Sicherung prüfen" })).toBeEnabled());
    choose(new File(["other backup"], "andere.eco"));
    expect(screen.getByRole("checkbox")).not.toBeChecked();
    expect(screen.queryByText(/2 Lok-Konfigurationen/)).not.toBeInTheDocument();
  });

  it("rejects oversized, empty and incorrectly named files without uploading", () => {
    const upload = vi.spyOn(api, "importECoSBackup");
    render(<ECoSBackupImport stationHost="ecos.local" onImported={vi.fn(async () => undefined)} />);
    const oversized = new File(["backup"], "gross.eco");
    Object.defineProperty(oversized, "size", { value: 8 * 1024 * 1024 + 1 });
    choose(oversized);
    expect(screen.getByRole("alert")).toHaveTextContent("größer als 8 MiB");
    expect(screen.getByRole("button", { name: "Sicherung prüfen" })).toBeDisabled();
    choose(new File([], "leer.eco"));
    expect(screen.getByRole("alert")).toHaveTextContent("ungültig");
    choose(new File(["not a backup"], "firmware.bci"));
    expect(screen.getByRole("alert")).toHaveTextContent("vollständige .eco-Sicherung");
    expect(upload).not.toHaveBeenCalled();
  });

  it.each([
    ["ecos_backup_unavailable", 503, "noch nicht eingerichtet"],
    ["ecos_backup_invalid", 422, "ungültig"],
    ["digital_center_backup_station_mismatch", 409, "ECoS hat sich geändert"]
  ])("shows a useful safe error for %s", async (code, status, message) => {
    vi.spyOn(api, "importECoSBackup").mockRejectedValue(new ApiError("private diagnostic", code, status));
    const onImported = vi.fn(async () => undefined);
    render(<ECoSBackupImport stationHost="ecos.local" onImported={onImported} />);
    choose(new File(["backup"], "bestand.eco"));
    fireEvent.click(screen.getByRole("button", { name: "Sicherung prüfen" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(message);
    expect(screen.queryByText("private diagnostic")).not.toBeInTheDocument();
    expect(onImported).not.toHaveBeenCalled();
  });

  it("disables repeated submissions and ignores a late result after leaving the view", async () => {
    let resolve: ((value: ECoSBackupImportResult) => void) | undefined;
    const upload = vi.spyOn(api, "importECoSBackup").mockImplementation(() => new Promise(done => { resolve = done; }));
    const onImported = vi.fn(async () => undefined);
    const { unmount } = render(<ECoSBackupImport stationHost="ecos.local" onImported={onImported} />);
    choose(new File(["backup"], "bestand.eco"));
    fireEvent.click(screen.getByRole("button", { name: "Sicherung prüfen" }));
    expect(screen.getByRole("button", { name: "Sicherung wird geprüft …" })).toBeDisabled();
    expect(screen.getByRole("checkbox")).toBeDisabled();
    unmount();
    await act(async () => { resolve?.(result); });
    expect(upload).toHaveBeenCalledOnce();
    expect(onImported).not.toHaveBeenCalled();
  });
});
