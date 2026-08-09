import { FormEvent, useCallback, useEffect, useState } from "react";
import { Cable, Pencil, Plus } from "lucide-react";

import {
  api,
  type LayoutUnit,
  type LayoutUnitPort,
  type LayoutUnitPortInput,
  type LayoutUnitPortKind
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";

type PortForm = LayoutUnitPortInput & { id?: string; version?: number };

const emptyPort: PortForm = {
  name: "", kind: "track", interfaceKey: "", xMm: 0, yMm: 0,
  directionDegrees: 0, notes: "", archived: false
};

export function LayoutModulePortsPanel({ unit, canPlan }: {
  unit: LayoutUnit | null;
  canPlan: boolean;
}) {
  const [ports, setPorts] = useState<LayoutUnitPort[]>([]);
  const [editing, setEditing] = useState<PortForm | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const { t } = useI18n();
  const unitID = unit?.id || "";
  const genericError = t("layouts.error.generic");

  const load = useCallback(async () => {
    if (!unitID) {
      setPorts([]);
      return;
    }
    setLoading(true);
    setMessage("");
    try {
      setPorts(await api.layoutUnitPorts(unitID));
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : genericError);
    } finally {
      setLoading(false);
    }
  }, [genericError, unitID]);

  useEffect(() => {
    setEditing(null);
    void load();
  }, [load]);

  const edit = (port: LayoutUnitPort) => setEditing({
    id: port.id, version: port.version, name: port.name, kind: port.kind,
    interfaceKey: port.interfaceKey, xMm: port.xMm, yMm: port.yMm,
    directionDegrees: port.directionDegrees, notes: port.notes || "", archived: port.archived
  });

  const save = async (event: FormEvent) => {
    event.preventDefault();
    if (!unit || !editing) return;
    setSaving(true);
    setMessage("");
    const input: LayoutUnitPortInput = {
      name: editing.name,
      kind: editing.kind,
      interfaceKey: editing.interfaceKey,
      xMm: Number(editing.xMm),
      yMm: Number(editing.yMm),
      directionDegrees: Number(editing.directionDegrees),
      notes: editing.notes?.trim() || undefined,
      archived: Boolean(editing.archived)
    };
    try {
      if (editing.id && editing.version) {
        await api.updateLayoutUnitPort(editing.id, { ...input, expectedVersion: editing.version });
      } else {
        await api.createLayoutUnitPort(unit.id, input);
      }
      setEditing(null);
      await load();
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally {
      setSaving(false);
    }
  };

  return <section className="panel layout-module-ports">
    <div className="layout-panel-head">
      <div className="panel-title"><Cable size={17} /><h3>{t("layouts.ports.title")}</h3></div>
      {canPlan && unit ? <button type="button" className="primary-button compact-action"
        onClick={() => setEditing({ ...emptyPort })}><Plus size={16} />{t("layouts.ports.create")}</button> : null}
    </div>
    {!unit ? <p className="layout-empty">{t("layouts.ports.noUnit")}</p> : <>
      <p className="layout-panel-subtitle">{t("layouts.ports.subtitle", { unit: unit.name })}</p>
      {message ? <p className="form-message" role="alert">{message}</p> : null}
      {loading ? <p>{t("layouts.ports.loading")}</p> : ports.length === 0
        ? <p className="layout-empty">{t("layouts.ports.empty")}</p>
        : <div className="table-wrap"><table className="layout-table layout-port-table">
          <thead><tr><th>{t("layouts.field.name")}</th><th>{t("layouts.ports.kind")}</th>
            <th>{t("layouts.ports.interface")}</th><th>{t("layouts.ports.position")}</th>
            <th>{t("layouts.ports.direction")}</th><th>{t("layouts.field.status")}</th>
            {canPlan ? <th><span className="sr-only">{t("layouts.plans.actions")}</span></th> : null}</tr></thead>
          <tbody>{ports.map((port) => <tr key={port.id}>
            <td><strong>{port.name}</strong>{port.notes ? <small>{port.notes}</small> : null}</td>
            <td>{t(`layouts.portKind.${port.kind}`)}</td><td className="layout-port-interface">{port.interfaceKey}</td>
            <td>{port.xMm} / {port.yMm} mm</td><td>{port.directionDegrees}°</td>
            <td><span className={port.archived ? "status-pill archived" : "status-pill"}>
              {t(port.archived ? "layouts.status.archived" : "layouts.status.active")}</span></td>
            {canPlan ? <td><button type="button" className="icon-button"
              aria-label={t("layouts.ports.editLabel", { name: port.name })} onClick={() => edit(port)}>
              <Pencil size={15} /></button></td> : null}
          </tr>)}</tbody>
        </table></div>}
      {canPlan && editing ? <form className="layout-form layout-port-form" onSubmit={save}>
        <label>{t("layouts.field.name")}<input required value={editing.name}
          onChange={(event) => setEditing({ ...editing, name: event.target.value })} /></label>
        <label>{t("layouts.ports.kind")}<AppSelect value={editing.kind} aria-label={t("layouts.ports.kind")}
          onChange={(event) => setEditing({ ...editing, kind: event.target.value as LayoutUnitPortKind })}>
          {(["track", "power", "digital", "feedback", "accessory", "other"] as const).map((kind) =>
            <option key={kind} value={kind}>{t(`layouts.portKind.${kind}`)}</option>)}</AppSelect></label>
        <label>{t("layouts.ports.interface")}<input required value={editing.interfaceKey}
          onChange={(event) => setEditing({ ...editing, interfaceKey: event.target.value })} /></label>
        <label>{t("layouts.ports.x")}<input type="number" min="0" max={unit.widthMm || undefined} step="0.1"
          value={editing.xMm} onChange={(event) => setEditing({ ...editing, xMm: Number(event.target.value) })} /></label>
        <label>{t("layouts.ports.y")}<input type="number" min="0" max={unit.heightMm || undefined} step="0.1"
          value={editing.yMm} onChange={(event) => setEditing({ ...editing, yMm: Number(event.target.value) })} /></label>
        <label>{t("layouts.ports.direction")}<input type="number" step="0.1" value={editing.directionDegrees}
          onChange={(event) => setEditing({ ...editing, directionDegrees: Number(event.target.value) })} /></label>
        <label className="layout-port-notes">{t("layouts.ports.notes")}<textarea value={editing.notes || ""}
          onChange={(event) => setEditing({ ...editing, notes: event.target.value })} /></label>
        {editing.id ? <label className="layout-check"><input type="checkbox" checked={Boolean(editing.archived)}
          onChange={(event) => setEditing({ ...editing, archived: event.target.checked })} />
          {t("layouts.field.archived")}</label> : null}
        <div className="layout-form-actions layout-port-actions">
          <button type="button" className="secondary-button" onClick={() => setEditing(null)}>{t("common.cancel")}</button>
          <button type="submit" className="primary-button" disabled={saving || !editing.name.trim() ||
            !editing.interfaceKey.trim()}>{saving ? t("common.saving") : t("layouts.ports.save")}</button>
        </div>
      </form> : null}
    </>}
  </section>;
}
