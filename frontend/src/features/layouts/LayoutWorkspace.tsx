import { useCallback, useEffect, useRef, useState } from "react";
import { AlertTriangle, Building2, ClipboardList, FileText, Pencil } from "lucide-react";

import {
  ApiError,
  api,
  type Layout,
  type LayoutConfiguration,
  type LayoutUnit
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { LayoutConfigurationsPanel } from "./LayoutConfigurationsPanel";
import { LayoutFormDialog, type LayoutFormValue } from "./LayoutFormDialog";
import { LayoutModulesPanel } from "./LayoutModulesPanel";
import { LayoutPlansPanel } from "./LayoutPlansPanel";
import { LayoutTechnicalPositionsPanel } from "./LayoutTechnicalPositionsPanel";
import { LayoutTwinPanel } from "./LayoutTwinPanel";

type LayoutTab = "overview" | "planner" | "modules" | "setups" | "technology" | "maintenance" | "documents";

function layoutFormValue(layout: Layout): LayoutFormValue {
  return {
    name: layout.name,
    kind: layout.kind,
    gauge: layout.gauge,
    scale: layout.scale,
    maxGradePercent: layout.maxGradePercent == null ? "" : String(layout.maxGradePercent),
    minimumTrackClearanceMm: layout.minimumTrackClearanceMm == null
      ? "" : String(layout.minimumTrackClearanceMm),
    description: layout.description || "",
    archived: layout.archived
  };
}

export function LayoutWorkspace({ layout, canPlan, onLayoutChanged }: {
  layout: Layout;
  canPlan: boolean;
  onLayoutChanged: (layout: Layout) => void;
}) {
  const [tab, setTab] = useState<LayoutTab>("overview");
  const [units, setUnits] = useState<LayoutUnit[]>([]);
  const [configurations, setConfigurations] = useState<LayoutConfiguration[]>([]);
  const [editOpen, setEditOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [conflict, setConflict] = useState(false);
  const editTriggerRef = useRef<HTMLButtonElement | null>(null);
  const { t, language } = useI18n();
  const formattedMaxGrade = layout.maxGradePercent == null
    ? t("layouts.field.maxGradePercentUnset")
    : `${new Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB", {
      minimumFractionDigits: 2, maximumFractionDigits: 2
    }).format(layout.maxGradePercent)} %`;
  const formattedMinimumTrackClearance = layout.minimumTrackClearanceMm == null
    ? t("layouts.field.minimumTrackClearanceMmUnset")
    : `${new Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB", {
      minimumFractionDigits: 2, maximumFractionDigits: 2
    }).format(layout.minimumTrackClearanceMm)} mm`;

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
      onLayoutChanged(nextLayout);
      return layoutFormValue(nextLayout);
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally {
      setLoading(false);
    }
  };

  const saveLayout = async (form: LayoutFormValue) => {
    setSaving(true); setMessage(""); setConflict(false);
    try {
      const updated = await api.updateLayout(layout.id, {
        name: form.name,
        kind: form.kind,
        gauge: form.gauge,
        scale: form.scale,
        maxGradePercent: form.maxGradePercent ? Number(form.maxGradePercent) : null,
        minimumTrackClearanceMm: form.minimumTrackClearanceMm
          ? Number(form.minimumTrackClearanceMm) : null,
        description: form.description?.trim() || undefined,
        archived: form.archived,
        expectedVersion: layout.version
      });
      onLayoutChanged(updated); setEditOpen(false);
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
    {message && !editOpen ? <div className={conflict ? "layout-conflict" : "form-message"}>
      {conflict ? <AlertTriangle size={16} /> : null}<span>{message}</span>
      {conflict ? <button type="button" className="secondary-button compact-action" onClick={reloadServerState}>
        {t("layouts.conflict.reload")}</button> : null}
    </div> : null}
    <div className="layout-tabs" role="tablist" aria-label={t("layouts.tabs.label")}>{tabs.map((item) =>
      <button type="button" role="tab" aria-selected={tab === item} key={item} className={tab === item ? "active" : ""}
        onClick={() => setTab(item)}>{t(`layouts.tabs.${item}`)}</button>)}</div>
    {loading ? <section className="panel"><p>{t("layouts.structure.loading")}</p></section> :
      tab === "overview" ? <section className="layout-overview-grid">
        <LayoutTwinPanel layout={layout} units={units} configurations={configurations} canPlan={canPlan} />
        <section className="panel layout-overview-summary">
          <div className="layout-panel-head">
            <div className="panel-title"><Building2 size={17} /><h3>{t("layouts.overview.title")}</h3></div>
            {canPlan ? <button ref={editTriggerRef} type="button" className="secondary-button compact-action"
              onClick={() => { setMessage(""); setConflict(false); setEditOpen(true); }}>
              <Pencil size={15} />{t("layouts.edit.action")}
            </button> : null}
          </div>
          <dl><div><dt>{t("layouts.field.kind")}</dt><dd>{t(`layouts.kind.${layout.kind}`)}</dd></div>
            <div><dt>{t("layouts.field.status")}</dt><dd>{layout.archived
              ? t("layouts.status.archived") : t("layouts.status.active")}</dd></div>
            <div><dt>{t("layouts.field.gauge")}</dt><dd>{layout.gauge}</dd></div>
            <div><dt>{t("layouts.field.scale")}</dt><dd>{layout.scale}</dd></div>
            <div><dt>{t("layouts.field.maxGradePercent")}</dt><dd>{formattedMaxGrade}</dd></div>
            <div><dt>{t("layouts.field.minimumTrackClearanceMm")}</dt>
              <dd>{formattedMinimumTrackClearance}</dd></div>
            <div><dt>{t("layouts.overview.version")}</dt><dd>{t("layouts.version", { version: layout.version })}</dd></div>
            <div><dt>{t("layouts.overview.units")}</dt><dd>{units.length}</dd></div>
            <div><dt>{t("layouts.overview.setups")}</dt><dd>{configurations.length}</dd></div>
            <div><dt>{t("layouts.overview.created")}</dt><dd>{new Date(layout.createdAt).toLocaleString()}</dd></div>
            <div><dt>{t("layouts.overview.updated")}</dt><dd>{new Date(layout.updatedAt).toLocaleString()}</dd></div></dl>
          <div className="layout-profile-description"><h4>{t("layouts.field.description")}</h4>
            <p className={layout.description ? "" : "layout-empty"}>
              {layout.description || t("layouts.overview.noDescription")}
            </p>
          </div>
        </section>
      </section> : tab === "modules" ? <LayoutModulesPanel units={units} layoutID={layout.id} canPlan={canPlan}
        onChanged={reloadStructure} /> : tab === "setups" ? <LayoutConfigurationsPanel configurations={configurations}
          units={units} layoutID={layout.id} canPlan={canPlan} onChanged={reloadStructure} />
        : tab === "planner" ? <LayoutPlansPanel units={units} gauge={layout.gauge} canPlan={canPlan} />
          : tab === "technology" ? <LayoutTechnicalPositionsPanel units={units} canPlan={canPlan} />
          : <LayoutDeferredPanel tab={tab} />}
    {editOpen ? <LayoutFormDialog mode="edit" initialValue={layoutFormValue(layout)} saving={saving}
      message={message} conflict={conflict} returnFocusTo={editTriggerRef.current}
      onSubmit={saveLayout} onReloadConflict={reloadServerState}
      onClose={() => { setEditOpen(false); setMessage(""); setConflict(false); }} /> : null}
  </section>;
}

function LayoutDeferredPanel({ tab }: { tab: "maintenance" | "documents" }) {
  const { t } = useI18n();
  const Icon = tab === "maintenance" ? ClipboardList : FileText;
  return <section className="panel layout-deferred"><Icon size={22} /><h3>{t(`layouts.tabs.${tab}`)}</h3>
    <p>{t(`layouts.deferred.${tab}`)}</p><span>{t("layouts.deferred.stage")}</span></section>;
}
