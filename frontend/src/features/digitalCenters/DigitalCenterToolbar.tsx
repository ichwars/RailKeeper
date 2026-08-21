import { Circle, Download, LoaderCircle, Monitor, MoreHorizontal, Play, Server, Square } from "lucide-react";

import type {
  DigitalCenterProvider,
  DigitalCenterReadSession,
  DigitalCenterSummary,
  ECoSLiveStatus
} from "./digitalCenterModel";

export function DigitalCenterToolbar({
  centers, selectedProvider, selectedCenter, liveStatus, actions, loading,
  readError, onSelectCenter, onRead, onStartLive, onStopLive, onConfigure
}: {
  centers: DigitalCenterSummary[];
  selectedProvider: DigitalCenterProvider | null;
  selectedCenter: DigitalCenterSummary | null;
  liveStatus: ECoSLiveStatus | null;
  actions: { canRead: boolean; canMonitor: boolean };
  loading: { workspace: boolean; live: boolean; read: boolean };
  readError: string;
  onSelectCenter: (provider: DigitalCenterProvider) => void;
  onRead: () => Promise<DigitalCenterReadSession>;
  onStartLive: () => Promise<ECoSLiveStatus>;
  onStopLive: () => Promise<ECoSLiveStatus>;
  onConfigure: () => void;
}) {
  const monitoring = liveStatus?.state === "running";
  const connected = Boolean(liveStatus?.connected);
  const endpoint = selectedCenter ? `${selectedCenter.host}:${selectedCenter.port}` : "Nicht eingerichtet";
  const monitorLabel = !actions.canMonitor ? "Live-Monitor nicht verfügbar" :
    monitoring ? "Live-Monitor stoppen" : "Live-Monitor starten";

  return (
    <section className="digital-center-toolbar" aria-label="Aktive Digitalzentrale">
      <label className="digital-center-select-wrap">
        <Server size={18} aria-hidden="true" />
        <span className="sr-only">Digitalzentrale wählen</span>
        <select aria-label="Digitalzentrale wählen" value={selectedProvider ?? ""}
          disabled={loading.workspace || centers.length === 0}
          onChange={(event) => {
            const center = centers.find((candidate) => candidate.provider === event.target.value);
            if (center) onSelectCenter(center.provider);
          }}>
          {centers.length === 0 && <option value="">Keine Digitalzentrale</option>}
          {centers.map((center) => <option key={center.provider} value={center.provider}>{center.name}</option>)}
        </select>
      </label>
      <span className={`digital-center-connection${connected ? " connected" : ""}`}>
        {loading.workspace ? <LoaderCircle size={15} aria-hidden="true" /> :
          <Circle size={13} fill="currentColor" aria-hidden="true" />}
        {loading.workspace ? "Wird geladen" : connected ? "Verbunden" : "Nicht verbunden"}
      </span>
      <span className="digital-center-endpoint" title={endpoint}>
        <Monitor size={17} aria-hidden="true" /><span>{endpoint}</span>
      </span>
      <div className="digital-center-toolbar-actions">
        <button type="button" className="digital-center-button" disabled={!actions.canMonitor || loading.live}
          onClick={() => void (monitoring ? onStopLive() : onStartLive()).catch(() => undefined)}>
          {!actions.canMonitor ? <Monitor size={16} aria-hidden="true" /> :
            monitoring ? <Square size={15} aria-hidden="true" /> : <Play size={16} aria-hidden="true" />}
          {monitorLabel}
        </button>
        <button type="button" className="digital-center-button" disabled={!actions.canRead || loading.read}
          onClick={() => void onRead().catch(() => undefined)}>
          <Download size={16} aria-hidden="true" />Daten lesen
        </button>
        <button type="button" className="digital-center-icon-button" aria-label="Digitalzentrale konfigurieren"
          title="Digitalzentrale konfigurieren" onClick={onConfigure}>
          <MoreHorizontal size={20} aria-hidden="true" />
        </button>
      </div>
      {readError && <p className="digital-center-toolbar-error" role="alert" aria-label="Lesefehler">
        <strong>Daten konnten nicht gelesen werden.</strong> {readError} Erneut über „Daten lesen“ versuchen.
      </p>}
    </section>
  );
}
