import { Archive, ArchiveRestore, MapPin, Pencil, X } from "lucide-react";
import { FormEvent, useState } from "react";

import { api, type StorageLocation, type StorageLocationInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppTextInput } from "../../shared/ui/AppTextInput";
import { availableStorageLocationParents, storageLocationPath } from "../../shared/storageLocations";

function locationInput(location: StorageLocation, archived = location.archived): StorageLocationInput {
  return {
    name: location.name,
    parentId: location.parentId,
    description: location.description,
    archived
  };
}

export function StorageLocationsSettings({ locations, canEdit, onChanged }: {
  locations: StorageLocation[];
  canEdit: boolean;
  onChanged: () => Promise<void>;
}) {
  const [editing, setEditing] = useState<StorageLocation | null>(null);
  const [name, setName] = useState("");
  const [parentID, setParentID] = useState("");
  const [description, setDescription] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const { t } = useI18n();
  const parentOptions = availableStorageLocationParents(locations, editing?.id);

  const resetForm = () => {
    setEditing(null);
    setName("");
    setParentID("");
    setDescription("");
    setMessage("");
  };

  const startEditing = (location: StorageLocation) => {
    setEditing(location);
    setName(location.name);
    setParentID(location.parentId || "");
    setDescription(location.description || "");
    setMessage("");
  };

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!canEdit) return;
    setBusy(true);
    setMessage("");
    const input: StorageLocationInput = {
      name: name.trim(),
      parentId: parentID || undefined,
      description: description.trim() || undefined,
      ...(editing ? { archived: editing.archived } : {})
    };
    try {
      if (editing) await api.updateStorageLocation(editing.id, input);
      else await api.createStorageLocation(input);
      resetForm();
      await onChanged();
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("accessories.error.generic"));
    } finally {
      setBusy(false);
    }
  };

  const setArchived = async (location: StorageLocation, archived: boolean) => {
    if (!canEdit) return;
    setBusy(true);
    setMessage("");
    try {
      await api.updateStorageLocation(location.id, locationInput(location, archived));
      if (editing?.id === location.id) resetForm();
      await onChanged();
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("accessories.error.generic"));
    } finally {
      setBusy(false);
    }
  };

  return <div className="storage-location-settings-grid">
    <section className="panel storage-location-list-panel">
      <div className="panel-head"><MapPin size={17} aria-hidden="true" />
        <h3>{t("settings.articleManagement.locations.title")}</h3>
      </div>
      {message ? <p className="form-message" role="alert">{message}</p> : null}
      <div className="table-wrap"><table><thead><tr><th>{t("accessories.field.location")}</th>
        <th>{t("accessories.field.status")}</th>{canEdit && <th className="is-right">
          {t("settings.articleManagement.actions")}</th>}</tr></thead><tbody>
        {locations.map((location) => <tr key={location.id} className={location.archived ? "muted-row" : ""}>
          <td>{storageLocationPath(location, locations)}</td>
          <td>{t(location.archived ? "settings.articleManagement.locations.archived" :
            "settings.articleManagement.locations.active")}</td>
          {canEdit && <td className="is-right"><div className="settings-card-actions">
            <button type="button" className="icon-button" onClick={() => startEditing(location)}
              aria-label={t("settings.articleManagement.edit", { label: location.name })}
              title={t("settings.articleManagement.editShort")}>
              <Pencil size={15} aria-hidden="true" />
            </button>
            <button type="button" className="icon-button" disabled={busy}
              onClick={() => void setArchived(location, !location.archived)}
              aria-label={t(location.archived ? "settings.articleManagement.reactivate" :
                "settings.articleManagement.archive", { label: location.name })}
              title={t(location.archived ? "settings.articleManagement.reactivateShort" :
                "settings.articleManagement.archiveShort")}>
              {location.archived ? <ArchiveRestore size={15} aria-hidden="true" /> :
                <Archive size={15} aria-hidden="true" />}
            </button>
          </div></td>}
        </tr>)}
      </tbody></table></div>
    </section>
    {canEdit ? <section className="panel storage-location-form-panel">
      <div className="panel-head"><MapPin size={17} aria-hidden="true" />
        <h3>{t(editing ? "settings.articleManagement.locations.edit" :
          "settings.articleManagement.locations.create")}</h3>
      </div>
      <form className="storage-location-form" onSubmit={submit}>
        <AppTextInput label={t("accessories.field.name")} required value={name}
          onChange={(event) => setName(event.target.value)} />
        <label>{t("accessories.field.parentLocation")}<AppSelect value={parentID}
          aria-label={t("accessories.field.parentLocation")}
          onChange={(event) => setParentID(event.target.value)}>
          <option value="">{t("settings.articleManagement.locations.noParent")}</option>
          {parentOptions.map((location) => <option key={location.id} value={location.id}>
            {storageLocationPath(location, locations)}</option>)}
        </AppSelect></label>
        <label>{t("accessories.field.description")}<textarea value={description}
          onChange={(event) => setDescription(event.target.value)} /></label>
        <div className="storage-location-form-actions">
          <button type="submit" className="primary-button" disabled={busy}>
            {t(editing ? "settings.articleManagement.saveChanges" :
              "settings.articleManagement.locations.save")}
          </button>
          {editing && <button type="button" className="icon-button" onClick={resetForm}
            aria-label={t("common.cancel")} title={t("common.cancel")}>
            <X size={16} aria-hidden="true" />
          </button>}
        </div>
      </form>
    </section> : null}
  </div>;
}
