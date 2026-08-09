import type { AccessoryTechnicalPlacement } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppTextInput } from "../../shared/ui/AppTextInput";

export const emptyTechnicalPlacement = (): AccessoryTechnicalPlacement => ({});

export function AccessoryTechnicalFields({ value, onChange }: {
  value: AccessoryTechnicalPlacement;
  onChange: (value: AccessoryTechnicalPlacement) => void;
}) {
  const { t } = useI18n();
  const field = (key: keyof AccessoryTechnicalPlacement) => ({
    value: value[key] || "",
    onChange: (event: React.ChangeEvent<HTMLInputElement>) => onChange({ ...value, [key]: event.target.value })
  });
  return <div className="accessory-technical-fields">
    <AppTextInput label={t("accessories.field.placement")} {...field("placement")} />
    <AppTextInput label={t("accessories.field.digitalAddress")} {...field("digitalAddress")} />
    <AppTextInput label={t("accessories.field.decoderOutput")} {...field("decoderOutput")} />
    <AppTextInput label={t("accessories.field.connection")} {...field("connection")} />
    <AppTextInput label={t("accessories.field.wiringNotes")} {...field("wiringNotes")} />
  </div>;
}
