import { useCallback, useEffect, useRef, useState } from "react";
import { observeAction } from "../../shared/asyncAction";
import type { MasterDataType } from "./MasterDataSettingsSection";
import { MasterDataSettingsSection } from "./MasterDataSettingsSection";

import { api, type MasterDataEntry, type StorageLocation } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { SettingsTabList } from "./SettingsTabList";
import { StorageLocationsSettings } from "./StorageLocationsSettings";

export type ArticleDataSection = "stock_unit" | "types" | "customFields" | "locations";

export type MasterDataConfirmationRequest = {
  title: string;
  body?: string;
  confirmLabel: string;
  danger?: boolean;
  onConfirm: () => void;
};

const sections: ArticleDataSection[] = ["stock_unit", "types", "customFields", "locations"];

const typesBySection: Record<ArticleDataSection, MasterDataType[]> = {
  stock_unit: ["stock_unit"],
  types: ["article_type", "accessory_subtype"],
  customFields: ["accessory_custom_field"],
  locations: []
};

export type ArticleManagementSettingsProps = {
  roles: string[];
  activeSection: ArticleDataSection;
  onSectionChange: (section: ArticleDataSection) => void;
  onConfirmAction: (request: MasterDataConfirmationRequest) => void;
};

export function ArticleManagementSettings({
  roles,
  activeSection,
  onSectionChange,
  onConfirmAction
}: ArticleManagementSettingsProps) {
  const { t } = useI18n();
  const canRead = roles.some((role) => ["Admin", "Editor", "Viewer"].includes(role));
  const canEdit = roles.includes("Admin") || roles.includes("Editor");
  const [entriesByType, setEntriesByType] = useState<Record<string, MasterDataEntry[]>>({});
  const [loadedTypes, setLoadedTypes] = useState<Record<string, boolean>>({});
  const [loadingSection, setLoadingSection] = useState<ArticleDataSection | null>(null);
  const [masterMessage, setMasterMessage] = useState("");
  const [locations, setLocations] = useState<StorageLocation[]>([]);
  const [locationsLoaded, setLocationsLoaded] = useState(false);
  const [locationsAttempted, setLocationsAttempted] = useState(false);
  const [locationsLoading, setLocationsLoading] = useState(false);
  const [locationsMessage, setLocationsMessage] = useState("");
  const masterRequestID = useRef(0);
  const locationRequestID = useRef(0);

  const loadStorageLocations = useCallback(async () => {
    const requestID = ++locationRequestID.current;
    setLocationsAttempted(true);
    setLocationsLoading(true);
    setLocationsMessage("");
    try {
      const loaded = await api.storageLocations();
      if (requestID !== locationRequestID.current) return;
      setLocations(loaded);
      setLocationsLoaded(true);
    } catch (reason) {
      if (requestID !== locationRequestID.current) return;
      setLocationsMessage(reason instanceof Error ? reason.message : t("settings.articleManagement.error.generic"));
    } finally {
      if (requestID === locationRequestID.current) setLocationsLoading(false);
    }
  }, [t]);

  useEffect(() => {
    if (!canRead) return;
    const requestID = ++masterRequestID.current;
    const typesToLoad = typesBySection[activeSection].filter((type) => !loadedTypes[type]);
    setMasterMessage("");
    if (typesToLoad.length === 0) {
      setLoadingSection(null);
      return;
    }
    setLoadingSection(activeSection);
    Promise.all(typesToLoad.map(async (type) => [type, await api.managedMasterData(type)] as const))
      .then((results) => {
        if (requestID !== masterRequestID.current) return;
        setEntriesByType((current) => ({ ...current, ...Object.fromEntries(results) }));
        setLoadedTypes((current) => ({
          ...current,
          ...Object.fromEntries(typesToLoad.map((type) => [type, true]))
        }));
      })
      .catch((reason: Error) => {
        if (requestID === masterRequestID.current) setMasterMessage(reason.message);
      })
      .finally(() => {
        if (requestID === masterRequestID.current) setLoadingSection(null);
      });
  }, [activeSection, canRead, loadedTypes]);

  useEffect(() => {
    if (!canRead || activeSection !== "locations" || locationsLoaded || locationsAttempted) return;
    observeAction(loadStorageLocations());
  }, [activeSection, canRead, loadStorageLocations, locationsAttempted, locationsLoaded]);

  const selectSection = (section: ArticleDataSection) => {
    masterRequestID.current += 1;
    setLoadingSection(null);
    if (activeSection === "locations" && section !== "locations") {
      locationRequestID.current += 1;
      setLocationsLoading(false);
      if (!locationsLoaded) setLocationsAttempted(false);
    }
    onSectionChange(section);
  };

  const updateEntry = (type: MasterDataType, entry: MasterDataEntry) => {
    setEntriesByType((current) => ({
      ...current,
      [type]: [...(current[type] || []).filter((item) => item.key !== entry.key), entry]
        .sort((left, right) => left.sortOrder - right.sortOrder || left.label.localeCompare(right.label))
    }));
  };

  const removeEntry = (type: MasterDataType, key: string) => {
    setEntriesByType((current) => ({
      ...current,
      [type]: (current[type] || []).filter((entry) => entry.key !== key)
    }));
  };

  const reloadType = async (type: MasterDataType) => {
    setMasterMessage("");
    try {
      const entries = await api.managedMasterData(type);
      setEntriesByType((current) => ({ ...current, [type]: entries }));
      setLoadedTypes((current) => ({ ...current, [type]: true }));
    } catch (reason) {
      setMasterMessage(reason instanceof Error ? reason.message :
        t("settings.articleManagement.error.generic"));
    }
  };

  const renderMasterData = (type: MasterDataType) => (
    <MasterDataSettingsSection type={type} entries={entriesByType[type] || []}
      canEdit={canEdit} onChanged={updateEntry} onRemoved={removeEntry}
      onReload={reloadType} onConfirmAction={onConfirmAction} />
  );

  return (
    <section className="article-management-settings" aria-label={t("settings.articleManagement.sections")}>
      {!canRead ? <p className="settings-read-only-note">{t("settings.articleManagement.unavailable")}</p> : (
        <>
          {!canEdit && <p className="settings-read-only-note">{t("settings.articleManagement.readOnly")}</p>}
          <SettingsTabList
            ariaLabel={t("settings.articleManagement.sections")}
            options={sections.map((section) => ({
              id: section,
              label: t(`settings.articleManagement.section.${section === "stock_unit" ? "units" : section}`)
            }))}
            value={activeSection}
            onChange={selectSection}
            className="article-management-tabs"
          />
          <div id="article-management-active-panel" role="tabpanel"
            aria-label={t(`settings.articleManagement.section.${activeSection === "stock_unit" ? "units" : activeSection}`)}
            className="article-management-panel">
            {masterMessage && <p className="form-message" role="alert">{masterMessage}</p>}
            {activeSection === "locations" && locationsMessage && <div className="form-message" role="alert">
              <span>{locationsMessage}</span>{" "}
              <button type="button" className="secondary-button compact-action"
                onClick={() => observeAction(loadStorageLocations())}>
                {t("settings.articleManagement.locations.retry")}
              </button>
            </div>}
            {loadingSection === activeSection ?
              <p className="loading-cell">{t("settings.articleManagement.loading")}</p> : (
              <div className="article-management-section">
                {activeSection === "stock_unit" && renderMasterData("stock_unit")}
                {activeSection === "types" && <>
                  {renderMasterData("article_type")}
                  {renderMasterData("accessory_subtype")}
                </>}
                {activeSection === "customFields" && renderMasterData("accessory_custom_field")}
                {activeSection === "locations" && (locationsLoading || (!locationsLoaded && !locationsAttempted) ?
                  <p className="loading-cell">{t("settings.articleManagement.locations.loading")}</p> :
                  !locationsMessage && <StorageLocationsSettings locations={locations} canEdit={canEdit}
                    onChanged={loadStorageLocations} />)}
              </div>
            )}
          </div>
        </>
      )}
    </section>
  );
}
