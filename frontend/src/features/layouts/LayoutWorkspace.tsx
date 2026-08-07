import { FormEvent, useCallback, useEffect, useState } from "react";
import { AlertTriangle, Building2, ClipboardList, FileText, Settings2, Wrench } from "lucide-react";

import {
  ApiError,
  api,
  type Layout,
  type LayoutConfiguration,
  type LayoutKind,
  type LayoutUnit
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { LayoutConfigurationsPanel } from "./LayoutConfigurationsPanel";
import { LayoutModulesPanel } from "./LayoutModulesPanel";
import { LayoutPlansPanel } from "./LayoutPlansPanel";

type LayoutTab = "overview" | "planner" | "modules" | "setups" | "technology" | "maintenance" | "documents";

export function LayoutWorkspace({ layout, canPlan, onLayoutChanged }: {
  layout: Layout;
  canPlan: boolean;
  onLayoutChanged: (layout: Layout) => void;
}) {
  const [tab, setTab] = useState<LayoutTab>("overview");
  const [units, setUnits] = useState<LayoutUnit[]>([]);
  const [configurations, setConfigurations] = useState<LayoutConfiguration[]>([]);
  const [form, setForm] = useState(layout);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [conflict, setConflict] = useState(false);
  const { t } = useI18n();

  const reloadStructure = useCallback(async () => {
    const [nextUnits, nextConfigurations] = await Promise.all([
      api.layoutUnits(layout.id), api.layoutConfigurations(layout.id)
    ]);
    setUnits(nextUnits); setConfigurations(nextConfigurations);
  }, [layout.id]);

  useEffect(() => {
    let active = true;
    Promise.all([api.layoutUnits(layout.id), api.layoutConfigurations(layout.id)])
      .then(([nextUnits, nextConfigurations]) => {
        if (!active) return;
        setUnits(nextUnits); setConfigurations(nextConfigurations);
      })
      .catch((reason: Error) => active && setMessage(reason.message))
      .finally(() => active && setLoading(false));
    return () => { active = false; };
  }, [layout.id]);

  const reloadServerState = async () => {
    setMessage(""); setConflict(false); setLoading(true);
    try {
      const [nextLayout] = await Promise.all([api.layout(layout.id), reloadStructure()]);
      setForm(nextLayout); onLayoutChanged(nextLayout);
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally {
      setLoading(false);
    }
  };

  const saveLayout = async (event: FormEvent) => {
    event.preventDefault(); setSaving(true); setMessage(""); setConflict(false);
    try {
      const updated = await api.updateLayout(layout.id, {
        name: form.name,
        kind: form.kind,
        gauge: form.gauge,
        scale: form.scale,
        description: form.description?.trim() || undefined,
        archived: form.archived,
        expectedVersion: form.version
      });
      setForm(updated); onLayoutChanged(updated);
    } catch (reason) {
      if (reason instanceof ApiError && reason.status === 409) {
        setConflict(true); setMessage(t("layouts.conflict.message"));
      } else setMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally {
      setSaving(false);
    }
  };

  const tabs: LayoutTab[] = ["overview", "planner", "modules", "setups", "technology", "maintenance", "documents"];
  return <section className="layout-workspace">
    <header className="layout-workspace-head">
      <div><p className="eyebrow">{t(`layouts.kind.${layout.kind}`)}</p><h2>{layout.name}</h2>
        <p>{layout.gauge} · {layout.scale} · {t("layouts.version", { version: layout.version })}</p></div>
      <span className={layout.archived ? "status-pill archived" : "status-pill"}>
        {layout.archived ? t("layouts.status.archived") : t("layouts.status.active")}
      </span>
    </header>
    {message ? <div className={conflict ? "layout-conflict" : "form-message"}>
      {conflict ? <AlertTriangle size={16} /> : null}<span>{message}</span>
      {conflict ? <button type="button" className="secondary-button compact-action" onClick={reloadServerState}>
        {t("layouts.conflict.reload")}</button> : null}
    </div> : null}
    <div className="layout-tabs" role="tablist" aria-label={t("layouts.tabs.label")}>{tabs.map((item) =>
      <button type="button" role="tab" aria-selected={tab === item} key={item} className={tab === item ? "active" : ""}
        onClick={() => setTab(item)}>{t(`layouts.tabs.${item}`)}</button>)}</div>
    {loading ? <section className="panel"><p>{t("layouts.structure.loading")}</p></section> :
      tab === "overview" ? <section className="layout-overview-grid">
        <section className="panel layout-overview-summary">
          <div className="panel-title"><Building2 size={17} /><h3>{t("layouts.overview.title")}</h3></div>
          <dl><div><dt>{t("layouts.field.kind")}</dt><dd>{t(`layouts.kind.${layout.kind}`)}</dd></div>
            <div><dt>{t("layouts.overview.units")}</dt><dd>{units.length}</dd></div>
            <div><dt>{t("layouts.overview.setups")}</dt><dd>{configurations.length}</dd></div>
            <div><dt>{t("layouts.overview.updated")}</dt><dd>{new Date(layout.updatedAt).toLocaleString()}</dd></div></dl>
          <p>{layout.description || t("layouts.overview.noDescription")}</p>
        </section>
        {canPlan ? <section className="panel">
          <div className="panel-title"><Settings2 size={17} /><h3>{t("layouts.edit.title")}</h3></div>
          <form className="layout-form" onSubmit={saveLayout}>
            <label>{t("layouts.field.name")}<input required value={form.name}
              onChange={(event) => setForm({ ...form, name: event.target.value })} /></label>
            <label>{t("layouts.field.kind")}<AppSelect value={form.kind}
              onChange={(event) => setForm({ ...form, kind: event.target.value as LayoutKind })}>
              <option value="private">{t("layouts.kind.private")}</option><option value="club">{t("layouts.kind.club")}</option>
            </AppSelect></label>
            <div className="layout-inline-fields"><label>{t("layouts.field.gauge")}<input required value={form.gauge}
              onChange={(event) => setForm({ ...form, gauge: event.target.value })} /></label>
              <label>{t("layouts.field.scale")}<input required value={form.scale}
                onChange={(event) => setForm({ ...form, scale: event.target.value })} /></label></div>
            <label>{t("layouts.field.description")}<textarea value={form.description || ""}
              onChange={(event) => setForm({ ...form, description: event.target.value })} /></label>
            <label className="layout-check"><input type="checkbox" checked={form.archived}
              onChange={(event) => setForm({ ...form, archived: event.target.checked })} />{t("layouts.field.archived")}</label>
            <button type="submit" className="primary-button" disabled={saving}>{saving ? t("common.saving") : t("layouts.edit.save")}</button>
          </form>
        </section> : null}
      </section> : tab === "modules" ? <LayoutModulesPanel units={units} layoutID={layout.id} canPlan={canPlan}
        onChanged={reloadStructure} /> : tab === "setups" ? <LayoutConfigurationsPanel configurations={configurations}
          units={units} layoutID={layout.id} canPlan={canPlan} onChanged={reloadStructure} />
        : tab === "planner" ? <LayoutPlansPanel units={units} canPlan={canPlan} />
          : <LayoutDeferredPanel tab={tab} />}
  </section>;
}

function LayoutDeferredPanel({ tab }: { tab: "technology" | "maintenance" | "documents" }) {
  const { t } = useI18n();
  const Icon = tab === "technology" ? Wrench : tab === "maintenance" ? ClipboardList : FileText;
  return <section className="panel layout-deferred"><Icon size={22} /><h3>{t(`layouts.tabs.${tab}`)}</h3>
    <p>{t(`layouts.deferred.${tab}`)}</p><span>{t("layouts.deferred.stage")}</span></section>;
}
