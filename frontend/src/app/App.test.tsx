import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, ApiError } from "../shared/api";
import { App, configuredStartView, currentView, pathForView } from "./App";

vi.mock("../features/importExport/ImportExportView", () => ({
  ImportExportView: ({ roles }: { roles: string[] }) => <div>transfer roles: {roles.join(",")}
    <input aria-label="Workspace draft" defaultValue="" /></div>
}));

vi.mock("./Shell", () => ({
  Shell: ({ children, onLogout }: { children: React.ReactNode; onLogout: () => void }) =>
    <div><button onClick={onLogout}>Logout</button>{children}</div>
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
    vi.spyOn(api, "setupStatus").mockResolvedValue({ setupRequired: false, cookieSecure: false });
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
    vi.spyOn(api, "setupStatus").mockResolvedValue({ setupRequired: false, cookieSecure: false });
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

  it("retains the current workspace after a failed logout and allows a successful retry", async () => {
    window.history.replaceState(null, "", "/import-export");
    vi.spyOn(api, "setupStatus").mockResolvedValue({ setupRequired: false, cookieSecure: false });
    vi.spyOn(api, "session").mockResolvedValue({
      username: "operator", roles: ["Viewer"], csrfToken: "csrf", twoFactorEnabled: false
    });
    const logout = vi.spyOn(api, "logout")
      .mockRejectedValueOnce(new Error("Connection lost. Please retry."))
      .mockResolvedValueOnce(undefined);

    render(<App />);
    expect(await screen.findByText("transfer roles: Viewer")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Workspace draft"), { target: { value: "unsaved draft" } });
    fireEvent.click(screen.getByRole("button", { name: "Logout" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Connection lost. Please retry.");
    expect(screen.getByLabelText("Workspace draft")).toHaveValue("unsaved draft");
    expect(screen.getByText("transfer roles: Viewer")).toBeInTheDocument();
    expect(window.location.pathname).toBe("/import-export");

    fireEvent.click(screen.getByRole("button", { name: "Logout" }));
    expect(await screen.findByLabelText("Benutzername")).toBeInTheDocument();
    expect(screen.queryByText("transfer roles: Viewer")).not.toBeInTheDocument();
    expect(logout).toHaveBeenCalledTimes(2);
  });

  it("renders the dedicated Digital Centers workspace for its direct route", async () => {
    window.history.replaceState(null, "", "/digital-centers");
    vi.spyOn(api, "setupStatus").mockResolvedValue({ setupRequired: false, cookieSecure: false });
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

  it.each([401, 403])("returns to login after logout rejection %s confirms an invalid session", async status => {
    window.history.replaceState(null, "", "/import-export");
    vi.spyOn(api, "setupStatus").mockResolvedValue({ setupRequired: false, cookieSecure: false });
    const session = vi.spyOn(api, "session")
      .mockResolvedValueOnce({ username: "operator", roles: ["Viewer"], csrfToken: "csrf", twoFactorEnabled: false })
      .mockRejectedValueOnce(new ApiError("Not logged in.", "unauthorized", 401));
    vi.spyOn(api, "logout").mockRejectedValueOnce(new ApiError("Rejected", "csrf_required", status));
    render(<App />);
    expect(await screen.findByText("transfer roles: Viewer")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Logout" }));
    expect(await screen.findByLabelText("Benutzername")).toBeInTheDocument();
    expect(screen.queryByText("transfer roles: Viewer")).not.toBeInTheDocument();
    expect(session).toHaveBeenCalledTimes(2);
  });

  it.each(["valid", "network", "storage", "forbidden"])("retains the workspace when session revalidation is %s", async result => {
    window.history.replaceState(null, "", "/import-export");
    vi.spyOn(api, "setupStatus").mockResolvedValue({ setupRequired: false, cookieSecure: false });
    const current = { username: "operator", roles: ["Viewer"], csrfToken: "csrf", twoFactorEnabled: false };
    const session = vi.spyOn(api, "session").mockResolvedValueOnce(current);
    if (result === "valid") session.mockResolvedValueOnce({ ...current, roles: ["Messe", "Viewer"], csrfToken: "renewed" });
    else session.mockRejectedValueOnce(result === "network" ? new Error("Network unavailable")
      : new ApiError("Lookup unavailable", "session_failed", result === "storage" ? 500 : 403));
    vi.spyOn(api, "logout").mockRejectedValueOnce(new ApiError("Retry logout", "csrf_required", 403))
      .mockResolvedValueOnce(undefined);
    render(<App />);
    expect(await screen.findByText("transfer roles: Viewer")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Workspace draft"), { target: { value: "unsaved retry draft" } });
    fireEvent.click(screen.getByRole("button", { name: "Logout" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Retry logout");
    expect(screen.getByLabelText("Workspace draft")).toHaveValue("unsaved retry draft");
    expect(screen.getByText(`transfer roles: ${result === "valid" ? "Messe,Viewer" : "Viewer"}`)).toBeInTheDocument();
    expect(window.location.pathname).toBe("/import-export");
    fireEvent.click(screen.getByRole("button", { name: "Logout" }));
    expect(await screen.findByLabelText("Benutzername")).toBeInTheDocument();
  });

  it("leaves the previous user's workspace when revalidation finds a different account", async () => {
    window.history.replaceState(null, "", "/import-export");
    vi.spyOn(api, "setupStatus").mockResolvedValue({ setupRequired: false, cookieSecure: false });
    vi.spyOn(api, "session")
      .mockResolvedValueOnce({ username: "operator", roles: ["Viewer"], csrfToken: "csrf", twoFactorEnabled: false })
      .mockResolvedValueOnce({ username: "other", roles: ["Viewer"], csrfToken: "other", twoFactorEnabled: false });
    vi.spyOn(api, "logout").mockRejectedValueOnce(new ApiError("Rejected", "csrf_required", 403));
    render(<App />);
    expect(await screen.findByText("transfer roles: Viewer")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Logout" }));
    expect(await screen.findByLabelText("Benutzername")).toBeInTheDocument();
    expect(screen.queryByText("transfer roles: Viewer")).not.toBeInTheDocument();
  });

  it("does not submit concurrent logout requests", async () => {
    window.history.replaceState(null, "", "/import-export");
    vi.spyOn(api, "setupStatus").mockResolvedValue({ setupRequired: false, cookieSecure: false });
    vi.spyOn(api, "session").mockResolvedValue({ username: "operator", roles: ["Viewer"], csrfToken: "csrf", twoFactorEnabled: false });
    let finish!: () => void;
    const logout = vi.spyOn(api, "logout").mockReturnValueOnce(new Promise(resolve => { finish = resolve; }));
    render(<App />);
    expect(await screen.findByText("transfer roles: Viewer")).toBeInTheDocument();
    const button = screen.getByRole("button", { name: "Logout" });
    fireEvent.click(button); fireEvent.click(button);
    expect(logout).toHaveBeenCalledTimes(1);
    finish();
    expect(await screen.findByLabelText("Benutzername")).toBeInTheDocument();
  });
});
