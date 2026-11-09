import { useEffect, useEffectEvent, useState } from "react";
import { DigitalCenterConnectionResult, DigitalCenterProbeResult, DigitalCenterSettings, ECoSConnectionResult, ECoSLiveStatus, ECoSLocomotiveSummary, api } from "../../shared/api";
import { observeAction } from "../../shared/asyncAction";
import { useI18n } from "../../shared/i18n";
import { DigitalAdapterDialog } from "./DigitalAdapterDialog";
import { DigitalCenterWorkflowView, type DigitalDiagnosticsTab, type DigitalProvider } from "./DigitalCenterWorkflowView";
import { digitalProviders, useDigitalProviderSettings } from "./useDigitalProviderSettings";

export type BusyState = "idle" | "saving" | "testing" | "diagnosing" | "probing" | "starting" | "stopping" | "refreshing" | "activating" | "deactivating" | "removing";
export type ConfigurationMeta = Partial<Record<DigitalProvider, { configuredAt: string; configuredBy: string }>>;

export type SettingsDigitalTabProps = {
  canManageUsers: boolean;
  formatDateTime: (value: string) => string;
  username: string;
};
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
  const {
    provider, setProvider, activeDialogProvider, setActiveDialogProvider, ecosEnabled,
    setEcosEnabled, ecosHost, setEcosHost, ecosPort, setEcosPort, z21Enabled, setZ21Enabled,
    z21Host, setZ21Host, z21Port, setZ21Port, intellibox3Enabled, setIntellibox3Enabled,
    intellibox3Host, setIntellibox3Host, intellibox3Port, setIntellibox3Port, cs3Enabled,
    setCS3Enabled, cs3Host, setCS3Host, cs3Port, setCS3Port, ecosInput, providerEnabled,
    providerHost, providerPort, providerPortDraft, providerInput, providerConfigured,
    providerActive, currentDigitalSettings, applyDigitalSettings, hasDigitalConfig,
    setActiveProviderHost, setActiveProviderPort, storeSettingsLocally,
  } = useDigitalProviderSettings();

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

  const loadDigitalOnEntry = useEffectEvent(() => {
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
      observeAction(refreshLiveStatus("idle", false));
    }
  });
  useEffect(() => { loadDigitalOnEntry(); }, [canManageUsers]);

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

  const saveAdapter = async () => {
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
    setBusy("saving");
    setDialogMessage("");
    try {
      await persistSettings(currentDigitalSettings(savedProvider));
      setConfigurationMeta(nextMeta);
      window.localStorage.setItem(configurationMetaKey, JSON.stringify(nextMeta));
      setLocomotiveSummary(null);
      setMessage(t("settings.digital.saved"));
      closeAdapter();
    } catch (error) {
      setDialogMessage(error instanceof Error ? error.message : t("settings.digital.error"));
    } finally {
      setBusy("idle");
    }
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

  const canProbeProvider = (providerId: DigitalProvider | null) =>
    providerId === "z21" || providerId === "intellibox3" || providerId === "cs3";

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
        : probedProvider === "intellibox3"
          ? await api.probeIntellibox3Connection(providerInput("intellibox3"))
          : await api.probeCS3Connection(providerInput("cs3"));
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
        observeAction(countLocomotives(true));
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

  const countConnectedLocomotives = useEffectEvent(() => { observeAction(countLocomotives(true)); });
  useEffect(() => {
    if (!liveStatus?.connected || !ecosReady || locomotiveSummary || busy !== "idle") return;
    countConnectedLocomotives();
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
        <DigitalAdapterDialog
              canManageUsers={canManageUsers}
              t={t}
              providerLabel={providerLabel}
              activeDialogProvider={activeDialogProvider}
              busy={busy}
              closeAdapter={closeAdapter}
              providerHost={providerHost}
              setActiveProviderHost={setActiveProviderHost}
              providerPortDraft={providerPortDraft}
              setActiveProviderPort={setActiveProviderPort}
              dialogMessage={dialogMessage}
              probeResult={probeResult}
              probeFieldEntries={probeFieldEntries}
              canProbeProvider={canProbeProvider}
              probeConnection={probeConnection}
              dialogCanTest={dialogCanTest}
              testConnection={testConnection}
              saveAdapter={saveAdapter}
              dialogCanSave={dialogCanSave}
            />
      )}
    </>
  );
}
