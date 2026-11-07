import { AppTextInput } from "../../shared/ui/AppTextInput";
import { AppTextArea } from "../../shared/ui/AppTextArea";
import { AppSelect } from "../../shared/ui/AppSelect";
import { useI18n } from "../../shared/i18n";
import { decoderDraftValid } from "./decoderConfigurationDraft";
import { decoderProfileLabel } from "./decoderProfileLabel";
import type { DecoderConfigurationController } from "./useDecoderConfigurationController";
import { useEffect, useRef } from "react";

export function DecoderConfigurationForm({ controller: model }: { controller: DecoderConfigurationController }) {
  const { t } = useI18n(), { context, draft } = model;
  const section = useRef<HTMLElement>(null);
  useEffect(() => {
    if (!draft) return;
    const invoker = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const container = section.current;
    container?.scrollIntoView?.({ block: "nearest" });
    container?.querySelector<HTMLElement>("input, textarea, button")?.focus();
    return () => {
      if (document.activeElement === document.body || container?.contains(document.activeElement)) {
        queueMicrotask(() => { if (invoker?.isConnected) invoker.focus(); });
      }
    };
  }, [draft?.id, draft?.kind]);
  if (!context || !draft) return null;
  const disabled = model.busy || model.loading || model.saving || model.readonly;
  const valid = decoderDraftValid(draft, context), instance = context.instances.find(value => value.id === draft.decoderId);
  const installation = context.installations.find(value => value.id === draft.targetId);
  const installed = context.installations.filter(value => value.vehicleId === context.vehicleId && value.state === "installed");
  const workspace = context.workspaces.find(value => value.id === draft.targetId);
  const instanceFields = draft.kind === "createInstance" || draft.kind === "editInstance";
  const dateField = (field: "installedAt" | "removedAt") => <AppTextInput label={t(`decoderConfig.${field}`)}
    value={draft[field]} onChange={event => model.update({ [field]: event.target.value })} disabled={disabled}
    placeholder={t("decoderManage.dateFormat")} helpText={t("decoderManage.dateHelp")} maxLength={10} inputMode="numeric" />;
  return <section className="decoder-command" ref={section} aria-label={t(`decoderManage.title.${draft.kind}`)}>
    <h4>{t(`decoderManage.title.${draft.kind}`)}</h4>
    <p>{t(`decoderManage.effect.${draft.kind}`)}</p>
    {workspace && <p><strong>{decoderProfileLabel(workspace.profile, t("decoderConfig.noProfile"))}</strong></p>}
    {(draft.kind === "remove" || draft.kind === "dates") && installation && <p>
      <strong>{context.instances.find(value => value.id === installation.decoderId)?.label}</strong>
      {` · ${t(`decoderConfig.state.${installation.state}`)}`}
    </p>}
    <div className="decoder-command-fields">
      {instanceFields && <>
        <AppTextInput label={<>{t("decoderManage.label")}<span aria-hidden="true"> *</span></>} value={draft.label} maxLength={160} disabled={disabled}
          onChange={event => model.update({ label: event.target.value })} helpText={t("decoderManage.labelHelp")} aria-required="true" />
        <AppTextInput label={t("decoderConfig.manufacturer")} value={draft.manufacturer} maxLength={160} disabled={disabled}
          onChange={event => model.update({ manufacturer: event.target.value })} />
        <AppTextInput label={t("decoderConfig.model")} value={draft.model} maxLength={160} disabled={disabled}
          onChange={event => model.update({ model: event.target.value })} />
        <AppTextInput label={t("decoderConfig.serial")} value={draft.serialNumber} maxLength={160} disabled={disabled}
          onChange={event => model.update({ serialNumber: event.target.value })} />
      </>}
      {draft.kind === "workspace" && <AppTextInput label={<>{t("decoderManage.profile")}<span aria-hidden="true"> *</span></>} value={draft.profile}
        maxLength={160} disabled={disabled} aria-required="true" onChange={event => model.update({ profile: event.target.value })} />}
      {draft.kind === "install" && <>
        <div><strong>{instance?.label}</strong><p className="decoder-record-id">{t("decoderManage.identifier")}: {draft.decoderId}</p></div>
        {dateField("installedAt")}
      </>}
      {draft.kind === "remove" && dateField("removedAt")}
      {draft.kind === "dates" && <>{dateField("installedAt")}
        {installation?.state !== "installed" && dateField("removedAt")}</>}
      {draft.kind === "assign" && <label className="decoder-command-select">{t("decoderManage.assignment")}
        <AppSelect aria-label={t("decoderManage.assignment")} value={draft.installationId} disabled={disabled} onChange={event => model.update({ installationId: event.target.value })}>
          <option value="">{t("decoderManage.unassigned")}</option>
          {draft.installationId && !installed.some(value => value.id === draft.installationId) &&
            <option value={draft.installationId} disabled>{t("decoderManage.previousInstallation")}</option>}
          {installed.map(value => { const decoder = context.instances.find(item => item.id === value.decoderId);
            return <option key={value.id} value={value.id}>{decoder?.label} · {decoder?.serialNumber || value.decoderId}</option>;
          })}
        </AppSelect>
      </label>}
      {(instanceFields || draft.kind === "install") && <AppTextArea label={t("decoderManage.notes")}
        value={draft.notes} maxLength={4000} disabled={disabled} onChange={event => model.update({ notes: event.target.value })} />}
      {(draft.kind === "remove" || draft.kind === "checkpoint") && <AppTextArea label={t(draft.kind === "remove" ? "decoderManage.removalNote" : "decoderConfig.note")}
        value={draft.note} maxLength={2000} disabled={disabled} onChange={event => model.update({ note: event.target.value })} />}
    </div>
    {model.conflicts.length > 0 && <section className="decoder-command-conflicts" aria-label={t("decoderManage.fieldConflicts")}>
      <p role="alert">{t("decoderManage.fieldConflicts")}</p>
      {model.conflicts.map(conflict => {
        const labels: Record<string, string> = { label: "decoderManage.label", manufacturer: "decoderConfig.manufacturer", model: "decoderConfig.model",
          serialNumber: "decoderConfig.serial", notes: "decoderManage.notes", installedAt: "decoderConfig.installedAt",
          removedAt: "decoderConfig.removedAt", installationId: "decoderManage.assignment" };
        const display = (value: unknown) => {
          if (conflict.field !== "installationId") return String(value || t("decoderConfig.unknown"));
          const item = context.installations.find(item => item.id === value);
          return item ? context.instances.find(unit => unit.id === item.decoderId)?.label || t("decoderConfig.unknown") : t("decoderManage.unassigned");
        };
        return <div className="decoder-command-conflict" key={conflict.field}>
          <strong>{t(labels[conflict.field] || "decoderManage.notes")}</strong>
          <dl><div><dt>{t("vehicles.saveState.stored")}</dt><dd>{display(conflict.saved)}</dd></div>
            <div><dt>{t("vehicles.saveState.local")}</dt><dd>{display(conflict.local)}</dd></div></dl>
          <div className="decoder-command-actions">
            <button type="button" className="secondary-button" disabled={disabled} onClick={() => model.resolve(conflict.field, true)}>{t("decoderManage.useSaved")}</button>
            <button type="button" className="secondary-button" disabled={disabled} onClick={() => model.resolve(conflict.field, false)}>{t("decoderManage.keepInput")}</button>
          </div>
        </div>;
      })}
    </section>}
    {(model.needsReview || model.stale) && <p role="alert">{t("decoderManage.reviewRequired")}</p>}
    {!valid && <p className="decoder-command-help">{t("decoderManage.checkFields")}</p>}
    <div className="decoder-command-actions">
      <button type="button" className="secondary-button" disabled={disabled} onClick={() => void model.load(true)}>{t("decoderManage.compare")}</button>
      <button type="button" className="primary-button" disabled={disabled || model.needsReview || model.stale || model.conflicts.length > 0 || !valid}
        onClick={() => void model.confirm()}>{t(`decoderManage.confirm.${draft.kind}`)}</button>
      <button type="button" className="secondary-button" disabled={model.busy || model.saving} onClick={model.discard}>{t("decoderManage.discard")}</button>
    </div>
  </section>;
}
