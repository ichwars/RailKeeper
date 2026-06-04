import { useEffect, useState } from "react";
import { Database, ListChecks, Power, PowerOff, RadioTower, RefreshCw, Save, Server, X } from "lucide-react";
import { api, DigitalCenterConnectionResult, ECoSConnectionResult, ECoSLiveStatus, ECoSRawLocomotive, ECoSRawProbe } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { localSettingKeys, readLocalBool, readLocalSetting } from "./settingsModel";

type BusyState = "idle" | "testing" | "probing" | "starting" | "stopping" | "refreshing";
type DigitalProvider = "ecos" | "z21" | "cs3";

type SettingsDigitalTabProps = {
  canManageUsers: boolean;
  formatDateTime: (value: string) => string;
};

const digitalProviders: DigitalProvider[] = ["ecos", "z21", "cs3"];

export function SettingsDigitalTab({ canManageUsers, formatDateTime }: SettingsDigitalTabProps) {
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
  const [cs3Enabled, setCS3Enabled] = useState(() => readLocalBool(localSettingKeys.digitalCS3Enabled, false));
  const [cs3Host, setCS3Host] = useState(() => readLocalSetting(localSettingKeys.digitalCS3Host, ""));
  const [cs3Port, setCS3Port] = useState(() => readLocalSetting(localSettingKeys.digitalCS3Port, "80"));
  const [busy, setBusy] = useState<BusyState>("idle");
  const [connectionResult, setConnectionResult] = useState<ECoSConnectionResult | DigitalCenterConnectionResult | null>(null);
  const [rawProbe, setRawProbe] = useState<ECoSRawProbe | null>(null);
  const [liveStatus, setLiveStatus] = useState<ECoSLiveStatus | null>(null);
  const [message, setMessage] = useState("");
  const [dialogMessage, setDialogMessage] = useState("");

  const ecosInput = () => ({
    host: ecosHost.trim(),
    port: Number(ecosPort) || 15471
  });

  const providerEnabled = (providerId: DigitalProvider) => {
    if (providerId === "ecos") return ecosEnabled;
    if (providerId === "z21") return z21Enabled;
    return cs3Enabled;
  };

  const providerHost = (providerId: DigitalProvider) => {
    if (providerId === "ecos") return ecosHost.trim();
    if (providerId === "z21") return z21Host.trim();
    return cs3Host.trim();
  };

  const providerPort = (providerId: DigitalProvider) => {
    if (providerId === "ecos") return ecosPort.trim() || "15471";
    if (providerId === "z21") return z21Port.trim() || "21105";
    return cs3Port.trim() || "80";
  };

  const providerInput = (providerId: DigitalProvider) => ({
    host: providerHost(providerId),
    port: Number(providerPort(providerId))
  });

  const providerConfigured = (providerId: DigitalProvider) => providerEnabled(providerId) && Boolean(providerHost(providerId));

  const setActiveProviderEnabled = (enabled: boolean) => {
    if (activeDialogProvider === "ecos") setEcosEnabled(enabled);
    if (activeDialogProvider === "z21") setZ21Enabled(enabled);
    if (activeDialogProvider === "cs3") setCS3Enabled(enabled);
  };

  const setActiveProviderHost = (host: string) => {
    if (activeDialogProvider === "ecos") setEcosHost(host);
    if (activeDialogProvider === "z21") setZ21Host(host);
    if (activeDialogProvider === "cs3") setCS3Host(host);
  };

  const setActiveProviderPort = (port: string) => {
    if (activeDialogProvider === "ecos") setEcosPort(port);
    if (activeDialogProvider === "z21") setZ21Port(port);
    if (activeDialogProvider === "cs3") setCS3Port(port);
  };

  const rememberSettings = (nextProvider: DigitalProvider = provider) => {
    window.localStorage.setItem(localSettingKeys.digitalProvider, nextProvider);
    window.localStorage.setItem(localSettingKeys.digitalEcosEnabled, String(ecosEnabled));
    window.localStorage.setItem(localSettingKeys.digitalEcosHost, ecosHost.trim());
    window.localStorage.setItem(localSettingKeys.digitalEcosPort, ecosPort.trim() || "15471");
    window.localStorage.setItem(localSettingKeys.digitalZ21Enabled, String(z21Enabled));
    window.localStorage.setItem(localSettingKeys.digitalZ21Host, z21Host.trim());
    window.localStorage.setItem(localSettingKeys.digitalZ21Port, z21Port.trim() || "21105");
    window.localStorage.setItem(localSettingKeys.digitalCS3Enabled, String(cs3Enabled));
    window.localStorage.setItem(localSettingKeys.digitalCS3Host, cs3Host.trim());
    window.localStorage.setItem(localSettingKeys.digitalCS3Port, cs3Port.trim() || "80");
    window.dispatchEvent(new Event("railkeeper-digital-settings-changed"));
  };

  const refreshLiveStatus = async (nextBusy: BusyState = "refreshing") => {
    if (!canManageUsers) return;
    setBusy(nextBusy);
    setMessage("");
    try {
      const status = await api.getECoSLiveStatus();
      setLiveStatus(status);
      setMessage(status.message);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t("settings.digital.error"));
    } finally {
      setBusy("idle");
    }
  };

  useEffect(() => {
    if (canManageUsers) {
      void refreshLiveStatus("idle");
    }
  }, [canManageUsers]);

  const openAdapter = (nextProvider: DigitalProvider) => {
    setActiveDialogProvider(nextProvider);
    setDialogMessage("");
  };

  const closeAdapter = () => {
    setActiveDialogProvider(null);
    setDialogMessage("");
  };

  const saveAdapter = () => {
    if (!activeDialogProvider) return;
    if (!providerHost(activeDialogProvider)) {
      setDialogMessage(t("settings.digital.hostRequired"));
      return;
    }
    setProvider(activeDialogProvider);
    rememberSettings(activeDialogProvider);
    setMessage(t("settings.digital.saved"));
    closeAdapter();
  };

  const testConnection = async () => {
    const testedProvider = activeDialogProvider || provider;
    rememberSettings(testedProvider);
    setBusy("testing");
    setMessage("");
    setDialogMessage("");
    try {
      const result = testedProvider === "ecos"
        ? await api.testECoSConnection(providerInput("ecos"))
        : testedProvider === "z21"
          ? await api.testZ21Connection(providerInput("z21"))
          : await api.testCS3Connection(providerInput("cs3"));
      setConnectionResult(result);
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

  const probeLocomotives = async () => {
    rememberSettings("ecos");
    setBusy("probing");
    setMessage("");
    try {
      const probe = await api.probeECoSLocomotiveRaw(ecosInput());
      setRawProbe(probe);
      setMessage(t("settings.digital.locomotiveProbeDone", { count: probe.locomotives.length }));
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t("settings.digital.error"));
    } finally {
      setBusy("idle");
    }
  };

  const startLive = async () => {
    rememberSettings("ecos");
    setBusy("starting");
    setMessage("");
    try {
      const status = await api.startECoSLive(ecosInput());
      setLiveStatus(status);
      setMessage(status.message);
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
  const providerReady = providerConfigured(provider);
  const ecosReady = provider === "ecos" && providerConfigured("ecos");
  const liveSince = liveStatus?.startedAt ? formatDateTime(liveStatus.startedAt) : t("settings.digital.notActive");
  const liveSeen = liveStatus?.lastSeenAt ? formatDateTime(liveStatus.lastSeenAt) : t("settings.digital.notSeen");
  const previewLocomotives = rawProbe?.locomotives.slice(0, 8) ?? [];
  const hiddenLocomotiveCount = rawProbe ? Math.max(rawProbe.locomotives.length - previewLocomotives.length, 0) : 0;
  const connectionDetail = connectionResult
    ? ("applicationVersion" in connectionResult && connectionResult.applicationVersion) || connectionResult.fields?.version || connectionResult.fields?.serialNumber || connectionResult.message
    : t("settings.digital.noTest");
  const dialogCanTest = canManageUsers && Boolean(activeDialogProvider);
  const dialogCanSave = Boolean(activeDialogProvider);

  const formatLocomotiveFunctions = (locomotive: ECoSRawLocomotive) => {
    const count = locomotive.numberOfFunctions ?? locomotive.functions?.length ?? 0;
    return count > 0 ? String(count) : "-";
  };
  const formatLocomotiveCVs = (locomotive: ECoSRawLocomotive) => {
    const values = locomotive.cvs?.slice(0, 4).map((cv) => `${cv.number}=${cv.value}`) ?? [];
    if (values.length === 0) return "-";
    return locomotive.cvs && locomotive.cvs.length > values.length ? `${values.join(", ")} ...` : values.join(", ");
  };

  return (
    <>
      <section className="digital-settings-grid">
        <section className="panel settings-card settings-tool-card digital-command-card">
          <div className="settings-section-head">
            <div className="settings-card-title">
              <RadioTower size={18} />
              <div>
                <h2>{t("settings.digital.title")}</h2>
                <p>{t("settings.digital.subtitle")}</p>
              </div>
            </div>
            <div className="settings-marker-row">
              <span className={providerReady ? "settings-pill active" : "settings-pill muted"}>
                {providerReady ? t("settings.digital.active") : t("settings.digital.inactive")}
              </span>
              <span className={liveStatus?.connected ? "settings-pill active" : "settings-pill muted"}>
                {liveStatus?.connected ? t("settings.digital.liveActive") : t("settings.digital.liveInactive")}
              </span>
            </div>
          </div>
          <div className="digital-provider-note">
            <Server size={17} />
            <span>{t("settings.digital.importVisibilityNote")}</span>
          </div>

          {!canManageUsers && (
            <div className="current-user-card">
              <strong>{t("settings.users.adminRequired")}</strong>
              <span>{t("settings.digital.adminHelp")}</span>
            </div>
          )}

          <div className="digital-overview-grid">
            <article className={providerReady ? "digital-active-provider ready" : "digital-active-provider"}>
              <span className="settings-pill">{t("settings.digital.configuredTitle")}</span>
              <strong>{providerLabel(provider)}</strong>
              <small>
                {providerReady
                  ? t("settings.digital.configuredEndpoint", { host: providerHost(provider), port: providerPort(provider) })
                  : t("settings.digital.notConfigured")}
              </small>
            </article>
            <article className="digital-active-provider">
              <span className="settings-pill">{t("settings.digital.connection")}</span>
              <strong>{connectionResult?.connected ? t("settings.digital.connected") : t("settings.digital.notConnected")}</strong>
              <small>{connectionDetail}</small>
            </article>
            <article className="digital-active-provider">
              <span className="settings-pill">{t("settings.digital.locomotives")}</span>
              <strong>{rawProbe?.locomotives.length ?? 0}</strong>
              <small>{rawProbe ? rawProbe.message : t("settings.digital.noLocomotiveProbe")}</small>
            </article>
          </div>

          <div className="settings-action-row digital-action-row">
            <button type="button" className="secondary-button" onClick={() => openAdapter(provider)}>
              <Server size={16} />
              {t("settings.digital.configure")}
            </button>
            <button type="button" className="secondary-button" onClick={probeLocomotives} disabled={!canManageUsers || busy !== "idle" || !ecosReady}>
              <ListChecks size={16} />
              {busy === "probing" ? t("settings.digital.fetchingLocomotives") : t("settings.digital.fetchLocomotives")}
            </button>
            <button type="button" className="primary-button" onClick={startLive} disabled={!canManageUsers || busy !== "idle" || !ecosReady || liveStatus?.connected}>
              <Power size={16} />
              {busy === "starting" ? t("settings.digital.starting") : t("settings.digital.startLive")}
            </button>
            <button type="button" className="secondary-button" onClick={stopLive} disabled={!canManageUsers || busy !== "idle" || !liveStatus?.connected}>
              <PowerOff size={16} />
              {busy === "stopping" ? t("settings.digital.stopping") : t("settings.digital.stopLive")}
            </button>
          </div>

          {message && <p className="form-message">{message}</p>}

          {rawProbe && (
            <section className="digital-loco-preview" aria-live="polite">
              <div className="digital-loco-preview-head">
                <div>
                  <strong>{t("settings.digital.locomotivePreviewTitle")}</strong>
                  <span>{t("settings.digital.locomotivePreviewSubtitle", { host: rawProbe.host, port: rawProbe.port })}</span>
                </div>
                <span className="settings-pill active">{t("settings.digital.locomotiveCount", { count: rawProbe.locomotives.length })}</span>
              </div>
              {previewLocomotives.length > 0 ? (
                <>
                  <div className="digital-loco-table">
                    <table>
                      <thead>
                        <tr>
                          <th>{t("settings.digital.objectId")}</th>
                          <th>{t("settings.digital.name")}</th>
                          <th>{t("settings.digital.address")}</th>
                          <th>{t("settings.digital.protocol")}</th>
                          <th>{t("settings.digital.functions")}</th>
                          <th>{t("settings.digital.cvs")}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {previewLocomotives.map((locomotive) => (
                          <tr key={locomotive.objectId}>
                            <td>{locomotive.objectId}</td>
                            <td>{locomotive.name || "-"}</td>
                            <td>{locomotive.address ?? "-"}</td>
                            <td>{locomotive.protocol || "-"}</td>
                            <td>{formatLocomotiveFunctions(locomotive)}</td>
                            <td>{formatLocomotiveCVs(locomotive)}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  {hiddenLocomotiveCount > 0 && (
                    <p className="digital-loco-more">{t("settings.digital.moreLocomotives", { count: hiddenLocomotiveCount })}</p>
                  )}
                </>
              ) : (
                <p className="digital-loco-empty">{t("settings.digital.noLocomotives")}</p>
              )}
            </section>
          )}
        </section>

        <aside className="settings-card-stack">
          <section className="panel settings-card settings-tool-card">
            <div className="settings-card-title">
              <Database size={18} />
              <div>
                <h2>{t("settings.digital.statusTitle")}</h2>
                <p>{t("settings.digital.statusSubtitle")}</p>
              </div>
            </div>
            <div className="digital-status-grid">
              <article>
                <span className="settings-pill">{t("settings.digital.live")}</span>
                <strong>{liveStatus?.connected ? t("settings.digital.liveActive") : t("settings.digital.liveInactive")}</strong>
                <small>{t("settings.digital.liveSince", { value: liveSince })}</small>
              </article>
              <article>
                <span className="settings-pill">{t("settings.digital.blocks")}</span>
                <strong>{liveStatus?.blocksReceived || 0}</strong>
                <small>{t("settings.digital.lastSeen", { value: liveSeen })}</small>
              </article>
            </div>
          </section>

          <section className="panel settings-card settings-tool-card">
            <div className="settings-card-title">
              <Server size={18} />
              <div>
                <h2>{t("settings.digital.roadmapTitle")}</h2>
                <p>{t("settings.digital.roadmapSubtitle")}</p>
              </div>
            </div>
            <div className="digital-roadmap-list">
              {digitalProviders.map((providerId) => (
                <button
                  type="button"
                  key={providerId}
                  className={`${providerId === provider ? "active" : ""} ${providerId !== "ecos" ? "prepared" : ""}`.trim()}
                  onClick={() => openAdapter(providerId)}
                >
                  <strong>
                    {providerLabel(providerId)}
                    <span className={providerConfigured(providerId) ? "work-status-badge compact active" : "work-status-badge compact open"}>
                      {providerConfigured(providerId) ? t("settings.digital.active") : t("settings.work.open")}
                    </span>
                  </strong>
                  <span>{providerId === "ecos" ? t("settings.digital.ecosReady") : t("settings.digital.adapterReady")}</span>
                </button>
              ))}
            </div>
          </section>
        </aside>
      </section>

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
              <label className="settings-toggle-row">
                <input type="checkbox" checked={activeDialogProvider ? providerEnabled(activeDialogProvider) : false} onChange={(event) => setActiveProviderEnabled(event.target.checked)} />
                <span>
                  <strong>{activeDialogProvider ? t(`settings.digital.enable.${activeDialogProvider}`) : t("settings.digital.configure")}</strong>
                  <small>{activeDialogProvider ? t(`settings.digital.enable.${activeDialogProvider}.help`) : t("settings.digital.dialogSubtitle")}</small>
                </span>
              </label>
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
                    placeholder={activeDialogProvider === "z21" ? "21105" : activeDialogProvider === "cs3" ? "80" : "15471"}
                  />
                </label>
              </div>

              {dialogMessage && <p className="form-message">{dialogMessage}</p>}
            </div>

            <footer className="modal-actions">
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
