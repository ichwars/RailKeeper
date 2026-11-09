import { Mail, RefreshCw, Save, Send } from "lucide-react";
import { AppSelect } from "../../shared/ui/AppSelect";
import type { SettingsAuthController } from "./useSettingsAuth";

type SettingsSMTPCardProps = {
  t: (key: string, values?: Record<string, string | number>) => string;
  loadSMTPSettings: SettingsAuthController["loadSMTPSettings"];
  canManageUsers: SettingsAuthController["canManageUsers"];
  smtpLoading: SettingsAuthController["smtpLoading"];
  saveSMTPSettings: SettingsAuthController["saveSMTPSettings"];
  smtpForm: SettingsAuthController["smtpForm"];
  setSmtpForm: SettingsAuthController["setSmtpForm"];
  smtpSettings: SettingsAuthController["smtpSettings"];
  smtpSaving: SettingsAuthController["smtpSaving"];
  testSMTPSettings: SettingsAuthController["testSMTPSettings"];
  smtpTesting: SettingsAuthController["smtpTesting"];
  smtpMessage: SettingsAuthController["smtpMessage"];
};

export function SettingsSMTPCard({
  t, loadSMTPSettings, canManageUsers, smtpLoading, saveSMTPSettings, smtpForm, setSmtpForm,
  smtpSettings, smtpSaving, testSMTPSettings, smtpTesting, smtpMessage,
}: SettingsSMTPCardProps) {
  return (
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
  );
}
