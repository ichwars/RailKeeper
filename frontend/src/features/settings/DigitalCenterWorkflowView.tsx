import {
  Check,
  CheckCircle2,
  ChevronDown,
  CircleDot,
  Database,
  LockKeyhole,
  Monitor,
  Network,
  Pencil,
  Play,
  Power,
  PowerOff,
  RadioTower,
  RefreshCw,
  Router,
  Server,
  ShieldCheck,
  Trash2,
  Wifi
} from "lucide-react";
import type { ComponentType } from "react";

import type {
  DigitalCenterConnectionResult,
  DigitalCenterProbeResult,
  ECoSConnectionResult,
  ECoSLiveStatus
} from "../../shared/api";

export type DigitalProvider = "ecos" | "z21" | "intellibox3" | "cs3";
export type DigitalDiagnosticsTab = "test" | "diagnosis" | "messages";

type Translate = (key: string, values?: Record<string, string | number>) => string;
type ConnectionResult = ECoSConnectionResult | DigitalCenterConnectionResult;

type DigitalCenterWorkflowViewProps = {
  provider: DigitalProvider;
  configuredProviders: Record<DigitalProvider, boolean>;
  activeProviders: Record<DigitalProvider, boolean>;
  host: string;
  port: string;
  protocol: string;
  timeout: string;
  adapterId: string;
  configuredAt?: string;
  configuredBy?: string;
  canManageUsers: boolean;
  busy: string;
  message: string;
  connectionResult: ConnectionResult | null;
  testError: string;
  testSucceeded: boolean;
  testedAt?: string;
  probeResult: DigitalCenterProbeResult | null;
  liveStatus: ECoSLiveStatus | null;
  diagnosticsTab: DigitalDiagnosticsTab;
  formatDateTime: (value: string) => string;
  t: Translate;
  onSelectProvider: (provider: DigitalProvider) => void;
  onConfigure: () => void;
  onTest: () => void;
  onActivate: () => void;
  onDeactivate: () => void;
  onStartLive: () => void;
  onStopLive: () => void;
  onRemove: () => void;
  onSelectDiagnosticsTab: (tab: DigitalDiagnosticsTab) => void;
  onProbe: () => void;
  onRefreshLive: () => void;
};

const providers: DigitalProvider[] = ["ecos", "z21", "intellibox3", "cs3"];

const providerIcons: Record<DigitalProvider, ComponentType<{ size?: number }>> = {
  ecos: Server,
  z21: RadioTower,
  intellibox3: Router,
  cs3: Network
};

function providerSupportsProbe(provider: DigitalProvider) {
  return provider === "z21" || provider === "intellibox3" || provider === "cs3";
}

function providerSupportsRead(provider: DigitalProvider) {
  return provider === "ecos" || provider === "cs3";
}

function resultFieldEntries(result: ConnectionResult | null) {
  if (!result) return [];
  const fields = { ...(result.fields || {}) };
  if ("protocolVersion" in result && result.protocolVersion) fields.protocolVersion = result.protocolVersion;
  if ("applicationVersion" in result && result.applicationVersion) fields.applicationVersion = result.applicationVersion;
  if ("hardwareVersion" in result && result.hardwareVersion) fields.hardwareVersion = result.hardwareVersion;
  return Object.entries(fields);
}

