import { useMemo, useState } from "react";
import { api, type StorageUsage } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { formatBytes } from "./settingsModel";

export function useSettingsStorage() {
  const { t } = useI18n();
  const [storageUsage, setStorageUsage] = useState<StorageUsage | null>(null);
  const [storageMessage, setStorageMessage] = useState("");
  const [storageLoading, setStorageLoading] = useState(false);
  const [storageOptimizing, setStorageOptimizing] = useState(false);
  const storageFileCount = useMemo(
    () => (storageUsage?.categories || []).reduce((total, category) => total + category.files, 0),
    [storageUsage]
  );

  const loadStorageUsage = () => {
    setStorageLoading(true);
    setStorageMessage("");
    api
      .storageUsage()
      .then(setStorageUsage)
      .catch((error: Error) => setStorageMessage(error.message))
      .finally(() => setStorageLoading(false));
  };

  const optimizeStorage = () => {
    setStorageOptimizing(true);
    setStorageMessage("");
    api
      .optimizeStorage()
      .then((result) => {
        setStorageMessage(t("settings.storage.optimizeDone", { bytes: formatBytes(result.reclaimedBytes) }));
        loadStorageUsage();
      })
      .catch((error: Error) => setStorageMessage(error.message))
      .finally(() => setStorageOptimizing(false));
  };

  return {
    storageUsage, storageMessage, storageLoading, storageOptimizing, storageFileCount,
    loadStorageUsage, optimizeStorage
  };
}

export type SettingsStorageController = ReturnType<typeof useSettingsStorage>;
