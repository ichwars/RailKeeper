import { Database, Download, HardDrive, RefreshCw, ShieldAlert, Upload } from "lucide-react";
import { api } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppFilePicker } from "../../shared/ui/AppFilePicker";
import { formatBytes } from "./settingsModel";
import type { SettingsStorageController } from "./useSettingsStorage";
import type { SettingsTransferController } from "./useSettingsTransfer";

type SettingsTransferTabProps = {
  transfer: SettingsTransferController;
  storage: SettingsStorageController;
};

export function SettingsTransferTab({ transfer, storage }: SettingsTransferTabProps) {
  const { t } = useI18n();
  const {
    backupFile, backupValidation, backupMessage, backupRestoreConfirm, setBackupRestoreConfirm,
    backupSaving, backupValidating, backupFileInputKey, backupRestorePhrase, backupRestoreConfirmed,
    selectBackupFile, restoreBackup,
    masterDataFile, setMasterDataFile, masterDataMessage, setMasterDataMessage,
    masterDataSaving, masterDataFileInputKey, importMasterData
  } = transfer;
  const { storageUsage, storageFileCount, storageLoading, loadStorageUsage } = storage;

  return (
    <section className="panel settings-card import-export-card backup-operational-lanes">
      <section className="backup-operational-lane">
        <div className="backup-lane-heading">
          <Database size={18} aria-hidden="true" />
          <div>
            <h3>{t("settings.masterTransfer.title")}</h3>
            <p>{t("settings.masterTransfer.subtitle")}</p>
          </div>
        </div>
        <div className="backup-lane-content">
            <AppFilePicker label={t("settings.masterTransfer.file")}
              triggerLabel={t("settings.file.pick")} emptyLabel={t("settings.file.none")}
              clearLabel={t("accessories.editor.documents.clear")} file={masterDataFile}
              key={masterDataFileInputKey}
              accept="application/json,.json"
              onFileChange={(selected) => {
                setMasterDataFile(selected);
                setMasterDataMessage("");
              }}
            />
          {masterDataSaving && <span className="inline-status">{t("settings.masterTransfer.saving")}</span>}
        </div>
        <div className="backup-lane-actions">
          <a className="secondary-button" href={api.masterDataExportUrl()}>
            <Download size={16} />
            {t("settings.masterTransfer.download")}
          </a>
          <button type="button" className="secondary-button" onClick={importMasterData} disabled={masterDataSaving || !masterDataFile}>
            <Upload size={16} />
            {t("settings.masterTransfer.upload")}
          </button>
        </div>
        {masterDataMessage && <p className="form-message">{masterDataMessage}</p>}
      </section>

      <section className="backup-operational-lane">
        <div className="backup-lane-heading">
          <HardDrive size={18} aria-hidden="true" />
          <div>
            <h3>{t("settings.backup.export.title")}</h3>
            <p>{t("settings.backup.export.subtitle")}</p>
          </div>
        </div>
        <div className="backup-lane-content">
          <div className="backup-summary-strip">
            <span>
              <strong>{formatBytes(storageUsage?.totalBytes || 0)}</strong>
              {t("settings.backup.storage")}
            </span>
            <span>
              <strong>{storageFileCount.toLocaleString("de-DE")}</strong>
              {t("settings.backup.files")}
            </span>
            <button type="button" className="icon-button" onClick={loadStorageUsage} disabled={storageLoading} aria-label={t("settings.backup.refreshStorage")} title={t("settings.backup.refreshStorage")}>
              <RefreshCw size={15} />
            </button>
          </div>
        </div>
        <div className="backup-lane-actions">
          <a className="primary-button backup-export-button" href={api.backupExportUrl()} download>
            <Download size={17} />
            {t("settings.backup.create")}
          </a>
        </div>
      </section>

      <section className="backup-operational-lane warning">
        <div className="backup-lane-heading">
          <ShieldAlert size={18} aria-hidden="true" />
          <div>
            <h3>{t("settings.backup.restore.title")}</h3>
            <p>{t("settings.backup.restore.subtitle")}</p>
          </div>
        </div>
        <div className="backup-lane-workspace">
          <div className="backup-restore-controls">
              <AppFilePicker label={t("settings.backup.file")}
                triggerLabel={t("settings.file.pick")} emptyLabel={t("settings.file.none")}
                clearLabel={t("accessories.editor.documents.clear")} file={backupFile}
                key={backupFileInputKey}
                accept="application/json,.json"
                onFileChange={selectBackupFile}
              />
            {backupValidating && <p className="backup-validation-status">{t("settings.backup.validating")}</p>}
            <button type="button" className="secondary-button danger" onClick={restoreBackup} disabled={backupSaving || backupValidating || !backupValidation?.compatible || !backupRestoreConfirmed}>
              {backupSaving ? (
                t("settings.backup.restoring")
              ) : (
                <>
                  <Upload size={17} />
                  {t("settings.backup.restoreButton")}
                </>
              )}
            </button>
          </div>
          {backupValidation && (
            <div className={backupValidation.compatible ? "backup-validation ok" : "backup-validation danger"}>
              <strong>{backupValidation.compatible ? t("settings.backup.compatible") : t("settings.backup.incompatible")}</strong>
              <dl>
                <div>
                  <dt>{t("settings.backup.version")}</dt>
                  <dd>{backupValidation.version || "-"}</dd>
                </div>
                <div>
                  <dt>{t("settings.backup.tables")}</dt>
                  <dd>{backupValidation.tableCount}</dd>
                </div>
                <div>
                  <dt>{t("settings.backup.rows")}</dt>
                  <dd>{backupValidation.rowCount}</dd>
                </div>
                <div>
                  <dt>{t("settings.backup.files")}</dt>
                  <dd>
                    {backupValidation.fileCount} / {formatBytes(backupValidation.fileBytes)}
                  </dd>
                </div>
              </dl>
              {backupValidation.errors.length > 0 && (
                <div className="backup-validation-list danger">
                  <strong>{t("settings.backup.errors")}</strong>
                  <ul>
                    {backupValidation.errors.map((error) => (
                      <li key={error}>{error}</li>
                    ))}
                  </ul>
                </div>
              )}
              {backupValidation.warnings.length > 0 && (
                <div className="backup-validation-list warning">
                  <strong>{t("settings.backup.warnings")}</strong>
                  <ul>
                    {backupValidation.warnings.map((warning) => (
                      <li key={warning}>{warning}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}
          {backupValidation?.compatible && (
            <label className="backup-confirm-field">
              {t("settings.backup.confirmLabel")}
              <span>{t("settings.backup.confirmHelp", { phrase: backupRestorePhrase })}</span>
              <input
                value={backupRestoreConfirm}
                onChange={(event) => setBackupRestoreConfirm(event.target.value)}
                placeholder={backupRestorePhrase}
                autoComplete="off"
              />
            </label>
          )}
          <p className="source-note backup-note">
            <ShieldAlert size={16} aria-hidden="true" />
            <span>{t("settings.backup.note")}</span>
          </p>
        </div>
      </section>

      {backupMessage && <p className="form-message">{backupMessage}</p>}
    </section>
  );
}
