import type { Dispatch, FormEvent, SetStateAction } from "react";
import { History, KeyRound, Mail, Pencil, RefreshCw, Save, Send, Shield, Trash2, UserCog, Users, X } from "lucide-react";
import type { AuditLogEntry, Role, Session, SessionRecord, SMTPSettings, SMTPSettingsInput, UserAccount } from "../../shared/api";
import { AppSelect } from "../../shared/ui/AppSelect";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type UserFormState = {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  roles: string[];
};

type PasswordFormState = {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
};

type SettingsAuthTabProps = {
  t: Translate;
  currentSession: Session | null;
  twoFactorPrepared: boolean;
  setTwoFactorPrepared: Dispatch<SetStateAction<boolean>>;
  setLocalBool: (key: string, value: boolean, setter: (value: boolean) => void) => void;
  twoFactorSettingKey: string;
  authMessage: string;
  loadCurrentSession: () => void;
  changePassword: (event: FormEvent<HTMLFormElement>) => void;
  passwordForm: PasswordFormState;
  setPasswordForm: Dispatch<SetStateAction<PasswordFormState>>;
  passwordSaving: boolean;
  passwordMessage: string;
  smtpSettings: SMTPSettings | null;
  smtpForm: SMTPSettingsInput & { password: string; testRecipient: string };
  setSmtpForm: Dispatch<SetStateAction<SMTPSettingsInput & { password: string; testRecipient: string }>>;
  smtpLoading: boolean;
  smtpSaving: boolean;
  smtpTesting: boolean;
  smtpMessage: string;
  loadSMTPSettings: () => void;
  saveSMTPSettings: (event: FormEvent<HTMLFormElement>) => void;
  testSMTPSettings: () => void;
  canManageUsers: boolean;
  startUserCreate: () => void;
  editingUser: UserAccount | null;
  saveUser: (event: FormEvent<HTMLFormElement>) => void;
  userForm: UserFormState;
  setUserForm: Dispatch<SetStateAction<UserFormState>>;
  availableRoles: Role[];
  toggleUserRole: (role: string, checked: boolean) => void;
  userSaving: boolean;
  usersLoading: boolean;
  users: UserAccount[];
  formatDateTime: (value: string) => string;
  startUserEdit: (user: UserAccount) => void;
  deleteUser: (user: UserAccount) => void;
  loadSessions: () => void;
  sessionsLoading: boolean;
  sessions: SessionRecord[];
  revokeSession: (session: SessionRecord) => void;
  sessionsMessage: string;
  loadAuditLog: () => void;
  auditLogLoading: boolean;
  auditLog: AuditLogEntry[];
  auditLabel: (action: string) => string;
  auditActor: (entry: AuditLogEntry) => string;
  auditTarget: (entry: AuditLogEntry) => string;
  auditLogMessage: string;
  roleDescription: (role: string) => string;
};

