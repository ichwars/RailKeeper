import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api, type MasterDataEntry } from "../../shared/api";
import { SettingsView } from "./SettingsView";

const scheme = {
  id: "scheme", category: "locomotive", prefix: "RK", nextNumber: 12, padding: 6,
  active: true, preview: "RK-000012", createdAt: "", updatedAt: ""
};
const smtp = {
  enabled: true, publicUrl: "", host: "mail.example.test", port: "587", username: "test",
  from: "test@example.test", tlsMode: "starttls", passwordConfigured: true
};

describe("Settings preservation across module boundaries", () => {
  beforeEach(() => {
    window.history.replaceState(null, "", "/settings");
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
    vi.spyOn(api, "updateProfileSettings").mockResolvedValue({ settings: {} });
    vi.spyOn(api, "session").mockResolvedValue({
      username: "admin", roles: ["Admin"], csrfToken: "test", twoFactorEnabled: false
    });
    vi.spyOn(api, "version").mockResolvedValue({
      version: "0.1.22", status: "local", checkedAt: "", updateAvailable: false, message: ""
    });
    vi.spyOn(api, "storageUsage").mockResolvedValue({ totalBytes: 0, categories: [], updatedAt: "" });
    vi.spyOn(api, "storageLocationInfo").mockResolvedValue({
      mode: "server", dataPath: "/synthetic/data", openFolderAvailable: false
    });
    vi.spyOn(api, "systemPrinters").mockResolvedValue({ status: "unavailable", message: "", printers: [] });
    vi.spyOn(api, "inventoryNumberSchemes").mockResolvedValue([scheme]);
    vi.spyOn(api, "roles").mockResolvedValue([]);
    vi.spyOn(api, "users").mockResolvedValue([]);
    vi.spyOn(api, "auditLog").mockResolvedValue({ entries: [] });
    vi.spyOn(api, "sessions").mockResolvedValue([]);
    vi.spyOn(api, "smtpSettings").mockResolvedValue(smtp);
    vi.spyOn(api, "twoFactorStatus").mockResolvedValue({ enabled: false, prepared: false });
    vi.spyOn(api, "managedMasterDataAll").mockResolvedValue({});
  });

  it("persists preferences immediately and preserves them across settings tabs", async () => {
    const user = userEvent.setup();
    render(<SettingsView username="admin" />);
    await user.click(await screen.findByRole("button", { name: /Standardansicht/ }));
    await user.click(screen.getByRole("option", { name: "Fahrzeugbestand" }));
    expect(api.updateProfileSettings).toHaveBeenCalledWith({ "railkeeper.settings.defaultView": "vehicles" });
    expect(localStorage.getItem("railkeeper.settings.defaultView")).toBe("vehicles");
    await user.click(screen.getByRole("button", { name: "Darstellung" }));
    await user.click(screen.getByRole("radio", { name: /^Dunkel/ }));
    expect(api.updateProfileSettings).toHaveBeenCalledWith({ "railkeeper.theme": "dark" });
    await user.click(screen.getByRole("button", { name: "Allgemein" }));
    expect(await screen.findByRole("button", { name: /Standardansicht/ })).toHaveTextContent("Fahrzeugbestand");
  });

  it("keeps beta update preferences separate from an explicit update check", async () => {
    const user = userEvent.setup();
    render(<SettingsView username="admin" />);
    await waitFor(() => expect(api.version).toHaveBeenCalledExactlyOnceWith(false, false));
    await user.click(screen.getByRole("checkbox", { name: /Beta-Versionen/ }));
    expect(api.updateProfileSettings).toHaveBeenCalledWith({ "railkeeper.settings.betaUpdates": "true" });
    expect(api.version).toHaveBeenCalledTimes(1);
    await user.click(screen.getByRole("button", { name: "Jetzt prüfen" }));
    await waitFor(() => expect(api.version).toHaveBeenLastCalledWith(true, true));
  });

  it("retains inventory-number drafts across tabs and writes only after saving", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "updateInventoryNumberScheme").mockResolvedValue({ ...scheme, prefix: "KEEP" });
    render(<SettingsView username="admin" />);
    const existingRow = (await screen.findByText("RK-000012")).closest("tr")!;
    const prefix = within(existingRow).getByDisplayValue("RK");
    const row = prefix.closest("tr");
    expect(row).not.toBeNull();
    await user.clear(prefix);
    await user.type(prefix, "KEEP");
    expect(api.updateInventoryNumberScheme).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Darstellung" }));
    await user.click(screen.getByRole("button", { name: "Allgemein" }));
    const preserved = await screen.findByDisplayValue("KEEP");
    const save = within(preserved.closest("tr")!).getByRole("button", { name: "Speichern" });
    await user.click(save);
    expect(api.updateInventoryNumberScheme).toHaveBeenCalledExactlyOnceWith("locomotive", {
      prefix: "KEEP", nextNumber: 12, padding: 6, active: true
    });
  });

  it("loads personal 2FA for viewers while keeping admin data inaccessible in the UI", async () => {
    window.history.replaceState(null, "", "/settings?tab=auth");
    vi.mocked(api.session).mockResolvedValue({
      username: "viewer", roles: ["Viewer"], csrfToken: "test", twoFactorEnabled: false
    });
    render(<SettingsView username="viewer" />);
    await waitFor(() => expect(api.twoFactorStatus).toHaveBeenCalledOnce());
    expect(api.users).not.toHaveBeenCalled();
    expect(api.roles).not.toHaveBeenCalled();
    expect(api.auditLog).not.toHaveBeenCalled();
    expect(api.sessions).not.toHaveBeenCalled();
    expect(api.smtpSettings).not.toHaveBeenCalled();
    expect(screen.queryByPlaceholderText("smtp.example.com")).not.toBeInTheDocument();
  });

  it("keeps a stored SMTP password unless the user explicitly clears it", async () => {
    const user = userEvent.setup();
    window.history.replaceState(null, "", "/settings?tab=auth");
    vi.spyOn(api, "updateSMTPSettings").mockResolvedValue(smtp);
    render(<SettingsView username="admin" />);
    await screen.findByDisplayValue(smtp.host);
    const form = document.querySelector(".smtp-settings-card form");
    expect(form).not.toBeNull();
    fireEvent.submit(form!);
    await waitFor(() => expect(api.updateSMTPSettings).toHaveBeenCalledTimes(1));
    expect(vi.mocked(api.updateSMTPSettings).mock.calls[0][0]).toMatchObject({
      password: undefined, clearPassword: false, host: smtp.host
    });
    await user.click(screen.getByRole("checkbox", { name: /^Gespeichertes Passwort löschen/ }));
    fireEvent.submit(form!);
    await waitFor(() => expect(api.updateSMTPSettings).toHaveBeenCalledTimes(2));
    expect(vi.mocked(api.updateSMTPSettings).mock.calls[1][0]).toMatchObject({ clearPassword: true });
  });

  it("rejects mismatched passwords locally and refreshes sessions/audit after a valid change", async () => {
    const user = userEvent.setup();
    window.history.replaceState(null, "", "/settings?tab=auth");
    vi.spyOn(api, "changePassword").mockResolvedValue(undefined);
    vi.mocked(api.session).mockResolvedValue({
      username: "viewer", roles: ["Viewer"], csrfToken: "test", twoFactorEnabled: false
    });
    render(<SettingsView username="admin" />);
    await user.type(await screen.findByLabelText("Aktuelles Passwort"), "synthetic-password");
    await user.type(screen.getByLabelText("Neues Passwort"), "synthetic-new-password");
    await user.type(screen.getByLabelText("Neues Passwort wiederholen"), "different-password");
    await user.click(screen.getByRole("button", { name: "Passwort speichern" }));
    expect(api.changePassword).not.toHaveBeenCalled();
    await user.clear(screen.getByLabelText("Neues Passwort wiederholen"));
    await user.type(screen.getByLabelText("Neues Passwort wiederholen"), "synthetic-new-password");
    await user.click(screen.getByRole("button", { name: "Passwort speichern" }));
    await waitFor(() => expect(api.changePassword).toHaveBeenCalledExactlyOnceWith({
      currentPassword: "synthetic-password", newPassword: "synthetic-new-password"
    }));
    await waitFor(() => expect(api.sessions).toHaveBeenCalledTimes(1));
    expect(api.auditLog).toHaveBeenCalledTimes(1);
    expect(screen.getByLabelText("Aktuelles Passwort")).toHaveValue("");
  });

  it("preserves unexposed manufacturer metadata during a label-only edit", async () => {
    const user = userEvent.setup();
    window.history.replaceState(null, "", "/settings?tab=data");
    const entry: MasterDataEntry = {
      id: "manufacturer", type: "manufacturer", key: "reference", label: "Reference maker",
      active: true, sortOrder: 7, createdAt: "", updatedAt: "", sourceUrl: "https://example.test/source",
      metadata: { website: "https://example.test", privateExtension: { original: 42 } }
    };
    vi.mocked(api.managedMasterDataAll).mockResolvedValue({ manufacturer: [entry] });
    vi.spyOn(api, "managedMasterData").mockResolvedValue([entry]);
    vi.spyOn(api, "updateMasterData").mockResolvedValue(entry);
    render(<SettingsView username="admin" />);
    const row = (await screen.findByText(entry.label)).closest("tr")!;
    await user.click(within(row).getByRole("button", { name: /bearbeiten/i }));
    const name = screen.getByDisplayValue(entry.label);
    await user.clear(name);
    await user.type(name, "Renamed maker");
    await user.click(screen.getByRole("button", { name: "Speichern" }));
    await waitFor(() => expect(api.updateMasterData).toHaveBeenCalledTimes(1));
    expect(vi.mocked(api.updateMasterData).mock.calls[0]).toEqual([
      "manufacturer", "reference", expect.objectContaining({
        key: "reference", label: "Renamed maker", active: true, sortOrder: 7,
        sourceUrl: entry.sourceUrl,
        metadata: expect.objectContaining({ privateExtension: { original: 42 }, website: "https://example.test" })
      })
    ]);
  });
});
