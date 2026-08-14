import { Check, Info, Map, Pencil, RefreshCw, TriangleAlert, X } from "lucide-react";
import { KeyboardEvent, PointerEvent, useEffect, useMemo, useRef, useState } from "react";

import {
  ApiError,
  api,
  type Layout,
  type LayoutConfiguration,
  type LayoutTwin,
  type LayoutTwinPosition,
  type LayoutTwinStatus,
  type LayoutUnit
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import { AppSelect } from "../../shared/ui/AppSelect";
import {
  screenToTwinPoint,
  twinGlobalToLocal,
  twinLocalToGlobal,
  twinPointInsideOutline,
  type TwinViewBox
} from "./layoutTwinGeometry";
import { LayoutTwinHistory } from "./LayoutTwinHistory";

const statuses: LayoutTwinStatus[] = ["planned", "reserved", "installed", "maintenance_due", "defective"];

function sourceSelection(source: string) {
  const [kind, id] = source.split(":", 2);
  return kind === "configuration" ? { configurationId: id } : { unitId: id };
}

function markerStatus(position: LayoutTwinPosition): LayoutTwinStatus {
  return [...position.statuses].reverse().find((status) => statuses.includes(status)) || "planned";
}

function positionSymbol(position: LayoutTwinPosition) {
  const symbols: Record<LayoutTwinPosition["kind"], string> = {
    turnout: "W", signal: "S", feedback: "R", decoder: "D",
    lighting: "L", power: "P", sensor: "F", other: "·"
  };
  return symbols[position.kind];
}

type TwinDrag =
  | { kind: "position"; id: string; unitID: string; pointerID: number }
  | { kind: "outline"; unitID: string; pointIndex: number; pointerID: number };

export function LayoutTwinPanel({ layout, units, configurations, canPlan }: {
  layout: Layout;
  units: LayoutUnit[];
  configurations: LayoutConfiguration[];
  canPlan: boolean;
}) {
  const { t } = useI18n();
  const sources = useMemo(() => [
    ...configurations.filter((configuration) => !configuration.archived).map((configuration) => ({
      value: `configuration:${configuration.id}`,
      label: t("layouts.twin.source.configuration", { name: configuration.name })
    })),
    ...units.filter((unit) => !unit.archived).map((unit) => ({
      value: `unit:${unit.id}`,
      label: t("layouts.twin.source.unit", { name: unit.name })
    }))
  ], [configurations, t, units]);
  const [source, setSource] = useState("");
  const [twin, setTwin] = useState<LayoutTwin | null>(null);
  const [serverTwin, setServerTwin] = useState<LayoutTwin | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [reloadKey, setReloadKey] = useState(0);
  const [activeStatuses, setActiveStatuses] = useState<Set<LayoutTwinStatus>>(() => new Set(statuses));
  const [hoveredID, setHoveredID] = useState("");
  const [selectedID, setSelectedID] = useState("");
  const [editing, setEditing] = useState(false);
  const [savingKey, setSavingKey] = useState("");
  const [editConflict, setEditConflict] = useState(false);
  const twinRef = useRef<LayoutTwin | null>(null);
  const dragRef = useRef<TwinDrag | null>(null);
  const saveTimerRef = useRef<number | null>(null);

  useEffect(() => {
    if (sources.some((option) => option.value === source)) return;
    setSource(sources[0]?.value || "");
  }, [source, sources]);

  useEffect(() => {
    if (!source) {
      setTwin(null);
      setLoading(false);
      return;
    }
    let active = true;
    setLoading(true);
    setError("");
    api.layoutTwin(layout.id, sourceSelection(source))
      .then((nextTwin) => {
        if (!active) return;
        setTwin(nextTwin);
        twinRef.current = nextTwin;
        setServerTwin(nextTwin);
        setEditConflict(false);
        setSelectedID((current) => nextTwin.units.some((unit) =>
          unit.positions.some((position) => position.id === current)) ? current : "");
      })
      .catch((reason: Error) => active && setError(reason.message))
      .finally(() => active && setLoading(false));
    return () => { active = false; };
  }, [layout.id, reloadKey, source]);

  useEffect(() => () => {
    if (saveTimerRef.current !== null) window.clearTimeout(saveTimerRef.current);
  }, []);

  useEffect(() => {
    setEditing(false);
    setEditConflict(false);
  }, [source]);

  useEffect(() => {
    if (!canPlan) setEditing(false);
  }, [canPlan]);

  const allPositions = twin?.units.flatMap((unit) => unit.positions) || [];
  const visiblePositions = allPositions.filter((position) =>
    position.statuses.some((status) => activeStatuses.has(status)));
  const hovered = allPositions.find((position) => position.id === hoveredID);
  const selected = allPositions.find((position) => position.id === selectedID);
  const selectedUnit = twin?.units.find((unit) => unit.id === selected?.layoutUnitId);
  const bounds = twin?.bounds;
  const baseWidth = Math.max(bounds?.widthMm || 0, 100);
  const baseHeight = Math.max(bounds?.heightMm || 0, 80);
  const padding = Math.max(baseWidth, baseHeight) * 0.06 + 20;
  const viewBoxModel: TwinViewBox = {
    x: (bounds?.minXMm || 0) - padding,
    y: (bounds?.minYMm || 0) - padding,
    width: baseWidth + padding * 2,
    height: baseHeight + padding * 2
  };
  const viewBox = `${viewBoxModel.x} ${viewBoxModel.y} ${viewBoxModel.width} ${viewBoxModel.height}`;
  const markerRadius = Math.max(8, Math.min(baseWidth, baseHeight) * 0.025);

  const toggleStatus = (status: LayoutTwinStatus, checked: boolean) => {
    setActiveStatuses((current) => {
      const next = new Set(current);
      if (checked) next.add(status); else next.delete(status);
      return next;
    });
  };
  const selectPosition = (position: LayoutTwinPosition) => setSelectedID(position.id);
  const applyTwin = (nextTwin: LayoutTwin) => {
    twinRef.current = nextTwin;
    setTwin(nextTwin);
  };

  const updatePositionDraft = (positionID: string, unitID: string, localX: number, localY: number) => {
    const current = twinRef.current;
    const unit = current?.units.find((item) => item.id === unitID);
    if (!current || !unit) return;
    const global = twinLocalToGlobal(unit, { xMm: localX, yMm: localY });
    applyTwin({ ...current, units: current.units.map((item) => item.id === unitID ? {
      ...item, positions: item.positions.map((position) => position.id === positionID ? {
        ...position, localXMm: localX, localYMm: localY, globalXMm: global.xMm, globalYMm: global.yMm,
        outsideOutline: !twinPointInsideOutline({ xMm: localX, yMm: localY }, unit.localOutline)
      } : position)
    } : item) });
  };

  const updateOutlineDraft = (unitID: string, pointIndex: number, localX: number, localY: number) => {
    const current = twinRef.current;
    const unit = current?.units.find((item) => item.id === unitID);
    if (!current || !unit) return;
    const localOutline = unit.localOutline.map((point, index) =>
      index === pointIndex ? { xMm: localX, yMm: localY } : point);
    const outline = localOutline.map((point) => twinLocalToGlobal(unit, point));
    applyTwin({ ...current, units: current.units.map((item) => item.id === unitID ? {
      ...item, localOutline, outline, positions: item.positions.map((position) => ({
        ...position,
        outsideOutline: !twinPointInsideOutline(
          { xMm: position.localXMm, yMm: position.localYMm },
          localOutline
        )
      }))
    } : item) });
  };

  const savePosition = async (positionID: string, unitID: string) => {
    const position = twinRef.current?.units.find((unit) => unit.id === unitID)?.positions
      .find((item) => item.id === positionID);
    if (!position) return;
    setSavingKey(`position:${positionID}`);
    try {
      await api.updateLayoutTechnicalPosition(positionID, {
        label: position.label, kind: position.kind, positionXMm: position.localXMm,
        positionYMm: position.localYMm, rotationDegrees: position.localRotationDegrees,
        productId: position.productId, description: position.description, expectedVersion: position.version
      });
      setReloadKey((current) => current + 1);
    } catch (reason) {
      if (reason instanceof ApiError && reason.status === 409) setEditConflict(true);
      else setError(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally {
      setSavingKey("");
    }
  };

  const saveOutline = async (unitID: string) => {
    const unit = twinRef.current?.units.find((item) => item.id === unitID);
    if (!unit) return;
    setSavingKey(`outline:${unitID}`);
    try {
      await api.updateLayoutUnitOutline(unitID, {
        points: unit.localOutline, expectedVersion: unit.version
      });
      setReloadKey((current) => current + 1);
    } catch (reason) {
      if (reason instanceof ApiError && reason.status === 409) setEditConflict(true);
      else setError(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally {
      setSavingKey("");
    }
  };

  const scheduleSave = (drag: TwinDrag) => {
    if (saveTimerRef.current !== null) window.clearTimeout(saveTimerRef.current);
    saveTimerRef.current = window.setTimeout(() => {
      saveTimerRef.current = null;
      if (drag.kind === "position") void savePosition(drag.id, drag.unitID);
      else void saveOutline(drag.unitID);
    }, 300);
  };

  const handlePointerMove = (event: PointerEvent<SVGSVGElement>) => {
    const drag = dragRef.current;
    if (!editing || !drag || drag.pointerID !== event.pointerId) return;
    const unit = twinRef.current?.units.find((item) => item.id === drag.unitID);
    if (!unit) return;
    const global = screenToTwinPoint(event.clientX, event.clientY,
      event.currentTarget.getBoundingClientRect(), viewBoxModel);
    const local = twinGlobalToLocal(unit, global);
    if (drag.kind === "position") updatePositionDraft(drag.id, drag.unitID, local.xMm, local.yMm);
    else updateOutlineDraft(drag.unitID, drag.pointIndex, local.xMm, local.yMm);
  };

  const finishPointerDrag = (event: PointerEvent<SVGSVGElement>) => {
    const drag = dragRef.current;
    if (!drag || drag.pointerID !== event.pointerId) return;
    dragRef.current = null;
    scheduleSave(drag);
  };

  const handlePositionKey = (event: KeyboardEvent<SVGGElement>, position: LayoutTwinPosition) => {
    const movements: Record<string, readonly [number, number]> = {
      ArrowLeft: [-1, 0], ArrowRight: [1, 0], ArrowUp: [0, -1], ArrowDown: [0, 1]
    };
    const movement = editing ? movements[event.key] : undefined;
    if (movement) {
      event.preventDefault();
      updatePositionDraft(position.id, position.layoutUnitId,
        position.localXMm + movement[0], position.localYMm + movement[1]);
      scheduleSave({ kind: "position", id: position.id, unitID: position.layoutUnitId, pointerID: -1 });
      return;
    }
    if (event.key !== "Enter" && event.key !== " ") return;
    event.preventDefault();
    selectPosition(position);
  };

  return <section className="panel layout-twin-panel">
    <div className="layout-panel-head layout-twin-head">
      <div className="panel-title"><Map size={17} /><div><h3>{t("layouts.twin.title")}</h3>
        <p>{t("layouts.twin.subtitle")}</p></div></div>
      <div className="layout-twin-head-actions">{canPlan ? <button type="button"
        className={editing ? "secondary-button compact-action active" : "secondary-button compact-action"}
        disabled={!source || loading || Boolean(savingKey)} onClick={() => setEditing((current) => !current)}>
        {editing ? <Check size={15} /> : <Pencil size={15} />}
        {t(editing ? "layouts.twin.edit.done" : "layouts.twin.edit.action")}</button> : null}
        <button type="button" className="icon-button" aria-label={t("layouts.twin.reload")}
          title={t("layouts.twin.reload")} disabled={!source || loading}
          onClick={() => setReloadKey((current) => current + 1)}><RefreshCw size={16} /></button></div>
    </div>
    {sources.length === 0 ? <div className="layout-twin-empty"><Info size={20} />
      <p>{t("layouts.twin.noSources")}</p></div> : <>
      <div className="layout-twin-toolbar">
        <label className="app-field layout-twin-source"><span className="app-field-label">
          {t("layouts.twin.source.label")}</span>
          <AppSelect value={source} aria-label={t("layouts.twin.source.label")}
            onChange={(event) => { setSource(event.target.value); setSelectedID(""); }}>
            {sources.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
          </AppSelect>
        </label>
        <fieldset className="layout-twin-filters"><legend>{t("layouts.twin.filters")}</legend>
          <div>{statuses.map((status) => <AppCheckbox key={status}
            className={`layout-twin-filter status-${status}`}
            label={`${t(`layouts.twin.status.${status}`)} (${allPositions.filter((position) =>
              position.statuses.includes(status)).length})`}
            checked={activeStatuses.has(status)} onChange={(event) => toggleStatus(status, event.target.checked)} />)}</div>
        </fieldset>
      </div>
      {error ? <div className="form-message" role="alert">{error}</div> : null}
      {editing ? <div className="layout-twin-edit-note"><Pencil size={15} />
        <span>{savingKey ? t("layouts.twin.edit.saving") : t("layouts.twin.edit.note")}</span></div> : null}
      {editConflict ? <div className="layout-conflict" role="alert"><TriangleAlert size={16} />
        <span>{t("layouts.twin.edit.conflict")}</span>
        <button type="button" className="secondary-button compact-action"
          onClick={() => setReloadKey((current) => current + 1)}>{t("layouts.conflict.reload")}</button>
        <button type="button" className="secondary-button compact-action" onClick={() => {
          if (serverTwin) applyTwin(serverTwin); setEditConflict(false);
        }}>{t("layouts.twin.edit.discard")}</button></div> : null}
      {loading && !twin ? <div className="layout-twin-empty"><p>{t("layouts.twin.loading")}</p></div> :
        twin && (!twin.hasGeometry || twin.units.length === 0) ? <div className="layout-twin-empty">
          <TriangleAlert size={20} /><p>{t("layouts.twin.noGeometry")}</p></div> : twin ?
          <div className={`layout-twin-body${selected ? " inspector-open" : ""}`}>
            <div className="layout-twin-stage">
              <svg className={`layout-twin-canvas${editing ? " editing" : ""}`} viewBox={viewBox} role="img"
                aria-label={t("layouts.twin.canvasLabel")} preserveAspectRatio="xMidYMid meet"
                onPointerMove={handlePointerMove} onPointerUp={finishPointerDrag}
                onPointerCancel={finishPointerDrag}>
                {twin.units.map((unit) => unit.outline.length > 2 ? <g key={unit.id}>
                  <polygon className="layout-twin-unit" points={unit.outline.map((point) =>
                    `${point.xMm},${point.yMm}`).join(" ")} />
                  <text className="layout-twin-unit-label" x={unit.outline[0]?.xMm} y={unit.outline[0]?.yMm}
                    dx={8} dy={18}>{unit.name}</text>
                  {editing ? unit.outline.map((point, pointIndex) => <circle
                    key={`${unit.id}:${pointIndex}`} className="layout-twin-outline-handle"
                    aria-label={t("layouts.twin.edit.outlinePoint", { point: pointIndex + 1, unit: unit.name })}
                    cx={point.xMm} cy={point.yMm} r={markerRadius * 0.62}
                    onPointerDown={(event) => {
                      event.preventDefault();
                      dragRef.current = { kind: "outline", unitID: unit.id,
                        pointIndex, pointerID: event.pointerId };
                      event.currentTarget.ownerSVGElement?.setPointerCapture?.(event.pointerId);
                    }} />) : null}
                </g> : null)}
                {visiblePositions.map((position) => <g key={position.id} role="button" tabIndex={0}
                  aria-label={`${position.label}, ${position.statuses.map((status) =>
                    t(`layouts.twin.status.${status}`)).join(", ")}` +
                    (position.outsideOutline ? `, ${t("layouts.twin.edit.outside")}` : "")}
                  className={`layout-twin-marker status-${markerStatus(position)}` +
                    (selectedID === position.id ? " selected" : "") +
                    (position.outsideOutline ? " outside" : "")}
                  transform={`translate(${position.globalXMm} ${position.globalYMm})`}
                  onMouseEnter={() => setHoveredID(position.id)} onMouseLeave={() => setHoveredID("")}
                  onFocus={() => setHoveredID(position.id)} onBlur={() => setHoveredID("")}
                  onPointerDown={(event) => {
                    if (!editing || savingKey === `position:${position.id}`) return;
                    event.preventDefault();
                    dragRef.current = { kind: "position", id: position.id,
                      unitID: position.layoutUnitId, pointerID: event.pointerId };
                    event.currentTarget.ownerSVGElement?.setPointerCapture?.(event.pointerId);
                  }}
                  onClick={() => selectPosition(position)}
                  onKeyDown={(event) => handlePositionKey(event, position)}>
                  <circle r={markerRadius} /><text textAnchor="middle" dominantBaseline="central"
                    fontSize={markerRadius * 1.05}>{positionSymbol(position)}</text>
                </g>)}
              </svg>
              <div className="layout-twin-stage-summary" aria-live="polite">
                {hovered ? <><strong>{hovered.label}</strong><span>{hovered.statuses.map((status) =>
                  t(`layouts.twin.status.${status}`)).join(" · ")}</span>
                  <small>{hovered.productName || t(`layouts.positionKind.${hovered.kind}`)}</small></> :
                  <span>{t("layouts.twin.visibleSummary", {
                    visible: visiblePositions.length, total: allPositions.length
                  })}</span>}
              </div>
            </div>
            {selected ? <LayoutTwinInspector position={selected} unitName={selectedUnit?.name || ""}
              onClose={() => setSelectedID("")} /> : null}
          </div> : null}
      {twin?.warnings.some((warning) => warning.code === "missing_geometry") ?
        <p className="layout-twin-warning"><TriangleAlert size={15} />{t("layouts.twin.geometryWarning")}</p> : null}
    </>}
  </section>;
}

function LayoutTwinInspector({ position, unitName, onClose }: {
  position: LayoutTwinPosition;
  unitName: string;
  onClose: () => void;
}) {
  const { t } = useI18n();
  const allocations = [...position.reservations.map((allocation) => ({ ...allocation, type: "reservation" as const })),
    ...position.installations.map((allocation) => ({ ...allocation, type: "installation" as const }))];
  return <aside className="layout-twin-inspector" aria-label={t("layouts.twin.inspector.title")}>
    <header><div><p className="eyebrow">{t(`layouts.positionKind.${position.kind}`)}</p>
      <h4>{position.label}</h4></div><button type="button" className="icon-button"
        aria-label={t("common.close")} title={t("common.close")} onClick={onClose}><X size={17} /></button></header>
    <div className="layout-twin-inspector-body">
      <div className="layout-twin-statuses">{position.statuses.map((status) => <span key={status}
        className={`status-pill status-${status}`}>{t(`layouts.twin.status.${status}`)}</span>)}</div>
      {position.outsideOutline ? <p className="layout-twin-position-warning"><TriangleAlert size={15} />
        {t("layouts.twin.edit.outsideWarning")}</p> : null}
      <dl><div><dt>{t("layouts.twin.inspector.unit")}</dt><dd>{unitName}</dd></div>
        <div><dt>{t("layouts.twin.inspector.coordinates")}</dt>
          <dd>{position.localXMm} / {position.localYMm} mm</dd></div>
        <div><dt>{t("layouts.field.rotation")}</dt><dd>{position.rotationDegrees}°</dd></div></dl>
      {position.description ? <p>{position.description}</p> : null}
      <section><h5>{t("layouts.twin.inspector.article")}</h5>{position.productId ? <dl>
        <div><dt>{t("accessories.field.inventoryNumber")}</dt><dd>{position.inventoryNumber || "–"}</dd></div>
        <div><dt>{t("accessories.field.manufacturer")}</dt><dd>{position.manufacturer || "–"}</dd></div>
        <div><dt>{t("accessories.field.articleNumber")}</dt><dd>{position.articleNumber || "–"}</dd></div>
        <div><dt>{t("accessories.field.name")}</dt><dd>{position.productName || "–"}</dd></div>
      </dl> : <p className="layout-empty">{t("layouts.technology.noArticle")}</p>}</section>
      <section><h5>{t("layouts.twin.inspector.allocations")}</h5>{allocations.length ?
        <div className="layout-twin-allocation-list">{allocations.map((allocation) => <article key={allocation.id}>
          <strong>{t(`layouts.twin.allocation.${allocation.type}`)}</strong>
          <span>{allocation.quantity} × {allocation.manufacturer} {allocation.articleNumber || allocation.productName}</span>
          {allocation.placement ? <small>{t("layouts.twin.inspector.placement")}: {allocation.placement}</small> : null}
          {allocation.digitalAddress ? <small>{t("layouts.twin.inspector.digitalAddress")}: {allocation.digitalAddress}</small> : null}
          {allocation.decoderOutput ? <small>{t("layouts.twin.inspector.decoderOutput")}: {allocation.decoderOutput}</small> : null}
          {allocation.connection ? <small>{t("layouts.twin.inspector.connection")}: {allocation.connection}</small> : null}
          {allocation.wiringNotes ? <small>{t("layouts.twin.inspector.wiringNotes")}: {allocation.wiringNotes}</small> : null}
        </article>)}</div> : <p className="layout-empty">{t("layouts.twin.inspector.noAllocations")}</p>}</section>
      <section><h5>{t("layouts.twin.history.title")}</h5>
        <LayoutTwinHistory key={position.id} positionID={position.id} productID={position.productId} /></section>
    </div>
  </aside>;
}
