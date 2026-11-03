import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../shared/api";
import type { BugReportCapability } from "../shared/apiBugReports";
import { Shell } from "./Shell";

const disabledBugReporting: BugReportCapability = {
  enabled: false,
  fallbackUrl: "https://github.com/ichwars/RailKeeper/issues/new",
  maxScreenshotBytes: 5 * 1024 * 1024,
  acceptedScreenshotTypes: ["image/png", "image/jpeg", "image/webp"],
  includedDiagnosticCategories: [],
  excludedDiagnosticCategories: [],
  contactAvailable: false,
  serverDiagnostics: {
    railkeeperVersion: "1.0.0",
    deploymentType: "native",
    goVersion: "go1.26.0",
    operatingSystem: "windows",
    architecture: "amd64",
    schemaVersion: 67,
    enabledDigitalProviderTypes: []
  }
};

describe("Shell article navigation", () => {
  beforeEach(() => {
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
    vi.spyOn(api, "version").mockResolvedValue({
      version: "1.0.0",
      updateAvailable: false,
      checkedAt: "2026-08-08T10:00:00Z",
      status: "current",
      message: "current"
    });
    vi.spyOn(api, "bugReportConfig").mockResolvedValue(disabledBugReporting);
  });

  it("uses the approved vehicle and accessory navigation labels", async () => {
    render(
      <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    const navigation = screen.getByRole("navigation", { name: "Hauptnavigation" });
    expect(navigation).toHaveTextContent("Fahrzeugbestand");
    expect(navigation).toHaveTextContent("Zubehör");
    expect(navigation).not.toHaveTextContent(/^Bestand$/);
    expect(navigation).not.toHaveTextContent("Artikelübersicht");
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });

  it("does not expose accessories to Messe users", () => {
    render(
      <Shell username="messe" roles={["Messe"]} activeView="exhibition" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    expect(screen.queryByRole("link", { name: "Zubehör" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Ausstellung" })).toBeInTheDocument();
  });

  it("exposes Import/Export to pure Messe users", () => {
    render(
      <Shell username="messe" roles={["Messe"]} activeView="importExport" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    expect(screen.getByRole("link", { name: "Import/Export" }))
      .toHaveAttribute("href", "/import-export");
  });

  it("uses the English accessories navigation label", async () => {
    window.localStorage.setItem("railkeeper.settings.language", "en");
    render(
      <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
        <p>Content</p>
      </Shell>
    );

    expect(screen.getByRole("link", { name: "Accessories" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Article overview" })).not.toBeInTheDocument();
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });

  it("does not expose the layout workspace in the navigation", async () => {
    render(
      <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    expect(screen.queryByRole("link", { name: "Anlage" })).not.toBeInTheDocument();
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });

  it("does not expose the layout workspace in the English navigation", async () => {
    window.localStorage.setItem("railkeeper.settings.language", "en");
    render(
      <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
        <p>Content</p>
      </Shell>
    );

    expect(screen.queryByRole("link", { name: "Layout" })).not.toBeInTheDocument();
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });

  it("links the Digital Centers navigation directly to its workspace", async () => {
    render(
      <Shell username="admin" roles={["Admin"]} activeView="digitalCenters" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    expect(screen.getByRole("link", { name: "Digitalzentralen" }))
      .toHaveAttribute("href", "/digital-centers");
    await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
  });

  it("places command stations directly before settings in the default navigation", () => {
    render(
      <Shell username="admin" roles={["Admin"]} activeView="overview" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    const navigation = screen.getByRole("navigation", { name: "Hauptnavigation" });
    const entries = Array.from(navigation.querySelectorAll(".nav-entry"));
    expect(entries.slice(-2)).toEqual([
      within(navigation).getByRole("link", { name: "Digitalzentralen" }),
      within(navigation).getByRole("link", { name: "Einstellungen" })
    ]);
  });

  it("migrates an older saved default order so command stations stay before settings", () => {
    window.localStorage.setItem("railkeeper.settings.sidebarPrefs:admin", JSON.stringify({
      order: ["overview", "vehicles", "accessories", "layouts", "exhibition", "importExport", "settings"],
      hidden: []
    }));

    render(
      <Shell username="admin" roles={["Admin"]} activeView="overview" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    const navigation = screen.getByRole("navigation", { name: "Hauptnavigation" });
    const entries = Array.from(navigation.querySelectorAll(".nav-entry"));
    expect(within(navigation).queryByRole("link", { name: "Anlage" })).not.toBeInTheDocument();
    expect(entries.slice(-2)).toEqual([
      within(navigation).getByRole("link", { name: "Digitalzentralen" }),
      within(navigation).getByRole("link", { name: "Einstellungen" })
    ]);
  });

  it("opens the in-app report dialog when reporting is enabled", async () => {
    vi.mocked(api.bugReportConfig).mockResolvedValue({ ...disabledBugReporting, enabled: true });
    render(
      <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    const trigger = await screen.findByRole("button", { name: "Fehler melden" });
    trigger.click();

    expect(await screen.findByRole("dialog", { name: "Fehler melden" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Bereich" })).toHaveTextContent("Zubehör");
  });

  it("keeps the external fallback when the capability check fails", async () => {
    vi.mocked(api.bugReportConfig).mockRejectedValue(new Error("not configured"));
    render(
      <Shell username="editor" roles={["Editor"]} activeView="overview" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    await waitFor(() => expect(api.bugReportConfig).toHaveBeenCalledOnce());
    expect(screen.getByRole("link", { name: "Fehler melden" })).toHaveAttribute(
      "href",
      "https://github.com/ichwars/RailKeeper/issues/new"
    );
  });

  it("rechecks a disabled capability when the window regains focus", async () => {
    vi.mocked(api.bugReportConfig)
      .mockResolvedValueOnce(disabledBugReporting)
      .mockResolvedValueOnce({ ...disabledBugReporting, enabled: true });
    render(
      <Shell username="editor" roles={["Editor"]} activeView="overview" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    await waitFor(() => expect(api.bugReportConfig).toHaveBeenCalledOnce());
    expect(screen.getByRole("link", { name: "Fehler melden" })).toBeInTheDocument();
    window.dispatchEvent(new Event("focus"));

    expect(await screen.findByRole("button", { name: "Fehler melden" })).toBeInTheDocument();
    expect(vi.mocked(api.bugReportConfig).mock.calls.filter((args) => args[0] !== true)).toHaveLength(2);
  });

  it("refreshes capability after submission failure without discarding the open draft", async () => {
    const user = userEvent.setup();
    const enabled = { ...disabledBugReporting, enabled: true };
    vi.mocked(api.bugReportConfig).mockResolvedValueOnce(enabled).mockResolvedValueOnce(disabledBugReporting);
    vi.spyOn(api, "submitBugReport").mockRejectedValue(new Error("GitHub unavailable"));
    render(
      <Shell username="editor" roles={["Editor"]} activeView="vehicles" onLogout={vi.fn()}>
        <p>Inhalt</p>
      </Shell>
    );

    await user.click(await screen.findByRole("button", { name: "Fehler melden" }));
    await user.type(screen.getByLabelText(/Was ist passiert/i),
      "Das Vorschaubild verschwindet nach dem Speichern.");
    await user.click(screen.getByRole("button", { name: "Fehlerbericht senden" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/konnte nicht gesendet/i);
    await waitFor(() => expect(vi.mocked(api.bugReportConfig).mock.calls.filter((args) => args[0] !== true)).toHaveLength(2));
    expect(screen.getByRole("dialog", { name: "Fehler melden" })).toBeInTheDocument();
    expect(screen.getByLabelText(/Was ist passiert/i)).toHaveValue(
      "Das Vorschaubild verschwindet nach dem Speichern."
    );
    expect(screen.getByRole("link", { name: "Fehler melden" })).toHaveAttribute(
      "href", "https://github.com/ichwars/RailKeeper/issues/new"
    );
  });
});
