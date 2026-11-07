import { useId } from "react";
import { useI18n } from "../../shared/i18n";
import type { VehicleValuationController } from "./useVehicleValuations";

export function VehicleValuationForm({ controller: model }: { controller: VehicleValuationController }) {
  const { t } = useI18n();
  const helpID = useId();
  const { draft, setDraft } = model;
  const set = (field: "amount" | "currency" | "valuedOn" | "source" | "notes", value: string) =>
    setDraft(previous => ({ ...previous, [field]: value }));
  return <fieldset className="vehicle-form valuation-form" disabled={model.saving || Boolean(model.uncertain)}>
    <legend>{t(draft.withdrawn ? "valuation.withdraw" : draft.replacesId ? "valuation.correct" : "valuation.new")}</legend>
    <p id={helpID}>{t(draft.withdrawn ? "valuation.withdrawHelp" : draft.replacesId ? "valuation.correctionHelp" : "valuation.inputHelp")}</p>
    <div className="valuation-fields">
      <label>{t("valuation.amount")} *
        <input type="text" inputMode="decimal" value={draft.amount} required maxLength={32}
          readOnly={draft.withdrawn} aria-describedby={helpID}
          onChange={event => set("amount", event.target.value)} />
      </label>
      <label>{t("valuation.currency")} *
        <input value={draft.currency} required maxLength={3} autoCapitalize="characters"
          readOnly={draft.withdrawn} onChange={event => set("currency", event.target.value)} />
      </label>
      <label>{t("valuation.valuedOn")} *
        <input type="date" value={draft.valuedOn} required readOnly={draft.withdrawn}
          onChange={event => set("valuedOn", event.target.value)} />
      </label>
      <label className="valuation-wide">{t(draft.withdrawn ? "valuation.reason" : "valuation.source")} *
        <textarea value={draft.source} required maxLength={1000} rows={2}
          onChange={event => set("source", event.target.value)} />
      </label>
      <label className="valuation-wide">{t("valuation.notes")}
        <textarea value={draft.notes} maxLength={4000} rows={2} onChange={event => set("notes", event.target.value)} />
      </label>
    </div>
    <div className="valuation-actions">
      <button type="button" className="primary-button" onClick={() => void model.save()}>
        {t(model.saving ? "valuation.saving" : draft.withdrawn ? "valuation.confirmWithdrawal" : "valuation.save")}
      </button>
      <button type="button" className="secondary-button" onClick={model.reset}>{t("valuation.discard")}</button>
    </div>
  </fieldset>;
}