export function DigitalCenterWorkflowView({
  provider,
  configuredProviders,
  activeProviders,
  host,
  port,
  protocol,
  timeout,
  adapterId,
  configuredAt,
  configuredBy,
  canManageUsers,
  busy,
  message,
  connectionResult,
  testError,
  testSucceeded,
  testedAt,
  probeResult,
  liveStatus,
  diagnosticsTab,
  formatDateTime,
  t,
  onSelectProvider,
  onConfigure,
  onTest,
  onActivate,
  onDeactivate,
  onStartLive,
  onStopLive,
  onRemove,
  onSelectDiagnosticsTab,
  onProbe,
  onRefreshLive
}: DigitalCenterWorkflowViewProps) {
  const configured = configuredProviders[provider];
  const active = activeProviders[provider];
  const liveConnected = provider === "ecos" && Boolean(liveStatus?.connected);
  const tested = testSucceeded || liveConnected;
  const isBusy = busy !== "idle";
  const canStartLive = canManageUsers && provider === "ecos" && configured && active && tested;
  const currentProbe = probeResult?.provider === provider ? probeResult : null;
  const resultFields = resultFieldEntries(connectionResult);
  const activeStep = !configured ? 2 : !tested ? 3 : !active ? 4 : 4;
  const testState = tested ? "done" : configured ? "current" : "locked";
  const activationState = active ? "done" : tested ? "current" : "locked";

  const renderStepStatus = (state: "done" | "current" | "locked") => {
    if (state === "done") return <span className="digital-workflow-status done"><Check size={13} />{t("settings.digital.workflow.done")}</span>;
    if (state === "current") return <span className="digital-workflow-status current"><CircleDot size={13} />{t("settings.digital.workflow.current")}</span>;
    return <span className="digital-workflow-status locked"><LockKeyhole size={13} />{t("settings.digital.workflow.locked")}</span>;
  };

  const renderTestPanel = () => (
    <div className="digital-diagnostic-content" role="tabpanel" id="digital-panel-test" aria-labelledby="digital-tab-test">
      <div className={`digital-test-summary ${tested ? "success" : testError ? "error" : "empty"}`}>
        <span className="digital-pulse-icon" aria-hidden="true">
          {tested ? <CheckCircle2 size={24} /> : <CircleDot size={24} />}
        </span>
        <div>
          <strong>
            {tested
              ? t("settings.digital.workflow.testSuccess")
              : testError
                ? t("settings.digital.notConnected")
                : t("settings.digital.workflow.noTestTitle")}
          </strong>
          <span>
            {connectionResult?.message || testError || t("settings.digital.workflow.noTestHelp")}
          </span>
          {testedAt && <small>{t("settings.digital.workflow.testedAt", { value: formatDateTime(testedAt) })}</small>}
        </div>
      </div>

      <div className="digital-technical-details">
        <strong>{t("settings.digital.workflow.technicalDetails")}</strong>
        <dl>
          <div><dt>{t("settings.digital.protocol")}</dt><dd>{protocol}</dd></div>
          <div><dt>{t("settings.digital.workflow.lastConfigured")}</dt><dd>{configuredAt ? formatDateTime(configuredAt) : "–"}</dd></div>
          <div><dt>{t("settings.digital.workflow.host")}</dt><dd>{host || "–"}</dd></div>
          <div><dt>{t("settings.digital.workflow.configuredBy")}</dt><dd>{configuredBy || "–"}</dd></div>
          <div><dt>{t("settings.digital.workflow.port")}</dt><dd>{port}</dd></div>
          <div><dt>{t("settings.digital.workflow.adapterId")}</dt><dd><code>{adapterId}</code></dd></div>
          <div><dt>{t("settings.digital.workflow.timeout")}</dt><dd>{timeout}</dd></div>
          <div>
            <dt>{t("settings.digital.workflow.driverVersion")}</dt>
            <dd>{connectionResult && "applicationVersion" in connectionResult ? connectionResult.applicationVersion || "–" : connectionResult?.fields?.version || "–"}</dd>
          </div>
        </dl>
      </div>
    </div>
  );

  const renderDiagnosisPanel = () => (
    <div className="digital-diagnostic-content single" role="tabpanel" id="digital-panel-diagnosis" aria-labelledby="digital-tab-diagnosis">
      <div className="digital-diagnostic-toolbar">
        <div>
          <strong>{t("settings.digital.workflow.diagnosisTitle")}</strong>
          <span>
            {providerSupportsProbe(provider)
              ? t("settings.digital.workflow.diagnosisHelp")
              : t("settings.digital.workflow.diagnosisUnavailable")}
          </span>
        </div>
        {providerSupportsProbe(provider) && (
          <button type="button" className="secondary-button" onClick={onProbe} disabled={!canManageUsers || !configured || isBusy}>
            <Database size={15} />
            {busy === "diagnosing" ? t("settings.digital.probing") : t("settings.digital.probe")}
          </button>
        )}
      </div>

      {currentProbe ? (
        <>
          <div className={`digital-probe-summary ${currentProbe.connected ? "success" : "error"}`}>
            <strong>{currentProbe.message}</strong>
            <span>{currentProbe.host}:{currentProbe.port}</span>
          </div>
          {Object.keys(currentProbe.fields || {}).length > 0 && (
            <dl className="digital-workflow-field-list">
              {Object.entries(currentProbe.fields || {}).map(([key, value]) => (
                <div key={key}><dt>{key}</dt><dd>{value}</dd></div>
              ))}
            </dl>
          )}
          <div className="digital-workflow-command-list">
            {currentProbe.commands.map((command) => (
              <details key={command.name} open={!command.ok}>
                <summary>
                  <span><strong>{command.name}</strong><small>{command.description}</small></span>
                  <em className={command.ok ? "success" : "error"}>{command.ok ? t("settings.digital.probeOk") : t("settings.digital.probeFailed")}</em>
                </summary>
                <dl>
                  {command.request && <div><dt>{t("settings.digital.probeRequest")}</dt><dd><code>{command.request}</code></dd></div>}
                  {command.commandHex && <div><dt>{t("settings.digital.probeCommand")}</dt><dd><code>{command.commandHex}</code></dd></div>}
                  {command.responseHex && <div><dt>{t("settings.digital.probeResponse")}</dt><dd><code>{command.responseHex}</code></dd></div>}
                  {Object.entries(command.fields || {}).map(([key, value]) => <div key={key}><dt>{key}</dt><dd>{value}</dd></div>)}
                </dl>
                {command.error && <p>{command.error}</p>}
              </details>
            ))}
          </div>
        </>
      ) : resultFields.length > 0 ? (
        <dl className="digital-workflow-field-list">
          {resultFields.map(([key, value]) => <div key={key}><dt>{key}</dt><dd>{value}</dd></div>)}
        </dl>
      ) : (
        <div className="digital-diagnostic-empty"><Database size={24} /><span>{t("settings.digital.workflow.noDiagnosis")}</span></div>
      )}
    </div>
  );

  const renderMessagesPanel = () => (
    <div className="digital-diagnostic-content single" role="tabpanel" id="digital-panel-messages" aria-labelledby="digital-tab-messages">
      <div className="digital-diagnostic-toolbar">
        <div>
          <strong>{t("settings.digital.workflow.messagesTitle")}</strong>
          <span>{provider === "ecos" ? t("settings.digital.workflow.messagesHelp") : t("settings.digital.workflow.messagesUnavailable")}</span>
        </div>
        {provider === "ecos" && (
          <button type="button" className="secondary-button" onClick={onRefreshLive} disabled={!canManageUsers || isBusy}>
            <RefreshCw size={15} />{t("settings.digital.refresh")}
          </button>
        )}
      </div>
      {provider === "ecos" && liveStatus ? (
        <div className="digital-message-status">
          <div><span>{t("settings.digital.live")}</span><strong>{liveStatus.connected ? t("settings.digital.liveActive") : t("settings.digital.liveInactive")}</strong></div>
          <div><span>{t("settings.digital.blocks")}</span><strong>{liveStatus.blocksReceived}</strong></div>
          <div><span>{t("settings.digital.workflow.replies")}</span><strong>{liveStatus.repliesReceived}</strong></div>
          <div><span>{t("settings.digital.workflow.events")}</span><strong>{liveStatus.eventsReceived}</strong></div>
          <p>{liveStatus.lastMessage || liveStatus.message || t("settings.digital.workflow.noMessages")}</p>
        </div>
      ) : (
        <div className="digital-diagnostic-empty"><Monitor size={24} /><span>{t("settings.digital.workflow.noMessages")}</span></div>
      )}
    </div>
  );

  return (
    <section className="digital-workflow-page">
      <div className="digital-workflow-layout">
        <div className="digital-workflow-main">
          <section className="panel digital-provider-selector" aria-labelledby="digital-provider-heading">
            <header>
              <h2 id="digital-provider-heading">{t("settings.digital.workflow.selectTitle")}</h2>
              <p>{t("settings.digital.workflow.selectHelp")}</p>
            </header>
            <div className="digital-provider-options">
              {providers.map((providerId) => {
                const Icon = providerIcons[providerId];
                const selected = providerId === provider;
                return (
                  <button
                    type="button"
                    key={providerId}
                    className={selected ? "selected" : ""}
                    aria-pressed={selected}
                    onClick={() => onSelectProvider(providerId)}
                  >
                    <span className="digital-provider-icon"><Icon size={21} /></span>
                    <span className="digital-provider-copy">
                      <strong>{t(`settings.digital.provider.${providerId}`)}</strong>
                      <small>
                        {activeProviders[providerId]
                          ? t("settings.digital.active")
                          : configuredProviders[providerId]
                            ? t("settings.digital.workflow.configured")
                            : t("settings.digital.inactive")}
                      </small>
                    </span>
                    {selected && <CheckCircle2 size={17} className="digital-provider-check" />}
                  </button>
                );
              })}
            </div>
          </section>

          {!canManageUsers && (
            <div className="current-user-card">
              <strong>{t("settings.users.adminRequired")}</strong>
              <span>{t("settings.digital.adminHelp")}</span>
            </div>
          )}

          <section className="panel digital-commissioning" aria-label={t("settings.digital.workflow.title")}>
          <article className="digital-workflow-step done">
            <span className="digital-step-number">1</span>
            <div className="digital-step-body">
              <header><div><h3>{t("settings.digital.workflow.step1")}</h3><p>{t("settings.digital.workflow.step1Help")}</p></div>{renderStepStatus("done")}</header>
              <div className="digital-step-summary"><span>{t("settings.digital.provider")}</span><strong>{t(`settings.digital.provider.${provider}`)}</strong></div>
            </div>
          </article>

          <article className={`digital-workflow-step ${configured ? "done" : activeStep === 2 ? "current" : "locked"}`}>
            <span className="digital-step-number">2</span>
            <div className="digital-step-body">
              <header>
                <div><h3>{t("settings.digital.workflow.step2")}</h3><p>{configured ? t("settings.digital.workflow.step2Done") : t("settings.digital.workflow.step2Help")}</p></div>
                {renderStepStatus(configured ? "done" : "current")}
              </header>
              <div className="digital-step-summary split">
                <span>{t("settings.digital.workflow.connectionType")}</span><strong>{protocol}</strong>
                <span>{t("settings.digital.workflow.address")}</span><strong>{configured ? `${host}:${port}` : "–"}</strong>
                <button type="button" className="inline-action" onClick={onConfigure} disabled={!canManageUsers}>
                  <Pencil size={14} />{configured ? t("settings.digital.workflow.changeConfiguration") : t("settings.digital.configure")}
                </button>
              </div>
            </div>
          </article>

          <article className={`digital-workflow-step expanded ${testState}`}>
            <span className="digital-step-number">3</span>
            <div className="digital-step-body">
              <header><div><h3>{t("settings.digital.workflow.step3")}</h3><p>{t("settings.digital.workflow.step3Help")}</p></div>{renderStepStatus(testState)}</header>
              <div className="digital-test-actions">
                <button type="button" className="primary-button" onClick={onTest} disabled={!canManageUsers || !configured || isBusy}>
                  <Play size={15} />{busy === "testing" ? t("settings.digital.testing") : t("settings.digital.test")}
                </button>
                <button type="button" className="secondary-button" onClick={onConfigure} disabled={!canManageUsers || isBusy}>
                  <Pencil size={15} />{t("settings.digital.workflow.changeConfiguration")}
                </button>
              </div>

              <section className="digital-diagnostics-panel">
                <div className="digital-diagnostic-tabs" role="tablist" aria-label={t("settings.digital.workflow.diagnostics") }>
                  <button id="digital-tab-test" type="button" role="tab" aria-selected={diagnosticsTab === "test"} aria-controls="digital-panel-test" onClick={() => onSelectDiagnosticsTab("test")}>{t("settings.digital.workflow.tabTest")}</button>
                  <button id="digital-tab-diagnosis" type="button" role="tab" aria-selected={diagnosticsTab === "diagnosis"} aria-controls="digital-panel-diagnosis" onClick={() => onSelectDiagnosticsTab("diagnosis")}>{t("settings.digital.workflow.tabDiagnosis")}</button>
                  <button id="digital-tab-messages" type="button" role="tab" aria-selected={diagnosticsTab === "messages"} aria-controls="digital-panel-messages" onClick={() => onSelectDiagnosticsTab("messages")}>{t("settings.digital.workflow.tabMessages")}</button>
                </div>
                {diagnosticsTab === "test" && renderTestPanel()}
                {diagnosticsTab === "diagnosis" && renderDiagnosisPanel()}
                {diagnosticsTab === "messages" && renderMessagesPanel()}
              </section>
            </div>
          </article>

          <article className={`digital-workflow-step ${activationState}`}>
            <span className="digital-step-number">4</span>
            <div className="digital-step-body activation">
              <header>
                <div><h3>{t("settings.digital.workflow.step4")}</h3><p>{active ? t("settings.digital.workflow.step4Done") : t("settings.digital.workflow.step4Help")}</p></div>
                {renderStepStatus(activationState)}
              </header>
              <button
                type="button"
                className={active ? "secondary-button" : "primary-button"}
                onClick={active ? onDeactivate : onActivate}
                disabled={!canManageUsers || isBusy || (!active && !tested)}
              >
                {active ? <PowerOff size={15} /> : <Power size={15} />}
                {busy === "activating"
                  ? t("settings.digital.workflow.activating")
                  : busy === "deactivating"
                    ? t("settings.digital.workflow.deactivating")
                    : active
                      ? t("settings.digital.workflow.deactivate")
                      : t("settings.digital.workflow.activate")}
              </button>
            </div>
          </article>
          </section>
        </div>

        <aside className="digital-workflow-aside">
          <section className="panel digital-safety-card">
            <header><ShieldCheck size={26} /><div><h2>{t("settings.digital.workflow.safetyTitle")}</h2><p>{t("settings.digital.workflow.safetySubtitle")}</p></div></header>
            <div className="digital-capability-list">
              <article><Wifi size={17} /><div><strong>{t("settings.digital.workflow.safetyTestTitle")}</strong><span>{t("settings.digital.workflow.safetyTestHelp")}</span></div><span className="capability-state available"><Check size={12} /></span></article>
              <article><Database size={17} /><div><strong>{t("settings.digital.workflow.safetyReadTitle")}</strong><span>{provider === "cs3" ? t("settings.digital.workflow.safetyReadCS3Help") : providerSupportsRead(provider) ? t("settings.digital.workflow.safetyReadHelp") : t("settings.digital.workflow.safetyReadLimited")}</span></div><span className={`capability-state ${providerSupportsRead(provider) ? "available" : "limited"}`}>{providerSupportsRead(provider) ? <Check size={12} /> : "–"}</span></article>
              <article><Monitor size={17} /><div><strong>{t("settings.digital.workflow.safetyMonitorTitle")}</strong><span>{provider === "ecos" ? t("settings.digital.workflow.safetyMonitorHelp") : t("settings.digital.workflow.safetyMonitorUnavailable")}</span></div><span className={`capability-state ${provider === "ecos" ? "available" : "limited"}`}>{provider === "ecos" ? <Check size={12} /> : "–"}</span></article>
              <article><LockKeyhole size={17} /><div><strong>{t("settings.digital.workflow.safetyWriteTitle")}</strong><span>{provider === "ecos" ? t("settings.digital.workflow.safetyWriteHelp") : t("settings.digital.workflow.safetyWriteUnavailable")}</span></div><span className="capability-state locked"><LockKeyhole size={11} /></span></article>
            </div>
            <details className="digital-safety-details"><summary>{t("settings.digital.workflow.details")}<ChevronDown size={15} /></summary><p>{t("settings.digital.workflow.scopeNote")}</p></details>
          </section>

          <section className="panel digital-actions-card">
            <header><Server size={19} /><h2>{t("settings.digital.workflow.actions")}</h2></header>
            <button
              type="button"
              className={liveConnected ? "secondary-button" : "primary-button"}
              onClick={liveConnected ? onStopLive : onStartLive}
              disabled={isBusy || (!liveConnected && !canStartLive)}
            >
              {liveConnected ? <PowerOff size={15} /> : <Play size={15} />}
              {busy === "starting"
                ? t("settings.digital.starting")
                : busy === "stopping"
                  ? t("settings.digital.stopping")
                  : liveConnected
                    ? t("settings.digital.stopLive")
                    : t("settings.digital.startLive")}
            </button>
            {!liveConnected && <p>{provider === "ecos" ? t("settings.digital.workflow.liveUnavailable") : t("settings.digital.workflow.liveProviderUnavailable")}</p>}
            <button type="button" className="secondary-button danger-outline" onClick={onRemove} disabled={!canManageUsers || !configured || isBusy}>
              <Trash2 size={15} />{t("settings.digital.workflow.remove")}
            </button>
          </section>
        </aside>
      </div>

      {message && <p className="form-message digital-workflow-message" role="status">{message}</p>}
      <div className="digital-workflow-footnote"><ShieldCheck size={15} />{t("settings.digital.workflow.footer")}</div>
    </section>
  );
}
