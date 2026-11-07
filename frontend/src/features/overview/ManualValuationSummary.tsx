import { useEffect, useId, useState } from "react";
import { api, type VehicleValuationSummary } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { valuationChangedEvent, valuationLocalDate } from "../vehicles/valuationModel";

export function ManualValuationSummary() {
  const { t } = useI18n();
  const dateID = useId();
  const [asOf, setAsOf] = useState(valuationLocalDate);
  const [summary, setSummary] = useState<VehicleValuationSummary | null>(null);
  const [failed, setFailed] = useState(false);
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    const refresh = () => setRevision(previous => previous + 1);
    window.addEventListener(valuationChangedEvent, refresh);
    return () => window.removeEventListener(valuationChangedEvent, refresh);
  }, []);
  useEffect(() => {
    let current = true;
    setSummary(null); setFailed(false);
    if (!asOf) return;
    void api.valuationSummary(asOf).then(result => {
      if (current) setSummary(result);
    }).catch(() => { if (current) setFailed(true); });
    return () => { current = false; };
  }, [asOf, revision]);
  return <article className="overview-card manual-valuation-summary" aria-label={t("valuation.summaryTitle")}>
    <h2>{t("valuation.summaryTitle")}</h2>
    <p>{t("valuation.summaryHelp")}</p>
    <div className="valuation-actions"><label htmlFor={dateID}>{t("valuation.asOf")}</label>
      <input id={dateID} type="date" value={asOf} onChange={event => setAsOf(event.target.value)} />
      <button type="button" className="secondary-button" onClick={() => setRevision(previous => previous + 1)}>{t("valuation.reload")}</button>
    </div>
    {!asOf ? <p>{t("valuation.chooseDate")}</p> : failed ? <p role="alert" className="form-message">{t("valuation.summaryFailed")}</p> :
      !summary ? <p role="status">{t("valuation.loading")}</p> : <>
        <p>{t("valuation.coverage", { rated: summary.valuedVehicles, total: summary.vehicleCount, missing: summary.missingVehicles })}</p>
        {summary.totals.length === 0 ? <p>{t("valuation.noValuesAtDate")}</p> : <dl className="valuation-currency-totals">
          {summary.totals.map(total => <div key={total.currency}><dt>{total.currency}</dt>
            <dd><strong>{total.amount}</strong><span>{t("valuation.currencyVehicles", { count: total.vehicles })}</span></dd></div>)}
        </dl>}
      </>}
    <p>{t("valuation.currenciesSeparate")}</p>
  </article>;
}
