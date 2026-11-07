import { api } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { VehicleValuationForm } from "./VehicleValuationForm";
import { useVehicleValuations } from "./useVehicleValuations";
import { formatDate, formatDateTime } from "./vehicleFormat";

export function VehicleValuations({ vehicleID, canEdit = false, active = true, journalOwner = "" }: {
  vehicleID: string; canEdit?: boolean; active?: boolean; journalOwner?: string;
}) {
  const { t } = useI18n();
  const model = useVehicleValuations(vehicleID, active, journalOwner);
  if (!active) return null;
  return <section className="vehicle-valuations" aria-label={t("valuation.title")}>
    <header className="valuation-heading"><h3>{t("valuation.title")}</h3>
      <a className="secondary-button" href={api.vehicleValuationsExportURL(vehicleID)} download>{t("valuation.export")}</a>
    </header>
    <p>{t("valuation.intro")}</p>
    {model.status && <p role="status">{t(model.status)}</p>}
    {model.error && <p role="alert" className="form-message">{t(model.error)}</p>}
    {model.uncertain && canEdit && <button type="button" className="secondary-button"
      disabled={model.saving} onClick={() => void model.review()}>{t("valuation.review")}</button>}
    {canEdit && <VehicleValuationForm controller={model} />}
    <div className="valuation-heading"><h4>{t("valuation.history")}</h4>
      <button type="button" className="secondary-button" disabled={model.loading || model.saving}
        onClick={() => void model.load()}>{t("valuation.reload")}</button>
    </div>
    {model.loading && <p role="status">{t("valuation.loading")}</p>}
    {model.loadFailed && <p role="alert" className="form-message">{t("valuation.loadFailed")}</p>}
    {!model.loading && !model.loadFailed && model.history.entries.length === 0 && <p>{t("valuation.empty")}</p>}
    <ol className="valuation-history">
      {model.history.entries.map(entry => <li key={entry.id}>
        <div className="valuation-heading"><strong>{entry.amount} {entry.currency}</strong>
          <span>{formatDate(entry.valuedOn)}</span>
          {entry.replacedById && <span>{t("valuation.replaced")}</span>}
          {entry.withdrawn && <span>{t("valuation.withdrawn")}</span>}
          {!entry.replacedById && !entry.withdrawn && <span>{t("valuation.current")}</span>}
        </div>
        <p className="valuation-provenance">{entry.source}</p>
        {entry.notes && <p className="valuation-provenance">{entry.notes}</p>}
        <small>{t("valuation.recordedAt", { date: formatDateTime(entry.recordedAt) })}</small>
        {canEdit && !entry.replacedById && <div className="valuation-actions">
          {model.error === "valuation.conflict" && model.draft.replacesId && <button type="button"
            className="secondary-button" disabled={model.saving || Boolean(model.uncertain)}
            onClick={() => model.rebaseCorrection(entry)}>{t("valuation.rebase")}</button>}
          <button type="button" className="secondary-button" disabled={model.saving || Boolean(model.uncertain)}
            onClick={() => model.startCorrection(entry, false)}>{t("valuation.correct")}</button>
          {!entry.withdrawn && <button type="button" className="secondary-button" disabled={model.saving || Boolean(model.uncertain)}
            onClick={() => model.startCorrection(entry, true)}>{t("valuation.withdraw")}</button>}
        </div>}
      </li>)}
    </ol>
    {model.history.entries.length < model.history.total && <button type="button" className="secondary-button"
      disabled={model.loading} onClick={() => void model.load(model.history.entries.length)}>{t("valuation.more")}</button>}
  </section>;
}
