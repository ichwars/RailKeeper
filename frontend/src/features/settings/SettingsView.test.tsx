import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { SettingsView } from "./SettingsView";
import { masterDataTypes, settingsTabs } from "./settingsModel";

describe("SettingsView data navigation", () => {
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

  it("keeps article management out of the main settings tabs", () => {
    expect(settingsTabs.map((tab) => tab.id)).not.toContain("articleManagement");
  });

  it("restores manufacturers as the first general master-data type", async () => {
    window.history.replaceState(null, "", "/settings?tab=data");
    vi.spyOn(api, "masterDataAll").mockResolvedValue({ manufacturer: [], vehicle_category: [] });

    render(<SettingsView username="viewer" />);

    expect(await screen.findByRole("button", { name: "Hersteller" })).toBeInTheDocument();
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
});
