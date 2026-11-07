import { useI18n } from "../../shared/i18n";
import { RefreshCw } from "lucide-react";
import type { VersionInfo } from "../../shared/api";
import { formatDateTime } from "./settingsModel";
import { SettingsBuildIdentity } from "./SettingsBuildIdentity";
import { WindowsUpdateDownload } from "./WindowsUpdateDownload";
import type { SettingsPreferencesController } from "./useSettingsPreferences";
import type { SettingsGeneralController } from "./useSettingsGeneral";

export function SettingsUpdatesCard({ preferences, general }: { preferences: SettingsPreferencesController; general: SettingsGeneralController }) {
  const { t } = useI18n();
  const { betaUpdates, updateBetaUpdates } = preferences;
  const { versionInfo, versionMessage, versionLoading, loadVersionInfo } = general;
  const versionStatusLabel = (info: VersionInfo) => {
    if (info.updateAvailable) return t("settings.updates.status.updateAvailable");
    if (info.status === "current") return betaUpdates ? t("settings.updates.status.currentBeta") : t("settings.updates.status.current");
    if (info.status === "no_release") return t("settings.updates.status.noRelease");
    if (info.status === "unavailable") return t("settings.updates.status.offline");
    return t("settings.updates.status.local");
  };
  const localizedStatusMessage = (message: string) => {
    if (message === "Lokale RailKeeper-Version gelesen.") return t("settings.updates.message.local");
    if (message.includes("Windows-Systemdrucker")) return t("settings.printer.messageLoaded");
    if (message.includes("Keine Release-Information")) return t("settings.updates.message.noRelease");
    if (message.includes("RailKeeper ist aktuell")) return t("settings.updates.message.current");
    return message;
  };
  const updateVersionLabel = (version?: string) => version?.replace(/^v/i, "") || t("settings.unknown");
  const updateCheckedAtLabel = versionInfo?.checkedAt && versionInfo.status !== "local"
    ? t("settings.updates.checkedAt", { date: formatDateTime(versionInfo.checkedAt) })
    : t("settings.updates.notChecked");

  return (
    <section className="panel settings-card settings-tool-card">
      <div className="settings-card-title">
        <RefreshCw size={17} />
        <h2>{t("settings.updates.title")}</h2>
      </div>
      <label className="settings-toggle-row">
        <span>
          <strong>{t("settings.updates.beta")}</strong>
          <small>{t("settings.updates.betaHelp")}</small>
        </span>
        <span className="switch-field">
          <input type="checkbox" checked={betaUpdates} onChange={(event) => updateBetaUpdates(event.target.checked)} />
          <span />
        </span>
      </label>
      <div className="settings-action-row">
        <p>
          {t("settings.updates.currentVersion")}: <strong>{updateVersionLabel(versionInfo?.version)}</strong>
          {versionInfo?.latestVersion && <> · {t("settings.updates.latestVersion")}: <strong>{updateVersionLabel(versionInfo.latestVersion)}</strong></>}
          <br />
          <small>{updateCheckedAtLabel}</small>
          <SettingsBuildIdentity build={versionInfo?.build} />
        </p>
        {versionInfo?.status && (
          <span className={`settings-pill ${versionInfo.updateAvailable ? "active" : ["unavailable", "no_release"].includes(versionInfo.status) ? "muted" : ""}`}>
            {versionStatusLabel(versionInfo)}
          </span>
        )}
        <button type="button" className="secondary-button" onClick={() => loadVersionInfo(true)} disabled={versionLoading}>
          <RefreshCw size={17} />
          {versionLoading ? t("settings.updates.checking") : t("settings.updates.checkNow")}
        </button>
      </div>
      {versionInfo && <WindowsUpdateDownload info={versionInfo} />}
      {versionInfo?.updateAvailable && (
        <div className="update-decision-panel">
          <div>
            <strong>{t("settings.updates.decisionTitle", { version: versionInfo.latestVersion || "" })}</strong>
          </div>
          {versionInfo.releaseNotes && <pre>{versionInfo.releaseNotes}</pre>}
        </div>
      )}
      {versionMessage && <p className="form-message">{localizedStatusMessage(versionMessage)}</p>}
    </section>
  );
}
