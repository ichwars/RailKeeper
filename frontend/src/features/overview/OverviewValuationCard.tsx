import { BarChart3 } from "lucide-react";

import type { OverviewValuation } from "../../shared/api";
import { type Language, useI18n } from "../../shared/i18n";

type OverviewValuationCardProps = {
  valuation: OverviewValuation | null;
  loading: boolean;
  error: string;
};

export function formatOverviewMoney(value: string, language: Language): string {
  if (!/^\d+\.\d{2}$/.test(value)) return "–";
  const [whole, cents] = value.split(".");
  const groupedWhole = whole.replace(/\B(?=(\d{3})+(?!\d))/g, language === "de" ? "." : ",");
  return language === "de" ? `${groupedWhole},${cents} €` : `€${groupedWhole}.${cents}`;
}

export function OverviewValuationCard({ valuation, loading, error }: OverviewValuationCardProps) {
  const { language, t } = useI18n();
  const values = valuation ? [
    ["overview.valuation.vehicleList", valuation.vehicleListValue],
    ["overview.valuation.vehiclePurchase", valuation.vehiclePurchaseValue],
    ["overview.valuation.accessoryList", valuation.accessoryListValue],
    ["overview.valuation.accessoryPurchase", valuation.accessoryPurchaseCost]
  ] as const : [];

  return (
    <div className="overview-valuation-card" aria-labelledby="overview-valuation-title">
      <div className="overview-valuation-head">
        <span className="overview-icon"><BarChart3 size={20} aria-hidden="true" /></span>
        <div>
          <h2 id="overview-valuation-title">{t("overview.valuation.title")}</h2>
          <p>{t("overview.valuation.basis")}</p>
        </div>
      </div>
      {loading ? <p className="overview-valuation-state" role="status">
        {t("overview.valuation.loading")}
      </p> : error ? <p className="overview-valuation-state error" role="alert">{error}</p> : valuation ? <>
        <dl className="overview-valuation-grid">
          {values.map(([label, value]) => <div key={label}>
            <dt>{t(label)}</dt>
            <dd>{formatOverviewMoney(value, language)}</dd>
          </div>)}
        </dl>
        {valuation.excludedForeignCurrencyPurchases > 0 ? (
          <p className="overview-valuation-hint">{t("overview.valuation.foreignExcluded", {
            count: valuation.excludedForeignCurrencyPurchases
          })}</p>
        ) : null}
      </> : null}
    </div>
  );
}
