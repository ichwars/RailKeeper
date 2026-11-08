import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type UserAccount } from "../../shared/api";
import { SettingsAuthTab } from "./SettingsAuthTab";
import type { SettingsConfirmDialog } from "./settingsModel";
import { useSettingsAuth } from "./useSettingsAuth";

const account: UserAccount = { id: "account", username: "operator", email: "operator@example.invalid", roles: ["Viewer"], createdAt: "2026-11-01T12:00:00Z", twoFactorEnabled: false };

function Workspace() {
  const [confirmation, setConfirmation] = useState<SettingsConfirmDialog | null>(null);
  const controller = useSettingsAuth({ username: "admin", activeSettingsTab: "auth", requestConfirmation: setConfirmation });
  return <><SettingsAuthTab controller={controller} />{confirmation && <div role="dialog" aria-label="Aktion bestätigen"><p>{confirmation.body}</p><button onClick={() => { confirmation.onConfirm(); setConfirmation(null); }}>Bestätigen</button><button onClick={() => setConfirmation(null)}>Verwerfen</button></div>}</>;
}

async function open() {
  render(<Workspace />);
  await screen.findByText(account.email!);
  return userEvent.setup();
}

describe("Authentication settings with the real controller", () => {
  beforeEach(() => {
    vi.spyOn(api, "session").mockResolvedValue({ username: "admin", roles: ["Admin"], csrfToken: "synthetic", twoFactorEnabled: false });
    vi.spyOn(api, "roles").mockResolvedValue(["Admin", "Editor", "Viewer", "Messe"].map(name => ({ id: name, name })));
    vi.spyOn(api, "users").mockResolvedValue([account]);
    vi.spyOn(api, "sessions").mockResolvedValue([{ id: "session", userId: account.id, username: account.username, createdAt: account.createdAt, expiresAt: "2026-12-01T12:00:00Z", active: true }]);
    vi.spyOn(api, "auditLog").mockResolvedValue({ entries: [{ id: "audit", actorUsername: "admin", action: "UserCreated", targetType: "user", targetId: account.id, createdAt: account.createdAt, detailsJson: "{}" }] });
    vi.spyOn(api, "smtpSettings").mockResolvedValue({ enabled: false, publicUrl: "", host: "", port: "587", username: "", from: "", tlsMode: "starttls", passwordConfigured: false });
    vi.spyOn(api, "twoFactorStatus").mockResolvedValue({ enabled: false, prepared: false });
  });

  it("keeps an existing password untouched when editing roles and retains rejected drafts", async () => {
    const update = vi.spyOn(api, "updateUser").mockRejectedValueOnce(new Error("Änderung abgelehnt")).mockResolvedValue({ ...account, roles: ["Viewer", "Editor"] });
    const user = await open();
    const row = within(screen.getByRole("row", { name: new RegExp(account.email!) }));
    await user.click(row.getByRole("button", { name: "Bearbeiten" }));
    const dialog = within(screen.getByRole("dialog", { name: "Benutzer bearbeiten" }));
    expect(dialog.getByLabelText("Passwort", { exact: true })).toHaveValue("");
    await user.click(dialog.getByRole("checkbox", { name: "Editor" }));
    await user.click(dialog.getByRole("button", { name: "Speichern" }));
    expect(await dialog.findByText("Änderung abgelehnt")).toBeInTheDocument();
    expect(dialog.getByRole("checkbox", { name: "Editor" })).toBeChecked();
    expect(update).toHaveBeenLastCalledWith(account.id, { username: account.username, email: account.email, password: undefined, roles: ["Viewer", "Editor"] });
    await user.click(dialog.getByRole("button", { name: "Speichern" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("requires a matching new password and at least one role before creating a user", async () => {
    const create = vi.spyOn(api, "createUser").mockResolvedValue({ ...account, id: "new", username: "new-user" });
    const user = await open();
    await user.click(screen.getByRole("button", { name: "Neuer Benutzer" }));
    const dialog = within(screen.getByRole("dialog", { name: "Benutzer anlegen" }));
    await user.type(dialog.getByLabelText("Benutzername"), "new-user");
    await user.type(dialog.getByLabelText("E-Mail-Adresse"), "new@example.invalid");
    await user.type(dialog.getByLabelText("Passwort", { exact: true }), "synthetic-password");
    expect(dialog.getByRole("button", { name: "Speichern" })).toBeDisabled();
    await user.type(dialog.getByLabelText("Passwort wiederholen"), "synthetic-password");
    await user.click(dialog.getByRole("checkbox", { name: "Viewer" }));
    expect(dialog.getByRole("button", { name: "Speichern" })).toBeDisabled();
    await user.click(dialog.getByRole("checkbox", { name: "Messe" }));
    expect(create).not.toHaveBeenCalled();
    await user.click(dialog.getByRole("button", { name: "Speichern" }));
    await waitFor(() => expect(create).toHaveBeenCalledWith({ username: "new-user", email: "new@example.invalid", password: "synthetic-password", roles: ["Messe"] }));
  });

  it("clears a cancelled account draft before another create dialog", async () => {
    const create = vi.spyOn(api, "createUser");
    const user = await open();
    await user.click(screen.getByRole("button", { name: "Neuer Benutzer" }));
    let dialog = within(screen.getByRole("dialog", { name: "Benutzer anlegen" }));
    await user.type(dialog.getByLabelText("Benutzername"), "discarded");
    await user.click(dialog.getAllByRole("button", { name: "Abbrechen" })[0]);
    await user.click(screen.getByRole("button", { name: "Neuer Benutzer" }));
    dialog = within(screen.getByRole("dialog", { name: "Benutzer anlegen" }));
    expect(dialog.getByLabelText("Benutzername")).toHaveValue("");
    expect(create).not.toHaveBeenCalled();
  });

  it("requires confirmation for user deletion and session revocation and reports rejection", async () => {
    const remove = vi.spyOn(api, "deleteUser").mockRejectedValueOnce(new Error("Letzter Administrator geschützt")).mockResolvedValue(undefined);
    const revoke = vi.spyOn(api, "revokeSession").mockResolvedValue(undefined);
    const user = await open();
    const row = within(screen.getByRole("row", { name: new RegExp(account.email!) }));
    await user.click(row.getByRole("button", { name: "Löschen" }));
    expect(remove).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Verwerfen" }));
    expect(remove).not.toHaveBeenCalled();
    await user.click(row.getByRole("button", { name: "Löschen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    expect(await screen.findByText("Letzter Administrator geschützt")).toBeInTheDocument();
    await user.click(row.getByRole("button", { name: "Löschen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    await waitFor(() => expect(remove).toHaveBeenCalledTimes(2));
    await user.click(screen.getByRole("button", { name: "Sitzung widerrufen" }));
    expect(revoke).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    await waitFor(() => expect(revoke).toHaveBeenCalledWith("session"));
    expect(api.sessions).toHaveBeenCalledTimes(2);
  });

  it("confirms login recovery, retains credentials and reports a rejected recovery", async () => {
    const reset = vi.spyOn(api, "resetLoginBudget").mockRejectedValueOnce(new Error("Freigabe abgelehnt")).mockResolvedValue(undefined);
    const user = await open();
    await user.click(screen.getByRole("button", { name: "Anmeldung freigeben" }));
    expect(reset).not.toHaveBeenCalled();
    expect(screen.getByText(/Gespeicherte Browserbelege werden widerrufen/)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Verwerfen" }));
    expect(reset).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Anmeldung freigeben" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    expect(await screen.findByText("Freigabe abgelehnt")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Anmeldung freigeben" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    expect(await screen.findByText(/Anmeldung freigegeben/)).toBeInTheDocument();
    expect(reset).toHaveBeenLastCalledWith(account.id);
    expect(api.auditLog).toHaveBeenCalledTimes(2);
  });

  it("prepares a second factor and activates only after an acknowledged fresh code", async () => {
    vi.spyOn(api, "setupTwoFactor").mockResolvedValue({ enabled: false, prepared: true, secret: "SYNTHETIC", otpauthUrl: "otpauth://totp/RailKeeper?secret=JBSWY3DPEHPK3PXP" });
    const enable = vi.spyOn(api, "enableTwoFactor").mockRejectedValueOnce(new Error("Code bereits verwendet")).mockResolvedValue({ enabled: true, prepared: true });
    const user = await open();
    await user.click(screen.getByRole("checkbox", { name: /Zwei-Faktor-Auth vorbereiten/ }));
    const dialog = within(await screen.findByRole("dialog", { name: "Zwei-Faktor-Auth einrichten" }));
    expect(dialog.getByRole("button", { name: "Aktivieren" })).toBeDisabled();
    expect(dialog.getByLabelText("Authenticator-Secret")).toHaveValue("SYNTHETIC");
    await user.type(dialog.getByLabelText("Bestätigungscode"), "123456");
    await user.click(dialog.getByRole("button", { name: "Aktivieren" }));
    expect(await dialog.findByText("Code bereits verwendet")).toBeInTheDocument();
    expect(dialog.getByLabelText("Bestätigungscode")).toHaveValue("123456");
    await user.clear(dialog.getByLabelText("Bestätigungscode"));
    await user.type(dialog.getByLabelText("Bestätigungscode"), "654321");
    vi.mocked(api.twoFactorStatus).mockResolvedValue({ enabled: true, prepared: true });
    await user.click(dialog.getByRole("button", { name: "Aktivieren" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(enable).toHaveBeenLastCalledWith({ code: "654321" });
  });

  it("requires password and code to disable a configured second factor", async () => {
    vi.mocked(api.twoFactorStatus).mockResolvedValue({ enabled: true, prepared: true });
    const disable = vi.spyOn(api, "disableTwoFactor").mockResolvedValue(undefined);
    const user = await open();
    await user.click(screen.getByRole("checkbox", { name: /Zwei-Faktor-Auth aktiv/ }));
    const dialog = within(screen.getByRole("dialog", { name: "Zwei-Faktor-Auth deaktivieren" }));
    expect(dialog.getByRole("button", { name: "Deaktivieren" })).toBeDisabled();
    await user.type(dialog.getByLabelText("Passwort"), "synthetic-password");
    expect(dialog.getByRole("button", { name: "Deaktivieren" })).toBeDisabled();
    await user.type(dialog.getByLabelText("Bestätigungscode"), "654321");
    await user.click(dialog.getByRole("button", { name: "Deaktivieren" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(disable).toHaveBeenCalledWith({ currentPassword: "synthetic-password", code: "654321" });
  });
});
