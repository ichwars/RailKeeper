import { History, RefreshCw } from "lucide-react";
import { auditActor, auditTarget, formatDateTime } from "./settingsModel";
import type { SettingsAuthController } from "./useSettingsAuth";

type SettingsAuditCardProps = {
  t: (key: string, values?: Record<string, string | number>) => string;
  canManageUsers: SettingsAuthController["canManageUsers"];
  loadAuditLog: SettingsAuthController["loadAuditLog"];
  auditLogLoading: SettingsAuthController["auditLogLoading"];
  auditLog: SettingsAuthController["auditLog"];
  auditLabel: (action: string) => string;
  auditLogMessage: SettingsAuthController["auditLogMessage"];
};

export function SettingsAuditCard({
  t, canManageUsers, loadAuditLog, auditLogLoading, auditLog, auditLabel, auditLogMessage,
}: SettingsAuditCardProps) {
  return (
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
  );
}
