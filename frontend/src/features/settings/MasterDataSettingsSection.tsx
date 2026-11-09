import { Plus, X } from "lucide-react";
import { FormEvent, useState } from "react";
import { api, type MasterDataEntry, type MasterDataInput } from "../../shared/api";
import { masterDataDisplayLabel, masterDataPersistedLabel } from "../../shared/articleMasterDataLabels";
import { observeAction } from "../../shared/asyncAction";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppTextInput } from "../../shared/ui/AppTextInput";
import type { MasterDataConfirmationRequest } from "./ArticleManagementSettings";
import {
  MasterDataBulkEntryCheckbox,
  MasterDataBulkSelectAllCheckbox,
  MasterDataBulkToolbar,
  useMasterDataBulkSelection
} from "./MasterDataBulkSelection";
import { MasterDataLifecycleActions } from "./MasterDataLifecycleActions";

export type MasterDataType = "stock_unit" | "article_type" | "accessory_subtype" |
  "accessory_custom_field";

export type CustomFieldKind = "text" | "number" | "boolean" | "date" | "single_select" | "multi_select";

export function entryInput(entry: MasterDataEntry, label: string, active = entry.active): MasterDataInput {
  return {
    label,
    active,
    sortOrder: entry.sortOrder,
    sourceUrl: entry.sourceUrl,
    metadata: entry.metadata
  };
}

export function MasterDataSettingsSection({
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
  const bulkSelection = useMasterDataBulkSelection(type, entries, entries);

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
      onConfirm: () => observeAction(setActive(entry, false))
    });
  };

  const requestBulkDeactivate = () => {
    const keys = [...bulkSelection.selectedKeys];
    const count = keys.length;
    if (count === 0) return;
    onConfirmAction({
      title: t("settings.master.deactivateManyTitle"),
      body: t("settings.master.deactivateManyBody", {
        count,
        type: t(`settings.articleManagement.master.${type}`)
      }),
      confirmLabel: t(
        count === 1 ? "settings.master.deactivateOneConfirm" : "settings.master.deactivateManyConfirm",
        { count }
      ),
      onConfirm: () => {
        setBusy(true);
        setMessage("");
        api.setMasterDataActiveMany(type, keys, false)
          .then((updated) => {
            for (const entry of updated) onChanged(type, entry);
            const updatedEditing = editing && updated.find((entry) => entry.key === editing.key);
            if (updatedEditing) setEditing(updatedEditing);
            bulkSelection.clear();
          })
          .catch((reason: Error) => setMessage(reason.message))
          .finally(() => setBusy(false));
      }
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
      {canEdit && <MasterDataBulkToolbar
        count={bulkSelection.selectedCount}
        busy={busy}
        onDeactivate={requestBulkDeactivate}
      />}
      <div className="table-wrap article-master-data-table">
        <table>
          <thead>
            <tr>
              {canEdit && <th className="master-data-selection-col">
                <MasterDataBulkSelectAllCheckbox selection={bulkSelection} disabled={busy} />
              </th>}
              <th>{t("settings.articleManagement.label")}</th>
              <th>{t("settings.articleManagement.key")}</th>
              <th>{t("settings.articleManagement.status")}</th>
              <th>{t("settings.master.origin")}</th>
              {canEdit && <th className="is-right">{t("settings.articleManagement.actions")}</th>}
            </tr>
          </thead>
          <tbody>
            {entries.length === 0 ? (
              <tr><td colSpan={canEdit ? 6 : 4} className="loading-cell">
                {t("settings.articleManagement.empty")}
              </td></tr>
            ) : entries.map((entry) => {
              const displayLabel = masterDataDisplayLabel(entry, t);
              return <tr key={entry.id} className={entry.active ? "" : "muted-row"}>
                {canEdit && <td className="master-data-selection-col">
                  <MasterDataBulkEntryCheckbox
                    selection={bulkSelection}
                    entry={entry}
                    label={displayLabel}
                    disabled={busy}
                  />
                </td>}
                <td><strong>{displayLabel}</strong></td>
                <td><code>{entry.key}</code></td>
                <td>{t(entry.active ? "settings.articleManagement.active" : "settings.articleManagement.inactive")}</td>
                <td>{t(`settings.master.origin.${entry.origin || "custom"}`)}</td>
                {canEdit && (
                  <td className="is-right">
                    <MasterDataLifecycleActions entry={entry} displayLabel={displayLabel} disabled={busy}
                      onEdit={() => startEditing(entry)}
                      onDeactivate={() => requestDeactivate(entry)}
                      onReactivate={() => observeAction(setActive(entry, true))}
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
