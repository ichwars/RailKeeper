import { FormEvent, useCallback, useEffect, useState } from "react";
import { LayoutTemplate, Plus, RefreshCw } from "lucide-react";

import { api, type Layout, type LayoutInput, type LayoutKind } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { LayoutWorkspace } from "./LayoutWorkspace";

const emptyLayout: LayoutInput = { name: "", kind: "private", gauge: "TT", scale: "1:120", description: "" };

export function LayoutsView({ roles }: { roles: string[] }) {
  const [layouts, setLayouts] = useState<Layout[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [form, setForm] = useState<LayoutInput>(emptyLayout);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
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

  const saveLayout = async (event: FormEvent) => {
    event.preventDefault();
    setSaving(true); setMessage("");
    try {
      const created = await api.createLayout({ ...form, description: form.description?.trim() || undefined });
      setForm(emptyLayout);
      await loadLayouts(created.id);
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
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
        <div className="panel-title"><LayoutTemplate size={17} /><h2>{t("layouts.catalog.title")}</h2></div>
        {layouts.length === 0 ? <p className="layout-empty">{t("layouts.catalog.empty")}</p> :
          <div className="layout-card-list">{layouts.map((layout) => <button type="button" key={layout.id}
            className={layout.id === selectedID ? "layout-card selected" : "layout-card"}
            onClick={() => setSelectedID(layout.id)}>
            <span><strong>{layout.name}</strong><small>{layout.gauge} · {layout.scale}</small></span>
            <span className={layout.archived ? "status-pill archived" : "status-pill"}>
              {layout.archived ? t("layouts.status.archived") : t(`layouts.kind.${layout.kind}`)}
            </span>
          </button>)}</div>}
      </section>
      {canPlan ? <section className="panel layout-create-panel">
        <div className="panel-title"><Plus size={17} /><h2>{t("layouts.create.title")}</h2></div>
        <form className="layout-form" onSubmit={saveLayout}>
          <label>{t("layouts.field.name")}<input required value={form.name}
            onChange={(event) => setForm({ ...form, name: event.target.value })} /></label>
          <label>{t("layouts.field.kind")}<AppSelect value={form.kind}
            onChange={(event) => setForm({ ...form, kind: event.target.value as LayoutKind })}>
            <option value="private">{t("layouts.kind.private")}</option><option value="club">{t("layouts.kind.club")}</option>
          </AppSelect></label>
          <div className="layout-inline-fields">
            <label>{t("layouts.field.gauge")}<input required value={form.gauge}
              onChange={(event) => setForm({ ...form, gauge: event.target.value })} /></label>
            <label>{t("layouts.field.scale")}<input required value={form.scale}
              onChange={(event) => setForm({ ...form, scale: event.target.value })} /></label>
          </div>
          <label>{t("layouts.field.description")}<textarea value={form.description || ""}
            onChange={(event) => setForm({ ...form, description: event.target.value })} /></label>
          <button type="submit" className="primary-button" disabled={saving || !form.name.trim()}>
            {saving ? t("common.saving") : t("layouts.create.save")}
          </button>
        </form>
      </section> : null}
    </section>
    {selected ? <LayoutWorkspace key={selected.id} layout={selected} canPlan={canPlan}
      onLayoutChanged={(updated) => {
        setLayouts((current) => current.map((layout) => layout.id === updated.id ? updated : layout));
      }} /> : null}
  </>;
}
