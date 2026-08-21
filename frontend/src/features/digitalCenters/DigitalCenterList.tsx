import { Circle, Plus, Server } from "lucide-react";

import { useI18n } from "../../shared/i18n";
import type { DigitalCenterProvider, DigitalCenterSummary } from "./digitalCenterModel";

export function DigitalCenterList({
  centers, selectedProvider, total, loading, error, onSelect, onConfigure, onRetry
}: {
  centers: DigitalCenterSummary[];
  selectedProvider: DigitalCenterProvider | null;
  total: number;
  loading: boolean;
  error: string;
  onSelect: (provider: DigitalCenterProvider) => void;
  onConfigure: () => void;
  onRetry: () => Promise<void>;
}) {
  const { t } = useI18n();
  return (
    <aside className="digital-centers-panel digital-center-list" aria-labelledby="digital-center-list-title">
      <header className="digital-centers-panel-head">
        <h2 id="digital-center-list-title">{t("digitalCenters.list.title")}</h2>
        <button type="button" className="digital-center-icon-button"
          aria-label={t("digitalCenters.list.add")} title={t("digitalCenters.list.add")} onClick={onConfigure}>
          <Plus size={19} aria-hidden="true" />
        </button>
      </header>
      <div className="digital-center-list-body">
        {loading && <p className="digital-centers-state">{t("digitalCenters.list.loading")}</p>}
        {!loading && error && <div className="digital-centers-state error" role="alert"
          aria-label={t("digitalCenters.error.workspaceLabel")}>
          <p>{error}</p>
          <button type="button" className="digital-center-button"
            aria-label={t("digitalCenters.error.retryStations")}
            onClick={() => void onRetry().catch(() => undefined)}>{t("digitalCenters.error.retry")}</button>
        </div>}
        {!loading && !error && centers.length === 0 &&
          <p className="digital-centers-state">{t("digitalCenters.list.empty")}</p>}
        {centers.map((center) => {
          const selected = center.provider === selectedProvider;
          return (
            <button key={center.provider} type="button"
              className={`digital-center-card${selected ? " selected" : ""}`}
              aria-pressed={selected} onClick={() => onSelect(center.provider)}>
              <Server size={22} aria-hidden="true" />
              <span className="digital-center-card-copy">
                <strong title={center.name}>{center.name}</strong>
                <small>{t("digitalCenters.common.locomotiveCount", { count: selected ? total : 0 })}</small>
              </span>
              <span className={`digital-center-card-state${center.active ? " active" : ""}`}>
                <Circle size={10} fill={center.active ? "currentColor" : "none"} aria-hidden="true" />
                {center.active ? t("digitalCenters.common.active") : t("digitalCenters.common.inactive")}
              </span>
            </button>
          );
        })}
      </div>
    </aside>
  );
}
