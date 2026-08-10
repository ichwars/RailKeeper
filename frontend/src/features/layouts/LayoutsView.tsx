import { useCallback, useEffect, useRef, useState } from "react";
import { LayoutTemplate, Plus, RefreshCw } from "lucide-react";

import { api, type Layout } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { LayoutFormDialog, type LayoutFormValue } from "./LayoutFormDialog";
import { LayoutWorkspace } from "./LayoutWorkspace";

const emptyLayout: LayoutFormValue = {
  name: "", kind: "private", gauge: "TT", scale: "1:120", maxGradePercent: "",
  minimumTrackClearanceMm: "", minimumFlexRadiusMm: "", description: "", archived: false
};

export function LayoutsView({ roles }: { roles: string[] }) {
  const [layouts, setLayouts] = useState<Layout[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [createMessage, setCreateMessage] = useState("");
  const createTriggerRef = useRef<HTMLButtonElement | null>(null);
  const { t } = useI18n();
  const canPlan = roles.includes("Admin") || roles.includes("Planner");
  const selected = layouts.find((layout) => layout.id === selectedID) || null;

  const loadLayouts = useCallback(async (preferID?: string) => {
    setMessage("");
    try {
      const next = await api.layouts();
      setLayouts(next);
      setSelectedID((current) => {
        const candidate = preferID || current;
        return next.some((layout) => layout.id === candidate) ? candidate : next[0]?.id || "";
      });
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    }
  }, [t]);

  useEffect(() => {
    let active = true;
    api.layouts().then((next) => {
      if (!active) return;
      setLayouts(next);
      setSelectedID(next[0]?.id || "");
    }).catch((reason: Error) => active && setMessage(reason.message))
      .finally(() => active && setLoading(false));
    return () => { active = false; };
  }, []);

  const createLayout = async (form: LayoutFormValue) => {
    setSaving(true); setCreateMessage("");
    try {
      const created = await api.createLayout({
        name: form.name,
        kind: form.kind,
        gauge: form.gauge,
        scale: form.scale,
        maxGradePercent: form.maxGradePercent ? Number(form.maxGradePercent) : null,
        minimumTrackClearanceMm: form.minimumTrackClearanceMm
          ? Number(form.minimumTrackClearanceMm) : null,
        minimumFlexRadiusMm: form.minimumFlexRadiusMm ? Number(form.minimumFlexRadiusMm) : null,
        description: form.description || undefined
      });
      await loadLayouts(created.id);
      setCreateOpen(false);
    } catch (reason) {
      setCreateMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <section className="panel"><p>{t("layouts.loading")}</p></section>;

  return <>
    <section className="inventory-head layout-head">
      <div><p className="eyebrow">{t("layouts.eyebrow")}</p><h1>{t("layouts.title")}</h1><p>{t("layouts.subtitle")}</p></div>
      <button type="button" className="icon-button" onClick={() => void loadLayouts()} aria-label={t("common.refresh")}
        title={t("common.refresh")}><RefreshCw size={16} /></button>
    </section>
    {message ? <p className="form-message">{message}</p> : null}
    <section className="layout-catalog-grid">
      <section className="panel layout-list-panel">
        <div className="layout-panel-head">
          <div className="panel-title"><LayoutTemplate size={17} /><h2>{t("layouts.catalog.title")}</h2></div>
          {canPlan ? <button ref={createTriggerRef} type="button" className="primary-button compact-action"
            onClick={() => { setCreateMessage(""); setCreateOpen(true); }}>
            <Plus size={16} />{t("layouts.create.action")}
          </button> : null}
        </div>
        {layouts.length === 0 ? <p className="layout-empty">{t("layouts.catalog.empty")}</p> :
          <div className="layout-card-list">{layouts.map((layout) => <button type="button" key={layout.id}
            className={layout.id === selectedID ? "layout-card selected" : "layout-card"}
            aria-pressed={layout.id === selectedID} onClick={() => setSelectedID(layout.id)}>
            <span className="layout-card-main"><strong>{layout.name}</strong>
              <span className="layout-card-facts">
                <small>{t(`layouts.kind.${layout.kind}`)}</small>
                <small>{layout.gauge} · {layout.scale}</small>
                <small>{t("layouts.version", { version: layout.version })}</small>
                <small>{new Date(layout.updatedAt).toLocaleString()}</small>
              </span>
            </span>
            <span className={layout.archived ? "status-pill archived" : "status-pill"}>
              {layout.archived ? t("layouts.status.archived") : t("layouts.status.active")}
            </span>
          </button>)}</div>}
      </section>
    </section>
    {selected ? <LayoutWorkspace key={selected.id} layout={selected} canPlan={canPlan}
      onLayoutChanged={(updated) => {
        setLayouts((current) => current.map((layout) => layout.id === updated.id ? updated : layout));
      }} /> : null}
    {createOpen ? <LayoutFormDialog mode="create" initialValue={emptyLayout} saving={saving}
      message={createMessage} conflict={false} returnFocusTo={createTriggerRef.current}
      onSubmit={createLayout} onClose={() => { setCreateOpen(false); setCreateMessage(""); }} /> : null}
  </>;
}
