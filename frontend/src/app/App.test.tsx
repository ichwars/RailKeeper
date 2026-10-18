import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../shared/api";
import { App, configuredStartView, currentView, pathForView } from "./App";

vi.mock("../features/importExport/ImportExportView", () => ({
  ImportExportView: ({ roles }: { roles: string[] }) => <div>transfer roles: {roles.join(",")}</div>
}));

vi.mock("./Shell", () => ({
  Shell: ({ children }: { children: React.ReactNode }) => <div>{children}</div>
}));

describe("App navigation availability", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    window.history.replaceState(null, "", "/");
  });

  it("falls back from a retired layout start view", () => {
    window.localStorage.setItem("railkeeper.settings.defaultView", "layouts");

    expect(configuredStartView()).toBe("overview");
    expect(window.localStorage.getItem("railkeeper.settings.defaultView")).toBe("layouts");
  });

  it("falls back from the retired layout route", () => {
    window.history.replaceState(null, "", "/layouts");
    expect(currentView()).toBe("overview");
  });

  it("uses the query-free Digital Centers history path", () => {
    expect(pathForView("digitalCenters")).toBe("/digital-centers");
  });

  it("passes session roles into the data transfer workspace", async () => {
    window.history.replaceState(null, "", "/import-export");
    vi.spyOn(api, "setupStatus").mockResolvedValue({ setupRequired: false });
    vi.spyOn(api, "session").mockResolvedValue({
      username: "operator",
      roles: ["Messe", "Viewer"],
      csrfToken: "csrf",
      twoFactorEnabled: false
    });

    render(<App />);

    expect(await screen.findByText("transfer roles: Messe,Viewer")).toBeInTheDocument();
  });

  it("keeps the data transfer workspace available to pure Messe users", async () => {
    window.history.replaceState(null, "", "/import-export");
    vi.spyOn(api, "setupStatus").mockResolvedValue({ setupRequired: false });
    vi.spyOn(api, "session").mockResolvedValue({
      username: "messe",
      roles: ["Messe"],
      csrfToken: "csrf",
      twoFactorEnabled: false
    });

    render(<App />);

    expect(await screen.findByText("transfer roles: Messe")).toBeInTheDocument();
    expect(window.location.pathname).toBe("/import-export");
  });

  it("renders the dedicated Digital Centers workspace for its direct route", async () => {
    window.history.replaceState(null, "", "/digital-centers");
    vi.spyOn(api, "setupStatus").mockResolvedValue({ setupRequired: false });
    vi.spyOn(api, "session").mockResolvedValue({
      username: "admin",
      roles: ["Admin"],
      csrfToken: "csrf",
      twoFactorEnabled: false
    });

    render(<App />);

    expect(await screen.findByRole("heading", { level: 1, name: "Digitalzentralen" }))
      .toBeInTheDocument();
    expect(screen.getByText("DIGITALBETRIEB")).toBeInTheDocument();
    expect(screen.queryByRole("navigation", { name: "Einstellungen" })).not.toBeInTheDocument();
  });
});
