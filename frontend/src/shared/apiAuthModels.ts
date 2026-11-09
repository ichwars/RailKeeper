import type { components } from "../generated/api";

export type SetupStatus = components["schemas"]["SetupStatus"];

export type CreateAdminRequest = {
  username: string;
  email: string;
  password: string;
  setupCode: string;
};

export type LoginRequest = {
  username: string;
  password: string;
  twoFactorCode?: string;
};

export type Session = {
  username: string;
  roles: string[];
  csrfToken: string;
  twoFactorEnabled: boolean;
};

export type Role = {
  id: string;
  name: string;
};

export type UserAccount = {
  id: string;
  username: string;
  email?: string;
  roles: string[];
  createdAt: string;
  twoFactorEnabled: boolean;
};

export type UserAccountInput = {
  username: string;
  email?: string;
  password?: string;
  roles: string[];
};

export type PasswordResetRequest = {
  email: string;
};

export type PasswordResetResult = {
  status: string;
  message: string;
};

export type PasswordResetConfirmRequest = {
  token: string;
  newPassword: string;
};

export type ChangePasswordInput = {
  currentPassword: string;
  newPassword: string;
};

export type TwoFactorStatus = {
  enabled: boolean;
  prepared: boolean;
  enabledAt?: string;
  username?: string;
  otpauthUrl?: string;
  secret?: string;
};

export type TwoFactorEnableInput = {
  code: string;
};

export type TwoFactorDisableInput = {
  currentPassword: string;
  code: string;
};

export type SMTPSettings = {
  enabled: boolean;
  publicUrl: string;
  host: string;
  port: string;
  username: string;
  from: string;
  tlsMode: string;
  passwordConfigured: boolean;
};

export type SMTPSettingsInput = {
  enabled: boolean;
  publicUrl: string;
  host: string;
  port: string;
  username: string;
  password?: string;
  from: string;
  tlsMode: string;
  clearPassword?: boolean;
};

export type SMTPTestRequest = {
  recipient: string;
};

export type SettingsPayload = {
  settings: Record<string, string>;
};

export type DigitalProviderSettings = {
  enabled: boolean;
  host: string;
  port: string;
};

export type DigitalCenterSettings = {
  provider: "ecos" | "z21" | "intellibox3" | "cs3";
  ecos: DigitalProviderSettings;
  z21: DigitalProviderSettings;
  intellibox3: DigitalProviderSettings;
  cs3: DigitalProviderSettings;
};

export type SessionRecord = {
  id: string;
  userId: string;
  username: string;
  createdAt: string;
  expiresAt: string;
  revokedAt?: string;
  active: boolean;
};
