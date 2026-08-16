import { Plus, X } from "lucide-react";
import { FormEvent, useCallback, useEffect, useRef, useState } from "react";

import { api, type MasterDataEntry, type MasterDataInput, type StorageLocation } from "../../shared/api";
import { masterDataDisplayLabel, masterDataPersistedLabel } from "../../shared/articleMasterDataLabels";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppTextInput } from "../../shared/ui/AppTextInput";
import { MasterDataLifecycleActions } from "./MasterDataLifecycleActions";
import { SettingsTabList } from "./SettingsTabList";
import { StorageLocationsSettings } from "./StorageLocationsSettings";

export type ArticleDataSection = "stock_unit" | "types" | "customFields" | "locations";
type MasterDataType = "stock_unit" | "article_type" | "accessory_subtype" |
  "accessory_custom_field";
type CustomFieldKind = "text" | "number" | "boolean" | "date" | "single_select" | "multi_select";

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

function entryInput(entry: MasterDataEntry, label: string, active = entry.active): MasterDataInput {
  return {
    label,
    active,
    sortOrder: entry.sortOrder,
    sourceUrl: entry.sourceUrl,
    metadata: entry.metadata
  };
}

function MasterDataSettingsSection({
  type,
  entries,
  canEdit,
  onChanged,
  onRemoved,
  onReload,
  onConfirmAction
}: {
  type: MasterDataType;
  entries: MasterDataEntry[];
  canEdit: boolean;
  onChanged: (type: MasterDataType, entry: MasterDataEntry) => void;
  onRemoved: (type: MasterDataType, key: string) => void;
  onReload: (type: MasterDataType) => Promise<void>;
  onConfirmAction: (request: MasterDataConfirmationRequest) => void;
}) {
  const { t } = useI18n();
  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<MasterDataEntry | null>(null);
  const [key, setKey] = useState("");
  const [label, setLabel] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const canCreate = canEdit && type !== "article_type";
  const [fieldKind, setFieldKind] = useState<CustomFieldKind>("text");
  const [fieldOptions, setFieldOptions] = useState("");
  const [fieldUnit, setFieldUnit] = useState("");

  const reset = () => {
    setCreating(false);
    setEditing(null);
    setKey("");
    setLabel("");
    setMessage("");
    setFieldKind("text");
    setFieldOptions("");
    setFieldUnit("");
  };

  const startEditing = (entry: MasterDataEntry) => {
    setCreating(false);
    setEditing(entry);
    setKey(entry.key);
    setLabel(masterDataDisplayLabel(entry, t));
    setMessage("");
    const kind = entry.metadata.kind;
    setFieldKind(typeof kind === "string" ? kind as CustomFieldKind : "text");
    setFieldOptions(Array.isArray(entry.metadata.options) ? entry.metadata.options.join(", ") : "");
    setFieldUnit(typeof entry.metadata.unit === "string" ? entry.metadata.unit : "");
  };

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!canEdit || (!editing && !canCreate)) return;
    setBusy(true);
    setMessage("");
    const options = fieldOptions.split(",").map((option) => option.trim()).filter(Boolean);
    const customFieldMetadata: Record<string, unknown> = {
      kind: fieldKind,
      ...(["single_select", "multi_select"].includes(fieldKind) ? { options } : {}),
      ...(fieldKind === "number" && fieldUnit.trim() ? { unit: fieldUnit.trim() } : {})
    };
    try {
      const result = editing
        ? await api.updateMasterData(type, editing.key, {
            ...entryInput(editing, masterDataPersistedLabel(editing, label, t)),
            metadata: type === "accessory_custom_field" ? customFieldMetadata : editing.metadata
          })
        : await api.createMasterData(type, {
            key: key.trim(),
            label: label.trim(),
            active: true,
            ...(type === "accessory_custom_field" ? { metadata: customFieldMetadata } : {})
          });
      onChanged(type, result);
      await onReload(type);
      reset();
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("settings.articleManagement.error.generic"));
    } finally {
      setBusy(false);
    }
  };

  const setActive = async (entry: MasterDataEntry, active: boolean) => {
    if (!canEdit) return;
    setBusy(true);
    setMessage("");
    try {
      const updated = await api.setMasterDataActive(type, entry.key, active);
      onChanged(type, updated);
      if (editing?.key === updated.key) setEditing(updated);
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("settings.articleManagement.error.generic"));
    } finally {
      setBusy(false);
    }
  };

  const requestDeactivate = (entry: MasterDataEntry) => {
    onConfirmAction({
      title: t("settings.master.deactivateTitle"),
      body: t("settings.master.deactivateBody"),
      confirmLabel: t("settings.master.deactivate"),
      onConfirm: () => void setActive(entry, false)
    });
  };

  const requestDelete = (entry: MasterDataEntry) => {
    onConfirmAction({
      title: t("settings.master.deleteTitle"),
      body: t("settings.master.deleteBody"),
      confirmLabel: t("settings.master.deletePermanently"),
      danger: true,
      onConfirm: () => {
        setBusy(true);
        setMessage("");
        api.deleteMasterData(type, entry.key)
          .then(() => {
            if (editing?.key === entry.key) reset();
            onRemoved(type, entry.key);
          })
          .catch((reason: Error) => setMessage(reason.message))
          .finally(() => setBusy(false));
      }
    });
  };

  return (
    <section className="article-master-data-panel" aria-labelledby={`article-master-${type}`}>
      <div className="settings-section-head">
        <div>
          <h3 id={`article-master-${type}`}>{t(`settings.articleManagement.master.${type}`)}</h3>
          <p>{t(`settings.articleManagement.master.${type}Help`)}</p>
        </div>
        {canCreate && !creating && !editing && (
          <button type="button" className="secondary-button compact-action" onClick={() => setCreating(true)}>
            <Plus size={15} aria-hidden="true" /> {t("settings.articleManagement.add")}
          </button>
        )}
      </div>

      {(creating || editing) && (
        <form className="article-master-data-form" onSubmit={submit}>
          {editing ? (
            <div className="article-master-data-key">
              <span>{t("settings.articleManagement.key")}</span>
              <code>{editing.key}</code>
            </div>
          ) : (
            <AppTextInput
              label={t("settings.articleManagement.key")}
              value={key}
              onChange={(event) => setKey(event.target.value)}
              required
            />
          )}
          <AppTextInput
            label={t("settings.articleManagement.label")}
            value={label}
            onChange={(event) => setLabel(event.target.value)}
            required
          />
          {type === "accessory_custom_field" && <>
            <label className="article-master-data-select">
              {t("settings.articleManagement.customField.kind")}
              <AppSelect value={fieldKind} aria-label={t("settings.articleManagement.customField.kind")}
                onChange={(event) => setFieldKind(event.target.value as CustomFieldKind)}>
                {(["text", "number", "boolean", "date", "single_select", "multi_select"] as const)
                  .map((kind) => <option key={kind} value={kind}>
                    {t(`settings.articleManagement.customField.kind.${kind}`)}
                  </option>)}
              </AppSelect>
            </label>
            {fieldKind === "number" && <AppTextInput
              label={t("settings.articleManagement.customField.unit")}
              value={fieldUnit}
              onChange={(event) => setFieldUnit(event.target.value)}
            />}
            {["single_select", "multi_select"].includes(fieldKind) && <AppTextInput
              label={t("settings.articleManagement.customField.options")}
              helpText={t("settings.articleManagement.customField.optionsHelp")}
              value={fieldOptions}
              onChange={(event) => setFieldOptions(event.target.value)}
              required
            />}
          </>}
          <div className="article-master-data-form-actions">
            <button type="submit" className="primary-button" disabled={busy}>
              {editing ? t("settings.articleManagement.saveChanges") : t("settings.articleManagement.create")}
            </button>
            <button type="button" className="icon-button" onClick={reset}
              aria-label={t("common.cancel")} title={t("common.cancel")}>
              <X size={16} aria-hidden="true" />
            </button>
          </div>
        </form>
      )}

      {message && <p className="form-message" role="alert">{message}</p>}
      <div className="table-wrap article-master-data-table">
        <table>
          <thead>
            <tr>
              <th>{t("settings.articleManagement.label")}</th>
              <th>{t("settings.articleManagement.key")}</th>
              <th>{t("settings.articleManagement.status")}</th>
              <th>{t("settings.master.origin")}</th>
              {canEdit && <th className="is-right">{t("settings.articleManagement.actions")}</th>}
            </tr>
          </thead>
          <tbody>
            {entries.length === 0 ? (
              <tr><td colSpan={canEdit ? 5 : 4} className="loading-cell">
                {t("settings.articleManagement.empty")}
              </td></tr>
            ) : entries.map((entry) => {
              const displayLabel = masterDataDisplayLabel(entry, t);
              return <tr key={entry.id} className={entry.active ? "" : "muted-row"}>
                <td><strong>{displayLabel}</strong></td>
                <td><code>{entry.key}</code></td>
                <td>{t(entry.active ? "settings.articleManagement.active" : "settings.articleManagement.inactive")}</td>
                <td>{t(`settings.master.origin.${entry.origin || "custom"}`)}</td>
                {canEdit && (
                  <td className="is-right">
                    <MasterDataLifecycleActions entry={entry} displayLabel={displayLabel} disabled={busy}
                      onEdit={() => startEditing(entry)}
                      onDeactivate={() => requestDeactivate(entry)}
                      onReactivate={() => void setActive(entry, true)}
                      onDelete={() => requestDelete(entry)} />
                  </td>
                )}
              </tr>;
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
}

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
  const canRead = roles.some((role) => ["Admin", "Editor", "Planner", "Viewer"].includes(role));
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
    void loadStorageLocations();
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
                onClick={() => void loadStorageLocations()}>
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
