import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api, type StorageLocationInfo } from "../../shared/api";
import { StorageLocationPanel } from "./StorageLocationPanel";

const standaloneInfo: StorageLocationInfo = {
  dataPath: String.raw`C:\Users\Ada\AppData\Local\RailKeeper\data`,
  mode: "windows_standalone",
  openFolderAvailable: true,
  migrationReceipt: {
    sourcePath: String.raw`C:\RailKeeper-alt\data`,
    targetPath: String.raw`C:\Users\Ada\AppData\Local\RailKeeper\data`,
    migratedAt: "2026-08-16T14:30:00Z",
    version: "0.1.18",
    filesVerified: 21,
    acknowledged: false
  }
};

describe("StorageLocationPanel", () => {
  afterEach(() => vi.restoreAllMocks());

  it("shows the exact paths and performs only the explicit server actions", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "storageLocationInfo").mockResolvedValue(standaloneInfo);
    vi.spyOn(api, "openStorageFolder").mockResolvedValue(undefined);
    vi.spyOn(api, "acknowledgeStorageMigration").mockResolvedValue(undefined);

    render(<StorageLocationPanel />);

    expect(await screen.findAllByText(standaloneInfo.dataPath)).not.toHaveLength(0);
    expect(screen.getByText("Windows Standalone (ohne Installation)")).toBeInTheDocument();
    expect(screen.getByText(standaloneInfo.migrationReceipt!.sourcePath)).toBeInTheDocument();
    expect(screen.getAllByText(standaloneInfo.migrationReceipt!.targetPath)).not.toHaveLength(0);

    await user.click(screen.getByRole("button", { name: "Datenordner öffnen" }));
    expect(api.openStorageFolder).toHaveBeenCalledOnce();
    await user.click(screen.getByRole("button", { name: "Verstanden" }));
    await waitFor(() => expect(api.acknowledgeStorageMigration).toHaveBeenCalledOnce());
    expect(screen.queryByRole("button", { name: "Verstanden" })).not.toBeInTheDocument();
    expect(screen.getByText("Bestätigt")).toBeInTheDocument();
  });

  it("wraps long paths and omits the folder action when the capability is unavailable", async () => {
    const longPath = "/srv/railkeeper/very/long/self-hosted/path/with/many/folders/for/persistent/data";
    vi.spyOn(api, "storageLocationInfo").mockResolvedValue({
      dataPath: longPath,
      mode: "server",
      openFolderAvailable: false
    });

    render(<StorageLocationPanel />);

    const path = await screen.findByText(longPath);
    expect(path).toHaveClass("storage-location-path");
    expect(screen.getByText("Server / Docker")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Datenordner öffnen" })).not.toBeInTheDocument();
  });

  it("shows a retryable error state", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "storageLocationInfo")
      .mockRejectedValueOnce(new Error("offline"))
      .mockResolvedValueOnce(standaloneInfo);

    render(<StorageLocationPanel />);

    expect(await screen.findByText("offline")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Erneut versuchen" }));
    expect(await screen.findAllByText(standaloneInfo.dataPath)).not.toHaveLength(0);
    expect(api.storageLocationInfo).toHaveBeenCalledTimes(2);
  });
});
