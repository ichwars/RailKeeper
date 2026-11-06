import { useState } from "react";
import { api, type BackupValidationResult } from "../../shared/api";
import { decoderLibraryChanged } from "../../shared/DecoderLibraryProvider";
import { useI18n } from "../../shared/i18n";
import type { SettingsConfirmDialog } from "./settingsModel";

type SettingsTransferOptions = {
  requestConfirmation: (dialog: SettingsConfirmDialog) => void;
  invalidateMasterData: () => void;
  clearMasterDataSearch: () => void;
  loadStorageUsage: () => void;
};

export function useSettingsTransfer({
  requestConfirmation,
  invalidateMasterData,
  clearMasterDataSearch,
  loadStorageUsage
}: SettingsTransferOptions) {
  const { language, t } = useI18n();
  const [backupFile, setBackupFile] = useState<File | null>(null);
  const [backupValidation, setBackupValidation] = useState<BackupValidationResult | null>(null);
  const [backupMessage, setBackupMessage] = useState("");
  const [backupRestoreConfirm, setBackupRestoreConfirm] = useState("");
  const [backupSaving, setBackupSaving] = useState(false);
  const [backupValidating, setBackupValidating] = useState(false);
  const [backupFileInputKey, setBackupFileInputKey] = useState(0);
  const [masterDataFile, setMasterDataFile] = useState<File | null>(null);
  const [masterDataMessage, setMasterDataMessage] = useState("");
  const [masterDataSaving, setMasterDataSaving] = useState(false);
  const [masterDataFileInputKey, setMasterDataFileInputKey] = useState(0);
  const languageLocale = language === "de" ? "de-DE" : "en-US";
  const backupRestorePhrase = t("settings.backup.confirmPhrase");
  const backupRestoreConfirmed =
    backupRestoreConfirm.trim().toLocaleUpperCase(languageLocale) === backupRestorePhrase.toLocaleUpperCase(languageLocale);

  const selectBackupFile = (file: File | null) => {
    setBackupFile(file);
    setBackupValidation(null);
    setBackupMessage("");
    setBackupRestoreConfirm("");
    if (!file) return;

    setBackupValidating(true);
    api
      .validateBackup(file)
      .then((result) => setBackupValidation(result))
      .catch((error: Error) => setBackupMessage(error.message))
      .finally(() => setBackupValidating(false));
  };

  const restoreBackup = () => {
    if (!backupFile) {
      setBackupMessage(t("settings.backup.error.noFile"));
      return;
    }
    if (!backupValidation?.compatible) {
      setBackupMessage(t("settings.backup.error.notValidated"));
      return;
    }
    if (!backupRestoreConfirmed) {
      setBackupMessage(t("settings.backup.error.confirmRequired", { phrase: backupRestorePhrase }));
      return;
    }
    requestConfirmation({
      title: t("settings.backup.restoreButton"),
      body: t("settings.backup.confirmRestore"),
      confirmLabel: t("settings.backup.restoreButton"),
      danger: true,
      onConfirm: () => {
        setBackupSaving(true);
        setBackupMessage("");
        api
          .restoreBackup(backupFile)
          .then((result) => {
            window.dispatchEvent(new Event(decoderLibraryChanged));
            setBackupMessage(t("settings.backup.restored", { rows: result.restoredRows, files: result.restoredFiles }));
            invalidateMasterData();
            setBackupFile(null);
            setBackupValidation(null);
            setBackupRestoreConfirm("");
            setBackupFileInputKey((current) => current + 1);
            loadStorageUsage();
          })
          .catch((error: Error) => setBackupMessage(error.message))
          .finally(() => setBackupSaving(false));
      }
    });
  };

  const importMasterData = () => {
    if (!masterDataFile) {
      setMasterDataMessage(t("settings.masterTransfer.fileMissing"));
      return;
    }
    requestConfirmation({
      title: t("settings.masterTransfer.upload"),
      body: t("settings.masterTransfer.confirm"),
      confirmLabel: t("settings.masterTransfer.upload"),
      danger: true,
      onConfirm: () => {
        setMasterDataSaving(true);
        setMasterDataMessage("");
        api
          .importMasterData(masterDataFile)
          .then((result) => {
            window.dispatchEvent(new Event(decoderLibraryChanged));
            setMasterDataMessage(t("settings.masterTransfer.imported", { entries: result.importedEntries, relations: result.importedRelations }));
            invalidateMasterData();
            setMasterDataFile(null);
            setMasterDataFileInputKey((current) => current + 1);
            clearMasterDataSearch();
          })
          .catch((error: Error) => setMasterDataMessage(error.message))
          .finally(() => setMasterDataSaving(false));
      }
    });
  };

  return {
    backupFile, backupValidation, backupMessage, backupRestoreConfirm, setBackupRestoreConfirm,
    backupSaving, backupValidating, backupFileInputKey, backupRestorePhrase, backupRestoreConfirmed,
    selectBackupFile, restoreBackup,
    masterDataFile, setMasterDataFile, masterDataMessage, setMasterDataMessage,
    masterDataSaving, masterDataFileInputKey, importMasterData
  };
}

export type SettingsTransferController = ReturnType<typeof useSettingsTransfer>;
