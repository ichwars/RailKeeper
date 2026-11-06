import { useEffect, useState } from "react";
import { useI18n } from "../../shared/i18n";
import { SettingsAppearanceTab } from "./SettingsAppearanceTab";
import { SettingsArticleSearchCard } from "./SettingsArticleSearchCard";
import { SettingsAuthTab } from "./SettingsAuthTab";
import { SettingsConfirmationDialog } from "./SettingsConfirmationDialog";
import { SettingsDataTab } from "./SettingsDataTab";
import { SettingsDigitalTab } from "./SettingsDigitalTab";
import { SettingsInventoryCard } from "./SettingsInventoryCard";
import { SettingsProfileCard } from "./SettingsProfileCard";
import { SettingsStorageCard } from "./SettingsStorageCard";
import { SettingsTransferTab } from "./SettingsTransferTab";
import { SettingsUpdatesCard } from "./SettingsUpdatesCard";
import { formatDateTime, settingsTabs, type SettingsConfirmDialog } from "./settingsModel";
import { useSettingsAuth } from "./useSettingsAuth";
import { useSettingsGeneral } from "./useSettingsGeneral";
import { useSettingsMasterData } from "./useSettingsMasterData";
import { useSettingsNavigation } from "./useSettingsNavigation";
import { useSettingsPreferences } from "./useSettingsPreferences";
import { useSettingsStorage } from "./useSettingsStorage";
import { useSettingsTransfer } from "./useSettingsTransfer";

export function SettingsView({ username }: { username: string }) {
  const { t } = useI18n();
  const navigation = useSettingsNavigation();
  const { activeSettingsTab, selectSettingsTab } = navigation;
  const [confirmDialog, setConfirmDialog] = useState<SettingsConfirmDialog | null>(null);
  // Keep controller state mounted when navigating between settings tabs.
  const storage = useSettingsStorage();
  const { storageUsage, storageLoading, loadStorageUsage } = storage;
  const preferences = useSettingsPreferences(username);
  const general = useSettingsGeneral(activeSettingsTab, preferences, loadStorageUsage);
  const auth = useSettingsAuth({ username, activeSettingsTab, requestConfirmation: setConfirmDialog });
  const { currentSession, canManageUsers, canEditMasterData } = auth;
  const masterData = useSettingsMasterData({
    navigation, canEditMasterData, requestConfirmation: setConfirmDialog
  });
  const transfer = useSettingsTransfer({
    requestConfirmation: setConfirmDialog,
    invalidateMasterData: masterData.invalidateMasterData,
    clearMasterDataSearch: masterData.clearMasterDataSearch,
    loadStorageUsage
  });
  useEffect(() => {
    if (activeSettingsTab !== "data" || storageUsage || storageLoading) return;
    loadStorageUsage();
  }, [activeSettingsTab, storageUsage, storageLoading]);
  const confirmSettingsAction = () => {
    const action = confirmDialog?.onConfirm;
    setConfirmDialog(null);
    action?.();
  };
  const activeSettingsDefinition =
    settingsTabs.find((tab) => tab.id === activeSettingsTab) || settingsTabs[0];

  return (
    <>
      <section className="settings-head">
        <h1>{t(activeSettingsDefinition.labelKey)}</h1>
        <p>{t(activeSettingsDefinition.descriptionKey)}</p>
      </section>

      <nav className="settings-primary-tabs" aria-label={t("settings.title")}>
        {settingsTabs.map((tab) => (
          <button
            key={tab.id}
            type="button"
            className={activeSettingsTab === tab.id ? "active" : ""}
            onClick={() => selectSettingsTab(tab.id)}
          >
            {t(tab.labelKey)}
          </button>
        ))}
      </nav>

      {activeSettingsTab === "general" && (
        <section className="settings-dashboard-grid">
          <div className="settings-card-stack">
            <SettingsProfileCard preferences={preferences} />
            <SettingsInventoryCard general={general} />
          </div>
          <aside className="settings-card-stack">
            <SettingsArticleSearchCard preferences={preferences} />
            <SettingsUpdatesCard preferences={preferences} general={general} />
            <SettingsStorageCard storage={storage} canManageUsers={canManageUsers} />
          </aside>
        </section>
      )}

      {activeSettingsTab === "data" && (
        <SettingsDataTab navigation={navigation} masterData={masterData}
          roles={currentSession?.roles || []} requestConfirmation={setConfirmDialog} />
      )}

      {activeSettingsTab === "importExport" && (
        <SettingsTransferTab transfer={transfer} storage={storage} />
      )}

      {activeSettingsTab === "digital" && (
        <SettingsDigitalTab canManageUsers={canManageUsers} formatDateTime={formatDateTime} username={username} />
      )}

      {activeSettingsTab === "appearance" && <SettingsAppearanceTab preferences={preferences} />}

      {activeSettingsTab === "auth" && (
        <SettingsAuthTab controller={auth} />
      )}

      {confirmDialog && (
        <SettingsConfirmationDialog confirmDialog={confirmDialog}
          onCancel={() => setConfirmDialog(null)} onConfirm={confirmSettingsAction} />
      )}

    </>
  );
}
