import { KeyRound, RefreshCw, X } from "lucide-react";
import { formatDateTime } from "./settingsModel";
import type { SettingsAuthController } from "./useSettingsAuth";

type SettingsSessionsCardProps = {
  t: (key: string, values?: Record<string, string | number>) => string;
  canManageUsers: SettingsAuthController["canManageUsers"];
  loadSessions: SettingsAuthController["loadSessions"];
  sessionsLoading: SettingsAuthController["sessionsLoading"];
  sessions: SettingsAuthController["sessions"];
  revokeSession: SettingsAuthController["revokeSession"];
  sessionsMessage: SettingsAuthController["sessionsMessage"];
};

export function SettingsSessionsCard({
  t, canManageUsers, loadSessions, sessionsLoading, sessions, revokeSession, sessionsMessage,
}: SettingsSessionsCardProps) {
  return (
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
  );
}
