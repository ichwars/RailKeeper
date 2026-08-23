import { Circle, Download, LoaderCircle, Monitor, MoreHorizontal, Play, Server, Square } from "lucide-react";

import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
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
  const { t } = useI18n();
  const monitoring = liveStatus?.state === "running";
  const connected = Boolean(liveStatus?.connected);
  const endpoint = selectedCenter ? `${selectedCenter.host}:${selectedCenter.port}` :
    t("digitalCenters.common.notConfigured");
  const monitorLabel = !actions.canMonitor ? t("digitalCenters.toolbar.monitorUnavailable") :
    monitoring ? t("digitalCenters.toolbar.stopMonitor") : t("digitalCenters.toolbar.startMonitor");
  const limitations = selectedCenter ? [
    !actions.canRead ? t("digitalCenters.toolbar.readUnsupported") : "",
    !actions.canMonitor ? t("digitalCenters.toolbar.monitorUnsupported") : ""
  ].filter(Boolean) : [];

  return (
    <section className="digital-center-toolbar" aria-label={t("digitalCenters.toolbar.regionLabel")}>
      <div className="digital-center-select-wrap">
        <Server size={18} aria-hidden="true" />
        <span className="sr-only">{t("digitalCenters.toolbar.select")}</span>
        <AppSelect aria-label={t("digitalCenters.toolbar.select")} value={selectedProvider ?? ""}
          disabled={loading.workspace || centers.length === 0}
          onChange={(event) => {
            const center = centers.find((candidate) => candidate.provider === event.target.value);
            if (center) onSelectCenter(center.provider);
          }}>
          {centers.length === 0 && <option value="">{t("digitalCenters.toolbar.noStation")}</option>}
          {centers.map((center) => <option key={center.provider} value={center.provider}>{center.name}</option>)}
        </AppSelect>
      </div>
      <span className={`digital-center-connection${connected ? " connected" : ""}`}>
        {loading.workspace ? <LoaderCircle size={15} aria-hidden="true" /> :
          <Circle size={13} fill="currentColor" aria-hidden="true" />}
        {loading.workspace ? t("digitalCenters.common.loading") : connected
          ? t("digitalCenters.common.connected") : t("digitalCenters.common.disconnected")}
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
          <Download size={16} aria-hidden="true" />
          {t(loading.read ? "digitalCenters.toolbar.reading" : "digitalCenters.toolbar.read")}
        </button>
        <button type="button" className="digital-center-icon-button"
          aria-label={t("digitalCenters.toolbar.configure")} title={t("digitalCenters.toolbar.configure")}
          onClick={onConfigure}>
          <MoreHorizontal size={20} aria-hidden="true" />
        </button>
      </div>
      {readError && <p className="digital-center-toolbar-error" role="alert"
        aria-label={t("digitalCenters.error.readLabel")}>
        <strong>{t("digitalCenters.error.readTitle")}</strong> {readError}{" "}
        {t("digitalCenters.error.readRetry")}
      </p>}
      {limitations.length > 0 && <p className="digital-center-capability-note">
        {limitations.map((limitation) => <span key={limitation}>{limitation}</span>)}
      </p>}
    </section>
  );
}
