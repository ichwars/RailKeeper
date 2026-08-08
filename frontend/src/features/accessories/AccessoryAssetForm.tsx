import type { FormEventHandler } from "react";

import type { AccessoryAssetInput, AccessoryCondition, AccessoryLifecycle, StorageLocation } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { activeStorageLocations, storageLocationPath } from "../../shared/storageLocations";
import { AppDateInput } from "../../shared/ui/AppDateInput";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppTextInput } from "../../shared/ui/AppTextInput";

const conditions: AccessoryCondition[] = ["ready", "maintenance_due", "defective", "unknown"];
const lifecycles: Array<Exclude<AccessoryLifecycle, "reserved" | "installed">> = ["stored", "maintenance", "retired"];

export function AccessoryAssetForm({ value, locations, locationId, submitLabel, onChange, onLocationChange, onSubmit }: {
  value: AccessoryAssetInput;
  locations: StorageLocation[];
  locationId: string;
  submitLabel: string;
  onChange: (value: AccessoryAssetInput) => void;
  onLocationChange: (value: string) => void;
  onSubmit: FormEventHandler<HTMLFormElement>;
}) {
  const { t } = useI18n();
  const activeLocations = activeStorageLocations(locations);
  return <form className="accessory-form article-asset-form" onSubmit={onSubmit}>
    <AppTextInput label={t("accessories.field.inventoryNumber")} value={value.inventoryNumber || ""}
      onChange={(event) => onChange({ ...value, inventoryNumber: event.target.value })} />
    <AppTextInput label={t("accessories.field.serialNumber")} value={value.serialNumber || ""}
      onChange={(event) => onChange({ ...value, serialNumber: event.target.value })} />
    <label className="app-field"><span className="app-field-label">{t("accessories.field.condition")}</span>
      <AppSelect value={value.condition || "ready"} aria-label={t("accessories.field.condition")}
        onChange={(event) => onChange({ ...value, condition: event.target.value as AccessoryCondition })}>
        {conditions.map((item) => <option key={item} value={item}>{t(`accessories.condition.${item}`)}</option>)}
      </AppSelect>
    </label>
    <label className="app-field"><span className="app-field-label">{t("accessories.field.lifecycle")}</span>
      <AppSelect value={value.lifecycle || "stored"} aria-label={t("accessories.field.lifecycle")}
        onChange={(event) => onChange({ ...value,
          lifecycle: event.target.value as Exclude<AccessoryLifecycle, "reserved" | "installed"> })}>
        {lifecycles.map((item) => <option key={item} value={item}>{t(`accessories.lifecycle.${item}`)}</option>)}
      </AppSelect>
    </label>
    <label className="app-field"><span className="app-field-label">{t("accessories.field.location")}</span>
      <AppSelect value={locationId} aria-label={t("accessories.field.location")}
        onChange={(event) => onLocationChange(event.target.value)}>
        {activeLocations.map((location) => <option key={location.id} value={location.id}>
          {storageLocationPath(location, locations)}</option>)}
      </AppSelect>
    </label>
    <label className="app-field"><span className="app-field-label">{t("accessories.field.purchaseDate")}</span>
      <AppDateInput value={value.purchaseDate || ""}
        onChange={(event) => onChange({ ...value, purchaseDate: event.target.value })} />
    </label>
    <AppNumberInput label={t("accessories.field.purchasePrice")} min="0" step="0.01"
      value={value.purchasePrice || ""} onValueChange={(purchasePrice) => onChange({ ...value, purchasePrice })} />
    <label className="app-field"><span className="app-field-label">{t("accessories.field.warrantyUntil")}</span>
      <AppDateInput value={value.warrantyUntil || ""}
        onChange={(event) => onChange({ ...value, warrantyUntil: event.target.value })} />
    </label>
    <AppTextInput label={t("accessories.field.notes")} value={value.notes || ""}
      onChange={(event) => onChange({ ...value, notes: event.target.value })} />
    <button type="submit" className="primary-button" disabled={!locationId}>{submitLabel}</button>
  </form>;
}
