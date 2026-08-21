import { Circle, ExternalLink, LockKeyhole, RadioTower } from "lucide-react";

import { useI18n, type Language } from "../../shared/i18n";
import type {
  DigitalCenterSessionMessage,
  DigitalCenterSummary,
  DigitalCenterWorkspaceTab,
  ECoSLiveStatus
} from "./digitalCenterModel";
import { LivePulseChart } from "./LivePulseChart";

const tabs: DigitalCenterWorkspaceTab[] = ["live", "diagnosis", "messages"];

type WorkspaceActions = {
  canTestConnection: boolean;
  canRead: boolean;
  canMonitor: boolean;
  canWrite: boolean;
  canWriteCVs: boolean;
  canDiagnose: boolean;
};

export function DigitalStatusPanel({
  tab, onTab, selectedCenter, liveStatus, messages, actions, loading, errors
}: {
  tab: DigitalCenterWorkspaceTab;
  onTab: (tab: DigitalCenterWorkspaceTab) => void;
  selectedCenter: DigitalCenterSummary | null;
  liveStatus: ECoSLiveStatus | null;
  messages: DigitalCenterSessionMessage[];
  actions: WorkspaceActions;
  loading: boolean;
  errors: { live: string; messages: string };
}) {
  const { t, language } = useI18n();
  return (
    <section className="digital-centers-panel digital-centers-status">
      <div className="digital-status-tabs" role="tablist" aria-label={t("digitalCenters.status.tablist")}>
        {tabs.map((item) => <button key={item} type="button" role="tab"
          aria-selected={tab === item} onClick={() => onTab(item)}>
          {t(`digitalCenters.status.${item}`)}
        </button>)}
      </div>
      <div className="digital-status-content">
        {tab === "live" && <LiveStatus status={liveStatus} canMonitor={actions.canMonitor}
          loading={loading} error={errors.live} t={t} language={language} />}
        {tab === "diagnosis" && <Diagnosis status={liveStatus} center={selectedCenter}
          actions={actions} messages={messages} t={t} language={language} />}
        {tab === "messages" && <Messages messages={messages} error={errors.messages}
          station={selectedCenter?.name ?? t("digitalCenters.messages.stationFallback")}
          t={t} language={language} />}
      </div>
      <div className="digital-write-lock">
        <LockKeyhole size={20} aria-hidden="true" />
        <span><strong>{t("digitalCenters.write.locked")}</strong><small>{actions.canWrite
          ? t("digitalCenters.write.lockedHelp")
          : t("digitalCenters.write.unsupported")}</small></span>
        <button type="button" className="digital-center-button" disabled={!actions.canDiagnose}
          onClick={() => onTab("diagnosis")}>{t("digitalCenters.status.openDiagnosis")}{" "}
          <ExternalLink size={14} aria-hidden="true" /></button>
      </div>
    </section>
  );
}

function LiveStatus({ status, canMonitor, loading, error, t, language }: {
  status: ECoSLiveStatus | null; canMonitor: boolean; loading: boolean; error: string;
  t: Translate; language: Language;
}) {
  if (!canMonitor) return <p className="digital-centers-state">{t("digitalCenters.toolbar.monitorUnavailable")}</p>;
  if (loading && !status) return <p className="digital-centers-state">{t("digitalCenters.status.loading")}</p>;
  if (error && !status) return <p className="digital-centers-state error">{error}</p>;
  if (status?.state === "interrupted") {
    return <p className="digital-centers-state error">{t("digitalCenters.status.interrupted")}</p>;
  }
  if (!status?.connected) return <p className="digital-centers-state">{t("digitalCenters.common.disconnected")}</p>;
  return <>
    <div className="digital-live-metrics">
      <Metric label={t("digitalCenters.status.blocks")} value={status.blocksReceived} />
      <Metric label={t("digitalCenters.status.replies")} value={status.repliesReceived} active />
      <Metric label={t("digitalCenters.status.events")} value={status.eventsReceived} />
      <Metric label={t("digitalCenters.status.lastMessage")} value={formatTime(status.lastSeenAt, language)} />
    </div>
    <LivePulseChart samples={status.pulseSamples} />
    <div className="digital-live-events">
      <header><strong>{t("digitalCenters.status.recentEvents")}</strong>
        <button type="button">{t("digitalCenters.status.showAll")}</button></header>
      <div>{status.recentEvents.length === 0 &&
        <p className="digital-centers-state">{t("digitalCenters.status.noEvents")}</p>}
        {status.recentEvents.map((event, index) => <p key={`${event.at}-${index}`} title={event.message}>
          <time>{formatTime(event.at, language)}</time><Circle size={8} fill="currentColor" aria-hidden="true" />
          <span>{event.message}</span><small>{event.protocol}</small></p>)}</div>
    </div>
  </>;
}

