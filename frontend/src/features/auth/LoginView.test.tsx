import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError, api, type Session } from "../../shared/api";
import { LoginView } from "./LoginView";

const session: Session = { username: "tester", roles: ["Editor"], csrfToken: "synthetic", twoFactorEnabled: true };

async function enterCredentials() {
  const user = userEvent.setup();
  await user.type(screen.getByLabelText("Benutzername"), "tester");
  await user.type(screen.getByLabelText("Passwort"), "synthetic-password");
  return user;
}

describe("LoginView authentication and recovery", () => {
  beforeEach(() => window.history.replaceState(null, "", "/"));

  it("waits for the login acknowledgment before opening the session", async () => {
    let acknowledge: (value: Session) => void = () => { };
    vi.spyOn(api, "login").mockReturnValue(new Promise(resolve => { acknowledge = resolve; }));
    const onLogin = vi.fn();
    render(<LoginView onLogin={onLogin} />);
    const user = await enterCredentials();
    await user.click(screen.getByRole("button", { name: "Anmelden" }));
    expect(screen.getByRole("button", { name: "Wird angemeldet..." })).toBeDisabled();
    expect(onLogin).not.toHaveBeenCalled();
    acknowledge(session);
    await waitFor(() => expect(onLogin).toHaveBeenCalledWith(session));
    expect(api.login).toHaveBeenCalledWith({ username: "tester", password: "synthetic-password", twoFactorCode: undefined });
  });

  it("requests a fresh second factor and resets its challenge when the password changes", async () => {
    const login = vi.spyOn(api, "login").mockRejectedValueOnce(new ApiError("Code erforderlich", "two_factor_required", 401)).mockResolvedValue(session);
    const onLogin = vi.fn();
    render(<LoginView onLogin={onLogin} />);
    const user = await enterCredentials();
    await user.click(screen.getByRole("button", { name: "Anmelden" }));
    const code = await screen.findByLabelText("Zwei-Faktor-Code");
    expect(code).toHaveAttribute("autocomplete", "one-time-code");
    expect(screen.getByText(/Jeder Code gilt einmal/)).toBeInTheDocument();
    await user.type(code, "123456");
    await user.type(screen.getByLabelText("Passwort"), "2");
    expect(screen.queryByLabelText("Zwei-Faktor-Code")).not.toBeInTheDocument();
    login.mockRejectedValueOnce(new ApiError("Code erforderlich", "two_factor_required", 401));
    await user.click(screen.getByRole("button", { name: "Anmelden" }));
    expect(await screen.findByLabelText("Zwei-Faktor-Code")).toHaveValue("");
    await user.type(screen.getByLabelText("Zwei-Faktor-Code"), "654321");
    await user.click(screen.getByRole("button", { name: "Anmelden" }));
    await waitFor(() => expect(onLogin).toHaveBeenCalledWith(session));
    expect(login).toHaveBeenLastCalledWith({ username: "tester", password: "synthetic-password2", twoFactorCode: "654321" });
  });

  it("shows rejected login without opening a session and permits an explicit retry", async () => {
    vi.spyOn(api, "login").mockRejectedValue(new ApiError("Bitte später erneut versuchen", "rate_limited", 429));
    const onLogin = vi.fn();
    render(<LoginView onLogin={onLogin} />);
    const user = await enterCredentials();
    await user.click(screen.getByRole("button", { name: "Anmelden" }));
    expect(await screen.findByText("Bitte später erneut versuchen")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Anmelden" })).toBeEnabled();
    expect(onLogin).not.toHaveBeenCalled();
  });

  it("requests recovery with Enter, shows the neutral response and clears it on close", async () => {
    vi.spyOn(api, "requestPasswordReset").mockResolvedValue({ status: "accepted", message: "" });
    render(<LoginView onLogin={vi.fn()} />);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Passwort vergessen?" }));
    expect(screen.getByRole("button", { name: "Reset anfordern" })).toBeDisabled();
    await user.type(screen.getByLabelText("E-Mail-Adresse"), "tester@example.invalid{Enter}");
    expect(await screen.findByText(/Wenn die E-Mail-Adresse bekannt ist/)).toBeInTheDocument();
    expect(api.requestPasswordReset).toHaveBeenCalledWith({ email: "tester@example.invalid" });
    await user.click(screen.getByRole("button", { name: "Passwort vergessen?" }));
    expect(screen.queryByLabelText("E-Mail-Adresse")).not.toBeInTheDocument();
    expect(screen.queryByText(/Wenn die E-Mail-Adresse bekannt ist/)).not.toBeInTheDocument();
  });

  it("shows recovery errors and the acknowledged server message on retry", async () => {
    vi.spyOn(api, "requestPasswordReset").mockRejectedValueOnce(new Error("SMTP nicht verfügbar")).mockResolvedValue({ status: "accepted", message: "Anfrage angenommen" });
    render(<LoginView onLogin={vi.fn()} />);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Passwort vergessen?" }));
    await user.type(screen.getByLabelText("E-Mail-Adresse"), "tester@example.invalid");
    await user.click(screen.getByRole("button", { name: "Reset anfordern" }));
    expect(await screen.findByText("SMTP nicht verfügbar")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Reset anfordern" }));
    expect(await screen.findByText("Anfrage angenommen")).toBeInTheDocument();
  });

  it("removes a reset token from the URL, checks confirmation and preserves it after rejection", async () => {
    window.history.replaceState(null, "", "/password-reset?token=synthetic-reset-token");
    const reset = vi.spyOn(api, "confirmPasswordReset").mockRejectedValueOnce(new Error("Reset abgelehnt")).mockResolvedValue(undefined);
    render(<LoginView onLogin={vi.fn()} />);
    expect(window.location.search).toBe("");
    const user = userEvent.setup();
    await user.type(screen.getByLabelText("Neues Passwort", { exact: true }), "new-synthetic-password");
    await user.type(screen.getByLabelText("Neues Passwort wiederholen"), "other-synthetic-password");
    await user.click(screen.getByRole("button", { name: "Passwort setzen" }));
    expect(await screen.findByText("Die Passwörter stimmen nicht überein.")).toBeInTheDocument();
    expect(reset).not.toHaveBeenCalled();
    await user.clear(screen.getByLabelText("Neues Passwort wiederholen"));
    await user.type(screen.getByLabelText("Neues Passwort wiederholen"), "new-synthetic-password");
    await user.click(screen.getByRole("button", { name: "Passwort setzen" }));
    expect(await screen.findByText("Reset abgelehnt")).toBeInTheDocument();
    expect(screen.getByLabelText("Neues Passwort", { exact: true })).toHaveValue("new-synthetic-password");
    await user.click(screen.getByRole("button", { name: "Passwort setzen" }));
    expect(await screen.findByRole("button", { name: "Anmelden" })).toBeInTheDocument();
    expect(screen.getByText(/Passwort wurde geändert/)).toBeInTheDocument();
    expect(window.location.pathname).toBe("/");
    expect(reset).toHaveBeenLastCalledWith({ token: "synthetic-reset-token", newPassword: "new-synthetic-password" });
  });

  it("can leave an unused reset link without submitting it", async () => {
    window.history.replaceState(null, "", "/password-reset?token=synthetic");
    const reset = vi.spyOn(api, "confirmPasswordReset");
    render(<LoginView onLogin={vi.fn()} />);
    await userEvent.setup().click(screen.getByRole("button", { name: "Zur Anmeldung" }));
    expect(screen.getByRole("button", { name: "Anmelden" })).toBeInTheDocument();
    expect(window.location.pathname).toBe("/");
    expect(reset).not.toHaveBeenCalled();
  });
});
