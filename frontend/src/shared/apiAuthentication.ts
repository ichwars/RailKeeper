import type {
  ChangePasswordInput, CreateAdminRequest, LoginRequest, PasswordResetConfirmRequest, PasswordResetRequest,
  PasswordResetResult, Role, Session, SessionRecord, SettingsPayload, SetupStatus, TwoFactorDisableInput,
  TwoFactorEnableInput, TwoFactorStatus, UserAccount, UserAccountInput
} from "./apiAuthModels";
import type { APIRequest } from "./apiTransport";
import { readCookie, setCSRFToken } from "./apiTransport";

export function createAuthenticationAPI(request: APIRequest) {
  return {
    setupStatus: () => request<SetupStatus>("/setup/status"),
    createAdmin: (input: CreateAdminRequest) =>
      request<{ status: string }>("/setup/admin", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    login: async (input: LoginRequest) => {
      const session = await request<Session>("/auth/login", {
        method: "POST",
        body: JSON.stringify(input)
      });
      setCSRFToken(session.csrfToken || readCookie("rk_csrf"));
      return session;
    },
    requestPasswordReset: (input: PasswordResetRequest) =>
      request<PasswordResetResult>("/auth/password-reset", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    confirmPasswordReset: (input: PasswordResetConfirmRequest) =>
      request<void>("/auth/password-reset/confirm", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    session: async () => {
      const session = await request<Session>("/auth/session");
      setCSRFToken(session.csrfToken || readCookie("rk_csrf"));
      return session;
    },
    profileSettings: () => request<SettingsPayload>("/profile/settings"),
    updateProfileSettings: (settings: Record<string, string>) =>
      request<SettingsPayload>("/profile/settings", {
        method: "PUT",
        body: JSON.stringify({ settings })
      }),
    logout: async () => {
      await request<void>("/auth/logout", { method: "POST" });
      setCSRFToken("");
    },
    changePassword: (input: ChangePasswordInput) =>
      request<void>("/auth/password", {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    twoFactorStatus: () => request<TwoFactorStatus>("/auth/two-factor"),
    setupTwoFactor: () =>
      request<TwoFactorStatus>("/auth/two-factor/setup", {
        method: "POST"
      }),
    enableTwoFactor: (input: TwoFactorEnableInput) =>
      request<TwoFactorStatus>("/auth/two-factor/enable", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    disableTwoFactor: (input: TwoFactorDisableInput) =>
      request<void>("/auth/two-factor/disable", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    roles: () => request<Role[]>("/roles"),
    users: () => request<UserAccount[]>("/users"),
    resetLoginBudget: (id: string) =>
      request<void>(`/users/${encodeURIComponent(id)}/login-budget/reset`, { method: "POST" }),
    createUser: (input: UserAccountInput) =>
      request<UserAccount>("/users", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    updateUser: (id: string, input: UserAccountInput) =>
      request<UserAccount>(`/users/${encodeURIComponent(id)}`, {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    deleteUser: (id: string) =>
      request<void>(`/users/${encodeURIComponent(id)}`, {
        method: "DELETE"
      }),
    sessions: (limit = 5) => request<SessionRecord[]>(`/sessions?limit=${encodeURIComponent(String(limit))}`),
    revokeSession: (id: string) =>
      request<void>(`/sessions/${encodeURIComponent(id)}/revoke`, {
        method: "PUT"
      })
  };
}