function Metric({ label, value, active = false }: { label: string; value: string | number; active?: boolean }) {
  return <span><small>{label}</small><strong>{value}{active && <Circle size={8} fill="currentColor" aria-hidden="true" />}</strong></span>;
}

function Diagnosis({ status, center, actions, messages, t, language }: {
  status: ECoSLiveStatus | null;
  center: DigitalCenterSummary | null;
  actions: WorkspaceActions;
  messages: DigitalCenterSessionMessage[];
  t: Translate;
  language: Language;
}) {
  const endpointHost = status?.host ?? center?.host;
  const endpointPort = status?.port ?? center?.port;
  const endpoint = endpointHost && endpointPort ? `${endpointHost}:${endpointPort}` :
    t("digitalCenters.common.notConfigured");
  const protocolErrors = messages.filter((message) =>
    message.severity === "error" || message.code === "parse.failed"
  ).length;
  const capabilities = [
    [t("digitalCenters.diagnosis.capabilityTest"), actions.canTestConnection],
    [t("digitalCenters.diagnosis.capabilityRead"), actions.canRead],
    [t("digitalCenters.diagnosis.capabilityLive"), actions.canMonitor],
    [t("digitalCenters.diagnosis.capabilityWrite"), actions.canWrite],
    [t("digitalCenters.diagnosis.capabilityCV"), actions.canWriteCVs],
    [t("digitalCenters.status.diagnosis"), actions.canDiagnose]
  ] as const;
  return <div className="digital-diagnosis">
    <RadioTower size={22} aria-hidden="true" />
    <h3>{t("digitalCenters.diagnosis.title")}</h3>
    <dl><div><dt>{t("digitalCenters.diagnosis.endpoint")}</dt><dd>{endpoint}</dd></div>
      <div><dt>{t("digitalCenters.diagnosis.state")}</dt>
        <dd>{connectionLabel(status?.diagnosis.connectionState, t)}</dd></div>
      <div><dt>{t("digitalCenters.diagnosis.latency")}</dt>
        <dd>{t("digitalCenters.diagnosis.notMeasured")}</dd></div>
      <div><dt>{t("digitalCenters.diagnosis.protocolErrors")}</dt><dd>{protocolErrors}</dd></div>
      <div><dt>{t("digitalCenters.diagnosis.lastSuccess")}</dt>
        <dd>{formatTime(status?.diagnosis.lastSuccessfulCommunication, language)}</dd></div>
      <div><dt>{t("digitalCenters.diagnosis.mode")}</dt>
        <dd>{status?.diagnosis.passive ? t("digitalCenters.diagnosis.passive")
          : t("digitalCenters.common.inactive")}</dd></div></dl>
    <h4>{t("digitalCenters.diagnosis.capabilities")}</h4>
    <ul>{capabilities.map(([label, supported]) => <li key={label}>
      {label}: {supported ? t("digitalCenters.diagnosis.supported")
        : t("digitalCenters.diagnosis.unsupported")}
    </li>)}</ul>
    {status?.diagnosis.lastError && <p className="digital-centers-state error">{status.diagnosis.lastError}</p>}
  </div>;
}

function Messages({ messages, error, station, t, language }: {
  messages: DigitalCenterSessionMessage[];
  error: string;
  station: string;
  t: Translate;
  language: Language;
}) {
  if (error) return <p className="digital-centers-state error">{error}</p>;
  if (messages.length === 0) return <p className="digital-centers-state">
    {t("digitalCenters.messages.empty")}</p>;
  return <div className="digital-session-messages">{messages.map((message) => {
    const severity = severityLabel(message.severity, t);
    return <article key={message.id} className={message.severity}
      aria-label={t("digitalCenters.messages.fromStation", { severity, station })}>
      <Circle size={9} fill="currentColor" aria-hidden="true" />
      <span><span className="digital-message-meta"><strong>{severity}</strong>
        <small>{station}</small></span>
      <span>{message.message}</span><small>{message.nextAction}</small></span>
      <time>{formatTime(message.createdAt, language)}</time></article>;
  })}</div>;
}

function severityLabel(severity: DigitalCenterSessionMessage["severity"], t: Translate) {
  return t(`digitalCenters.messages.severity.${severity}`);
}

function connectionLabel(state: ECoSLiveStatus["state"] | undefined, t: Translate) {
  if (state === "running") return t("digitalCenters.common.connected");
  if (state === "interrupted") return t("digitalCenters.common.interrupted");
  return t("digitalCenters.common.stopped");
}

function formatTime(value: string | undefined, language: Language) {
  if (!value) return "–";
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleTimeString(language === "de" ? "de-DE" : "en-GB", {
    hour: "2-digit", minute: "2-digit", second: "2-digit"
  });
}

type Translate = (key: string, values?: Record<string, string | number>) => string;
