import { Circle, ExternalLink, LockKeyhole, RadioTower } from "lucide-react";

import type {
  DigitalCenterSessionMessage,
  DigitalCenterSummary,
  DigitalCenterWorkspaceTab,
  ECoSLiveStatus
} from "./digitalCenterModel";
import { LivePulseChart } from "./LivePulseChart";

const tabs: Array<{ value: DigitalCenterWorkspaceTab; label: string }> = [
  { value: "live", label: "Live-Status" },
  { value: "diagnosis", label: "Diagnose" },
  { value: "messages", label: "Meldungen" }
];

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
  return (
    <section className="digital-centers-panel digital-centers-status">
      <div className="digital-status-tabs" role="tablist" aria-label="Digitalzentralen-Status">
        {tabs.map((item) => <button key={item.value} type="button" role="tab"
          aria-selected={tab === item.value} onClick={() => onTab(item.value)}>{item.label}</button>)}
      </div>
      <div className="digital-status-content">
        {tab === "live" && <LiveStatus status={liveStatus} canMonitor={actions.canMonitor}
          loading={loading} error={errors.live} />}
        {tab === "diagnosis" && <Diagnosis status={liveStatus} center={selectedCenter}
          actions={actions} messages={messages} />}
        {tab === "messages" && <Messages messages={messages} error={errors.messages}
          station={selectedCenter?.name ?? "Digitalzentrale"} />}
      </div>
      <div className="digital-write-lock">
        <LockKeyhole size={20} aria-hidden="true" />
        <span><strong>Schreiben gesperrt</strong><small>{actions.canWrite
          ? "In dieser Zentrale sind Schreibbefehle gesperrt."
          : "Diese Digitalzentrale unterstützt keine Schreibbefehle."}</small></span>
        <button type="button" className="digital-center-button" disabled={!actions.canDiagnose}
          onClick={() => onTab("diagnosis")}>Diagnose öffnen <ExternalLink size={14} aria-hidden="true" /></button>
      </div>
    </section>
  );
}

function LiveStatus({ status, canMonitor, loading, error }: {
  status: ECoSLiveStatus | null; canMonitor: boolean; loading: boolean; error: string;
}) {
  if (!canMonitor) return <p className="digital-centers-state">Live-Monitor nicht verfügbar</p>;
  if (loading && !status) return <p className="digital-centers-state">Live-Status wird geladen</p>;
  if (error && !status) return <p className="digital-centers-state error">{error}</p>;
  if (status?.state === "interrupted") {
    return <p className="digital-centers-state error">Verbindung unterbrochen</p>;
  }
  if (!status?.connected) return <p className="digital-centers-state">Nicht verbunden</p>;
  return <>
    <div className="digital-live-metrics">
      <Metric label="Blöcke" value={status.blocksReceived} />
      <Metric label="Antworten" value={status.repliesReceived} active />
      <Metric label="Ereignisse" value={status.eventsReceived} />
      <Metric label="Letzte Nachricht" value={formatTime(status.lastSeenAt)} />
    </div>
    <LivePulseChart samples={status.pulseSamples} />
    <div className="digital-live-events">
      <header><strong>Letzte Ereignisse</strong><button type="button">Alle anzeigen</button></header>
      <div>{status.recentEvents.length === 0 && <p className="digital-centers-state">Keine Ereignisse</p>}
        {status.recentEvents.map((event, index) => <p key={`${event.at}-${index}`} title={event.message}>
          <time>{formatTime(event.at)}</time><Circle size={8} fill="currentColor" aria-hidden="true" />
          <span>{event.message}</span><small>{event.protocol}</small></p>)}</div>
    </div>
  </>;
}

function Metric({ label, value, active = false }: { label: string; value: string | number; active?: boolean }) {
  return <span><small>{label}</small><strong>{value}{active && <Circle size={8} fill="currentColor" aria-hidden="true" />}</strong></span>;
}

function Diagnosis({ status, center, actions, messages }: {
  status: ECoSLiveStatus | null;
  center: DigitalCenterSummary | null;
  actions: WorkspaceActions;
  messages: DigitalCenterSessionMessage[];
}) {
  const endpointHost = status?.host ?? center?.host;
  const endpointPort = status?.port ?? center?.port;
  const endpoint = endpointHost && endpointPort ? `${endpointHost}:${endpointPort}` : "Nicht eingerichtet";
  const protocolErrors = messages.filter((message) =>
    message.severity === "error" || message.code === "parse.failed"
  ).length;
  const capabilities = [
    ["Testverbindung", actions.canTestConnection],
    ["Lesen", actions.canRead],
    ["Live-Monitor", actions.canMonitor],
    ["Schreiben", actions.canWrite],
    ["CV-Schreiben", actions.canWriteCVs],
    ["Diagnose", actions.canDiagnose]
  ] as const;
  return <div className="digital-diagnosis">
    <RadioTower size={22} aria-hidden="true" />
    <h3>Verbindungsdiagnose</h3>
    <dl><div><dt>Endpunkt</dt><dd>{endpoint}</dd></div>
      <div><dt>Zustand</dt><dd>{connectionLabel(status?.diagnosis.connectionState)}</dd></div>
      <div><dt>Latenz</dt><dd>Nicht gemessen</dd></div>
      <div><dt>Protokollfehler</dt><dd>{protocolErrors}</dd></div>
      <div><dt>Letzte erfolgreiche Kommunikation</dt>
        <dd>{formatTime(status?.diagnosis.lastSuccessfulCommunication)}</dd></div>
      <div><dt>Modus</dt><dd>{status?.diagnosis.passive ? "Passiv" : "Inaktiv"}</dd></div></dl>
    <h4>Unterstützte Funktionen</h4>
    <ul>{capabilities.map(([label, supported]) => <li key={label}>
      {label}: {supported ? "unterstützt" : "nicht unterstützt"}
    </li>)}</ul>
    {status?.diagnosis.lastError && <p className="digital-centers-state error">{status.diagnosis.lastError}</p>}
  </div>;
}

function Messages({ messages, error, station }: {
  messages: DigitalCenterSessionMessage[];
  error: string;
  station: string;
}) {
  if (error) return <p className="digital-centers-state error">{error}</p>;
  if (messages.length === 0) return <p className="digital-centers-state">Keine Meldungen</p>;
  return <div className="digital-session-messages">{messages.map((message) => {
    const severity = severityLabel(message.severity);
    return <article key={message.id} className={message.severity}
      aria-label={`${severity} von ${station}`}>
      <Circle size={9} fill="currentColor" aria-hidden="true" />
      <span><span className="digital-message-meta"><strong>{severity}</strong>
        <small>{station}</small></span>
      <span>{message.message}</span><small>{message.nextAction}</small></span>
      <time>{formatTime(message.createdAt)}</time></article>;
  })}</div>;
}

function severityLabel(severity: DigitalCenterSessionMessage["severity"]) {
  if (severity === "error") return "Fehler";
  if (severity === "warning") return "Warnung";
  return "Information";
}

function connectionLabel(state?: ECoSLiveStatus["state"]) {
  if (state === "running") return "Verbunden";
  if (state === "interrupted") return "Unterbrochen";
  return "Gestoppt";
}

function formatTime(value?: string) {
  if (!value) return "–";
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleTimeString("de-DE", {
    hour: "2-digit", minute: "2-digit", second: "2-digit"
  });
}
