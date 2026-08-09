import { PointerEvent as ReactPointerEvent, useCallback, useEffect, useRef, useState } from "react";
import { ArrowLeft, ExternalLink, RotateCcw, RotateCw, Trash2 } from "lucide-react";

import {
  ApiError,
  api,
  type LayoutUnit,
  type PlanRevision,
  type PlanTrackObject,
  type TrackGeometryDefinition,
  type TrackPlanAnalysis
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { LayoutConfirmDialog, type LayoutPendingAction } from "./LayoutConfirmDialog";
import { TrackPlanAnalysisPanel } from "./TrackPlanAnalysisPanel";
import {
  findTrackSnap,
  normalizedRotation,
  routePolylinePoints,
  trackObjectTransform
} from "./trackPlannerGeometry";

type DragState = {
  object: PlanTrackObject;
  offsetX: number;
  offsetY: number;
};

type TrackPortStatus = "connected" | "open" | "incompatible" | "unknown";

const trackPortSymbols: Record<TrackPortStatus, string> = {
  connected: "✓",
  open: "○",
  incompatible: "!",
  unknown: "·"
};

function trackPortStatus(
  analysis: TrackPlanAnalysis | null,
  objectID: string,
  portID: string
): TrackPortStatus {
  if (!analysis) return "unknown";
  if (analysis.connections.some((connection) =>
    (connection.objectAId === objectID && connection.portAId === portID) ||
    (connection.objectBId === objectID && connection.portBId === portID))) return "connected";
  for (const issue of analysis.issues) {
    const objectIndex = issue.objectIds.indexOf(objectID);
    if (objectIndex < 0 || !issue.portIds?.includes(portID)) continue;
    if (issue.portIds.length === issue.objectIds.length && issue.portIds[objectIndex] !== portID) continue;
    if (issue.code === "incompatible_connection") return "incompatible";
    if (issue.code === "open_end") return "open";
  }
  return "unknown";
}

export function TrackPlannerCanvas({ unit, gauge, revision, canPlan, onClose }: {
  unit: LayoutUnit;
  gauge: string;
  revision: PlanRevision;
  canPlan: boolean;
  onClose: () => void;
}) {
  const [geometries, setGeometries] = useState<TrackGeometryDefinition[]>([]);
  const [objects, setObjects] = useState<PlanTrackObject[]>([]);
  const [analysis, setAnalysis] = useState<TrackPlanAnalysis | null>(null);
  const [selectedID, setSelectedID] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [conflict, setConflict] = useState(false);
  const [pending, setPending] = useState<LayoutPendingAction | null>(null);
  const canvasRef = useRef<SVGSVGElement | null>(null);
  const dragRef = useRef<DragState | null>(null);
  const { t } = useI18n();
  const genericError = t("layouts.error.generic");
  const editable = canPlan && revision.status === "draft";
  const width = unit.widthMm > 0 ? unit.widthMm : 1000;
  const height = unit.heightMm > 0 ? unit.heightMm : 600;
  const selected = objects.find((object) => object.id === selectedID);

  const refreshAnalysis = useCallback(async () => {
    const nextAnalysis = await api.trackPlanAnalysis(revision.id);
    setAnalysis(nextAnalysis);
  }, [revision.id]);

  const load = useCallback(async () => {
    setLoading(true); setMessage(""); setConflict(false);
    try {
      const [nextGeometries, plan, nextAnalysis] = await Promise.all([
        api.trackGeometries(gauge), api.trackPlan(revision.id), api.trackPlanAnalysis(revision.id)
      ]);
      setGeometries(nextGeometries.filter((geometry) => geometry.status === "verified"));
      setObjects(plan.objects);
      setAnalysis(nextAnalysis);
      setSelectedID((current) => plan.objects.some((object) => object.id === current) ? current : "");
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : genericError);
    } finally {
      setLoading(false);
    }
  }, [gauge, revision.id, revision.status, genericError]);

  useEffect(() => { void load(); }, [load]);

  const showError = (reason: unknown) => {
    const changed = reason instanceof ApiError && reason.status === 409;
    setConflict(changed);
    setMessage(changed ? t("layouts.trackPlanner.conflict")
      : reason instanceof Error ? reason.message : genericError);
  };

  const place = async (geometry: TrackGeometryDefinition) => {
    setSaving(true); setMessage(""); setConflict(false);
    try {
      const created = await api.createPlanTrackObject(revision.id, {
        geometryId: geometry.id,
        positionXMm: Math.max(0, width / 2 - geometry.lengthMm / 2),
        positionYMm: height / 2,
        rotationDegrees: 0
      });
      setObjects((current) => [...current, created]);
      setSelectedID(created.id);
      await refreshAnalysis();
    } catch (reason) { showError(reason); }
    finally { setSaving(false); }
  };

  const update = async (object: PlanTrackObject, positionXMm: number, positionYMm: number,
    rotationDegrees: number) => {
    setSaving(true); setMessage(""); setConflict(false);
    try {
      const updated = await api.updatePlanTrackObject(object.id, {
        positionXMm, positionYMm, rotationDegrees, expectedVersion: object.version
      });
      setObjects((current) => current.map((item) => item.id === updated.id ? updated : item));
      await refreshAnalysis();
    } catch (reason) { showError(reason); }
    finally { setSaving(false); }
  };

  const canvasPoint = (event: Pick<PointerEvent, "clientX" | "clientY">) => {
    const bounds = canvasRef.current?.getBoundingClientRect();
    if (!bounds || bounds.width <= 0 || bounds.height <= 0) return { x: 0, y: 0 };
    return {
      x: (event.clientX - bounds.left) * width / bounds.width,
      y: (event.clientY - bounds.top) * height / bounds.height
    };
  };

  const startDrag = (event: ReactPointerEvent<SVGGElement>, object: PlanTrackObject) => {
    if (!editable || saving) return;
    event.preventDefault();
    const point = canvasPoint(event.nativeEvent);
    dragRef.current = { object, offsetX: point.x - object.positionXMm, offsetY: point.y - object.positionYMm };
    setSelectedID(object.id);
  };

  const moveDrag = (event: ReactPointerEvent<SVGSVGElement>) => {
    const drag = dragRef.current;
    if (!drag) return;
    const point = canvasPoint(event.nativeEvent);
    const x = Math.max(0, Math.min(width, point.x - drag.offsetX));
    const y = Math.max(0, Math.min(height, point.y - drag.offsetY));
    const pose = findTrackSnap({ ...drag.object, positionXMm: x, positionYMm: y }, objects).pose;
    setObjects((current) => current.map((item) => item.id === drag.object.id
      ? { ...item, ...pose } : item));
  };

  const finishDrag = (event: ReactPointerEvent<SVGSVGElement>) => {
    const drag = dragRef.current;
    if (!drag) return;
    dragRef.current = null;
    const point = canvasPoint(event.nativeEvent);
    const x = Math.max(0, Math.min(width, point.x - drag.offsetX));
    const y = Math.max(0, Math.min(height, point.y - drag.offsetY));
    const pose = findTrackSnap({ ...drag.object, positionXMm: x, positionYMm: y }, objects).pose;
    setObjects((current) => current.map((item) => item.id === drag.object.id
      ? { ...item, ...pose } : item));
    void update(drag.object, pose.positionXMm, pose.positionYMm, pose.rotationDegrees);
  };

  const rotate = (degrees: number) => {
    if (!selected) return;
    void update(selected, selected.positionXMm, selected.positionYMm,
      normalizedRotation(selected.rotationDegrees + degrees));
  };

  const askDelete = () => {
    if (!selected) return;
    const object = selected;
    setPending({
      title: t("layouts.trackPlanner.deleteTitle"),
      body: t("layouts.trackPlanner.deleteBody", { article: object.geometry.articleNumber }),
      confirmLabel: t("common.delete"),
      dangerous: true,
      run: async () => {
        await api.deletePlanTrackObject(object.id, object.version);
        setObjects((current) => current.filter((item) => item.id !== object.id));
        setSelectedID("");
        await refreshAnalysis();
      }
    });
  };

  return <section className="panel track-planner">
    <header className="track-planner-head">
      <div>
        <button type="button" className="icon-button" aria-label={t("layouts.trackPlanner.back")} onClick={onClose}>
          <ArrowLeft size={17} />
        </button>
        <div><p className="eyebrow">{t(`layouts.revisionStatus.${revision.status}`)}</p>
          <h3>{t("layouts.trackPlanner.title", { unit: unit.name, revision: revision.revisionNumber })}</h3></div>
      </div>
      <span className={`status-pill revision-${revision.status}`}>
        {editable ? t("layouts.trackPlanner.editMode") : t("layouts.trackPlanner.readMode")}
      </span>
    </header>
    {message ? <div className={conflict ? "layout-conflict" : "form-message"}><span>{message}</span>
      {conflict ? <button type="button" className="secondary-button compact-action" onClick={() => void load()}>
        {t("layouts.conflict.reload")}</button> : null}</div> : null}
    {loading ? <p className="layout-empty">{t("layouts.trackPlanner.loading")}</p> : <div className="track-planner-body">
      {editable ? <aside className="track-planner-palette" aria-label={t("layouts.trackPlanner.palette")}>
        <h4>{t("layouts.trackPlanner.palette")}</h4>
        <p>{t("layouts.trackPlanner.paletteHint")}</p>
        {geometries.map((geometry) => <button key={geometry.id} type="button" className="track-palette-item"
          aria-label={`Tillig ${geometry.articleNumber} · ${geometry.name}`}
          disabled={saving} onClick={() => void place(geometry)}>
          <strong>Tillig {geometry.articleNumber} · {geometry.name}</strong>
          <small>{geometry.lengthMm} mm · {t(`layouts.trackPlanner.kind.${geometry.kind}`)}</small>
        </button>)}
      </aside> : null}
      <div className="track-planner-stage">
        <svg ref={canvasRef} className="track-planner-canvas" viewBox={`0 0 ${width} ${height}`}
          role="img" aria-label={t("layouts.trackPlanner.canvasLabel", { unit: unit.name })}
          onPointerMove={moveDrag} onPointerUp={finishDrag}>
          <defs><pattern id={`track-grid-${revision.id}`} width="50" height="50" patternUnits="userSpaceOnUse">
            <path d="M 50 0 L 0 0 0 50" className="track-grid-line" /></pattern></defs>
          <rect width={width} height={height} className="track-canvas-surface" />
          <rect width={width} height={height} fill={`url(#track-grid-${revision.id})`} />
          {objects.map((object) => {
            const objectIssue = analysis?.issues.find((issue) =>
              issue.objectIds.includes(object.id) && ["overlap", "broken_geometry"].includes(issue.code));
            return <g key={object.id} role="button" tabIndex={0}
              aria-label={`Gleis Tillig ${object.geometry.articleNumber} G1`}
              className={object.id === selectedID ? "track-object selected" : "track-object"}
              transform={trackObjectTransform(object)} onClick={() => setSelectedID(object.id)}
              onPointerDown={(event) => startDrag(event, object)}>
              {object.geometry.geometry.routes.map((route) => <polyline key={route.id}
                points={routePolylinePoints(route.points)} />)}
              {object.geometry.geometry.ports.map((port) => {
                const status = trackPortStatus(analysis, object.id, port.id);
                return <g key={port.id} className={`track-port status-${status}`} data-status={status}>
                  <circle cx={port.xMm} cy={port.yMm} r="5" />
                  <text x={port.xMm} y={port.yMm} aria-hidden="true">{trackPortSymbols[status]}</text>
                </g>;
              })}
              {objectIssue ? <text className={`track-object-issue severity-${objectIssue.severity}`}
                x={object.geometry.lengthMm / 2} y={-12} aria-hidden="true">
                {objectIssue.code === "broken_geometry" ? "×" : "!"}
              </text> : null}
            </g>;
          })}
        </svg>
        {objects.length === 0 ? <p className="track-planner-empty">{editable
          ? t("layouts.trackPlanner.emptyDraft") : t("layouts.trackPlanner.emptyRead")}</p> : null}
      </div>
      <aside className="track-planner-inspector" aria-label={t("layouts.trackPlanner.inspector")}>
        <h4>{t("layouts.trackPlanner.inspector")}</h4>
        {selected ? <>
          <strong>Tillig {selected.geometry.articleNumber} · {selected.geometry.name}</strong>
          <dl><div><dt>{t("layouts.trackPlanner.position")}</dt>
            <dd>{selected.positionXMm.toFixed(1)} / {selected.positionYMm.toFixed(1)} mm</dd></div>
            <div><dt>{t("layouts.trackPlanner.rotation")}</dt><dd>{selected.rotationDegrees}°</dd></div>
            <div><dt>{t("layouts.trackPlanner.length")}</dt><dd>{selected.geometry.lengthMm} mm</dd></div></dl>
          <a href={selected.geometry.sourceUrl} target="_blank" rel="noreferrer">
            {t("layouts.trackPlanner.source")}<ExternalLink size={13} />
          </a>
          {editable ? <div className="track-planner-actions">
            <button type="button" className="secondary-button compact-action" disabled={saving}
              onClick={() => rotate(-15)}><RotateCcw size={14} />-15°</button>
            <button type="button" className="secondary-button compact-action" disabled={saving}
              onClick={() => rotate(15)}><RotateCw size={14} />+15°</button>
            <button type="button" className="danger-button compact-action" disabled={saving}
              onClick={askDelete}><Trash2 size={14} />{t("layouts.trackPlanner.delete")}</button>
          </div> : null}
        </> : <p className="layout-empty">{t("layouts.trackPlanner.select")}</p>}
      </aside>
      {analysis ? <TrackPlanAnalysisPanel analysis={analysis} /> : null}
    </div>}
    <LayoutConfirmDialog action={pending} onClose={() => setPending(null)} />
  </section>;
}
