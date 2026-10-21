import type { VehicleFunctionInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";

export function FunctionFlagsFields({ functionKey, edit, disabled, update }: {
  functionKey: string; edit: VehicleFunctionInput; disabled: boolean;
  update: (patch: Partial<VehicleFunctionInput>) => void;
}) {
  const { t } = useI18n();
  return <div className="function-flags">
    {(["directionDependent", "inverted"] as const).map(field => <label key={field}>
      {t(`vehicles.functions.${field}`)}
      <AppSelect value={edit[field] == null ? "unknown" : String(edit[field])} disabled={disabled}
        aria-label={`${functionKey} ${t(`vehicles.functions.${field}`)}`}
        onChange={event => update({ [field]: event.target.value === "unknown" ? null : event.target.value === "true", flagsVersion: 2 })}>
        <option value="unknown">{t("vehicles.decoder.unknown")}</option>
        <option value="true">{t("common.yes")}</option>
        <option value="false">{t("common.no")}</option>
      </AppSelect>
    </label>)}
    {edit.legacyDirectionDependent != null && <small className="function-legacy-flag">
      {t("vehicles.functions.legacyFlag", { value: t(edit.legacyDirectionDependent ? "common.yes" : "common.no") })}
    </small>}
  </div>;
}
