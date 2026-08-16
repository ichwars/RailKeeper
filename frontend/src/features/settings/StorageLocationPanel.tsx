import { useCallback, useEffect, useState } from "react";
import { CheckCircle2, DatabaseBackup, FolderOpen, RefreshCw } from "lucide-react";

import { api, type StorageLocationInfo } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export function StorageLocationPanel() {
  const { t } = useI18n();
  const [info, setInfo] = useState<StorageLocationInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [busyAction, setBusyAction] = useState<"open" | "acknowledge" | null>(null);
  const [message, setMessage] = useState("");
  const loadFailedLabel = t("settings.storage.location.loadFailed");

  const load = useCallback(async () => {
    setLoading(true);
    setMessage("");
    try {
      setInfo(await api.storageLocationInfo());
    } catch (error) {
      setInfo(null);
      setMessage(error instanceof Error ? error.message : loadFailedLabel);
    } finally {
      setLoading(false);
    }
  }, [loadFailedLabel]);

  useEffect(() => {
    void load();
  }, [load]);

  const openFolder = async () => {
    setBusyAction("open");
    setMessage("");
    try {
      await api.openStorageFolder();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t("settings.storage.location.openFailed"));
    } finally {
      setBusyAction(null);
    }
  };

  const acknowledgeMigration = async () => {
    setBusyAction("acknowledge");
    setMessage("");
    try {
      await api.acknowledgeStorageMigration();
      setInfo((current) => current?.migrationReceipt ? {
        ...current,
        migrationReceipt: { ...current.migrationReceipt, acknowledged: true }
      } : current);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t("settings.storage.location.ackFailed"));
    } finally {
      setBusyAction(null);
    }
  };

  if (loading) {
    return <div className="storage-location-panel">
      <p className="empty-state compact">{t("settings.storage.location.loading")}</p>
    </div>;
  }

  if (!info) {
    return <div className="storage-location-panel storage-location-error">
      <p className="form-message">{message || t("settings.storage.location.loadFailed")}</p>
      <button type="button" className="secondary-button" onClick={() => void load()}>
        <RefreshCw size={15} aria-hidden="true" />
        {t("settings.storage.location.retry")}
      </button>
    </div>;
  }

  const modeLabel = t(`settings.storage.location.mode.${info.mode}`);
  const receipt = info.migrationReceipt;

  return <div className="storage-location-panel">
    <div className="storage-location-summary">
      <div>
        <span>{t("settings.storage.location.mode")}</span>
        <strong>{modeLabel}</strong>
      </div>
      {info.openFolderAvailable && <button type="button" className="secondary-button"
        onClick={() => void openFolder()} disabled={busyAction !== null}>
        <FolderOpen size={15} aria-hidden="true" />
        {t("settings.storage.location.open")}
      </button>}
    </div>
    <div className="storage-location-path-block">
      <span>{t("settings.storage.location.path")}</span>
      <code className="storage-location-path">{info.dataPath}</code>
    </div>

    {receipt && <section className={`storage-migration-receipt${receipt.acknowledged ? " acknowledged" : ""}`}>
      <div className="storage-migration-title">
        {receipt.acknowledged
          ? <CheckCircle2 size={17} aria-hidden="true" />
          : <DatabaseBackup size={17} aria-hidden="true" />}
        <strong>{t("settings.storage.location.migration.title")}</strong>
        {receipt.acknowledged && <span>{t("settings.storage.location.migration.acknowledged")}</span>}
      </div>
      <p>{t("settings.storage.location.migration.body")}</p>
      <dl>
        <div>
          <dt>{t("settings.storage.location.migration.source")}</dt>
          <dd><code className="storage-location-path">{receipt.sourcePath}</code></dd>
        </div>
        <div>
          <dt>{t("settings.storage.location.migration.target")}</dt>
          <dd><code className="storage-location-path">{receipt.targetPath}</code></dd>
        </div>
      </dl>
      {!receipt.acknowledged && <button type="button" className="secondary-button"
        onClick={() => void acknowledgeMigration()} disabled={busyAction !== null}>
        {t("settings.storage.location.migration.acknowledge")}
      </button>}
    </section>}
    {message && <p className="form-message">{message}</p>}
  </div>;
}
