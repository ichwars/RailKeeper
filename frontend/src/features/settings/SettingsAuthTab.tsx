import { userAccountValidation } from "./userAccountValidation";
import { useEffect, useState } from "react";
import { useI18n } from "../../shared/i18n";
import { AppChoiceInput } from "../../shared/ui/AppChoiceInput";
import { SettingsAuditCard } from "./SettingsAuditCard";
import { SettingsSMTPCard } from "./SettingsSMTPCard";
import { SettingsSessionsCard } from "./SettingsSessionsCard";
import { SettingsUsersCard } from "./SettingsUsersCard";
import type { SettingsAuthController } from "./useSettingsAuth";

import { KeyRound, RefreshCw, Save, Shield, UserCog, Users, X } from "lucide-react";
import type { UserAccount } from "../../shared/api";

export function SettingsAuthTab({ controller }: { controller: SettingsAuthController }) {
  const { t } = useI18n();
  const {
    currentSession,
    twoFactorStatus,
    twoFactorCode,
    setTwoFactorCode,
    twoFactorPassword,
    setTwoFactorPassword,
    twoFactorLoading,
    twoFactorSaving,
    twoFactorMessage,
    setupTwoFactor,
    enableTwoFactor,
    disableTwoFactor,
    authMessage,
    loadCurrentSession,
    changePassword,
    passwordForm,
    setPasswordForm,
    passwordSaving,
    passwordMessage,
    smtpSettings,
    smtpForm,
    setSmtpForm,
    smtpLoading,
    smtpSaving,
    smtpTesting,
    smtpMessage,
    loadSMTPSettings,
    saveSMTPSettings,
    testSMTPSettings,
    canManageUsers,
    startUserCreate,
    editingUser,
    saveUser,
    userForm,
    setUserForm,
    availableRoles,
    toggleUserRole,
    userSaving,
    usersLoading,
    users,
    startUserEdit,
    deleteUser,
    resetLoginBudget,
    loadSessions,
    sessionsLoading,
    sessions,
    revokeSession,
    sessionsMessage,
    loadAuditLog,
    auditLogLoading,
    auditLog,
    auditLogMessage
  } = controller;
  const auditLabel = (action: string) => {
    const label = t(`settings.auditAction.${action}`);
    return label === `settings.auditAction.${action}` ? action : label;
  };
  const roleDescription = (role: string) => {
    const description = t(`settings.role.${role}`);
    return description === `settings.role.${role}` ? t("settings.roles.custom") : description;
  };

  const [userDialogOpen, setUserDialogOpen] = useState(false);
  const [twoFactorDialogOpen, setTwoFactorDialogOpen] = useState(false);
  const [twoFactorDialogMode, setTwoFactorDialogMode] = useState<"setup" | "disable">("setup");
  const [twoFactorQrCode, setTwoFactorQrCode] = useState("");
  const currentRoles = currentSession?.roles || [];
  const userCountLabel = t("settings.users.count", { count: users.length });
  const twoFactorEnabled = Boolean(twoFactorStatus?.enabled || currentSession?.twoFactorEnabled);
  const { canSaveUser, userPasswordMessage, passwordRequired, passwordTouched } =
    userAccountValidation(userForm, editingUser, userSaving, t);

  const openUserCreateDialog = () => {
    startUserCreate();
    setUserDialogOpen(true);
  };

  const openUserEditDialog = (user: UserAccount) => {
    startUserEdit(user);
    setUserDialogOpen(true);
  };

  const closeUserDialog = () => {
    setUserDialogOpen(false);
    startUserCreate();
  };

  useEffect(() => {
    let active = true;
    if (!twoFactorDialogOpen || twoFactorDialogMode !== "setup" || !twoFactorStatus?.otpauthUrl) {
      setTwoFactorQrCode("");
      return () => {
        active = false;
      };
    }
    import("qrcode")
      .then((module) => module.default.toDataURL(twoFactorStatus.otpauthUrl || "", {
        width: 220,
        margin: 2,
        color: {
          dark: "#0b1114",
          light: "#ffffff"
        }
      }))
      .then((dataUrl) => {
        if (active) setTwoFactorQrCode(dataUrl);
      })
      .catch(() => {
        if (active) setTwoFactorQrCode("");
      });
    return () => {
      active = false;
    };
  }, [twoFactorDialogMode, twoFactorDialogOpen, twoFactorStatus?.otpauthUrl]);

  const openTwoFactorSetupDialog = () => {
    setTwoFactorDialogMode("setup");
    setTwoFactorPassword("");
    setupTwoFactor(() => setTwoFactorDialogOpen(true));
  };

  const openTwoFactorDisableDialog = () => {
    setTwoFactorDialogMode("disable");
    setTwoFactorCode("");
    setTwoFactorPassword("");
    setTwoFactorDialogOpen(true);
  };

  const closeTwoFactorDialog = () => {
    setTwoFactorDialogOpen(false);
    setTwoFactorCode("");
    setTwoFactorPassword("");
  };

  const toggleTwoFactor = (checked: boolean) => {
    if (checked) {
      if (!twoFactorEnabled) openTwoFactorSetupDialog();
      return;
    }
    if (twoFactorEnabled) openTwoFactorDisableDialog();
  };

  return (
        <section className="auth-settings-grid">
          <section className="panel settings-card settings-tool-card auth-status-card">
            <div className="auth-status-grid" aria-label={t("settings.auth.status")}>
              <article>
                <span className={twoFactorEnabled ? "settings-pill muted" : "settings-pill active"}>{twoFactorEnabled ? t("settings.auth.inactive") : t("common.active")}</span>
                <strong>{t("settings.auth.local")}</strong>
                <small>{twoFactorEnabled ? t("settings.auth.localDisabledHelp") : t("settings.auth.localHelp")}</small>
              </article>
              <article>
                <span className="settings-pill">{t("common.roles", { count: currentSession?.roles.length || 0 })}</span>
                <strong>{t("settings.auth.currentSession")}</strong>
                <small>{currentSession?.username ? t("settings.auth.signedInAs", { username: currentSession.username }) : t("settings.auth.loading")}</small>
              </article>
              <article>
                <span className={twoFactorEnabled ? "settings-pill active" : "settings-pill muted"}>
                  {twoFactorEnabled ? t("common.active") : t("settings.auth.inactive")}
                </span>
                <strong>{t("settings.auth.twoFactor")}</strong>
                <small>{t("settings.auth.twoFactorHelp")}</small>
              </article>
            </div>
            <label className="settings-toggle-row">
              <span>
                <strong>{t("settings.auth.local")}</strong>
                <small>{twoFactorEnabled ? t("settings.auth.localDisabledHelp") : t("settings.auth.localToggleHelp")}</small>
              </span>
              <span className="switch-field">
                <input type="checkbox" checked={!twoFactorEnabled} readOnly disabled />
                <span />
              </span>
            </label>
            <label className="settings-toggle-row">
              <span>
                <strong>{twoFactorEnabled ? t("settings.auth.twoFactorEnabledTitle") : t("settings.auth.twoFactorPrepare")}</strong>
                <small>{twoFactorEnabled ? t("settings.auth.twoFactorActiveHelp") : t("settings.auth.twoFactorPrepareHelp")}</small>
              </span>
              <span className="switch-field">
                <input
                  type="checkbox"
                  checked={twoFactorEnabled}
                  onChange={(event) => toggleTwoFactor(event.target.checked)}
                  disabled={twoFactorSaving || twoFactorLoading}
                />
                <span />
              </span>
            </label>
            {twoFactorMessage && <p className="form-message">{twoFactorMessage}</p>}
            {authMessage && <p className="form-message">{authMessage}</p>}
          </section>

          {twoFactorDialogOpen && (
            <div className="modal-layer two-factor-dialog-layer" role="dialog" aria-modal="true" aria-label={twoFactorDialogMode === "setup" ? t("settings.auth.twoFactorDialogTitle") : t("settings.auth.twoFactorDisableDialogTitle")}>
              <section className="vehicle-modal compact-modal two-factor-dialog">
                <header className="modal-head">
                  <div>
                    <h2>{twoFactorDialogMode === "setup" ? t("settings.auth.twoFactorDialogTitle") : t("settings.auth.twoFactorDisableDialogTitle")}</h2>
                    <p>{twoFactorDialogMode === "setup" ? t("settings.auth.twoFactorDialogSubtitle") : t("settings.auth.twoFactorDisableDialogSubtitle")}</p>
                  </div>
                  <button type="button" className="icon-button" onClick={closeTwoFactorDialog} aria-label={t("vehicles.cancel")} title={t("vehicles.cancel")}>
                    <X size={18} />
                  </button>
                </header>
                <div className="modal-body two-factor-dialog-body">
                  {twoFactorDialogMode === "setup" ? (
                    <>
                      <div className="two-factor-qr-panel">
                        {twoFactorQrCode ? (
                          <img src={twoFactorQrCode} alt={t("settings.auth.twoFactorQrAlt")} />
                        ) : (
                          <div className="two-factor-qr-placeholder">{t("settings.auth.twoFactorQrLoading")}</div>
                        )}
                        <small>{t("settings.auth.twoFactorQrHelp")}</small>
                      </div>
                      <label>
                        {t("settings.auth.twoFactorSecret")}
                        <input value={twoFactorStatus?.secret || ""} readOnly />
                      </label>
                      <label>
                        {t("settings.auth.twoFactorCode")}
                        <input value={twoFactorCode} inputMode="numeric" autoComplete="one-time-code" onChange={(event) => setTwoFactorCode(event.target.value)} />
                      </label>
                    </>
                  ) : (
                    <>
                      <label>
                        {t("settings.auth.twoFactorCode")}
                        <input value={twoFactorCode} inputMode="numeric" autoComplete="one-time-code" onChange={(event) => setTwoFactorCode(event.target.value)} />
                      </label>
                      <label>
                        {t("auth.password")}
                        <input type="password" value={twoFactorPassword} autoComplete="current-password" onChange={(event) => setTwoFactorPassword(event.target.value)} />
                      </label>
                    </>
                  )}
                  {twoFactorMessage && <p className="form-message">{twoFactorMessage}</p>}
                </div>
                <footer className="modal-actions">
                  <button type="button" className="secondary-button" onClick={closeTwoFactorDialog}>
                    {t("vehicles.cancel")}
                  </button>
                  {twoFactorDialogMode === "setup" ? (
                    <button type="button" className="primary-button" onClick={() => enableTwoFactor(() => setTwoFactorDialogOpen(false))} disabled={twoFactorSaving || !twoFactorCode.trim()}>
                      <Shield size={16} />
                      {twoFactorSaving ? t("vehicles.saving") : t("settings.auth.twoFactorEnable")}
                    </button>
                  ) : (
                    <button type="button" className="secondary-button danger" onClick={() => disableTwoFactor(() => setTwoFactorDialogOpen(false))} disabled={twoFactorSaving || !twoFactorPassword.trim() || !twoFactorCode.trim()}>
                      <X size={16} />
                      {twoFactorSaving ? t("vehicles.saving") : t("settings.auth.twoFactorDisable")}
                    </button>
                  )}
                </footer>
              </section>
            </div>
          )}

          <div className="auth-side-stack">
            <SettingsSMTPCard
            t={t}
            loadSMTPSettings={loadSMTPSettings}
            canManageUsers={canManageUsers}
            smtpLoading={smtpLoading}
            saveSMTPSettings={saveSMTPSettings}
            smtpForm={smtpForm}
            setSmtpForm={setSmtpForm}
            smtpSettings={smtpSettings}
            smtpSaving={smtpSaving}
            testSMTPSettings={testSMTPSettings}
            smtpTesting={smtpTesting}
            smtpMessage={smtpMessage}
          />

            <section className="panel settings-card settings-tool-card role-settings-card">
              <div className="settings-section-head">
                <div className="settings-card-title">
                  <Users size={18} />
                  <h2>{t("settings.roles.title")}</h2>
                </div>
                <span className="settings-pill active">{t("common.active")}</span>
              </div>
              <div className="role-list">
                {(availableRoles.length > 0 ? availableRoles.map((role) => role.name) : ["Admin", "Editor", "Viewer", "Messe"]).map((role) => (
                  <article key={role}>
                    <strong>{role}</strong>
                    <span>{roleDescription(role)}</span>
                  </article>
                ))}
              </div>
            </section>
          </div>

          {!canManageUsers && <section className="panel settings-card settings-tool-card current-user-settings-card">
            <div className="settings-section-head">
              <div className="settings-card-title">
                <UserCog size={18} />
                <div>
                  <h2>{t("settings.currentUser")}</h2>
                  <p>{t("settings.session.refreshHelp")}</p>
                </div>
              </div>
              <button type="button" className="icon-button" onClick={loadCurrentSession} aria-label={t("settings.session.refresh")} title={t("settings.session.refresh")}>
                <RefreshCw size={16} />
              </button>
            </div>
            <div className="account-security-grid">
              <div className="current-user-card account-summary-card">
                <span className="settings-pill active">{t("settings.auth.currentSession")}</span>
                <strong>{currentSession?.username || t("settings.notLoaded")}</strong>
                <div className="role-chip-row">
                  {currentRoles.map((role) => <span className="settings-pill" key={role}>{role}</span>)}
                  {currentRoles.length === 0 && <span className="settings-pill muted">{t("common.noRoles")}</span>}
                </div>
              </div>
              <form className="password-change-form password-change-panel" onSubmit={changePassword}>
                <div className="user-form-head">
                  <h3>{t("settings.password.title")}</h3>
                </div>
                <div className="password-field-grid">
                  <label>
                    {t("settings.password.current")}
                    <input
                      type="password"
                      value={passwordForm.currentPassword}
                      onChange={(event) => setPasswordForm((current) => ({ ...current, currentPassword: event.target.value }))}
                      autoComplete="current-password"
                    />
                  </label>
                  <label>
                    {t("settings.password.new")}
                    <input
                      type="password"
                      value={passwordForm.newPassword}
                      onChange={(event) => setPasswordForm((current) => ({ ...current, newPassword: event.target.value }))}
                      autoComplete="new-password"
                      minLength={12}
                    />
                  </label>
                  <label>
                    {t("settings.password.confirm")}
                    <input
                      type="password"
                      value={passwordForm.confirmPassword}
                      onChange={(event) => setPasswordForm((current) => ({ ...current, confirmPassword: event.target.value }))}
                      autoComplete="new-password"
                      minLength={12}
                    />
                  </label>
                </div>
                <button type="submit" className="primary-button" disabled={passwordSaving || !passwordForm.currentPassword || !passwordForm.newPassword || !passwordForm.confirmPassword}>
                  <KeyRound size={16} />
                  {t("settings.password.save")}
                </button>
                {passwordMessage && <p className="form-message">{passwordMessage}</p>}
              </form>
            </div>
          </section>}

          <SettingsUsersCard
            t={t}
            userCountLabel={userCountLabel}
            canManageUsers={canManageUsers}
            openUserCreateDialog={openUserCreateDialog}
            usersLoading={usersLoading}
            users={users}
            editingUser={editingUser}
            currentSession={currentSession}
            resetLoginBudget={resetLoginBudget}
            openUserEditDialog={openUserEditDialog}
            deleteUser={deleteUser}
          />

          {canManageUsers && userDialogOpen && (
            <div className="modal-layer user-dialog-layer" role="dialog" aria-modal="true" aria-label={editingUser ? t("settings.users.edit") : t("settings.users.create")}>
              <form className="vehicle-modal compact-modal user-dialog" onSubmit={(event) => saveUser(event, () => setUserDialogOpen(false))}>
                <header className="modal-head">
                  <div>
                    <h2>{editingUser ? t("settings.users.edit") : t("settings.users.create")}</h2>
                    {editingUser && <p>{editingUser.username}</p>}
                  </div>
                  <button type="button" className="icon-button" onClick={closeUserDialog} aria-label={t("vehicles.cancel")} title={t("vehicles.cancel")}>
                    <X size={18} />
                  </button>
                </header>
                <div className="modal-body settings-form user-form user-dialog-body">
                  <label>
                    {t("settings.users.username")}
                    <input
                      value={userForm.username}
                      onChange={(event) => setUserForm((current) => ({ ...current, username: event.target.value }))}
                      placeholder={t("settings.users.usernamePlaceholder")}
                    />
                  </label>
                  <label>
                    {t("auth.email")}
                    <input
                      type="email"
                      value={userForm.email}
                      onChange={(event) => setUserForm((current) => ({ ...current, email: event.target.value }))}
                      placeholder={t("settings.users.emailPlaceholder")}
                    />
                  </label>
                  <label>
                    {t("auth.password")}
                    <input
                      type="password"
                      value={userForm.password}
                      onChange={(event) => setUserForm((current) => ({ ...current, password: event.target.value }))}
                      placeholder={editingUser ? t("settings.users.passwordPlaceholderEdit") : t("settings.users.passwordPlaceholderNew")}
                      autoComplete="new-password"
                      minLength={passwordRequired || passwordTouched ? 12 : undefined}
                    />
                  </label>
                  <label>
                    {t("settings.users.passwordConfirm")}
                    <input
                      type="password"
                      value={userForm.confirmPassword}
                      onChange={(event) => setUserForm((current) => ({ ...current, confirmPassword: event.target.value }))}
                      placeholder={editingUser ? t("settings.users.passwordPlaceholderEdit") : t("settings.users.passwordPlaceholderNew")}
                      autoComplete="new-password"
                      minLength={passwordRequired || passwordTouched ? 12 : undefined}
                    />
                  </label>
                  <div className="role-select-grid" aria-label={t("settings.users.roles")}>
                    {availableRoles.map((role) => (
                      <label className="checkbox-field" key={role.id}>
                        <AppChoiceInput
                          type="checkbox"
                          checked={userForm.roles.includes(role.name)}
                          onChange={(event) => toggleUserRole(role.name, event.target.checked)}
                        />
                        {role.name}
                      </label>
                    ))}
                  </div>
                  {userPasswordMessage && <p className="form-message">{userPasswordMessage}</p>}
                  {authMessage && <p className="form-message">{authMessage}</p>}
                </div>
                <footer className="modal-actions">
                  <button type="button" className="secondary-button" onClick={closeUserDialog}>
                    {t("vehicles.cancel")}
                  </button>
                  <button type="submit" className="primary-button" disabled={!canSaveUser}>
                    <Save size={16} />
                    {userSaving ? t("vehicles.saving") : t("vehicles.save")}
                  </button>
                </footer>
              </form>
            </div>
          )}

          <SettingsSessionsCard
            t={t}
            canManageUsers={canManageUsers}
            loadSessions={loadSessions}
            sessionsLoading={sessionsLoading}
            sessions={sessions}
            revokeSession={revokeSession}
            sessionsMessage={sessionsMessage}
          />

          <SettingsAuditCard
            t={t}
            canManageUsers={canManageUsers}
            loadAuditLog={loadAuditLog}
            auditLogLoading={auditLogLoading}
            auditLog={auditLog}
            auditLabel={auditLabel}
            auditLogMessage={auditLogMessage}
          />

        </section>
  );
}
