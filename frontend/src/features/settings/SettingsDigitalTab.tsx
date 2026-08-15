import { useEffect, useState } from "react";
import { Database, RefreshCw, Save, Server, X } from "lucide-react";
import { api, DigitalCenterConnectionResult, DigitalCenterProbeResult, DigitalCenterSettings, ECoSConnectionResult, ECoSLiveStatus, ECoSLocomotiveSummary } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { DigitalCenterWorkflowView, type DigitalDiagnosticsTab, type DigitalProvider } from "./DigitalCenterWorkflowView";
import { localSettingKeys, readLocalBool, readLocalSetting } from "./settingsModel";

type BusyState = "idle" | "testing" | "diagnosing" | "probing" | "starting" | "stopping" | "refreshing" | "activating" | "deactivating" | "removing";
type ConfigurationMeta = Partial<Record<DigitalProvider, { configuredAt: string; configuredBy: string }>>;

type SettingsDigitalTabProps = {
  canManageUsers: boolean;
  formatDateTime: (value: string) => string;
  username: string;
};

const digitalProviders: DigitalProvider[] = ["ecos", "z21", "intellibox3", "cs3"];
const configurationMetaKey = "railkeeper.digital.configurationMeta";

function readConfigurationMeta(): ConfigurationMeta {
  try {
    return JSON.parse(window.localStorage.getItem(configurationMetaKey) || "{}") as ConfigurationMeta;
  } catch {
    return {};
  }
}

