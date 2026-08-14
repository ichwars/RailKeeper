import { FormEvent, useCallback, useEffect, useState } from "react";
import { Eye, GitBranch, PencilRuler, Plus } from "lucide-react";

import { ApiError, api, type LayoutUnit, type PlanRevision, type PlanVariant } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { LayoutConfirmDialog, type LayoutPendingAction } from "./LayoutConfirmDialog";
import { TrackPlannerCanvas } from "./TrackPlannerCanvas";
import { TrackLibraryPanel } from "./TrackLibraryPanel";

export function LayoutPlansPanel({ units, gauge, canPlan, canManageLibraries = false }: {
  units: LayoutUnit[];
  gauge: string;
  canPlan: boolean;
  canManageLibraries?: boolean;
}) {
  const [unitID, setUnitID] = useState(() => units[0]?.id || "");
  const [variants, setVariants] = useState<PlanVariant[]>([]);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [conflict, setConflict] = useState(false);
  const [pending, setPending] = useState<LayoutPendingAction | null>(null);
  const [openRevision, setOpenRevision] = useState<PlanRevision | null>(null);
  const { t } = useI18n();
  const genericError = t("layouts.error.generic");
  const selectableUnits = units.filter((unit) => !unit.archived || unit.id === unitID);

  const loadVariants = useCallback(async () => {
    if (!unitID) { setVariants([]); return; }
    setLoading(true); setConflict(false); setMessage("");
    try { setVariants(await api.planVariants(unitID)); }
    catch (reason) { setMessage(reason instanceof Error ? reason.message : genericError); }
    finally { setLoading(false); }
  }, [genericError, unitID]);

  useEffect(() => { void loadVariants(); }, [loadVariants]);

  const createVariant = async (event: FormEvent) => {
    event.preventDefault(); if (!unitID) return;
    setSaving(true); setMessage("");
    try {
      await api.createPlanVariant(unitID, { name, description: description.trim() || undefined });
      setName(""); setDescription(""); await loadVariants();
    } catch (reason) { setMessage(reason instanceof Error ? reason.message : genericError); }
    finally { setSaving(false); }
  };

  const runRevisionAction = async (action: () => Promise<PlanRevision>) => {
    setMessage(""); setConflict(false);
    try { await action(); await loadVariants(); }
    catch (reason) {
      if (reason instanceof ApiError && reason.status === 409) {
        setConflict(true); setMessage(t("layouts.conflict.revision"));
      } else setMessage(reason instanceof Error ? reason.message : genericError);
      throw reason;
    }
  };

  const createDraft = async (variant: PlanVariant) => {
    const base = [...variant.revisions].reverse().find((revision) => revision.status === "published");
    await runRevisionAction(() => api.createPlanRevision(variant.id, { baseRevisionId: base?.id }));
  };
  const submit = (revision: PlanRevision) => runRevisionAction(() => api.submitPlanRevision(revision.id, revision.version));
  const askPublish = (variant: PlanVariant, revision: PlanRevision) => setPending({
    title: t("layouts.plans.publishTitle"),
    body: t("layouts.plans.publishBody", { variant: variant.name, revision: revision.revisionNumber }),
    run: () => runRevisionAction(() => api.publishPlanRevision(revision.id, revision.version))
  });

  const selectedUnit = units.find((unit) => unit.id === unitID);
  if (openRevision && selectedUnit) return <TrackPlannerCanvas unit={selectedUnit} gauge={gauge}
    revision={openRevision} canPlan={canPlan} onClose={() => setOpenRevision(null)} />;

  return <section className="layout-plans-stack">
    <TrackLibraryPanel canManage={canManageLibraries} />
    <section className="panel layout-plan-toolbar">
      <div className="panel-title"><GitBranch size={17} /><h3>{t("layouts.plans.title")}</h3></div>
      <label>{t("layouts.plans.unit")}<AppSelect value={unitID}
        onChange={(event) => setUnitID(event.target.value)} disabled={selectableUnits.length === 0}>
        {selectableUnits.length === 0 ? <option value="">{t("layouts.plans.noUnits")}</option> : null}
        {selectableUnits.map((unit) => <option key={unit.id} value={unit.id}>{unit.name}</option>)}
      </AppSelect></label>
    </section>
    {message ? <div className={conflict ? "layout-conflict" : "form-message"}><span>{message}</span>
      {conflict ? <button type="button" className="secondary-button compact-action" onClick={loadVariants}>
        {t("layouts.conflict.reload")}</button> : null}</div> : null}
    {loading ? <section className="panel"><p>{t("layouts.plans.loading")}</p></section> : variants.length === 0 ?
      <section className="panel"><p className="layout-empty">{unitID ? t("layouts.plans.empty") : t("layouts.plans.noUnits")}</p></section> :
      <div className="layout-variant-list">{variants.map((variant) => <section className="panel" key={variant.id}>
        <div className="layout-variant-head"><div><h3>{variant.name}</h3><p>{variant.description || t("layouts.plans.noDescription")}</p></div>
          {canPlan ? <button type="button" className="secondary-button compact-action" onClick={() => void createDraft(variant)}>
            <Plus size={14} />{t("layouts.plans.createDraft")}</button> : null}</div>
        {variant.revisions.length === 0 ? <p className="layout-empty">{t("layouts.plans.noRevisions")}</p> : <div className="table-wrap">
          <table className="layout-table"><thead><tr><th>{t("layouts.plans.revision")}</th><th>{t("layouts.field.status")}</th>
            <th>{t("layouts.plans.author")}</th><th>{t("layouts.plans.timestamps")}</th><th>{t("layouts.plans.actions")}</th></tr></thead>
            <tbody>{variant.revisions.map((revision) => <tr key={revision.id}>
              <td><strong>R{revision.revisionNumber}</strong>{revision.baseRevisionId ? <small>{t("layouts.plans.hasBase")}</small> : null}</td>
              <td><span className={`status-pill revision-${revision.status}`}>{t(`layouts.revisionStatus.${revision.status}`)}</span></td>
              <td>{revision.createdBy || "-"}{revision.publishedBy ? <small>{t("layouts.plans.publishedBy", { actor: revision.publishedBy })}</small> : null}</td>
              <td>{new Date(revision.createdAt).toLocaleString()}{revision.publishedAt ?
                <small>{t("layouts.plans.publishedAt", { date: new Date(revision.publishedAt).toLocaleString() })}</small> : null}</td>
              <td><div className="layout-plan-actions">
                <button type="button" className="secondary-button compact-action" onClick={() => setOpenRevision(revision)}>
                  {canPlan && revision.status === "draft" ? <PencilRuler size={14} /> : <Eye size={14} />}
                  {canPlan && revision.status === "draft"
                    ? t("layouts.trackPlanner.openEdit") : t("layouts.trackPlanner.openRead")}
                </button>
                {canPlan && revision.status === "draft" ? <button type="button" className="secondary-button compact-action"
                  onClick={() => void submit(revision)}>{t("layouts.plans.submit")}</button> : null}
                {canPlan && revision.status === "review" ? <button type="button" className="primary-button compact-action"
                  onClick={() => askPublish(variant, revision)}>{t("layouts.plans.publish")}</button> : null}
              </div></td>
            </tr>)}</tbody></table></div>}
      </section>)}</div>}
    {canPlan && unitID ? <section className="panel layout-plan-create">
      <div className="panel-title"><Plus size={17} /><h3>{t("layouts.plans.createVariant")}</h3></div>
      <form className="layout-form" onSubmit={createVariant}>
        <label>{t("layouts.field.name")}<input required value={name} onChange={(event) => setName(event.target.value)} /></label>
        <label>{t("layouts.field.description")}<textarea value={description}
          onChange={(event) => setDescription(event.target.value)} /></label>
        <button type="submit" className="primary-button" disabled={saving || !name.trim()}>
          {saving ? t("common.saving") : t("layouts.plans.saveVariant")}</button>
      </form>
    </section> : null}
    <LayoutConfirmDialog action={pending} onClose={() => setPending(null)} />
  </section>;
}
