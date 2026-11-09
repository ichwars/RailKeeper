import { Database, RefreshCw, Save, X } from "lucide-react";
import { DigitalCenterProbeResult } from "../../shared/api";
import { observeAction } from "../../shared/asyncAction";
import { type DigitalProvider } from "./DigitalCenterWorkflowView";
import type { BusyState } from "./SettingsDigitalTab";

type DigitalAdapterDialogProps = {
  canManageUsers: boolean;
  t: (key: string, values?: Record<string, string | number>) => string;
  providerLabel: (providerId: DigitalProvider) => string;
  activeDialogProvider: NonNullable<DigitalProvider | null>;
  busy: BusyState;
  closeAdapter: () => void;
  providerHost: (providerId: DigitalProvider) => string;
  setActiveProviderHost: (host: string) => void;
  providerPortDraft: (providerId: DigitalProvider) => string;
  setActiveProviderPort: (port: string) => void;
  dialogMessage: string;
  probeResult: DigitalCenterProbeResult | null;
  probeFieldEntries: [string, string][];
  canProbeProvider: (providerId: DigitalProvider | null) => providerId is "z21" | "intellibox3" | "cs3";
  probeConnection: () => Promise<void>;
  dialogCanTest: boolean;
  testConnection: () => Promise<void>;
  saveAdapter: () => Promise<void>;
  dialogCanSave: boolean;
};

export function DigitalAdapterDialog({
  canManageUsers,
  t, providerLabel, activeDialogProvider, busy, closeAdapter, providerHost, setActiveProviderHost,
  providerPortDraft, setActiveProviderPort, dialogMessage, probeResult, probeFieldEntries,
  canProbeProvider, probeConnection, dialogCanTest, testConnection, saveAdapter, dialogCanSave,
}: DigitalAdapterDialogProps) {
  return (
    <div className="modal-layer digital-adapter-layer" role="dialog" aria-modal="true" aria-label={t("settings.digital.dialogTitle", { provider: providerLabel(activeDialogProvider) })}>
      <section className="vehicle-modal digital-adapter-dialog">
        <header className="modal-head">
          <div>
            <h2>{t("settings.digital.dialogTitle", { provider: providerLabel(activeDialogProvider) })}</h2>
            <p>{t("settings.digital.dialogSubtitle")}</p>
          </div>
          <button type="button" className="icon-button" disabled={busy !== "idle"} onClick={closeAdapter} aria-label={t("vehicles.close")} title={t("vehicles.close")}>
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
                disabled={busy !== "idle"}
                onChange={(event) => setActiveProviderHost(event.target.value)}
                placeholder={activeDialogProvider ? t(`settings.digital.${activeDialogProvider}.hostPlaceholder`) : ""}
              />
            </label>
            <label>
              {activeDialogProvider ? t(`settings.digital.${activeDialogProvider}.port`) : t("settings.digital.ecos.port")}
              <input
                value={activeDialogProvider ? providerPortDraft(activeDialogProvider) : ""}
                disabled={busy !== "idle"}
                onChange={(event) => setActiveProviderPort(event.target.value)}
                inputMode="numeric"
                placeholder={activeDialogProvider === "z21" || activeDialogProvider === "intellibox3" ? "21105" : activeDialogProvider === "cs3" ? "80" : "15471"}
              />
            </label>
          </div>

          {dialogMessage && <p className="form-message" role="alert">{dialogMessage}</p>}

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
                          {command.request && (
                            <div>
                              <dt>{t("settings.digital.probeRequest")}</dt>
                              <dd><code>{command.request}</code></dd>
                            </div>
                          )}
                          {command.commandHex && (
                            <div>
                              <dt>{t("settings.digital.probeCommand")}</dt>
                              <dd><code>{command.commandHex}</code></dd>
                            </div>
                          )}
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
          <button type="button" className="primary-button" onClick={() => observeAction(saveAdapter())} disabled={busy !== "idle" || !dialogCanSave || !activeDialogProvider || !providerHost(activeDialogProvider)}>
            <Save size={16} />
            {t(busy === "saving" ? "common.saving" : "settings.digital.save")}
          </button>
        </footer>
      </section>
    </div>
  );
}
