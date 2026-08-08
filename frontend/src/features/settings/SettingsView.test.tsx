import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { SettingsView } from "./SettingsView";
import { masterDataTypes, settingsTabs } from "./settingsModel";

describe("SettingsView data navigation", () => {
  beforeEach(() => {
    window.history.replaceState(null, "", "/settings");
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
    vi.spyOn(api, "masterData").mockResolvedValue([]);
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

  it("restores manufacturers as the first general master-data type", async () => {
    window.history.replaceState(null, "", "/settings?tab=data");
    vi.spyOn(api, "masterDataAll").mockResolvedValue({ manufacturer: [], vehicle_category: [] });

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
});
