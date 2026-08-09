import { FormEvent, useEffect, useMemo, useState } from "react";
import { Layers3, Magnet, Plus } from "lucide-react";

import {
  api,
  type ConfigurationUnitInput,
  type LayoutConfiguration,
  type LayoutUnit,
  type PlanRevision
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { LayoutConfigurationPortAnalysis } from "./LayoutConfigurationPortAnalysis";

type ConfigurationForm = {
  id?: string;
  version?: number;
  name: string;
  description: string;
  archived: boolean;
  units: ConfigurationUnitInput[];
};

type RevisionOption = { revision: PlanRevision; label: string };
const emptyConfiguration: ConfigurationForm = { name: "", description: "", archived: false, units: [] };

export function LayoutConfigurationsPanel({ configurations, units, layoutID, canPlan, onChanged }: {
  configurations: LayoutConfiguration[];
  units: LayoutUnit[];
  layoutID: string;
  canPlan: boolean;
  onChanged: () => Promise<void>;
}) {
  const [selectedID, setSelectedID] = useState("");
  const [form, setForm] = useState<ConfigurationForm>(emptyConfiguration);
  const [revisionOptions, setRevisionOptions] = useState<Record<string, RevisionOption[]>>({});
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const { t } = useI18n();
  const selected = configurations.find((configuration) => configuration.id === selectedID);
  const activeUnits = useMemo(() => units.filter((unit) => !unit.archived || form.units.some((item) => item.unitId === unit.id)),
    [form.units, units]);

  useEffect(() => {
    let active = true;
    Promise.all(units.map(async (unit) => ({ unitID: unit.id, variants: await api.planVariants(unit.id) })))
      .then((results) => {
        if (!active) return;
        setRevisionOptions(Object.fromEntries(results.map(({ unitID, variants }) => [unitID, variants.flatMap((variant) =>
          variant.revisions.filter((revision) => revision.status === "published").map((revision) => ({
            revision, label: `${variant.name} · R${revision.revisionNumber}`
          })))])));
      }).catch((reason: Error) => active && setMessage(reason.message));
    return () => { active = false; };
  }, [units]);

  useEffect(() => {
    if (!selected) return;
    setForm({ id: selected.id, version: selected.version, name: selected.name,
      description: selected.description || "", archived: selected.archived,
      units: selected.units.map((unit) => ({ unitId: unit.unitId, planRevisionId: unit.planRevisionId,
        positionXMm: unit.positionXMm, positionYMm: unit.positionYMm, rotationDegrees: unit.rotationDegrees })) });
  }, [selected]);

  const reset = () => { setSelectedID(""); setForm(emptyConfiguration); setMessage(""); };
  const assignment = (unitID: string) => form.units.find((item) => item.unitId === unitID);
  const toggleUnit = (unitID: string, enabled: boolean) => setForm((current) => ({ ...current,
    units: enabled ? [...current.units, { unitId: unitID, positionXMm: 0, positionYMm: 0, rotationDegrees: 0 }]
      : current.units.filter((item) => item.unitId !== unitID) }));
  const updateUnit = (unitID: string, patch: Partial<ConfigurationUnitInput>) => setForm((current) => ({ ...current,
    units: current.units.map((item) => item.unitId === unitID ? { ...item, ...patch } : item) }));

  const snapUnit = async (unitID: string) => {
    const item = assignment(unitID);
    if (!form.id || !item) return;
    setMessage("");
    try {
      const preview = await api.previewLayoutConfigurationUnitSnap(form.id, {
        unitId: unitID,
        positionXMm: item.positionXMm || 0,
        positionYMm: item.positionYMm || 0,
        rotationDegrees: item.rotationDegrees || 0
      });
      if (!preview.snapped) {
        setMessage(t("layouts.setups.snapMissing"));
        return;
      }
      updateUnit(unitID, preview.pose);
      setMessage(t("layouts.setups.snapApplied"));
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    }
  };

  const save = async (event: FormEvent) => {
    event.preventDefault(); setSaving(true); setMessage("");
    const input = { name: form.name, description: form.description.trim() || undefined, archived: form.archived,
      units: form.units };
    try {
      if (form.id && form.version) await api.updateLayoutConfiguration(form.id, { ...input, expectedVersion: form.version });
      else await api.createLayoutConfiguration(layoutID, input);
      reset(); await onChanged();
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally { setSaving(false); }
  };

  return <section className="layout-panel-grid">
    <section className="panel">
      <div className="panel-title"><Layers3 size={17} /><h3>{t("layouts.setups.title")}</h3></div>
      {configurations.length === 0 ? <p className="layout-empty">{t("layouts.setups.empty")}</p> :
        <div className="layout-card-list">{configurations.map((configuration) => <button type="button"
          key={configuration.id} className={configuration.id === selectedID ? "layout-card selected" : "layout-card"}
          onClick={() => setSelectedID(configuration.id)}><span><strong>{configuration.name}</strong>
            <small>{t("layouts.setups.unitCount", { count: configuration.units.length })}</small></span>
          <span className={configuration.archived ? "status-pill archived" : "status-pill"}>
            {configuration.archived ? t("layouts.status.archived") : t("layouts.status.active")}</span></button>)}</div>}
      <LayoutConfigurationPortAnalysis configuration={selected || null} />
    </section>
    {canPlan ? <section className="panel layout-setup-form-panel">
      <div className="panel-title"><Plus size={17} /><h3>{form.id ? t("layouts.setups.edit") : t("layouts.setups.create")}</h3></div>
      <form className="layout-form" onSubmit={save}>
        <label>{t("layouts.field.name")}<input required value={form.name}
          onChange={(event) => setForm({ ...form, name: event.target.value })} /></label>
        <label>{t("layouts.field.description")}<textarea value={form.description}
          onChange={(event) => setForm({ ...form, description: event.target.value })} /></label>
        <fieldset className="layout-unit-placements"><legend>{t("layouts.setups.units")}</legend>
          {activeUnits.length === 0 ? <p className="layout-empty">{t("layouts.setups.noUnits")}</p> : activeUnits.map((unit) => {
            const item = assignment(unit.id);
            return <div key={unit.id} className="layout-placement-row">
              <label className="layout-check"><input type="checkbox" checked={Boolean(item)}
                onChange={(event) => toggleUnit(unit.id, event.target.checked)} /><span><strong>{unit.name}</strong>
                  <small>{t(`layouts.unitKind.${unit.kind}`)}</small></span></label>
              {item ? <div className="layout-placement-fields">
                <label>{t("layouts.field.planRevision")}<AppSelect value={item.planRevisionId || ""}
                  onChange={(event) => updateUnit(unit.id, { planRevisionId: event.target.value || undefined })}>
                  <option value="">{t("layouts.setups.noRevision")}</option>
                  {(revisionOptions[unit.id] || []).map((option) => <option key={option.revision.id}
                    value={option.revision.id}>{option.label}</option>)}
                </AppSelect></label>
                <label>X (mm)<input type="number" value={item.positionXMm || 0}
                  onChange={(event) => updateUnit(unit.id, { positionXMm: Number(event.target.value) })} /></label>
                <label>Y (mm)<input type="number" value={item.positionYMm || 0}
                  onChange={(event) => updateUnit(unit.id, { positionYMm: Number(event.target.value) })} /></label>
                <label>{t("layouts.field.rotation")}<input type="number" step="0.1" value={item.rotationDegrees || 0}
                  onChange={(event) => updateUnit(unit.id, { rotationDegrees: Number(event.target.value) })} /></label>
              </div> : null}
              {item && form.id ? <div className="layout-placement-actions"><button type="button" className="secondary-button"
                onClick={() => void snapUnit(unit.id)}><Magnet size={14} />{t("layouts.setups.snap", { name: unit.name })}
              </button></div> : null}
            </div>;
          })}</fieldset>
        {form.id ? <label className="layout-check"><input type="checkbox" checked={form.archived}
          onChange={(event) => setForm({ ...form, archived: event.target.checked })} />{t("layouts.field.archived")}</label> : null}
        {message ? <p className="form-message">{message}</p> : null}
        <div className="layout-form-actions">{form.id ? <button type="button" className="secondary-button" onClick={reset}>
          {t("layouts.setups.new")}</button> : null}<button type="submit" className="primary-button" disabled={saving || !form.name.trim()}>
          {saving ? t("common.saving") : t("layouts.setups.save")}</button></div>
      </form>
    </section> : null}
  </section>;
}
