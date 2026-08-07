import { useState } from "react";
import type { FormEvent } from "react";
import { MapPin } from "lucide-react";

import { api, type StorageLocation } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { accessoryLocationPath, activeAccessoryLocations } from "./accessoryLocations";

export function AccessoryLocationsPanel({ locations, canEdit, onChanged }: {
  locations: StorageLocation[];
  canEdit: boolean;
  onChanged: () => Promise<void>;
}) {
  const [name, setName] = useState("");
  const [parentID, setParentID] = useState("");
  const [description, setDescription] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const { t } = useI18n();
  const activeLocations = activeAccessoryLocations(locations);

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    setBusy(true); setMessage("");
    try {
      await api.createStorageLocation({
        name,
        parentId: parentID || undefined,
        description: description || undefined
      });
      setName(""); setParentID(""); setDescription("");
      await onChanged();
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("accessories.error.generic"));
    } finally {
      setBusy(false);
    }
  };

  return <div className="accessory-work-grid">
    <section className="panel accessory-list-panel">
      <div className="panel-head"><MapPin size={17} aria-hidden="true" />
        <h2>{t("accessories.locations.title")}</h2>
      </div>
      <div className="table-wrap"><table><thead><tr><th>{t("accessories.field.location")}</th>
        <th>{t("accessories.field.status")}</th></tr></thead><tbody>
        {locations.map((location) => <tr key={location.id}>
          <td>{accessoryLocationPath(location, locations)}</td>
          <td>{t(location.archived ? "accessories.locations.archived" : "accessories.locations.active")}</td>
        </tr>)}
      </tbody></table></div>
    </section>
    {canEdit ? <section className="panel accessory-form-panel">
      <div className="panel-head"><MapPin size={17} aria-hidden="true" />
        <h2>{t("accessories.locations.create")}</h2>
      </div>
      <form className="accessory-form" onSubmit={submit}>
        <label>{t("accessories.field.name")}<input required value={name}
          onChange={(event) => setName(event.target.value)} /></label>
        <label>{t("accessories.field.parentLocation")}<AppSelect value={parentID}
          onChange={(event) => setParentID(event.target.value)}>
          <option value="">{t("accessories.locations.noParent")}</option>
          {activeLocations.map((location) => <option key={location.id} value={location.id}>
            {accessoryLocationPath(location, locations)}</option>)}
        </AppSelect></label>
        <label>{t("accessories.field.description")}<textarea value={description}
          onChange={(event) => setDescription(event.target.value)} /></label>
        {message ? <p className="form-message">{message}</p> : null}
        <button type="submit" className="primary-button" disabled={busy}>{t("accessories.locations.save")}</button>
      </form>
    </section> : null}
  </div>;
}
