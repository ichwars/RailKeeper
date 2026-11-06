import { HardDrive, RefreshCw, Wrench } from "lucide-react";
import { useI18n } from "../../shared/i18n";
import { formatBytes, formatDateTime } from "./settingsModel";
import { StorageLocationPanel } from "./StorageLocationPanel";
import type { SettingsStorageController } from "./useSettingsStorage";

type SettingsStorageCardProps = {
  storage: SettingsStorageController;
  canManageUsers: boolean;
};

export function SettingsStorageCard({ storage, canManageUsers }: SettingsStorageCardProps) {
  const { language, t } = useI18n();
  const { storageUsage, storageMessage, storageLoading, storageOptimizing, loadStorageUsage, optimizeStorage } = storage;
  const storageCategoryLabel = (key: string, label: string) => {
    const normalized = `${key} ${label}`.toLocaleLowerCase("de-DE");
    const categoryKey =
      normalized.includes("database") || normalized.includes("datenbank") ? "database" :
      normalized.includes("thumbnail") || normalized.includes("vorschau") ? "thumbnails" :
      normalized.includes("image") || normalized.includes("bild") ? "images" :
      normalized.includes("attachment") || normalized.includes("beilage") ? "attachments" :
      normalized.includes("other") || normalized.includes("sonstig") ? "other" : "";
    if (!categoryKey) return label;
    return t(`settings.storage.category.${categoryKey}`);
  };
  const fileCountLabel = (count: number, bytes: number) =>
    t("settings.storage.fileCount", {
      count,
      suffix: count === 1 ? "" : language === "de" ? "en" : "s",
      size: formatBytes(bytes)
    });

  return (
    <section className="panel settings-card settings-tool-card">
      <div className="settings-section-head">
        <div className="settings-card-title">
          <HardDrive size={17} />
          <h2>{t("settings.storage.title")}</h2>
        </div>
        <div className="settings-card-actions">
          <button type="button" className="icon-button subtle-icon-action" onClick={optimizeStorage} aria-label={t("settings.storage.optimize")} title={t("settings.storage.optimize")} disabled={storageLoading || storageOptimizing}>
            <Wrench size={15} />
          </button>
          <button type="button" className="icon-button" onClick={loadStorageUsage} aria-label={t("settings.storage.refresh")} title={t("settings.data.refresh")} disabled={storageLoading || storageOptimizing}>
            <RefreshCw size={16} />
          </button>
        </div>
      </div>
      <p>{t("settings.storage.subtitle")}</p>
      {canManageUsers && <StorageLocationPanel />}
      <div className="storage-total">
        <strong>{formatBytes(storageUsage?.totalBytes || 0)}</strong>
        <span>{storageUsage?.updatedAt ? t("settings.storage.updated", { date: formatDateTime(storageUsage.updatedAt) }) : t("settings.storage.notUpdated")}</span>
      </div>
      <div className="storage-list">
        {(storageUsage?.categories || []).map((category) => {
          const percent = storageUsage?.totalBytes ? Math.round((category.bytes / storageUsage.totalBytes) * 100) : 0;
          return (
            <div className="storage-row" key={category.key}>
              <div>
                <strong>{storageCategoryLabel(category.key, category.label)}</strong>
                <span>{fileCountLabel(category.files, category.bytes)}</span>
              </div>
              <div className="storage-bar" aria-label={storageCategoryLabel(category.key, category.label) + ": " + percent + "%"}>
                <span style={{ width: Math.max(2, percent) + "%" }} />
              </div>
            </div>
          );
        })}
        {storageLoading && <p className="empty-state compact">{t("settings.storage.loading")}</p>}
        {!storageLoading && !storageUsage && <p className="empty-state compact">{t("settings.storage.empty")}</p>}
      </div>
      {storageMessage && <p className="form-message">{storageMessage}</p>}
    </section>
  );
}
