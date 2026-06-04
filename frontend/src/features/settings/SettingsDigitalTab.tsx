import { useEffect, useState } from "react";
import { Database, ListChecks, Power, PowerOff, RadioTower, RefreshCw, Save, Server, X } from "lucide-react";
import { api, ECoSConnectionResult, ECoSLiveStatus, ECoSRawLocomotive, ECoSRawProbe } from "../../shared/api";
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
  const [busy, setBusy] = useState<BusyState>("idle");
  const [connectionResult, setConnectionResult] = useState<ECoSConnectionResult | null>(null);
  const [rawProbe, setRawProbe] = useState<ECoSRawProbe | null>(null);
  const [liveStatus, setLiveStatus] = useState<ECoSLiveStatus | null>(null);
  const [message, setMessage] = useState("");
  const [dialogMessage, setDialogMessage] = useState("");

  const ecosInput = () => ({
    host: ecosHost.trim(),
    port: Number(ecosPort) || 15471
  });

  const rememberSettings = (nextProvider: DigitalProvider = provider) => {
    window.localStorage.setItem(localSettingKeys.digitalProvider, nextProvider);
    window.localStorage.setItem(localSettingKeys.digitalEcosEnabled, String(ecosEnabled));
    window.localStorage.setItem(localSettingKeys.digitalEcosHost, ecosHost.trim());
    window.localStorage.setItem(localSettingKeys.digitalEcosPort, ecosPort.trim() || "15471");
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
    if (!activeDialogProvider || activeDialogProvider !== "ecos") return;
    if (!ecosHost.trim()) {
      setDialogMessage(t("settings.digital.hostRequired"));
      return;
    }
    setProvider(activeDialogProvider);
    window.localStorage.setItem(localSettingKeys.digitalEcosEnabled, String(ecosEnabled));
    rememberSettings(activeDialogProvider);
    setMessage(t("settings.digital.saved"));
    closeAdapter();
  };

  const testConnection = async () => {
    rememberSettings("ecos");
    setBusy("testing");
    setMessage("");
    setDialogMessage("");
    try {
      const result = await api.testECoSConnection(ecosInput());
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
  const providerReady = provider === "ecos" && ecosEnabled && Boolean(ecosHost.trim());
  const liveSince = liveStatus?.startedAt ? formatDateTime(liveStatus.startedAt) : t("settings.digital.notActive");
  const liveSeen = liveStatus?.lastSeenAt ? formatDateTime(liveStatus.lastSeenAt) : t("settings.digital.notSeen");
  const previewLocomotives = rawProbe?.locomotives.slice(0, 8) ?? [];
  const hiddenLocomotiveCount = rawProbe ? Math.max(rawProbe.locomotives.length - previewLocomotives.length, 0) : 0;
  const dialogProviderActive = activeDialogProvider === "ecos";
  const dialogCanTest = canManageUsers && dialogProviderActive;
  const dialogCanSave = dialogProviderActive;

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
                  ? t("settings.digital.configuredEndpoint", { host: ecosHost.trim(), port: ecosPort.trim() || "15471" })
                  : t("settings.digital.notConfigured")}
              </small>
            </article>
            <article className="digital-active-provider">
              <span className="settings-pill">{t("settings.digital.connection")}</span>
              <strong>{connectionResult?.connected ? t("settings.digital.connected") : t("settings.digital.notConnected")}</strong>
              <small>{connectionResult?.applicationVersion || connectionResult?.message || t("settings.digital.noTest")}</small>
            </article>
            <article className="digital-active-provider">
              <span className="settings-pill">{t("settings.digital.locomotives")}</span>
              <strong>{rawProbe?.locomotives.length ?? 0}</strong>
              <small>{rawProbe ? rawProbe.message : t("settings.digital.noLocomotiveProbe")}</small>
            </article>
          </div>

          <div className="settings-action-row digital-action-row">
            <button type="button" className="secondary-button" onClick={() => openAdapter("ecos")}>
              <Server size={16} />
              {t("settings.digital.configure")}
            </button>
            <button type="button" className="secondary-button" onClick={probeLocomotives} disabled={!canManageUsers || busy !== "idle" || !providerReady}>
              <ListChecks size={16} />
              {busy === "probing" ? t("settings.digital.fetchingLocomotives") : t("settings.digital.fetchLocomotives")}
            </button>
            <button type="button" className="primary-button" onClick={startLive} disabled={!canManageUsers || busy !== "idle" || !providerReady || liveStatus?.connected}>
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
                    <span className={providerId === "ecos" && providerReady ? "work-status-badge compact active" : "work-status-badge compact open"}>
                      {providerId === "ecos" && providerReady ? t("settings.digital.active") : t("settings.work.open")}
                    </span>
                  </strong>
                  <span>{providerId === "ecos" ? t("settings.digital.ecosReady") : t("settings.digital.futureProvider")}</span>
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
              {!dialogProviderActive ? (
                <div className="digital-provider-note">
                  <Server size={17} />
                  <span>{t("settings.digital.futureProvider")} {t("settings.work.futureProviderNote")}</span>
                </div>
              ) : (
                <>
                  <label className="settings-toggle-row">
                    <input type="checkbox" checked={ecosEnabled} onChange={(event) => setEcosEnabled(event.target.checked)} />
                    <span>
                      <strong>{t("settings.digital.enableEcos")}</strong>
                      <small>{t("settings.digital.enableEcosHelp")}</small>
                    </span>
                  </label>
                  <div className="settings-field-grid">
                    <label>
                      {t("settings.digital.ecos.host")}
                      <input value={ecosHost} onChange={(event) => setEcosHost(event.target.value)} placeholder={t("settings.digital.ecos.hostPlaceholder")} />
                    </label>
                    <label>
                      {t("settings.digital.ecos.port")}
                      <input value={ecosPort} onChange={(event) => setEcosPort(event.target.value)} inputMode="numeric" placeholder="15471" />
                    </label>
                  </div>
                </>
              )}

              {dialogMessage && <p className="form-message">{dialogMessage}</p>}
            </div>

            <footer className="modal-actions">
              <button type="button" className="secondary-button" onClick={testConnection} disabled={!dialogCanTest || busy !== "idle" || !ecosHost.trim()}>
                <RefreshCw size={16} />
                {busy === "testing" ? t("settings.digital.testing") : t("settings.digital.test")}
              </button>
              <button type="button" className="primary-button" onClick={saveAdapter} disabled={!dialogCanSave || !ecosHost.trim()}>
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
