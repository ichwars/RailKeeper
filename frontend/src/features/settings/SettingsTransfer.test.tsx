import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type BackupValidationResult } from "../../shared/api";
import { decoderLibraryChanged } from "../../shared/DecoderLibraryProvider";
import { SettingsView } from "./SettingsView";

const compatibleBackup: BackupValidationResult = {
  compatible: true,
  version: 24,
  tableCount: 1,
  rowCount: 2,
  fileCount: 1,
  fileBytes: 100,
  tables: [],
  warnings: [],
  errors: []
};

const backup = () => new File(["{}"], "reference.json", { type: "application/json" });

describe("Settings transfer safety and state", () => {
  beforeEach(() => {
    window.history.replaceState(null, "", "/settings?tab=importExport");
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
    vi.spyOn(api, "session").mockResolvedValue({
      username: "editor", roles: ["Editor"], csrfToken: "test", twoFactorEnabled: false
    });
    vi.spyOn(api, "managedMasterDataAll").mockResolvedValue({});
    vi.spyOn(api, "version").mockRejectedValue(new Error("offline"));
    vi.spyOn(api, "systemPrinters").mockRejectedValue(new Error("offline"));
    vi.spyOn(api, "inventoryNumberSchemes").mockResolvedValue([{
      id: "scheme", category: "locomotive", prefix: "RK", nextNumber: 1, padding: 6,
      active: true, preview: "RK000001", createdAt: "", updatedAt: ""
    }]);
    vi.spyOn(api, "storageUsage").mockResolvedValue({
      totalBytes: 1024, categories: [{ key: "images", label: "Images", bytes: 1024, files: 3 }],
      updatedAt: "2026-01-01T00:00:00Z"
    });
    vi.spyOn(api, "validateBackup").mockResolvedValue(compatibleBackup);
    vi.spyOn(api, "restoreBackup").mockResolvedValue({ restoredTables: 1, restoredRows: 2, restoredFiles: 1 });
    vi.spyOn(api, "importMasterData").mockResolvedValue({
      importedTypes: 1, importedEntries: 2, importedRelations: 3
    });
  });

  it("waits for compatibility and the typed phrase, then requires a cancellable confirmation", async () => {
    const user = userEvent.setup();
    let finishValidation!: (value: BackupValidationResult) => void;
    vi.mocked(api.validateBackup).mockReturnValue(new Promise((resolve) => { finishValidation = resolve; }));
    render(<SettingsView username="editor" />);

    const file = backup();
    await user.upload(screen.getByLabelText(/^Backup-Datei/), file);
    expect(screen.getByText("Backup wird geprüft...")).toBeInTheDocument();
    const restore = screen.getByRole("button", { name: "Backup einspielen" });
    expect(restore).toBeDisabled();
    await act(async () => finishValidation(compatibleBackup));
    expect(restore).toBeDisabled();
    await user.type(screen.getByPlaceholderText("WIEDERHERSTELLEN"), "wiederherstellen");
    await user.click(restore);
    expect(api.restoreBackup).not.toHaveBeenCalled();
    await user.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Abbrechen" }));
    expect(api.restoreBackup).not.toHaveBeenCalled();
    expect(screen.getByText(file.name)).toBeInTheDocument();

    const decoderChange = vi.fn();
    window.addEventListener(decoderLibraryChanged, decoderChange);
    try {
      await user.click(restore);
      await user.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Backup einspielen" }));
      await waitFor(() => expect(api.restoreBackup).toHaveBeenCalledExactlyOnceWith(file));
      expect(await screen.findByText("Backup wiederhergestellt: 2 Datensätze, 1 Dateien.")).toBeInTheDocument();
      expect(api.storageUsage).toHaveBeenCalledTimes(1);
      expect(decoderChange).toHaveBeenCalledOnce();
      expect(screen.queryByText(file.name)).not.toBeInTheDocument();
      expect(restore).toBeDisabled();
    } finally {
      window.removeEventListener(decoderLibraryChanged, decoderChange);
    }
  });

  it("blocks incompatible files and validation errors without submitting a restore", async () => {
    const user = userEvent.setup();
    vi.mocked(api.validateBackup)
      .mockResolvedValueOnce({ ...compatibleBackup, compatible: false, errors: ["Future backup version"] })
      .mockRejectedValueOnce(new Error("Invalid backup JSON"));
    render(<SettingsView username="editor" />);

    await user.upload(screen.getByLabelText(/^Backup-Datei/), backup());
    expect(await screen.findByText("Future backup version")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Backup einspielen" })).toBeDisabled();
    expect(screen.queryByPlaceholderText("WIEDERHERSTELLEN")).not.toBeInTheDocument();
    await user.upload(screen.getByLabelText(/^Backup-Datei/), new File(["bad"], "invalid.json"));
    expect(await screen.findByText("Invalid backup JSON")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Backup einspielen" })).toBeDisabled();
    expect(api.restoreBackup).not.toHaveBeenCalled();
  });

  it("keeps the selected backup and phrase across tabs and after a failed restore", async () => {
    const user = userEvent.setup();
    vi.mocked(api.restoreBackup).mockRejectedValueOnce(new Error("Restore rejected"));
    render(<SettingsView username="editor" />);
    await user.upload(screen.getByLabelText(/^Backup-Datei/), backup());
    await user.type(await screen.findByPlaceholderText("WIEDERHERSTELLEN"), "WIEDERHERSTELLEN");
    await user.click(screen.getByRole("button", { name: "Darstellung" }));
    await user.click(screen.getByRole("button", { name: "Backup" }));
    expect(screen.getByText("reference.json")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("WIEDERHERSTELLEN")).toHaveValue("WIEDERHERSTELLEN");
    expect(api.validateBackup).toHaveBeenCalledOnce();

    await user.click(screen.getByRole("button", { name: "Backup einspielen" }));
    await user.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Backup einspielen" }));
    expect(await screen.findByText("Restore rejected")).toBeInTheDocument();
    expect(screen.getByText("reference.json")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Backup einspielen" })).toBeEnabled();
  });

  it("confirms master-data imports and reloads cached master data afterwards", async () => {
    const user = userEvent.setup();
    render(<SettingsView username="editor" />);
    await user.click(screen.getByRole("button", { name: "Daten" }));
    await waitFor(() => expect(api.managedMasterDataAll).toHaveBeenCalledTimes(1));
    await user.click(screen.getByRole("button", { name: "Backup" }));
    const file = backup();
    await user.upload(screen.getByLabelText(/^Stammdaten-Datei/), file);
    const importButton = screen.getByRole("button", { name: "Stammdaten einspielen" });
    await user.click(importButton);
    expect(api.importMasterData).not.toHaveBeenCalled();
    await user.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Abbrechen" }));
    await user.click(importButton);
    await user.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Stammdaten einspielen" }));
    await waitFor(() => expect(api.importMasterData).toHaveBeenCalledExactlyOnceWith(file));
    expect(screen.queryByText(file.name)).not.toBeInTheDocument();
    expect(importButton).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Daten" }));
    await waitFor(() => expect(api.managedMasterDataAll).toHaveBeenCalledTimes(2));
  });

  it("retains storage results across tabs and loads them at the existing navigation points", async () => {
    const user = userEvent.setup();
    render(<SettingsView username="editor" />);
    expect(api.storageUsage).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Speichernutzung aktualisieren" }));
    await waitFor(() => expect(api.storageUsage).toHaveBeenCalledTimes(1));
    expect(screen.getByText("3")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Daten" }));
    expect(api.storageUsage).toHaveBeenCalledTimes(1);
    await user.click(screen.getByRole("button", { name: "Allgemein" }));
    await waitFor(() => expect(api.storageUsage).toHaveBeenCalledTimes(2));
    expect(screen.getByText("Bilder")).toBeInTheDocument();
    expect(screen.queryByText("Images")).not.toBeInTheDocument();
  });
});
