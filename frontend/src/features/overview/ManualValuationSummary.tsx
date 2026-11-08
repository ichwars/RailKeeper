import { useEffect, useId, useState } from "react";
import { ChevronDown, RefreshCw } from "lucide-react";
import { api, type VehicleValuationSummary } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppDateInput } from "../../shared/ui/AppDateInput";
import { valuationChangedEvent, valuationLocalDate } from "../vehicles/valuationModel";

export function ManualValuationSummary() {
  const { t } = useI18n();
  const dateID = useId();
  const contentID = useId();
  const [expanded, setExpanded] = useState(false);
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
    if (!expanded) return;
    let current = true;
    setSummary(null); setFailed(false);
    if (!asOf) return;
    void api.valuationSummary(asOf).then(result => {
      if (current) setSummary(result);
    }).catch(() => { if (current) setFailed(true); });
    return () => { current = false; };
  }, [asOf, expanded, revision]);
  return <article className="overview-card manual-valuation-summary" aria-label={t("valuation.summaryTitle")}>
    <header className="overview-analysis-head">
      <h2><button type="button" className="manual-valuation-toggle" aria-expanded={expanded}
        aria-controls={contentID} onClick={() => setExpanded(previous => !previous)}>
        <span>{t("valuation.summaryTitle")}</span><ChevronDown size={16} aria-hidden="true" />
      </button></h2>
    </header>
    <div id={contentID} className="manual-valuation-body" hidden={!expanded}>
    {expanded && <>
    <p>{t("valuation.summaryHelp")}</p>
    <div className="valuation-actions"><label htmlFor={dateID}>{t("valuation.asOf")}</label>
      <AppDateInput id={dateID} value={asOf} onChange={event => setAsOf(event.target.value)} />
      <button type="button" className="secondary-button" onClick={() => setRevision(previous => previous + 1)}>
        <RefreshCw size={16} aria-hidden="true" />{t("valuation.reload")}
      </button>
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
    </>}
    </div>
  </article>;
}
