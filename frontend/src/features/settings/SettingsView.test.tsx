import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type MasterDataEntry } from "../../shared/api";
import { SettingsView } from "./SettingsView";
import { readSettingsLocation } from "./settingsDataModel";
import {
  defaultSidebarOrder,
  masterDataTypes,
  normalizeSidebarOrder,
  readSidebarPrefs,
  settingsTabs,
  sidebarPrefsKey
} from "./settingsModel";

describe("SettingsView data navigation", () => {
  beforeEach(() => {
    window.history.replaceState(null, "", "/settings");
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
    vi.spyOn(api, "masterData").mockResolvedValue([]);
    vi.spyOn(api, "managedMasterData").mockResolvedValue([]);
    vi.spyOn(api, "managedMasterDataAll").mockResolvedValue({});
    vi.spyOn(api, "version").mockRejectedValue(new Error("offline"));
    vi.spyOn(api, "storageUsage").mockRejectedValue(new Error("offline"));
    vi.spyOn(api, "systemPrinters").mockRejectedValue(new Error("offline"));
    vi.spyOn(api, "session").mockResolvedValue({
      username: "viewer",
      roles: ["Viewer"],
      csrfToken: "test",
      twoFactorEnabled: false
    });
  });

  it("keeps article management out of the main settings tabs", () => {
    expect(settingsTabs.map((tab) => tab.id)).not.toContain("articleManagement");
  });

  it("keeps master-data transfer in Settings", () => {
    expect(settingsTabs.map((tab) => tab.id)).toContain("importExport");
    expect(readSettingsLocation("?tab=importExport").tab).toBe("importExport");
  });

  it("uses command stations instead of layouts in the current sidebar order", () => {
    expect(defaultSidebarOrder).toEqual([
      "overview",
      "vehicles",
      "accessories",
      "exhibition",
      "importExport",
      "digitalCenters",
      "settings"
    ]);
  });

  it("migrates legacy sidebar preferences without losing the custom relative order", () => {
    window.localStorage.setItem(sidebarPrefsKey("viewer"), JSON.stringify({
      order: ["settings", "layouts", "vehicles"],
      hidden: ["layouts"]
    }));

    const preferences = readSidebarPrefs("viewer");

    expect(preferences.order).not.toContain("layouts");
    expect(preferences.order).toContain("digitalCenters");
    expect(preferences.order.indexOf("settings")).toBeLessThan(preferences.order.indexOf("vehicles"));
    expect(preferences.hidden).toEqual([]);
  });

  it("inserts missing command stations before settings", () => {
    const normalized = normalizeSidebarOrder([
      "overview",
      "vehicles",
      "accessories",
      "exhibition",
      "importExport",
      "settings"
    ]);

    expect(normalized.slice(-2)).toEqual(["digitalCenters", "settings"]);
  });

  it("normalizes legacy sidebar preferences loaded from the profile", async () => {
    vi.mocked(api.profileSettings).mockResolvedValue({
      settings: {
        [sidebarPrefsKey("viewer")]: JSON.stringify({
          order: ["settings", "layouts", "vehicles"],
          hidden: ["layouts"]
        })
      }
    });

    render(<SettingsView username="viewer" />);

    await waitFor(() => {
      const stored = JSON.parse(window.localStorage.getItem(sidebarPrefsKey("viewer")) || "{}") as {
        order?: string[];
        hidden?: string[];
      };
      expect(stored.order).not.toContain("layouts");
      expect(stored.order).toContain("digitalCenters");
      expect(stored.hidden).toEqual([]);
    });
  });

  it("restores manufacturers as the first general master-data type", async () => {
    window.history.replaceState(null, "", "/settings?tab=data");
    vi.mocked(api.managedMasterDataAll).mockResolvedValue({ manufacturer: [], vehicle_category: [] });

    render(<SettingsView username="viewer" />);

    expect(await screen.findByRole("tab", { name: "Hersteller" })).toBeInTheDocument();
    expect(masterDataTypes[0]).toEqual({ type: "manufacturer" });
  });

  it("normalizes the removed article-management route", async () => {
    window.history.replaceState(null, "", "/settings?tab=articleManagement");
    render(<SettingsView username="viewer" />);

    await waitFor(() => {
      const query = new URLSearchParams(window.location.search);
      expect(query.get("tab")).toBe("data");
      expect(query.get("group")).toBe("article");
      expect(query.get("type")).toBe("stock_unit");
    });
  });

  it("groups general and article master data under Data", async () => {
    const user = userEvent.setup();
    window.history.replaceState(null, "", "/settings?tab=data");
    render(<SettingsView username="viewer" />);

    const groups = await screen.findByRole("tablist", { name: "Datengruppen" });
    expect(within(groups).getAllByRole("tab")).toHaveLength(2);
    expect(within(groups).getByRole("tab", { name: "Allgemeine Stammdaten" }))
      .toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("tab", { name: "Hersteller" })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("columnheader", { name: "Herstellerseite" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "Suchdomains" })).toBeInTheDocument();
    expect(document.querySelector(".master-data-table-manufacturer")).toBeInTheDocument();

    await user.click(within(groups).getByRole("tab", { name: "Artikelstammdaten" }));
    expect(screen.getByRole("tab", { name: "Bestandseinheiten" })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("tab", { name: "Artikelarten und Unterarten" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Kontrollierte Zusatzfelder" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Lagerorte" })).toBeInTheDocument();
    expect(screen.queryByRole("tab", { name: "Hersteller" })).not.toBeInTheDocument();

    const query = new URLSearchParams(window.location.search);
    expect(query.get("tab")).toBe("data");
    expect(query.get("group")).toBe("article");
    expect(query.get("type")).toBe("stock_unit");
  });

  it("shows management metadata and confirms safe deactivation", async () => {
    const user = userEvent.setup();
    window.history.replaceState(null, "", "/settings?tab=data");
    vi.mocked(api.session).mockResolvedValue({
      username: "editor",
      roles: ["Editor"],
      csrfToken: "test",
      twoFactorEnabled: false
    });
    const manufacturer = {
      id: "manufacturer-tillig",
      type: "manufacturer",
      key: "tillig",
      label: "Tillig",
      active: true,
      sortOrder: 10,
      metadata: {},
      origin: "bundled" as const,
      capabilities: { canDeactivate: true, canReactivate: false, canDelete: false },
      createdAt: "2026-08-16T10:00:00Z",
      updatedAt: "2026-08-16T10:00:00Z"
    };
    vi.mocked(api.managedMasterDataAll).mockResolvedValue({ manufacturer: [manufacturer] });
    vi.spyOn(api, "setMasterDataActive").mockResolvedValue({
      ...manufacturer,
      active: false,
      capabilities: { canDeactivate: false, canReactivate: true, canDelete: false }
    });

    render(<SettingsView username="editor" />);

    expect(await screen.findByRole("columnheader", { name: "Status" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "Herkunft" })).toBeInTheDocument();
    expect(screen.getByText("RailKeeper")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Tillig deaktivieren" }));

    expect(screen.getByRole("dialog", { name: "Stammdateneintrag deaktivieren" }))
      .toHaveTextContent("Bestehende gespeicherte Verwendungen bleiben unverändert");
    await user.click(screen.getByRole("button", { name: "Deaktivieren" }));

    await waitFor(() => expect(api.setMasterDataActive)
      .toHaveBeenCalledWith("manufacturer", "tillig", false));
  });

  it("deactivates selected visible master data with one confirmed batch request", async () => {
    const user = userEvent.setup();
    window.history.replaceState(null, "", "/settings?tab=data");
    vi.mocked(api.session).mockResolvedValue({
      username: "editor",
      roles: ["Editor"],
      csrfToken: "test",
      twoFactorEnabled: false
    });
    const manufacturer = (key: string, label: string): MasterDataEntry => ({
      id: `manufacturer-${key}`,
      type: "manufacturer",
      key,
      label,
      active: true,
      sortOrder: 10,
      metadata: {},
      origin: "bundled",
      capabilities: { canDeactivate: true, canReactivate: false, canDelete: false },
      createdAt: "2026-08-16T10:00:00Z",
      updatedAt: "2026-08-16T10:00:00Z"
    });
    const tillig = manufacturer("tillig", "Tillig");
    const roco = manufacturer("roco", "Roco");
    vi.mocked(api.managedMasterDataAll).mockResolvedValue({ manufacturer: [tillig, roco] });
    vi.spyOn(api, "setMasterDataActiveMany").mockResolvedValue([tillig, roco].map((item) => ({
      ...item,
      active: false,
      capabilities: { canDeactivate: false, canReactivate: true, canDelete: false }
    })));

    render(<SettingsView username="editor" />);

    await user.click(await screen.findByRole("checkbox", {
      name: "Alle sichtbaren aktiven Einträge auswählen"
    }));
    expect(screen.getByText("2 Einträge ausgewählt")).toBeVisible();
    await user.click(screen.getByRole("button", { name: "Ausgewählte deaktivieren" }));
    expect(screen.getByRole("dialog", { name: "Ausgewählte Stammdateneinträge deaktivieren" }))
      .toHaveTextContent("2 ausgewählte Einträge aus „Hersteller“");
    await user.click(screen.getByRole("button", { name: "2 Einträge deaktivieren" }));

    await waitFor(() => expect(api.setMasterDataActiveMany)
      .toHaveBeenCalledWith("manufacturer", ["tillig", "roco"], false));
    expect(screen.queryByText("2 Einträge ausgewählt")).not.toBeInTheDocument();
    expect(screen.getAllByText("Inaktiv")).toHaveLength(2);
  });

  it.each([
    ["", "Allgemein", "Sprache, Startseite, Datumsformat und Druckausgabe."],
    ["?tab=data", "Daten", "Stammdaten für Fahrzeuge, Artikel und Anlagen zentral pflegen."],
    ["?tab=digital", "Digitalzentralen", "Zentrale Verbindungen konfigurieren, testen und für spätere Live-Aktualisierungen vorbereiten."],
    ["?tab=importExport", "Backup", "Stammdaten übertragen, vollständige Sicherungen erstellen und kontrolliert wiederherstellen."],
    ["?tab=appearance", "Darstellung", "Design-Optionen und Anzeigeeinstellungen werden hier gebündelt."],
    ["?tab=auth", "Authentifizierung", "Ihre Instanz ist mit lokaler Benutzeranmeldung geschützt."]
  ])("uses the active tab as the only page header for %s", async (search, title, description) => {
    window.history.replaceState(null, "", `/settings${search}`);
    const view = render(<SettingsView username="viewer" />);

    expect(await screen.findByRole("heading", { level: 1, name: title })).toBeInTheDocument();
    expect(screen.getAllByText(description)).toHaveLength(1);
    expect(screen.queryByRole("heading", { level: 2, name: title })).not.toBeInTheDocument();
    view.unmount();
  });

  it("presents backup operations as three aligned operational lanes", async () => {
    window.history.replaceState(null, "", "/settings?tab=importExport");

    render(<SettingsView username="viewer" />);

    expect(await screen.findByRole("button", { name: "Backup" })).toBeInTheDocument();
    const workspace = document.querySelector(".backup-operational-lanes");
    expect(workspace).toBeInTheDocument();
    expect(workspace?.querySelectorAll(".backup-operational-lane")).toHaveLength(3);
    expect(within(workspace as HTMLElement).getByRole("heading", { name: "Stammdaten" }))
      .toBeInTheDocument();
    expect(within(workspace as HTMLElement).getByRole("heading", { name: "Backup erstellen" }))
      .toBeInTheDocument();
    expect(within(workspace as HTMLElement).getByRole("heading", { name: "Backup wiederherstellen" }))
      .toBeInTheDocument();
    expect(workspace?.querySelector(".backup-restore-controls"))
      .toContainElement(screen.getByRole("button", { name: "Backup einspielen" }));
  });

  it("keeps the Digital Centers configuration workflow inside the full Settings navigation", async () => {
    window.history.replaceState(null, "", "/settings?tab=digital");

    render(<SettingsView username="viewer" />);

    expect(await screen.findByRole("heading", { level: 3, name: "Adapter wählen" }))
      .toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: "Einstellungen" })).toBeInTheDocument();
  });

  it("keeps the version out of the active page heading", async () => {
    vi.mocked(api.version).mockResolvedValue({
      version: "9.9.9",
      latestVersion: "9.9.9",
      updateAvailable: false,
      checkedAt: "2026-08-09T00:00:00Z",
      status: "current",
      message: ""
    });
    window.history.replaceState(null, "", "/settings?tab=data");
    render(<SettingsView username="viewer" />);

    expect(await screen.findByRole("heading", { level: 1, name: "Daten" })).toHaveTextContent(/^Daten$/);
  });

  it("composes the trusted Windows package action in the updates card", async () => {
    vi.mocked(api.version).mockResolvedValue({
      version: "0.1.0",
      latestVersion: "v0.2.0",
      updateAvailable: true,
      releaseUrl: "https://github.com/ichwars/RailKeeper/releases/tag/v0.2.0",
      windowsPackage: {
        version: "v0.2.0",
        name: "RailKeeper-windows-x64-v0.2.0.zip",
        url: "https://github.com/ichwars/RailKeeper/releases/download/v0.2.0/RailKeeper-windows-x64-v0.2.0.zip"
      },
      checkedAt: "2026-08-16T16:00:00Z",
      status: "update_available",
      message: "Eine neuere RailKeeper-Version ist verfügbar."
    });

    render(<SettingsView username="viewer" />);

    expect(await screen.findByRole("link", { name: "Version v0.2.0 herunterladen" }))
      .toHaveAttribute("href", expect.stringContaining("RailKeeper-windows-x64-v0.2.0.zip"));
  });

  it("enables Anlage as a start page and displays a stored layout preference", async () => {
    const user = userEvent.setup();
    window.localStorage.setItem("railkeeper.settings.defaultView", "layouts");
    render(<SettingsView username="viewer" />);

    const trigger = await screen.findByRole("button", { name: /Standardansicht/ });
    expect(trigger).toHaveTextContent("Anlage");
    await user.click(trigger);
    expect(screen.getByRole("option", { name: "Anlage" })).toBeEnabled();
    expect(window.localStorage.getItem("railkeeper.settings.defaultView")).toBe("layouts");
  });
});
