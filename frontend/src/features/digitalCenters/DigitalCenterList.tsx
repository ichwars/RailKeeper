import { Circle, Plus, Server } from "lucide-react";

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
  return (
    <aside className="digital-centers-panel digital-center-list" aria-labelledby="digital-center-list-title">
      <header className="digital-centers-panel-head">
        <h2 id="digital-center-list-title">Zentralen</h2>
        <button type="button" className="digital-center-icon-button" aria-label="Digitalzentrale hinzufügen"
          title="Digitalzentrale hinzufügen" onClick={onConfigure}>
          <Plus size={19} aria-hidden="true" />
        </button>
      </header>
      <div className="digital-center-list-body">
        {loading && <p className="digital-centers-state">Digitalzentralen werden geladen</p>}
        {!loading && error && <div className="digital-centers-state error" role="alert"
          aria-label="Arbeitsbereich konnte nicht geladen werden">
          <p>{error}</p>
          <button type="button" className="digital-center-button"
            aria-label="Digitalzentralen erneut laden"
            onClick={() => void onRetry().catch(() => undefined)}>Erneut laden</button>
        </div>}
        {!loading && !error && centers.length === 0 &&
          <p className="digital-centers-state">Keine Digitalzentrale konfiguriert</p>}
        {centers.map((center) => {
          const selected = center.provider === selectedProvider;
          return (
            <button key={center.provider} type="button"
              className={`digital-center-card${selected ? " selected" : ""}`}
              aria-pressed={selected} onClick={() => onSelect(center.provider)}>
              <Server size={22} aria-hidden="true" />
              <span className="digital-center-card-copy">
                <strong title={center.name}>{center.name}</strong><small>{selected ? total : 0} Loks</small>
              </span>
              <span className={`digital-center-card-state${center.active ? " active" : ""}`}>
                <Circle size={10} fill={center.active ? "currentColor" : "none"} aria-hidden="true" />
                {center.active ? "Aktiv" : "Inaktiv"}
              </span>
            </button>
          );
        })}
      </div>
    </aside>
  );
}