export function SettingsAuthTab({
  t,
  currentSession,
  twoFactorPrepared,
  setTwoFactorPrepared,
  setLocalBool,
  twoFactorSettingKey,
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
  formatDateTime,
  startUserEdit,
  deleteUser,
  loadSessions,
  sessionsLoading,
  sessions,
  revokeSession,
  sessionsMessage,
  loadAuditLog,
  auditLogLoading,
  auditLog,
  auditLabel,
  auditActor,
  auditTarget,
  auditLogMessage,
  roleDescription
}: SettingsAuthTabProps) {
  const currentRoles = currentSession?.roles || [];
  const userCountLabel = t("settings.users.count", { count: users.length });

  return (
        <section className="auth-settings-grid">
          <section className="panel settings-card settings-tool-card auth-status-card">
            <div className="settings-card-title">
              <Shield size={18} />
              <div>
                <h2>{t("settings.auth.title")}</h2>
                <p>{t("settings.auth.subtitle")}</p>
              </div>
            </div>
            <div className="auth-provider-tabs" aria-label={t("settings.auth.provider")}>
              <button type="button" className="active"><Mail size={15} /> {t("settings.auth.emailLocal")}</button>
              <button type="button" disabled><KeyRound size={15} /> {t("settings.auth.twoFactor")}</button>
            </div>
            <div className="auth-status-grid" aria-label={t("settings.auth.status")}>
              <article>
                <span className="settings-pill active">{t("common.active")}</span>
                <strong>{t("settings.auth.local")}</strong>
                <small>{t("settings.auth.localHelp")}</small>
              </article>
              <article>
                <span className="settings-pill">{t("common.roles", { count: currentSession?.roles.length || 0 })}</span>
                <strong>{t("settings.auth.currentSession")}</strong>
                <small>{currentSession?.username ? t("settings.auth.signedInAs", { username: currentSession.username }) : t("settings.auth.loading")}</small>
              </article>
              <article>
                <span className={twoFactorPrepared ? "settings-pill active" : "settings-pill muted"}>{twoFactorPrepared ? t("settings.auth.prepared") : t("settings.auth.open")}</span>
                <strong>{t("settings.auth.twoFactor")}</strong>
                <small>{t("settings.auth.twoFactorHelp")}</small>
              </article>
            </div>
            <label className="settings-toggle-row">
              <span>
                <strong>{t("settings.auth.local")}</strong>
                <small>{t("settings.auth.localToggleHelp")}</small>
              </span>
              <span className="switch-field">
                <input type="checkbox" checked readOnly disabled />
                <span />
              </span>
            </label>
            <label className="settings-toggle-row disabled">
              <span>
                <strong>{t("settings.auth.twoFactorPrepare")}</strong>
                <small>{t("settings.auth.twoFactorPrepareHelp")}</small>
              </span>
              <span className="switch-field">
                <input type="checkbox" checked={twoFactorPrepared} onChange={(event) => setLocalBool(twoFactorSettingKey, event.target.checked, setTwoFactorPrepared)} />
                <span />
              </span>
            </label>
            {authMessage && <p className="form-message">{authMessage}</p>}
          </section>

          <section className="panel settings-card settings-tool-card smtp-settings-card">
            <div className="settings-section-head">
              <div className="settings-card-title">
                <Mail size={18} />
                <div>
                  <h2>{t("settings.smtp.title")}</h2>
                  <p>{t("settings.smtp.subtitle")}</p>
                </div>
              </div>
              <button type="button" className="icon-button" onClick={loadSMTPSettings} disabled={!canManageUsers || smtpLoading} aria-label={t("settings.smtp.refresh")} title={t("settings.smtp.refresh")}>
                <RefreshCw size={16} />
              </button>
            </div>
            {!canManageUsers ? (
              <div className="current-user-card">
                <strong>{t("settings.users.adminRequired")}</strong>
                <span>{t("settings.smtp.adminHelp")}</span>
              </div>
            ) : (
              <form className="settings-form" onSubmit={saveSMTPSettings}>
                <label className="settings-toggle-row smtp-toggle-row">
                  <span>
                    <strong>{t("settings.smtp.enabled")}</strong>
                    <small>{t("settings.smtp.enabledHelp")}</small>
                  </span>
                  <span className="switch-field">
                    <input type="checkbox" checked={smtpForm.enabled} onChange={(event) => setSmtpForm((current) => ({ ...current, enabled: event.target.checked }))} />
                    <span />
                  </span>
                </label>
                <div className="settings-field-grid">
                  <label>
                    {t("settings.smtp.publicUrl")}
                    <input value={smtpForm.publicUrl} onChange={(event) => setSmtpForm((current) => ({ ...current, publicUrl: event.target.value }))} placeholder="http://localhost:8080" />
                  </label>
                  <label>
                    {t("settings.smtp.host")}
                    <input value={smtpForm.host} onChange={(event) => setSmtpForm((current) => ({ ...current, host: event.target.value }))} placeholder="smtp.example.com" />
                  </label>
                  <label>
                    {t("settings.smtp.port")}
                    <input inputMode="numeric" value={smtpForm.port} onChange={(event) => setSmtpForm((current) => ({ ...current, port: event.target.value }))} placeholder="587" />
                  </label>
                  <label>
                    {t("settings.smtp.tlsMode")}
                    <AppSelect value={smtpForm.tlsMode} onChange={(event) => setSmtpForm((current) => ({ ...current, tlsMode: event.target.value }))}>
                      <option value="starttls">STARTTLS</option>
                      <option value="implicit">{t("settings.smtp.tlsImplicit")}</option>
                      <option value="none">{t("settings.smtp.tlsNone")}</option>
                    </AppSelect>
                  </label>
                  <label>
                    {t("settings.smtp.from")}
                    <input type="email" value={smtpForm.from} onChange={(event) => setSmtpForm((current) => ({ ...current, from: event.target.value }))} placeholder="railkeeper@example.com" />
                  </label>
                  <label>
                    {t("settings.smtp.username")}
                    <input value={smtpForm.username} onChange={(event) => setSmtpForm((current) => ({ ...current, username: event.target.value }))} autoComplete="username" />
                  </label>
                  <label>
                    {t("settings.smtp.password")}
                    <input type="password" value={smtpForm.password} onChange={(event) => setSmtpForm((current) => ({ ...current, password: event.target.value, clearPassword: false }))} placeholder={smtpSettings?.passwordConfigured ? t("settings.smtp.passwordKeep") : ""} autoComplete="new-password" />
                  </label>
                  <label>
                    {t("settings.smtp.testRecipient")}
                    <input type="email" value={smtpForm.testRecipient} onChange={(event) => setSmtpForm((current) => ({ ...current, testRecipient: event.target.value }))} placeholder="admin@example.com" />
                  </label>
                </div>
                <label className="settings-toggle-row smtp-toggle-row smtp-clear-password-row">
                  <span>
                    <strong>{t("settings.smtp.clearPassword")}</strong>
                    <small>{smtpSettings?.passwordConfigured ? t("settings.smtp.passwordConfigured") : t("settings.smtp.passwordMissing")}</small>
                  </span>
                  <span className="switch-field">
                    <input type="checkbox" checked={Boolean(smtpForm.clearPassword)} onChange={(event) => setSmtpForm((current) => ({ ...current, clearPassword: event.target.checked, password: event.target.checked ? "" : current.password }))} disabled={!smtpSettings?.passwordConfigured} />
                    <span />
                  </span>
                </label>
                <div className="settings-action-row">
                  <span className={smtpForm.enabled ? "settings-pill active" : "settings-pill muted"}>
                    {smtpForm.enabled ? t("common.active") : t("settings.smtp.disabled")}
                  </span>
                  <button type="submit" className="primary-button" disabled={smtpSaving || smtpLoading}>
                    <Save size={16} />
                    {smtpSaving ? t("vehicles.saving") : t("vehicles.save")}
                  </button>
                  <button type="button" className="secondary-button" onClick={testSMTPSettings} disabled={smtpTesting || smtpLoading || !smtpForm.testRecipient.trim()}>
                    <Send size={16} />
                    {smtpTesting ? t("settings.smtp.testing") : t("settings.smtp.test")}
                  </button>
                </div>
                {smtpMessage && <p className="form-message">{smtpMessage}</p>}
              </form>
            )}
          </section>

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

          <section className="panel settings-card settings-tool-card user-management-card">
            <div className="settings-section-head">
              <div className="settings-card-title">
                <Users size={18} />
                <div>
                  <h2>{t("settings.users.title")}</h2>
                  <p>{t("settings.users.subtitle")}</p>
                </div>
              </div>
              <div className="user-management-toolbar">
                <span className="settings-pill">{userCountLabel}</span>
                {canManageUsers && (
                  <button type="button" className="secondary-button" onClick={startUserCreate}>
                    <UserCog size={16} />
                    {t("settings.users.new")}
                  </button>
                )}
              </div>
            </div>

            {!canManageUsers ? (
              <div className="current-user-card">
                <strong>{t("settings.users.adminRequired")}</strong>
                <span>{t("settings.users.adminHelp")}</span>
              </div>
            ) : (
              <div className="user-management-grid">
                <form className="settings-form user-form" onSubmit={saveUser}>
                  <div className="user-form-head">
                    <h3>{editingUser ? t("settings.users.edit") : t("settings.users.create")}</h3>
                    {editingUser && <span className="settings-pill active">{editingUser.username}</span>}
                  </div>
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
                    />
                  </label>
                  <div className="role-select-grid" aria-label={t("settings.users.roles")}>
                    {availableRoles.map((role) => (
                      <label className="checkbox-field" key={role.id}>
                        <input
                          type="checkbox"
                          checked={userForm.roles.includes(role.name)}
                          onChange={(event) => toggleUserRole(role.name, event.target.checked)}
                        />
                        {role.name}
                      </label>
                    ))}
                  </div>
                  <div className="settings-action-row">
                    <button type="submit" className="primary-button" disabled={userSaving || userForm.roles.length === 0}>
                      {userSaving ? t("vehicles.saving") : t("vehicles.save")}
                    </button>
                    {editingUser && (
                      <button type="button" className="secondary-button" onClick={startUserCreate}>
                        {t("vehicles.cancel")}
                      </button>
                    )}
                  </div>
                </form>

                <div className="table-wrap settings-inline-table user-table">
                  <table>
                    <thead>
                      <tr>
                        <th>{t("settings.users.username")}</th>
                        <th>{t("auth.email")}</th>
                        <th>{t("settings.users.roles")}</th>
                        <th>{t("settings.users.created")}</th>
                        <th>{t("settings.sessions.actions")}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {usersLoading ? (
                        <tr><td colSpan={5} className="loading-cell">{t("settings.users.loading")}</td></tr>
                      ) : users.length === 0 ? (
                        <tr><td colSpan={5} className="loading-cell">{t("settings.users.empty")}</td></tr>
                      ) : (
                        users.map((user) => (
                          <tr key={user.id} className={editingUser?.id === user.id ? "selected-row" : ""}>
                            <td>
                              <div className="user-cell">
                                <strong>{user.username}</strong>
                                {currentSession?.username === user.username && <span className="settings-pill active">{t("settings.users.current")}</span>}
                              </div>
                            </td>
                            <td>{user.email || "-"}</td>
                            <td>
                              <div className="role-chip-row">
                                {user.roles.map((role) => <span className="settings-pill" key={role}>{role}</span>)}
                              </div>
                            </td>
                            <td>{formatDateTime(user.createdAt)}</td>
                            <td>
                              <div className="table-actions">
                                <button type="button" className="icon-button" onClick={() => startUserEdit(user)} aria-label={t("vehicles.edit")} title={t("vehicles.edit")}>
                                  <Pencil size={16} />
                                </button>
                                <button type="button" className="icon-button danger" onClick={() => deleteUser(user)} aria-label={t("vehicles.delete")} title={t("vehicles.delete")}>
                                  <Trash2 size={16} />
                                </button>
                              </div>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </section>

          <section className="panel settings-card settings-tool-card session-management-card">
            <div className="settings-section-head">
              <div className="settings-card-title">
                <KeyRound size={18} />
                <div>
                  <h2>{t("settings.sessions.title")}</h2>
                  <p>{t("settings.sessions.subtitle")}</p>
                </div>
              </div>
              {canManageUsers && (
                <button type="button" className="icon-button" onClick={loadSessions} disabled={sessionsLoading} aria-label={t("settings.sessions.refresh")} title={t("settings.sessions.refresh")}>
                  <RefreshCw size={16} />
                </button>
              )}
            </div>

            {!canManageUsers ? (
              <div className="current-user-card">
                <strong>{t("settings.users.adminRequired")}</strong>
                <span>{t("settings.sessions.adminHelp")}</span>
              </div>
            ) : (
              <>
                <div className="table-wrap settings-inline-table session-table">
                  <table>
                    <thead>
                      <tr>
                        <th>{t("settings.sessions.user")}</th>
                        <th>{t("settings.sessions.status")}</th>
                        <th>{t("settings.sessions.created")}</th>
                        <th>{t("settings.sessions.expires")}</th>
                        <th>{t("settings.sessions.actions")}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {sessionsLoading ? (
                        <tr><td colSpan={5} className="loading-cell">{t("settings.sessions.loading")}</td></tr>
                      ) : sessions.length === 0 ? (
                        <tr><td colSpan={5} className="loading-cell">{t("settings.sessions.empty")}</td></tr>
                      ) : (
                        sessions.map((session) => (
                          <tr key={session.id} className={session.active ? "" : "muted-row"}>
                            <td><strong>{session.username}</strong></td>
                            <td><span className={session.active ? "settings-pill active" : "settings-pill muted"}>{session.active ? t("common.active") : t("settings.sessions.ended")}</span></td>
                            <td>{formatDateTime(session.createdAt)}</td>
                            <td>{session.revokedAt ? t("settings.sessions.revoked", { date: formatDateTime(session.revokedAt) }) : formatDateTime(session.expiresAt)}</td>
                            <td>
                              <button type="button" className="icon-button danger" onClick={() => revokeSession(session)} disabled={!session.active || sessionsLoading} aria-label={t("settings.sessions.revoke")} title={t("settings.sessions.revoke")}>
                                <X size={16} />
                              </button>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
                {sessionsMessage && <p className="form-message">{sessionsMessage}</p>}
              </>
            )}
          </section>

          <section className="panel settings-card settings-tool-card audit-log-card">
            <div className="settings-section-head">
              <div className="settings-card-title">
                <History size={18} />
                <div>
                  <h2>{t("settings.audit.title")}</h2>
                  <p>{t("settings.audit.subtitle")}</p>
                </div>
              </div>
              {canManageUsers && (
                <button type="button" className="icon-button" onClick={loadAuditLog} disabled={auditLogLoading} aria-label={t("settings.audit.refresh")} title={t("settings.audit.refresh")}>
                  <RefreshCw size={16} />
                </button>
              )}
            </div>

            {!canManageUsers ? (
              <div className="current-user-card">
                <strong>{t("settings.users.adminRequired")}</strong>
                <span>{t("settings.audit.adminHelp")}</span>
              </div>
            ) : (
              <>
                <div className="table-wrap settings-inline-table audit-log-table">
                  <table>
                    <thead>
                      <tr>
                        <th>{t("settings.audit.time")}</th>
                        <th>{t("settings.audit.event")}</th>
                        <th>{t("settings.sessions.user")}</th>
                        <th>{t("settings.audit.target")}</th>
                        <th>{t("settings.audit.details")}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {auditLogLoading ? (
                        <tr><td colSpan={5} className="loading-cell">{t("settings.audit.loading")}</td></tr>
                      ) : auditLog.length === 0 ? (
                        <tr><td colSpan={5} className="loading-cell">{t("settings.audit.empty")}</td></tr>
                      ) : (
                        auditLog.slice(0, 10).map((entry) => (
                          <tr key={entry.id}>
                            <td>{formatDateTime(entry.createdAt)}</td>
                            <td><span className="settings-pill">{auditLabel(entry.action)}</span></td>
                            <td>{auditActor(entry)}</td>
                            <td>{auditTarget(entry)}</td>
                            <td><code>{entry.detailsJson && entry.detailsJson !== "{}" ? entry.detailsJson : "-"}</code></td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
                {auditLogMessage && <p className="form-message">{auditLogMessage}</p>}
              </>
            )}
          </section>

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

        </section>
  );
}
