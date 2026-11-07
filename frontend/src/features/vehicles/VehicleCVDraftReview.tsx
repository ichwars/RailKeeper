import { Save, Trash2 } from "lucide-react";
import type { DecoderConfigurationContext, VehicleCVFile, VehicleCVValue, VehicleCVValueInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { cvIndexLabel } from "./CVPageControls";
import { cvDraftCollision, cvDraftCurrent, type CVDraftChange } from "./cvDraft";
import { decoderProfileLabel } from "./decoderProfileLabel";
import { cvWorkingGroupLabel } from "./cvWorkingGroup";

export type ManualCVDraftState = { changes: CVDraftChange[]; values: VehicleCVValue[]; conflicts: string[]; bindingChanged?: boolean; lastSavedCount?: number };

export function VehicleCVDraftReview({ draft, stored, files, context, disabled, save, discard, discardChange, rebase }: {
  draft: ManualCVDraftState; stored: VehicleCVValue[]; files: VehicleCVFile[]; disabled: boolean;
  context?: DecoderConfigurationContext | null;
  save: () => void; discard: () => void; discardChange: (id: string) => void; rebase: (id: string) => void;
}) {
  const { t } = useI18n();
  const conflicts = new Set(draft.conflicts);
  const empty = t("vehicles.saveState.empty");
  const identity = (value: VehicleCVValueInput) => `CV ${value.cvNumber} · ${decoderProfileLabel(value.decoderProfile || "", t("vehicles.cv.noProfile"))} · ${
    cvIndexLabel(value, t("vehicles.cv.base"))}`;
  const display = (field: string, value: unknown) => field === "sourceFileId" && value
    ? files.find(file => file.id === value)?.originalName || t("vehicles.cv.draft.unavailableSource")
    : value == null || value === "" ? empty : String(value);
  const fields = ["description", "category", "protocol", "sourceFileId", "sourceLine", "sourceExportedAt"] as const;
  return <section className="cv-draft-review" aria-label={t("vehicles.cv.draft.title", { count: draft.changes.length })}>
    <h3>{t("vehicles.cv.draft.title", { count: draft.changes.length })}</h3>
    <p>{t("vehicles.cv.draft.help")}</p>
    {conflicts.size > 0 && <p className="cv-draft-conflict" role="alert">{t(draft.bindingChanged ? "cvScope.draftBindingChanged" : "vehicles.cv.draft.conflict")}</p>}
    <div className="table-wrap cv-draft-table" tabIndex={0} aria-label={t("vehicles.cv.draft.review")}>
      <table><thead><tr><th>{t("vehicles.saveState.stored")}</th><th>{t("vehicles.saveState.local")}</th>
        <th>{t("vehicles.actions")}</th></tr></thead>
        <tbody>{draft.changes.map(change => {
          const current = cvDraftCurrent(change, stored, draft.changes);
          const collision = change.original && cvDraftCollision(change, stored, draft.changes);
          const compare = current || change.original;
          const conflict = conflicts.has(change.id);
          return <tr key={change.id} className={conflict ? "cv-draft-conflict" : undefined}>
            <td><span className="cv-draft-mobile-label" aria-hidden="true">{t("vehicles.saveState.stored")}</span>
              {compare ? <><strong>{identity(compare)}</strong><p>{compare.value}</p></> : t("vehicles.cv.draft.new")}
              {context && change.scope && <small>{t("cvScope.currentAssignment")}: {cvWorkingGroupLabel(context, change.scope.workspaceId, compare?.decoderProfile, t)}</small>}
              {change.original && !current && <small>{t("vehicles.cv.draft.original")}</small>}
              {collision ? <p>{t("vehicles.cv.draft.error.occupied")} {identity(collision)}: {collision.value}</p>
                : conflict && <p>{t(current ? "vehicles.cv.draft.changedElsewhere" : "vehicles.cv.draft.missing")}</p>}</td>
            <td><span className="cv-draft-mobile-label" aria-hidden="true">{t("vehicles.saveState.local")}</span>
              {change.value ? <><strong>{identity(change.value)}</strong><p>{change.value.value}</p>
              {context && change.value.workspaceId && <small>{t("cvScope.currentAssignment")}: {cvWorkingGroupLabel(context, change.value.workspaceId, change.value.decoderProfile, t)}</small>}
              {fields.filter(field => (compare?.[field] ?? "") !== (change.value?.[field] ?? "")).map(field =>
                <small key={field}>{t(`vehicles.cv.draft.field.${field}`)}: {display(field, compare?.[field])}
                  {" → "}{display(field, change.value?.[field])}</small>)}
            </> : t("vehicles.cv.draft.remove")}</td>
            <td className="cv-draft-row-controls"><div className="cv-draft-row-actions">
              {conflict && (current || !change.original) && !collision && <button type="button" className="secondary-button" disabled={disabled}
                onClick={() => rebase(change.id)}>{t("vehicles.cv.draft.rebase")}</button>}
              <button type="button" className="icon-button" disabled={disabled} onClick={() => discardChange(change.id)}
                aria-label={t("vehicles.cv.draft.discardChange")} title={t("vehicles.cv.draft.discardChange")}><Trash2 size={15} /></button>
            </div></td>
          </tr>;
        })}</tbody>
      </table>
    </div>
    <div className="cv-draft-actions">
      <button type="button" className="primary-button" disabled={disabled || conflicts.size > 0} onClick={save}>
        <Save size={15} aria-hidden="true" />{t("vehicles.cv.draft.save")}</button>
      <button type="button" className="secondary-button" disabled={disabled} onClick={discard}>{t("vehicles.cv.draft.discard")}</button>
    </div>
  </section>;
}
