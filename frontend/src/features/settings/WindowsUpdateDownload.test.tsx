import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { VersionInfo } from "../../shared/api";
import { WindowsUpdateDownload } from "./WindowsUpdateDownload";

const trustedURL =
  "https://github.com/ichwars/RailKeeper/releases/download/v0.2.0/" +
  "RailKeeper-windows-x64-v0.2.0.zip";

const availableUpdate: VersionInfo = {
  version: "0.1.0",
  latestVersion: "v0.2.0",
  updateAvailable: true,
  releaseUrl: "https://github.com/ichwars/RailKeeper/releases/tag/v0.2.0",
  windowsPackage: {
    version: "v0.2.0",
    name: "RailKeeper-windows-x64-v0.2.0.zip",
    url: trustedURL
  },
  checkedAt: "2026-08-16T16:00:00Z",
  status: "update_available",
  message: "Eine neuere RailKeeper-Version ist verfügbar."
};

describe("WindowsUpdateDownload", () => {
  it("renders the exact trusted ZIP action without an install or confirmation flow", async () => {
    const user = userEvent.setup();
    const confirm = vi.spyOn(window, "confirm");
    render(<WindowsUpdateDownload info={availableUpdate} />);

    const download = screen.getByRole("link", { name: "Version v0.2.0 herunterladen" });
    expect(download).toHaveAttribute("href", trustedURL);
    expect(download).toHaveAttribute("rel", "noreferrer");
    expect(screen.getByText(
      "Das ZIP wird nur heruntergeladen. RailKeeper installiert oder ersetzt keine Dateien."
    )).toBeInTheDocument();

    download.addEventListener("click", (event) => event.preventDefault(), { once: true });
    await user.click(download);
    expect(confirm).not.toHaveBeenCalled();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("shows only the release-page fallback when no trusted package is available", () => {
    render(<WindowsUpdateDownload info={{ ...availableUpdate, windowsPackage: undefined }} />);

    expect(screen.queryByRole("link", { name: /herunterladen/i })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Release öffnen" }))
      .toHaveAttribute("href", availableUpdate.releaseUrl);
    expect(screen.queryByText(/Windows Standalone/i)).not.toBeInTheDocument();
  });

  it("uses the English label and wraps a long prerelease version", () => {
    window.localStorage.setItem("railkeeper.settings.language", "en");
    const version = "v0.3.0-beta.12-long-validation-label";
    render(<WindowsUpdateDownload info={{
      ...availableUpdate,
      latestVersion: version,
      windowsPackage: {
        version,
        name: `RailKeeper-windows-x64-${version}.zip`,
        url: trustedURL
      }
    }} />);

    const download = screen.getByRole("link", { name: `Download version ${version}` });
    expect(download).toHaveClass("windows-update-download-button");
    expect(screen.getByText(
      "This only downloads the ZIP. RailKeeper does not install or replace files."
    )).toBeInTheDocument();
  });

  it("renders no action when the update check is offline without a release page", () => {
    render(<WindowsUpdateDownload info={{
      version: "0.1.0",
      updateAvailable: false,
      checkedAt: "2026-08-16T16:00:00Z",
      status: "unavailable",
      message: "Updatequelle konnte nicht erreicht werden."
    }} />);

    expect(screen.queryByRole("link")).not.toBeInTheDocument();
    expect(screen.queryByText(/ZIP/i)).not.toBeInTheDocument();
  });
});
