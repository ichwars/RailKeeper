import { FormEvent, useEffect, useState } from "react";
import { Boxes, Plus } from "lucide-react";

import { api, type LayoutUnit, type LayoutUnitInput, type LayoutUnitKind } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";

type UnitForm = LayoutUnitInput & { id?: string; version?: number };
const emptyUnit: UnitForm = { name: "", kind: "module", ownerLabel: "", widthMm: 0, heightMm: 0, archived: false };

export function LayoutModulesPanel({ units, layoutID, canPlan, onChanged }: {
  units: LayoutUnit[];
  layoutID: string;
  canPlan: boolean;
  onChanged: () => Promise<void>;
}) {
  const [selectedID, setSelectedID] = useState("");
  const [form, setForm] = useState<UnitForm>(emptyUnit);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const { t } = useI18n();
  const selected = units.find((unit) => unit.id === selectedID);

  useEffect(() => {
    if (!selected) return;
    setForm({ id: selected.id, version: selected.version, name: selected.name, kind: selected.kind,
      ownerLabel: selected.ownerLabel || "", widthMm: selected.widthMm, heightMm: selected.heightMm,
      archived: selected.archived });
  }, [selected]);

  const reset = () => { setSelectedID(""); setForm(emptyUnit); setMessage(""); };
  const save = async (event: FormEvent) => {
    event.preventDefault(); setSaving(true); setMessage("");
    const input: LayoutUnitInput = { name: form.name, kind: form.kind, ownerLabel: form.ownerLabel?.trim() || undefined,
      widthMm: Number(form.widthMm || 0), heightMm: Number(form.heightMm || 0), archived: Boolean(form.archived) };
    try {
      if (form.id && form.version) await api.updateLayoutUnit(form.id, { ...input, expectedVersion: form.version });
      else await api.createLayoutUnit(layoutID, input);
      reset(); await onChanged();
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally { setSaving(false); }
  };

  return <section className="layout-panel-grid">
    <section className="panel">
      <div className="panel-title"><Boxes size={17} /><h3>{t("layouts.modules.title")}</h3></div>
      {units.length === 0 ? <p className="layout-empty">{t("layouts.modules.empty")}</p> : <div className="table-wrap">
        <table className="layout-table"><thead><tr><th>{t("layouts.field.name")}</th><th>{t("layouts.field.kind")}</th>
          <th>{t("layouts.field.dimensions")}</th><th>{t("layouts.field.owner")}</th><th>{t("layouts.field.status")}</th></tr></thead>
          <tbody>{units.map((unit) => <tr key={unit.id} className={unit.id === selectedID ? "selected-row" : ""}>
            <td><button type="button" className="inventory-name-link" onClick={() => setSelectedID(unit.id)}>{unit.name}</button></td>
            <td>{t(`layouts.unitKind.${unit.kind}`)}</td><td>{unit.widthMm} × {unit.heightMm} mm</td>
            <td>{unit.ownerLabel || "-"}</td><td>{unit.archived ? t("layouts.status.archived") : t("layouts.status.active")}</td>
          </tr>)}</tbody></table></div>}
    </section>
    {canPlan ? <section className="panel">
      <div className="panel-title"><Plus size={17} /><h3>{form.id ? t("layouts.modules.edit") : t("layouts.modules.create")}</h3></div>
      <form className="layout-form" onSubmit={save}>
        <label>{t("layouts.field.name")}<input required value={form.name}
          onChange={(event) => setForm({ ...form, name: event.target.value })} /></label>
        <label>{t("layouts.field.kind")}<AppSelect value={form.kind}
          onChange={(event) => setForm({ ...form, kind: event.target.value as LayoutUnitKind })}>
          {(["baseboard", "module", "segment", "area"] as const).map((kind) =>
            <option key={kind} value={kind}>{t(`layouts.unitKind.${kind}`)}</option>)}
        </AppSelect></label>
        <label>{t("layouts.field.owner")}<input value={form.ownerLabel || ""}
          onChange={(event) => setForm({ ...form, ownerLabel: event.target.value })} /></label>
        <div className="layout-inline-fields"><label>{t("layouts.field.width")}<input type="number" min="0" step="1"
          value={form.widthMm || 0} onChange={(event) => setForm({ ...form, widthMm: Number(event.target.value) })} /></label>
          <label>{t("layouts.field.height")}<input type="number" min="0" step="1" value={form.heightMm || 0}
            onChange={(event) => setForm({ ...form, heightMm: Number(event.target.value) })} /></label></div>
        {form.id ? <label className="layout-check"><input type="checkbox" checked={Boolean(form.archived)}
          onChange={(event) => setForm({ ...form, archived: event.target.checked })} />{t("layouts.field.archived")}</label> : null}
        {message ? <p className="form-message">{message}</p> : null}
        <div className="layout-form-actions">{form.id ? <button type="button" className="secondary-button" onClick={reset}>
          {t("layouts.modules.new")}</button> : null}<button type="submit" className="primary-button" disabled={saving || !form.name.trim()}>
          {saving ? t("common.saving") : t("layouts.modules.save")}</button></div>
      </form>
    </section> : null}
  </section>;
}
