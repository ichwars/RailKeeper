import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { SettingsView } from "./SettingsView";
import { masterDataTypes, settingsTabs } from "./settingsModel";

describe("SettingsView article management navigation", () => {
  beforeEach(() => {
    window.history.replaceState(null, "", "/settings");
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
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

  it("exposes a stable article-management tab identifier", () => {
    expect(settingsTabs).toContainEqual({
      id: "articleManagement",
      labelKey: "settings.tabs.articleManagement"
    });
  });

  it("does not expose manufacturers in the legacy master-data administration", async () => {
    window.history.replaceState(null, "", "/settings?tab=data");
    vi.spyOn(api, "masterDataAll").mockResolvedValue({ manufacturer: [], vehicle_category: [] });

    render(<SettingsView username="viewer" />);

    expect(await screen.findByRole("button", { name: "Kategorie" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Hersteller" })).not.toBeInTheDocument();
    expect(masterDataTypes).not.toContainEqual({ type: "manufacturer" });
  });

  it("keeps the selected settings tab in the query path", async () => {
    const user = userEvent.setup();
    render(<SettingsView username="viewer" />);

    await user.click(screen.getByRole("button", { name: "Artikelverwaltung" }));

    expect(new URLSearchParams(window.location.search).get("tab")).toBe("articleManagement");
  });

  it("opens article management directly from its query path", async () => {
    window.history.replaceState(null, "", "/settings?tab=articleManagement");
    render(<SettingsView username="viewer" />);

    expect(await screen.findByRole("heading", { name: "Artikelverwaltung" })).toBeInTheDocument();
  });
});
