import { useState } from "react";
import { Cpu, RefreshCw } from "lucide-react";
import { useI18n } from "../../shared/i18n";
import { AppTextInput } from "../../shared/ui/AppTextInput";
import { DecoderConfigurationForm } from "./DecoderConfigurationForm";
import { decoderProfileLabel } from "./decoderProfileLabel";
import type { DecoderConfigurationController } from "./useDecoderConfigurationController";

export function DecoderConfigurationPanel({ controller: model }: { controller: DecoderConfigurationController }) {
  const { t } = useI18n(), { context } = model;
  const [search, setSearch] = useState(""), [page, setPage] = useState(0);
  const disabled = model.busy || model.loading || model.saving, locked = disabled || Boolean(model.draft);
  const instances = context?.instances.filter(value => `${value.label} ${value.manufacturer} ${value.model} ${value.serialNumber} ${value.id}`
    .toLowerCase().includes(search.toLowerCase())) || [];
  const currentPage = Math.min(page, Math.max(0, Math.ceil(instances.length / 50) - 1));
  const installations = context?.installations.filter(value => value.vehicleId === context.vehicleId || value.vehicleReferenceId === context.vehicleId) || [];
  return <section className="decoder-configurations" aria-label={t("decoderManage.title")} onKeyDown={event => {
    if (event.key === "Enter" && event.target instanceof HTMLInputElement && !event.nativeEvent.isComposing) event.preventDefault();
  }}>
    <p className="decoder-configurations-intro">{t("decoderManage.intro")}</p>
    <div className="decoder-configurations-toolbar">
      <button type="button" className="secondary-button" disabled={disabled} onClick={() => void model.load()}>
        <RefreshCw size={16} aria-hidden="true" />{t("decoderConfig.reload")}</button>
      {!model.readonly && <>
        <button type="button" className="secondary-button" disabled={locked || !context} onClick={() => model.start("createInstance")}>{t("decoderManage.newInstance")}</button>
        <button type="button" className="secondary-button" disabled={locked || !context} onClick={() => model.start("workspace")}>{t("decoderManage.newWorkspace")}</button>
      </>}
    </div>
    {model.loading && <p role="status">{t("decoderManage.loading")}</p>}
    {model.error && <p className="error-text" role="alert">{model.error}</p>}
    {model.status && <p role="status">{model.status}</p>}
    <DecoderConfigurationForm controller={model} />
    {context && <>
      <section className="decoder-installations"><h4><Cpu size={18} aria-hidden="true" />{t("decoderManage.installations")}</h4>
        {!installations.length && <p>{t("decoderManage.noInstallations")}</p>}
        {installations.map(item => { const decoder = context.instances.find(value => value.id === item.decoderId);
          return <div className="decoder-installation" key={item.id}><div>
            <strong>{decoder?.label || t("decoderConfig.unknown")}</strong>
            <small>{[decoder?.manufacturer, decoder?.model, decoder?.serialNumber].filter(Boolean).join(" · ")}</small>
            <p>{t(`decoderConfig.state.${item.state}`)}</p>
            <dl><div><dt>{t("decoderConfig.installedAt")}</dt><dd>{item.installedAt || t("decoderConfig.unknown")}</dd></div>
              <div><dt>{t("decoderConfig.removedAt")}</dt><dd>{item.removedAt || t("decoderConfig.unknown")}</dd></div></dl>
            {item.notes && <p className="decoder-record-notes">{item.notes}</p>}
          </div>{!model.readonly && <div className="decoder-record-actions">
            <button type="button" className="secondary-button" disabled={locked} onClick={() => model.start("editInstance", item.decoderId)}>{t("decoderManage.editInstance")}</button>
            <button type="button" className="secondary-button" disabled={locked} onClick={() => model.start("dates", item.id)}>{t("decoderManage.correctDates")}</button>
            {item.state === "installed" && <button type="button" className="secondary-button" disabled={locked} onClick={() => model.start("remove", item.id)}>{t("decoderManage.remove")}</button>}
          </div>}</div>;
        })}
      </section>
      <section className="decoder-working-groups"><h4>{t("decoderManage.workspaces")}</h4>
        <p>{t("decoderManage.workspaceHelp")}</p>
        {!context.workspaces.length && <p>{t("decoderManage.noWorkspaces")}</p>}
        {context.workspaces.map(item => {
          const installed = context.installations.find(value => value.id === item.installationId);
          const decoder = context.instances.find(value => value.id === installed?.decoderId);
          return <div className="decoder-working-group" key={item.id}><div>
            <strong>{decoderProfileLabel(item.profile, t("decoderConfig.noProfile"))}</strong>
            <p>{t("decoderManage.workspaceCount", { values: item.valueCount, files: item.sourceCount })}</p>
            <p>{decoder ? `${decoder.label} · ${t(`decoderConfig.state.${installed!.state}`)}` : t("decoderManage.unassigned")}</p>
          </div>{!model.readonly && <div className="decoder-record-actions">
            <button type="button" className="secondary-button" disabled={locked} onClick={() => model.start("assign", item.id)}>{t("decoderManage.assign")}</button>
            <button type="button" className="secondary-button" disabled={locked} onClick={() => model.start("checkpoint", item.id)}>{t("decoderManage.checkpoint")}</button>
          </div>}</div>;
        })}
      </section>
      <details className="decoder-instance-pool"><summary>{t("decoderManage.instances")}</summary>
        <AppTextInput label={t("decoderManage.searchInstances")} value={search} onChange={event => { setSearch(event.target.value); setPage(0); }} />
        <div className="decoder-instance-pages"><span>{t("decoderManage.instanceCount", { count: instances.length })}</span>
          <button type="button" className="secondary-button" disabled={currentPage === 0} onClick={() => setPage(currentPage - 1)}>{t("vehicles.cv.previous")}</button>
          <button type="button" className="secondary-button" disabled={(currentPage + 1) * 50 >= instances.length} onClick={() => setPage(currentPage + 1)}>{t("vehicles.cv.next")}</button></div>
        {instances.slice(currentPage * 50, (currentPage + 1) * 50).map(item => {
          const installation = context.installations.find(value => value.decoderId === item.id && value.state === "installed");
          return <div className="decoder-instance" key={item.id}><div><strong>{item.label}</strong>
            <small>{[item.manufacturer, item.model, item.serialNumber].filter(Boolean).join(" · ")}</small>
            <small className="decoder-record-id">{t("decoderManage.identifier")}: {item.id}</small>
            {item.notes && <p className="decoder-record-notes">{item.notes}</p>}
            <p>{installation ? t("decoderManage.installedIn", { vehicle: installation.vehicleName || installation.vehicleInventoryNumber }) : t("decoderManage.available")}</p>
          </div>{!model.readonly && <div className="decoder-record-actions">
            <button type="button" className="secondary-button" disabled={locked} onClick={() => model.start("editInstance", item.id)}>{t("decoderManage.editInstance")}</button>
            {!installation && <button type="button" className="secondary-button" disabled={locked} onClick={() => model.start("install", item.id)}>{t("decoderManage.install")}</button>}
          </div>}</div>;
        })}
      </details>
    </>}
  </section>;
}