export function SettingsDigitalTab({ canManageUsers, formatDateTime, username }: SettingsDigitalTabProps) {
  const { t } = useI18n();
  const [provider, setProvider] = useState<DigitalProvider>(() => {
    const stored = readLocalSetting(localSettingKeys.digitalProvider, "ecos");
    return digitalProviders.includes(stored as DigitalProvider) ? stored as DigitalProvider : "ecos";
  });
  const [activeDialogProvider, setActiveDialogProvider] = useState<DigitalProvider | null>(null);
  const [ecosEnabled, setEcosEnabled] = useState(() => {
    const stored = window.localStorage.getItem(localSettingKeys.digitalEcosEnabled);
    if (stored !== null) return readLocalBool(localSettingKeys.digitalEcosEnabled, false);
    return Boolean(readLocalSetting(localSettingKeys.digitalEcosHost, "").trim());
  });
  const [ecosHost, setEcosHost] = useState(() => readLocalSetting(localSettingKeys.digitalEcosHost, ""));
  const [ecosPort, setEcosPort] = useState(() => readLocalSetting(localSettingKeys.digitalEcosPort, "15471"));
  const [z21Enabled, setZ21Enabled] = useState(() => readLocalBool(localSettingKeys.digitalZ21Enabled, false));
  const [z21Host, setZ21Host] = useState(() => readLocalSetting(localSettingKeys.digitalZ21Host, ""));
  const [z21Port, setZ21Port] = useState(() => readLocalSetting(localSettingKeys.digitalZ21Port, "21105"));
  const [intellibox3Enabled, setIntellibox3Enabled] = useState(() => readLocalBool(localSettingKeys.digitalIntellibox3Enabled, false));
  const [intellibox3Host, setIntellibox3Host] = useState(() => readLocalSetting(localSettingKeys.digitalIntellibox3Host, ""));
  const [intellibox3Port, setIntellibox3Port] = useState(() => readLocalSetting(localSettingKeys.digitalIntellibox3Port, "21105"));
  const [cs3Enabled, setCS3Enabled] = useState(() => readLocalBool(localSettingKeys.digitalCS3Enabled, false));
  const [cs3Host, setCS3Host] = useState(() => readLocalSetting(localSettingKeys.digitalCS3Host, ""));
  const [cs3Port, setCS3Port] = useState(() => readLocalSetting(localSettingKeys.digitalCS3Port, "80"));
  const [busy, setBusy] = useState<BusyState>("idle");
  const [connectionResult, setConnectionResult] = useState<ECoSConnectionResult | DigitalCenterConnectionResult | null>(null);
  const [probeResult, setProbeResult] = useState<DigitalCenterProbeResult | null>(null);
  const [locomotiveSummary, setLocomotiveSummary] = useState<ECoSLocomotiveSummary | null>(null);
  const [liveStatus, setLiveStatus] = useState<ECoSLiveStatus | null>(null);
  const [message, setMessage] = useState("");
  const [dialogMessage, setDialogMessage] = useState("");
  const [diagnosticsTab, setDiagnosticsTab] = useState<DigitalDiagnosticsTab>("test");
  const [lastTestedProvider, setLastTestedProvider] = useState<DigitalProvider | null>(null);
  const [lastTestedAt, setLastTestedAt] = useState("");
  const [testError, setTestError] = useState("");
  const [configurationMeta, setConfigurationMeta] = useState<ConfigurationMeta>(() => readConfigurationMeta());

  const ecosInput = () => ({
    host: ecosHost.trim(),
    port: Number(ecosPort) || 15471
  });

  const providerEnabled = (providerId: DigitalProvider) => {
    if (providerId === "ecos") return ecosEnabled;
    if (providerId === "z21") return z21Enabled;
    if (providerId === "intellibox3") return intellibox3Enabled;
    return cs3Enabled;
  };

  const providerHost = (providerId: DigitalProvider) => {
    if (providerId === "ecos") return ecosHost.trim();
    if (providerId === "z21") return z21Host.trim();
    if (providerId === "intellibox3") return intellibox3Host.trim();
    return cs3Host.trim();
  };

  const providerPort = (providerId: DigitalProvider) => {
    if (providerId === "ecos") return ecosPort.trim() || "15471";
    if (providerId === "z21") return z21Port.trim() || "21105";
    if (providerId === "intellibox3") return intellibox3Port.trim() || "21105";
    return cs3Port.trim() || "80";
  };

  const providerInput = (providerId: DigitalProvider) => ({
    host: providerHost(providerId),
    port: Number(providerPort(providerId))
  });

  const providerConfigured = (providerId: DigitalProvider) => Boolean(providerHost(providerId));
  const providerActive = (providerId: DigitalProvider) => providerEnabled(providerId) && providerConfigured(providerId);

  const currentDigitalSettings = (nextProvider: DigitalProvider = provider): DigitalCenterSettings => ({
    provider: nextProvider,
    ecos: { enabled: ecosEnabled, host: ecosHost.trim(), port: ecosPort.trim() || "15471" },
    z21: { enabled: z21Enabled, host: z21Host.trim(), port: z21Port.trim() || "21105" },
    intellibox3: { enabled: intellibox3Enabled, host: intellibox3Host.trim(), port: intellibox3Port.trim() || "21105" },
    cs3: { enabled: cs3Enabled, host: cs3Host.trim(), port: cs3Port.trim() || "80" }
  });

  const applyDigitalSettings = (settings: DigitalCenterSettings) => {
    const nextProvider = digitalProviders.includes(settings.provider) ? settings.provider : "ecos";
    setProvider(nextProvider);
    setEcosEnabled(Boolean(settings.ecos?.enabled));
    setEcosHost(settings.ecos?.host || "");
    setEcosPort(settings.ecos?.port || "15471");
    setZ21Enabled(Boolean(settings.z21?.enabled));
    setZ21Host(settings.z21?.host || "");
    setZ21Port(settings.z21?.port || "21105");
    setIntellibox3Enabled(Boolean(settings.intellibox3?.enabled));
    setIntellibox3Host(settings.intellibox3?.host || "");
    setIntellibox3Port(settings.intellibox3?.port || "21105");
    setCS3Enabled(Boolean(settings.cs3?.enabled));
    setCS3Host(settings.cs3?.host || "");
    setCS3Port(settings.cs3?.port || "80");
  };

  const hasDigitalConfig = (settings: DigitalCenterSettings) =>
    Boolean(
      settings.ecos?.host ||
      settings.z21?.host ||
      settings.intellibox3?.host ||
      settings.cs3?.host ||
      settings.ecos?.enabled ||
      settings.z21?.enabled ||
      settings.intellibox3?.enabled ||
      settings.cs3?.enabled
    );

  const setActiveProviderHost = (host: string) => {
    if (activeDialogProvider === "ecos") setEcosHost(host);
    if (activeDialogProvider === "z21") setZ21Host(host);
    if (activeDialogProvider === "intellibox3") setIntellibox3Host(host);
    if (activeDialogProvider === "cs3") setCS3Host(host);
  };

  const setActiveProviderPort = (port: string) => {
    if (activeDialogProvider === "ecos") setEcosPort(port);
    if (activeDialogProvider === "z21") setZ21Port(port);
    if (activeDialogProvider === "intellibox3") setIntellibox3Port(port);
    if (activeDialogProvider === "cs3") setCS3Port(port);
  };

  const storeSettingsLocally = (settings: DigitalCenterSettings) => {
    const nextProvider = settings.provider;
    window.localStorage.setItem(localSettingKeys.digitalProvider, nextProvider);
    window.localStorage.setItem(localSettingKeys.digitalEcosEnabled, String(settings.ecos.enabled));
    window.localStorage.setItem(localSettingKeys.digitalEcosHost, settings.ecos.host);
    window.localStorage.setItem(localSettingKeys.digitalEcosPort, settings.ecos.port);
    window.localStorage.setItem(localSettingKeys.digitalZ21Enabled, String(settings.z21.enabled));
    window.localStorage.setItem(localSettingKeys.digitalZ21Host, settings.z21.host);
    window.localStorage.setItem(localSettingKeys.digitalZ21Port, settings.z21.port);
    window.localStorage.setItem(localSettingKeys.digitalIntellibox3Enabled, String(settings.intellibox3.enabled));
    window.localStorage.setItem(localSettingKeys.digitalIntellibox3Host, settings.intellibox3.host);
    window.localStorage.setItem(localSettingKeys.digitalIntellibox3Port, settings.intellibox3.port);
    window.localStorage.setItem(localSettingKeys.digitalCS3Enabled, String(settings.cs3.enabled));
    window.localStorage.setItem(localSettingKeys.digitalCS3Host, settings.cs3.host);
    window.localStorage.setItem(localSettingKeys.digitalCS3Port, settings.cs3.port);
    window.dispatchEvent(new Event("railkeeper-digital-settings-changed"));
  };

  const persistSettings = async (settings: DigitalCenterSettings) => {
    storeSettingsLocally(settings);
    const saved = await api.updateDigitalSettings(settings);
    applyDigitalSettings(saved);
    return saved;
  };

  const rememberSettings = (nextProvider: DigitalProvider = provider) => {
    const settings = currentDigitalSettings(nextProvider);
    storeSettingsLocally(settings);
    void api.updateDigitalSettings(settings).then(applyDigitalSettings).catch(() => undefined);
  };

  const settingsWithProviderEnabled = (providerId: DigitalProvider, enabled: boolean): DigitalCenterSettings => {
    const settings = currentDigitalSettings(providerId);
    return {
      ...settings,
      [providerId]: {
        ...settings[providerId],
        enabled
      }
    };
  };

  const settingsWithoutProviderConnection = (providerId: DigitalProvider): DigitalCenterSettings => {
    const settings = currentDigitalSettings(providerId);
    const defaultPort = providerId === "ecos" ? "15471" : providerId === "cs3" ? "80" : "21105";
    return {
      ...settings,
      [providerId]: {
        enabled: false,
        host: "",
        port: defaultPort
      }
    };
  };

  const refreshLiveStatus = async (nextBusy: BusyState = "refreshing", announce = true) => {
    if (!canManageUsers) return;
    setBusy(nextBusy);
    if (announce) setMessage("");
    try {
      const status = await api.getECoSLiveStatus();
      setLiveStatus(status);
      if (announce) setMessage(status.message);
    } catch (error) {
      if (announce) setMessage(error instanceof Error ? error.message : t("settings.digital.error"));
    } finally {
      setBusy("idle");
    }
  };

  useEffect(() => {
    if (canManageUsers) {
      api
        .digitalSettings()
        .then((settings) => {
          if (hasDigitalConfig(settings)) {
            applyDigitalSettings(settings);
            return;
          }
          const localSettings = currentDigitalSettings();
          if (hasDigitalConfig(localSettings)) {
            void api.updateDigitalSettings(localSettings).then(applyDigitalSettings).catch(() => undefined);
          }
        })
        .catch(() => undefined);
      void refreshLiveStatus("idle", false);
    }
  }, [canManageUsers]);

  const openAdapter = (nextProvider: DigitalProvider) => {
    setActiveDialogProvider(nextProvider);
    setDialogMessage("");
    setProbeResult(null);
  };

  const selectProvider = (nextProvider: DigitalProvider) => {
    if (nextProvider === provider) return;
    setProvider(nextProvider);
    setConnectionResult(null);
    setProbeResult(null);
    setLastTestedProvider(null);
    setLastTestedAt("");
    setTestError("");
    setDiagnosticsTab("test");
    rememberSettings(nextProvider);
  };

  const closeAdapter = () => {
    setActiveDialogProvider(null);
    setDialogMessage("");
    setProbeResult(null);
  };

  const saveAdapter = () => {
    if (!activeDialogProvider) return;
    if (!providerHost(activeDialogProvider)) {
      setDialogMessage(t("settings.digital.hostRequired"));
      return;
    }
    const savedProvider = activeDialogProvider;
    const nextMeta = {
      ...configurationMeta,
      [savedProvider]: {
        configuredAt: new Date().toISOString(),
        configuredBy: username
      }
    };
    setProvider(savedProvider);
    rememberSettings(savedProvider);
    setConfigurationMeta(nextMeta);
    window.localStorage.setItem(configurationMetaKey, JSON.stringify(nextMeta));
    setLocomotiveSummary(null);
    setMessage(t("settings.digital.saved"));
    closeAdapter();
  };

  const testConnection = async () => {
    const testedProvider = activeDialogProvider || provider;
    rememberSettings(testedProvider);
    setBusy("testing");
    setMessage("");
    setDialogMessage("");
    setProbeResult(null);
    setConnectionResult(null);
    setLastTestedProvider(testedProvider);
    setLastTestedAt("");
    setTestError("");
    setDiagnosticsTab("test");
    try {
      const result = testedProvider === "ecos"
        ? await api.testECoSConnection(providerInput("ecos"))
        : testedProvider === "z21"
          ? await api.testZ21Connection(providerInput("z21"))
          : testedProvider === "intellibox3"
            ? await api.testIntellibox3Connection(providerInput("intellibox3"))
          : await api.testCS3Connection(providerInput("cs3"));
      setConnectionResult(result);
      setLastTestedAt(new Date().toISOString());
      setMessage(result.message);
      setDialogMessage(result.message);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t("settings.digital.error");
      setTestError(errorMessage);
      setLastTestedAt(new Date().toISOString());
      setMessage(errorMessage);
      setDialogMessage(errorMessage);
    } finally {
      setBusy("idle");
    }
  };

  const canProbeProvider = (providerId: DigitalProvider | null) => providerId === "z21" || providerId === "intellibox3";

  const probeConnection = async () => {
    const probedProvider = activeDialogProvider || provider;
    if (!canProbeProvider(probedProvider)) return;
    rememberSettings(probedProvider);
    setBusy("diagnosing");
    setMessage("");
    setDialogMessage("");
    setProbeResult(null);
    try {
      const result = probedProvider === "z21"
        ? await api.probeZ21Connection(providerInput("z21"))
        : await api.probeIntellibox3Connection(providerInput("intellibox3"));
      setProbeResult(result);
      setDiagnosticsTab("diagnosis");
      setMessage(result.message);
      setDialogMessage(result.message);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t("settings.digital.error");
      setMessage(errorMessage);
      setDialogMessage(errorMessage);
    } finally {
      setBusy("idle");
    }
  };

  const activateProvider = async () => {
    const hasSuccessfulTest = (lastTestedProvider === provider && Boolean(connectionResult?.connected))
      || (provider === "ecos" && Boolean(liveStatus?.connected));
    if (!providerConfigured(provider) || !hasSuccessfulTest) return;
    setBusy("activating");
    setMessage("");
    try {
      await persistSettings(settingsWithProviderEnabled(provider, true));
      setMessage(t("settings.digital.workflow.activated"));
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t("settings.digital.error"));
    } finally {
      setBusy("idle");
    }
  };

  const deactivateProvider = async () => {
    setBusy("deactivating");
    setMessage("");
    try {
      if (provider === "ecos" && liveStatus?.connected) {
        const status = await api.stopECoSLive();
        setLiveStatus(status);
      }
      await persistSettings(settingsWithProviderEnabled(provider, false));
      setMessage(t("settings.digital.workflow.deactivated"));
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t("settings.digital.error"));
    } finally {
      setBusy("idle");
    }
  };

  const removeProviderConnection = async () => {
    if (!providerConfigured(provider)) return;
    if (!window.confirm(t("settings.digital.workflow.removeConfirm", { provider: providerLabel(provider) }))) return;
    setBusy("removing");
    setMessage("");
    try {
      if (provider === "ecos" && liveStatus?.connected) {
        const status = await api.stopECoSLive();
        setLiveStatus(status);
      }
      await persistSettings(settingsWithoutProviderConnection(provider));
      const nextMeta = { ...configurationMeta };
      delete nextMeta[provider];
      setConfigurationMeta(nextMeta);
      window.localStorage.setItem(configurationMetaKey, JSON.stringify(nextMeta));
      setConnectionResult(null);
      setProbeResult(null);
      setLastTestedProvider(null);
      setLastTestedAt("");
      setTestError("");
      setLocomotiveSummary(null);
      setMessage(t("settings.digital.workflow.removed"));
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t("settings.digital.error"));
    } finally {
      setBusy("idle");
    }
  };

  const countLocomotives = async (silent = false) => {
    rememberSettings("ecos");
    if (!silent) {
      setBusy("probing");
      setMessage("");
    }
    try {
      const summary = await api.countECoSLocomotives(ecosInput());
      setLocomotiveSummary(summary);
      if (!silent) {
        setMessage(t("settings.digital.locomotiveProbeDone", { count: summary.count }));
      }
    } catch (error) {
      if (!silent) {
        setMessage(error instanceof Error ? error.message : t("settings.digital.error"));
      }
    } finally {
      if (!silent) {
        setBusy("idle");
      }
    }
  };

  const startLive = async () => {
    rememberSettings("ecos");
    setBusy("starting");
    setMessage("");
    setLocomotiveSummary(null);
    try {
      const status = await api.startECoSLive(ecosInput());
      setLiveStatus(status);
      setMessage(status.message);
      if (status.connected) {
        void countLocomotives(true);
      }
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t("settings.digital.error"));
    } finally {
      setBusy("idle");
    }
  };

  const stopLive = async () => {
    setBusy("stopping");
    setMessage("");
    try {
      const status = await api.stopECoSLive();
      setLiveStatus(status);
      setMessage(status.message);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t("settings.digital.error"));
    } finally {
      setBusy("idle");
    }
  };

  const providerLabel = (providerId: DigitalProvider) => t(`settings.digital.provider.${providerId}`);
  const ecosReady = provider === "ecos" && providerActive("ecos");
  const dialogCanTest = canManageUsers && Boolean(activeDialogProvider);
  const dialogCanSave = Boolean(activeDialogProvider);
  const probeFieldEntries = Object.entries(probeResult?.fields || {});
  const configuredProviders = Object.fromEntries(
    digitalProviders.map((providerId) => [providerId, providerConfigured(providerId)])
  ) as Record<DigitalProvider, boolean>;
  const activeProviders = Object.fromEntries(
    digitalProviders.map((providerId) => [providerId, providerActive(providerId)])
  ) as Record<DigitalProvider, boolean>;
  const currentConnectionResult = lastTestedProvider === provider ? connectionResult : null;
  const currentTestError = lastTestedProvider === provider ? testError : "";
  const testSucceeded = Boolean(
    currentConnectionResult?.connected || (provider === "ecos" && liveStatus?.connected)
  );
  const providerProtocols = Object.fromEntries(
    digitalProviders.map((providerId) => [providerId, t(`settings.digital.workflow.protocol.${providerId}`)])
  ) as Record<DigitalProvider, string>;
  const providerTimeouts: Record<DigitalProvider, string> = {
    ecos: "8 s",
    z21: "4 s",
    intellibox3: "4 s",
    cs3: "4 s"
  };

  useEffect(() => {
    if (!liveStatus?.connected || !ecosReady || locomotiveSummary || busy !== "idle") return;
    void countLocomotives(true);
  }, [liveStatus?.connected, ecosReady, locomotiveSummary, busy]);

  return (
    <>
      <DigitalCenterWorkflowView
        provider={provider}
        configuredProviders={configuredProviders}
        activeProviders={activeProviders}
        host={providerHost(provider)}
        port={providerPort(provider)}
        protocol={providerProtocols[provider]}
        timeout={providerTimeouts[provider]}
        adapterId={provider === "intellibox3" ? "intellibox3" : provider}
        configuredAt={configurationMeta[provider]?.configuredAt}
        configuredBy={configurationMeta[provider]?.configuredBy}
        canManageUsers={canManageUsers}
        busy={busy}
        message={message}
        connectionResult={currentConnectionResult}
        testError={currentTestError}
        testSucceeded={testSucceeded}
        testedAt={lastTestedProvider === provider ? lastTestedAt : undefined}
        probeResult={probeResult}
        liveStatus={liveStatus}
        diagnosticsTab={diagnosticsTab}
        formatDateTime={formatDateTime}
        t={t}
        onSelectProvider={selectProvider}
        onConfigure={() => openAdapter(provider)}
        onTest={testConnection}
        onActivate={activateProvider}
        onDeactivate={deactivateProvider}
        onStartLive={startLive}
        onStopLive={stopLive}
        onRemove={removeProviderConnection}
        onSelectDiagnosticsTab={setDiagnosticsTab}
        onProbe={probeConnection}
        onRefreshLive={() => refreshLiveStatus()}
      />

      {activeDialogProvider && (
        <div className="modal-layer digital-adapter-layer" role="dialog" aria-modal="true" aria-label={t("settings.digital.dialogTitle", { provider: providerLabel(activeDialogProvider) })}>
          <section className="vehicle-modal digital-adapter-dialog">
            <header className="modal-head">
              <div>
                <h2>{t("settings.digital.dialogTitle", { provider: providerLabel(activeDialogProvider) })}</h2>
                <p>{t("settings.digital.dialogSubtitle")}</p>
              </div>
              <button type="button" className="icon-button" onClick={closeAdapter} aria-label={t("vehicles.close")} title={t("vehicles.close")}>
                <X size={18} />
              </button>
            </header>

            <div className="digital-adapter-form">
              {!canManageUsers && (
                <div className="current-user-card">
                  <strong>{t("settings.users.adminRequired")}</strong>
                  <span>{t("settings.digital.adminHelp")}</span>
                </div>
              )}
              <div className="settings-field-grid">
                <label>
                  {activeDialogProvider ? t(`settings.digital.${activeDialogProvider}.host`) : t("settings.digital.ecos.host")}
                  <input
                    value={activeDialogProvider ? providerHost(activeDialogProvider) : ""}
                    onChange={(event) => setActiveProviderHost(event.target.value)}
                    placeholder={activeDialogProvider ? t(`settings.digital.${activeDialogProvider}.hostPlaceholder`) : ""}
                  />
                </label>
                <label>
                  {activeDialogProvider ? t(`settings.digital.${activeDialogProvider}.port`) : t("settings.digital.ecos.port")}
                  <input
                    value={activeDialogProvider ? providerPort(activeDialogProvider) : ""}
                    onChange={(event) => setActiveProviderPort(event.target.value)}
                    inputMode="numeric"
                    placeholder={activeDialogProvider === "z21" || activeDialogProvider === "intellibox3" ? "21105" : activeDialogProvider === "cs3" ? "80" : "15471"}
                  />
                </label>
              </div>

              {dialogMessage && <p className="form-message">{dialogMessage}</p>}

              {probeResult && (
                <section className="digital-probe-panel">
                  <div className="digital-probe-head">
                    <div>
                      <strong>{t("settings.digital.probeTitle")}</strong>
                      <span>{t("settings.digital.probeEndpoint", { host: probeResult.host, port: String(probeResult.port) })}</span>
                    </div>
                    <span className={probeResult.connected ? "digital-probe-state active" : "digital-probe-state inactive"}>
                      {probeResult.connected ? t("settings.digital.connected") : t("settings.digital.notConnected")}
                    </span>
                  </div>

                  {probeFieldEntries.length > 0 && (
                    <dl className="digital-probe-fields">
                      {probeFieldEntries.map(([key, value]) => (
                        <div key={key}>
                          <dt>{key}</dt>
                          <dd>{value}</dd>
                        </div>
                      ))}
                    </dl>
                  )}

                  <div className="digital-probe-command-list">
                    {probeResult.commands.map((command) => {
                      const commandFields = Object.entries(command.fields || {});
                      return (
                        <details className="digital-probe-command" key={command.name} open={command.ok}>
                          <summary>
                            <span>
                              <strong>{command.name}</strong>
                              <small>{command.description}</small>
                            </span>
                            <em>{command.ok ? t("settings.digital.probeOk") : t("settings.digital.probeFailed")}</em>
                          </summary>
                          <div className="digital-probe-command-body">
                            <dl>
                              <div>
                                <dt>{t("settings.digital.probeCommand")}</dt>
                                <dd><code>{command.commandHex}</code></dd>
                              </div>
                              {command.header && (
                                <div>
                                  <dt>{t("settings.digital.probeHeader")}</dt>
                                  <dd><code>{command.header}</code></dd>
                                </div>
                              )}
                              {command.payloadHex && (
                                <div>
                                  <dt>{t("settings.digital.probePayload")}</dt>
                                  <dd><code>{command.payloadHex}</code></dd>
                                </div>
                              )}
                              {command.responseHex && (
                                <div>
                                  <dt>{t("settings.digital.probeResponse")}</dt>
                                  <dd><code>{command.responseHex}</code></dd>
                                </div>
                              )}
                              {commandFields.map(([key, value]) => (
                                <div key={key}>
                                  <dt>{key}</dt>
                                  <dd>{value}</dd>
                                </div>
                              ))}
                            </dl>
                            {command.error && <p className="digital-probe-error">{command.error}</p>}
                          </div>
                        </details>
                      );
                    })}
                  </div>
                </section>
              )}
            </div>

            <footer className="modal-actions">
              {canProbeProvider(activeDialogProvider) && (
                <button type="button" className="secondary-button" onClick={probeConnection} disabled={!dialogCanTest || busy !== "idle" || !activeDialogProvider || !providerHost(activeDialogProvider)}>
                  <Database size={16} />
                  {busy === "diagnosing" ? t("settings.digital.probing") : t("settings.digital.probe")}
                </button>
              )}
              <button type="button" className="secondary-button" onClick={testConnection} disabled={!dialogCanTest || busy !== "idle" || !activeDialogProvider || !providerHost(activeDialogProvider)}>
                <RefreshCw size={16} />
                {busy === "testing" ? t("settings.digital.testing") : t("settings.digital.test")}
              </button>
              <button type="button" className="primary-button" onClick={saveAdapter} disabled={!dialogCanSave || !activeDialogProvider || !providerHost(activeDialogProvider)}>
                <Save size={16} />
                {t("settings.digital.save")}
              </button>
            </footer>
          </section>
        </div>
      )}
    </>
  );
}
