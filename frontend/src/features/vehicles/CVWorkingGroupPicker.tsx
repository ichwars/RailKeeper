import { Check } from "lucide-react";
import { AppSelect } from "../../shared/ui/AppSelect";
import { useI18n } from "../../shared/i18n";
import { decoderProfileLabel } from "./decoderProfileLabel";
import { cvWorkingGroupLabel } from "./cvWorkingGroup";
import type { DecoderConfigurationController } from "./useDecoderConfigurationController";

export function CVWorkingGroupPicker({ controller: model }: { controller: DecoderConfigurationController }) {
  const { t } = useI18n(), { context, activeTarget, requestedWorkspaceId } = model;
  if (model.readonly) return null;
  const requested = context?.workspaces.find(group => group.id === requestedWorkspaceId);
  const disabled = model.busy || model.saving || model.loading;
  const unchanged = (activeTarget?.workspaceId || "") === requestedWorkspaceId && model.targetCurrent;
  return <section className="cv-working-group-picker" aria-label={t("cvScope.title")}>
    {!context && <p role={model.error ? "alert" : "status"}>{t(model.error ? "cvScope.loadFailed" : "cvScope.loading")}</p>}
    {context && model.error && <p role="alert">{t("cvScope.loadFailed")}</p>}
    {context && <>
      <div className="cv-working-group-choice"><label>{t("cvScope.title")}
        <AppSelect aria-label={t("cvScope.title")} value={requestedWorkspaceId} disabled={disabled}
          onChange={event => model.chooseWorkspace(event.target.value)}>
          <option value="">{t("cvScope.fromInput")}</option>
          {requestedWorkspaceId && !requested && <option value={requestedWorkspaceId} disabled>{t("cvScope.unavailable")}</option>}
          {context.workspaces.map(group => <option key={group.id} value={group.id}>
            {decoderProfileLabel(group.profile, t("decoderConfig.noProfile"))}</option>)}
        </AppSelect></label>
        <button type="button" className="secondary-button" disabled={disabled || unchanged || Boolean(requestedWorkspaceId && !requested)}
          onClick={model.activateWorkspace}>{t("cvScope.confirm")}</button>
      </div>
      {requested && !unchanged && <p className="cv-working-assignments">{t("cvScope.requestedAssignment")}: {cvWorkingGroupLabel(context, requested.id, requested.profile, t)}</p>}
      <p className="cv-working-group-current"><Check size={16} aria-hidden="true" />{activeTarget
        ? t("cvScope.active", { profile: decoderProfileLabel(activeTarget.profile, t("decoderConfig.noProfile")) }) : t("cvScope.fromInput")}
        {activeTarget && <small>{t("cvScope.currentAssignment")}: {cvWorkingGroupLabel(context, activeTarget.workspaceId, activeTarget.profile, t)}</small>}</p>
      <p className="cv-working-group-help">{t("cvScope.effect")}</p>
      {!model.targetCurrent && <p role="alert">{t("cvScope.bindingChanged")}</p>}
    </>}
    <button type="button" className="secondary-button" disabled={disabled} onClick={() => void model.load()}>{t("cvScope.reload")}</button>
  </section>;
}
