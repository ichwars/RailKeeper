import { PointerEvent as ReactPointerEvent, useCallback, useEffect, useRef, useState } from "react";
import { ArrowLeft, ExternalLink, PackageCheck, Pencil, Plus, RotateCcw, RotateCw, Trash2 } from "lucide-react";

import {
  ApiError,
  api,
  type LayoutUnit,
  type CreateFreePlanObjectInput,
  type PlanFreeObject,
  type PlanRevision,
  type PlanTrackObject,
  type FlexTrackPath,
  type TransitionCurvePath,
  type TrackGeometryDefinition,
  type TrackPlanAnalysis,
  type TrackPlanChangePreview,
  type TrackMaterialStatus
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { LayoutConfirmDialog, type LayoutPendingAction } from "./LayoutConfirmDialog";
import { FlexTrackEditorDialog } from "./FlexTrackEditorDialog";
import { FreePlanObjectDialog } from "./FreePlanObjectDialog";
import { FreePlanObjectInspector } from "./FreePlanObjectInspector";
import { FreePlanObjectLayer } from "./FreePlanObjectLayer";
import { TransitionCurveEditorDialog } from "./TransitionCurveEditorDialog";
import { TrackPlanAnalysisPanel } from "./TrackPlanAnalysisPanel";
import { TrackPlanChangePreviewPanel } from "./TrackPlanChangePreviewPanel";
import { TrackPlanReservationDialog } from "./TrackPlanReservationDialog";
import { trackGeometryLabel } from "./trackGeometryLabel";
import {
  normalizedRotation,
  routePolylinePoints,
  snapTrackPose,
  trackObjectTransform
} from "./trackPlannerGeometry";

type DragState = {
  object: PlanTrackObject;
  offsetX: number;
  offsetY: number;
};

type FreeDragState = {
  object: PlanFreeObject;
  offsetX: number;
  offsetY: number;
};

type TrackPortStatus = "connected" | "open" | "incompatible" | "unknown";

type TrackPathSelection = {
  flexPath: FlexTrackPath | null;
  transitionPath: TransitionCurvePath | null;
};

function defaultFlexPath(object: PlanTrackObject): FlexTrackPath {
  const end = object.effectiveGeometry.ports.find((port) => port.id === "b");
  const length = object.effectiveLengthMm || object.geometry.lengthMm;
  return {
    schemaVersion: 1,
    endXMm: end?.xMm ?? object.geometry.lengthMm,
    endYMm: end?.yMm ?? 0,
    endDirectionDegrees: end?.directionDegrees ?? 0,
    startHandleMm: length / 3,
    endHandleMm: length / 3
  };
}

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
  const [freeObjects, setFreeObjects] = useState<PlanFreeObject[]>([]);
  const [analysis, setAnalysis] = useState<TrackPlanAnalysis | null>(null);
  const [preview, setPreview] = useState<TrackPlanChangePreview | null>(null);
  const [reservationMaterial, setReservationMaterial] = useState<TrackMaterialStatus | null>(null);
  const [selectedID, setSelectedID] = useState("");
  const [selectedFreeObjectID, setSelectedFreeObjectID] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [conflict, setConflict] = useState(false);
  const [pending, setPending] = useState<LayoutPendingAction | null>(null);
  const [elevationStart, setElevationStart] = useState("0");
  const [elevationEnd, setElevationEnd] = useState("0");
  const [flexEditorID, setFlexEditorID] = useState("");
  const [transitionEditorID, setTransitionEditorID] = useState("");
  const [freeObjectDialog, setFreeObjectDialog] = useState<"" | "create" | "edit">("");
  const canvasRef = useRef<SVGSVGElement | null>(null);
  const dragRef = useRef<DragState | null>(null);
  const freeDragRef = useRef<FreeDragState | null>(null);
  const { language, t } = useI18n();
  const genericError = t("layouts.error.generic");
  const editable = canPlan && revision.status === "draft";
  const reservable = canPlan && (revision.status === "draft" || revision.status === "review");
  const width = unit.widthMm > 0 ? unit.widthMm : 1000;
  const height = unit.heightMm > 0 ? unit.heightMm : 600;
  const selected = objects.find((object) => object.id === selectedID);
  const selectedFreeObject = freeObjects.find((object) => object.id === selectedFreeObjectID);
  const analyzedGrade = analysis?.grades.find((grade) => grade.objectId === selected?.id);
  const selectedLength = selected?.effectiveLengthMm || selected?.geometry.lengthMm || 0;
  const selectedGrade = analyzedGrade?.gradePercent ?? (selected && selectedLength > 0
    ? (selected.elevationEndMm - selected.elevationStartMm) / selectedLength * 100
    : 0);
  const elevationStartValue = elevationStart.trim() === "" ? Number.NaN : Number(elevationStart);
  const elevationEndValue = elevationEnd.trim() === "" ? Number.NaN : Number(elevationEnd);
  const elevationValid = Number.isFinite(elevationStartValue) && Number.isFinite(elevationEndValue);
  const formatDecimal = (value: number, digits = 2) => new Intl.NumberFormat(
    language === "de" ? "de-DE" : "en-GB",
    { minimumFractionDigits: digits, maximumFractionDigits: digits }
  ).format(value);

  useEffect(() => {
    setElevationStart(selected ? String(selected.elevationStartMm) : "0");
    setElevationEnd(selected ? String(selected.elevationEndMm) : "0");
  }, [selected?.id, selected?.version]);

  const refreshDerived = useCallback(async () => {
    const [nextAnalysis, nextPreview] = await Promise.all([
      api.trackPlanAnalysis(revision.id), api.trackPlanChangePreview(revision.id)
    ]);
    setAnalysis(nextAnalysis); setPreview(nextPreview);
  }, [revision.id]);

  const load = useCallback(async () => {
    setLoading(true); setMessage(""); setConflict(false);
    try {
      const [nextGeometries, plan, nextAnalysis, nextPreview] = await Promise.all([
        api.trackGeometries(gauge), api.trackPlan(revision.id), api.trackPlanAnalysis(revision.id),
        api.trackPlanChangePreview(revision.id)
      ]);
      setGeometries(nextGeometries.filter((geometry) => geometry.status === "verified"));
      setObjects(plan.objects);
      setFreeObjects(plan.freeObjects);
      setAnalysis(nextAnalysis);
      setPreview(nextPreview);
      setSelectedID((current) => plan.objects.some((object) => object.id === current) ? current : "");
      setSelectedFreeObjectID((current) => plan.freeObjects.some((object) => object.id === current)
        ? current : "");
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
        rotationDegrees: 0,
        elevationStartMm: 0,
        elevationEndMm: 0,
        ...(geometry.kind === "flex" ? { flexPath: {
          schemaVersion: 1 as const,
          endXMm: geometry.lengthMm,
          endYMm: 0,
          endDirectionDegrees: 0,
          startHandleMm: geometry.lengthMm / 3,
          endHandleMm: geometry.lengthMm / 3
        } } : {})
      });
      setObjects((current) => [...current, created]);
      setSelectedID(created.id);
      await refreshDerived();
    } catch (reason) { showError(reason); }
    finally { setSaving(false); }
  };

  const update = async (object: PlanTrackObject, positionXMm: number, positionYMm: number,
    rotationDegrees: number, paths: TrackPathSelection = {
      flexPath: object.flexPath ?? null,
      transitionPath: object.transitionPath ?? null
    }) => {
    setSaving(true); setMessage(""); setConflict(false);
    try {
      const updated = await api.updatePlanTrackObject(object.id, {
        positionXMm, positionYMm, rotationDegrees,
        elevationStartMm: object.elevationStartMm, elevationEndMm: object.elevationEndMm,
        ...(object.geometry.kind === "flex" ? paths : {}),
        expectedVersion: object.version
      });
      setObjects((current) => current.map((item) => item.id === updated.id ? updated : item));
      await refreshDerived();
      return true;
    } catch (reason) { showError(reason); }
    finally { setSaving(false); }
    return false;
  };

  const updateFreeObject = async (object: PlanFreeObject, input: CreateFreePlanObjectInput) => {
    setSaving(true); setMessage(""); setConflict(false);
    try {
      const updated = await api.updateFreePlanObject(object.id, {
        ...input, expectedVersion: object.version
      });
      setFreeObjects((current) => current.map((item) => item.id === updated.id ? updated : item));
      await refreshDerived();
      return true;
    } catch (reason) { showError(reason); }
    finally { setSaving(false); }
    return false;
  };

  const createFreeObject = async (input: CreateFreePlanObjectInput) => {
    setSaving(true); setMessage(""); setConflict(false);
    try {
      const created = await api.createFreePlanObject(revision.id, input);
      setFreeObjects((current) => [...current, created]);
      setSelectedID("");
      setSelectedFreeObjectID(created.id);
      setFreeObjectDialog("");
      await refreshDerived();
    } catch (reason) { showError(reason); }
    finally { setSaving(false); }
  };

  const freeObjectInput = (object: PlanFreeObject, values: Partial<CreateFreePlanObjectInput> = {}) => ({
    name: object.name,
    category: object.category,
    positionXMm: object.positionXMm,
    positionYMm: object.positionYMm,
    rotationDegrees: object.rotationDegrees,
    shape: object.shape,
    ...values
  });

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
    event.currentTarget.setPointerCapture?.(event.pointerId);
    const point = canvasPoint(event.nativeEvent);
    dragRef.current = { object, offsetX: point.x - object.positionXMm, offsetY: point.y - object.positionYMm };
    setSelectedID(object.id);
    setSelectedFreeObjectID("");
  };

  const startFreeDrag = (event: ReactPointerEvent<SVGGElement>, object: PlanFreeObject) => {
    if (!editable || saving) return;
    event.preventDefault();
    event.currentTarget.setPointerCapture?.(event.pointerId);
    const point = canvasPoint(event.nativeEvent);
    freeDragRef.current = { object, offsetX: point.x - object.positionXMm, offsetY: point.y - object.positionYMm };
    setSelectedID("");
    setSelectedFreeObjectID(object.id);
  };

  const moveDrag = (event: ReactPointerEvent<SVGSVGElement>) => {
    const freeDrag = freeDragRef.current;
    if (freeDrag) {
      const point = canvasPoint(event.nativeEvent);
      const positionXMm = Math.max(0, Math.min(width, point.x - freeDrag.offsetX));
      const positionYMm = Math.max(0, Math.min(height, point.y - freeDrag.offsetY));
      setFreeObjects((current) => current.map((item) => item.id === freeDrag.object.id
        ? { ...item, positionXMm, positionYMm } : item));
      return;
    }
    const drag = dragRef.current;
    if (!drag) return;
    const point = canvasPoint(event.nativeEvent);
    const x = Math.max(0, Math.min(width, point.x - drag.offsetX));
    const y = Math.max(0, Math.min(height, point.y - drag.offsetY));
    const pose = snapTrackPose({ ...drag.object, positionXMm: x, positionYMm: y }, objects).pose;
    setObjects((current) => current.map((item) => item.id === drag.object.id
      ? { ...item, ...pose } : item));
  };

  const finishDrag = (event: ReactPointerEvent<SVGSVGElement>) => {
    const freeDrag = freeDragRef.current;
    if (freeDrag) {
      freeDragRef.current = null;
      const point = canvasPoint(event.nativeEvent);
      const positionXMm = Math.max(0, Math.min(width, point.x - freeDrag.offsetX));
      const positionYMm = Math.max(0, Math.min(height, point.y - freeDrag.offsetY));
      setFreeObjects((current) => current.map((item) => item.id === freeDrag.object.id
        ? { ...item, positionXMm, positionYMm } : item));
      if (positionXMm !== freeDrag.object.positionXMm || positionYMm !== freeDrag.object.positionYMm) {
        void updateFreeObject(freeDrag.object, freeObjectInput(freeDrag.object, { positionXMm, positionYMm }));
      }
      return;
    }
    const drag = dragRef.current;
    if (!drag) return;
    dragRef.current = null;
    const point = canvasPoint(event.nativeEvent);
    const x = Math.max(0, Math.min(width, point.x - drag.offsetX));
    const y = Math.max(0, Math.min(height, point.y - drag.offsetY));
    const pose = snapTrackPose({ ...drag.object, positionXMm: x, positionYMm: y }, objects).pose;
    setObjects((current) => current.map((item) => item.id === drag.object.id
      ? { ...item, ...pose } : item));
    if (pose.positionXMm === drag.object.positionXMm && pose.positionYMm === drag.object.positionYMm &&
      pose.rotationDegrees === drag.object.rotationDegrees) return;
    void update(drag.object, pose.positionXMm, pose.positionYMm, pose.rotationDegrees);
  };

  const rotate = (degrees: number) => {
    if (!selected) return;
    void update(selected, selected.positionXMm, selected.positionYMm,
      normalizedRotation(selected.rotationDegrees + degrees));
  };

  const rotateFreeObject = (degrees: number) => {
    if (!selectedFreeObject) return;
    void updateFreeObject(selectedFreeObject, freeObjectInput(selectedFreeObject, {
      rotationDegrees: normalizedRotation(selectedFreeObject.rotationDegrees + degrees)
    }));
  };

  const saveElevation = () => {
    if (!selected || !elevationValid) return;
    const object = { ...selected, elevationStartMm: elevationStartValue, elevationEndMm: elevationEndValue };
    void update(object, object.positionXMm, object.positionYMm, object.rotationDegrees);
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
        await refreshDerived();
      }
    });
  };

  const askDeleteFreeObject = () => {
    if (!selectedFreeObject) return;
    const object = selectedFreeObject;
    setPending({
      title: t("layouts.freeObject.deleteTitle"),
      body: t("layouts.freeObject.deleteBody", { name: object.name }),
      confirmLabel: t("common.delete"),
      dangerous: true,
      run: async () => {
        await api.deleteFreePlanObject(object.id, object.version);
        setFreeObjects((current) => current.filter((item) => item.id !== object.id));
        setSelectedFreeObjectID("");
        await refreshDerived();
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
        <button type="button" className="track-palette-item free-object-add" disabled={saving}
          onClick={() => setFreeObjectDialog("create")}><strong><Plus size={15} />
            {t("layouts.freeObject.add")}</strong></button>
        {geometries.map((geometry) => <button key={geometry.id} type="button" className="track-palette-item"
          aria-label={trackGeometryLabel(geometry)}
          disabled={saving} onClick={() => void place(geometry)}>
          <strong>{trackGeometryLabel(geometry)}</strong>
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
          <FreePlanObjectLayer objects={freeObjects} selectedID={selectedFreeObjectID || null}
            onSelect={(object) => { setSelectedID(""); setSelectedFreeObjectID(object.id); }}
            onPointerDown={startFreeDrag} />
          {objects.map((object) => {
            const effectiveGeometry = object.effectiveGeometry ?? object.geometry.geometry;
            const objectIssue = analysis?.issues.find((issue) =>
              issue.objectIds.includes(object.id) && ["overlap", "broken_geometry"].includes(issue.code));
            return <g key={object.id} role="button" tabIndex={0}
              aria-label={`Gleis ${trackGeometryLabel(object.geometry)}`}
              className={object.id === selectedID ? "track-object selected" : "track-object"}
              transform={trackObjectTransform(object)} onClick={() => {
                setSelectedFreeObjectID(""); setSelectedID(object.id);
              }}
              onPointerDown={(event) => startDrag(event, object)}>
              {effectiveGeometry.routes.map((route) => <polyline key={route.id}
                points={routePolylinePoints(route.points)} />)}
              {effectiveGeometry.ports.map((port) => {
                const status = trackPortStatus(analysis, object.id, port.id);
                return <g key={port.id} className={`track-port status-${status}`} data-status={status}>
                  <circle cx={port.xMm} cy={port.yMm} r="5" />
                  <text x={port.xMm} y={port.yMm} aria-hidden="true">{trackPortSymbols[status]}</text>
                </g>;
              })}
              {objectIssue ? <text className={`track-object-issue severity-${objectIssue.severity}`}
                x={(object.effectiveLengthMm || object.geometry.lengthMm) / 2} y={-12} aria-hidden="true">
                {objectIssue.code === "broken_geometry" ? "×" : "!"}
              </text> : null}
            </g>;
          })}
        </svg>
        {objects.length === 0 && freeObjects.length === 0 ? <p className="track-planner-empty">{editable
          ? t("layouts.trackPlanner.emptyDraft") : t("layouts.trackPlanner.emptyRead")}</p> : null}
      </div>
      <aside className="track-planner-inspector" aria-label={t("layouts.trackPlanner.inspector")}>
        <h4>{t("layouts.trackPlanner.inspector")}</h4>
        {selectedFreeObject ? <FreePlanObjectInspector object={selectedFreeObject} editable={editable}
          saving={saving} onEdit={() => setFreeObjectDialog("edit")}
          onRotate={rotateFreeObject} onDelete={askDeleteFreeObject} /> : selected ? <>
          <strong>{trackGeometryLabel(selected.geometry)}</strong>
          <dl><div><dt>{t("layouts.trackPlanner.position")}</dt>
            <dd>{selected.positionXMm.toFixed(1)} / {selected.positionYMm.toFixed(1)} mm</dd></div>
            <div><dt>{t("layouts.trackPlanner.rotation")}</dt><dd>{selected.rotationDegrees}°</dd></div>
            <div><dt>{t("layouts.trackPlanner.length")}</dt><dd>{formatDecimal(selectedLength)} mm</dd></div>
            {selected.geometry.kind === "flex" ? <div><dt>{t("layouts.flexEditor.radius")}</dt>
              <dd>{selected.effectiveMinimumRadiusMm == null ? t("layouts.flexEditor.radiusInfinite")
                : `${formatDecimal(selected.effectiveMinimumRadiusMm)} mm`}</dd></div> : null}
            {!editable ? <>
              <div><dt>{t("layouts.trackPlanner.elevationStart")}</dt>
                <dd>{formatDecimal(selected.elevationStartMm)} mm</dd></div>
              <div><dt>{t("layouts.trackPlanner.elevationEnd")}</dt>
                <dd>{formatDecimal(selected.elevationEndMm)} mm</dd></div>
            </> : null}
            <div><dt>{t("layouts.trackPlanner.grade")}</dt><dd>{formatDecimal(selectedGrade)} %</dd></div>
          </dl>
          {editable ? <div className="track-height-editor">
            <div className="track-height-fields">
              <AppNumberInput label={t("layouts.trackPlanner.elevationStart")} value={elevationStart}
                step="0.1" onValueChange={setElevationStart} disabled={saving} />
              <AppNumberInput label={t("layouts.trackPlanner.elevationEnd")} value={elevationEnd}
                step="0.1" onValueChange={setElevationEnd} disabled={saving} />
            </div>
            <button type="button" className="primary-button compact-action" disabled={saving || !elevationValid}
              onClick={saveElevation}>{t("layouts.trackPlanner.saveElevation")}</button>
          </div> : null}
          <a href={selected.geometry.sourceUrl} target="_blank" rel="noreferrer">
            {t("layouts.trackPlanner.source")}<ExternalLink size={13} />
          </a>
          {editable || reservable ? <div className="track-planner-actions">
            {editable ? <>
            <button type="button" className="secondary-button compact-action" disabled={saving}
              onClick={() => rotate(-15)}><RotateCcw size={14} />-15°</button>
            <button type="button" className="secondary-button compact-action" disabled={saving}
              onClick={() => rotate(15)}><RotateCw size={14} />+15°</button>
            {selected.geometry.kind === "flex" ? <button type="button"
              className="secondary-button compact-action" disabled={saving}
              onClick={() => setFlexEditorID(selected.id)}><Pencil size={14} />
              {t("layouts.flexEditor.open")}</button> : null}
            {selected.geometry.kind === "flex" ? <button type="button"
              className="secondary-button compact-action" disabled={saving}
              onClick={() => setTransitionEditorID(selected.id)}><Pencil size={14} />
              {t("layouts.transitionEditor.open")}</button> : null}
            <button type="button" className="danger-button compact-action" disabled={saving}
              onClick={askDelete}><Trash2 size={14} />{t("layouts.trackPlanner.delete")}</button>
            </> : null}
            {analysis?.reservations.some((item) => item.trackObjectId === selected.id)
              ? <span className="track-reservation-status">✓ {t("layouts.trackReservation.reserved")}</span>
              : <button type="button" className="secondary-button compact-action" disabled={saving}
                onClick={() => setReservationMaterial(analysis?.materials.find(
                  (material) => material.geometryId === selected.geometryId
                ) ?? null)}><PackageCheck size={14} />{t("layouts.trackReservation.open")}</button>}
          </div> : null}
        </> : <p className="layout-empty">{t("layouts.trackPlanner.select")}</p>}
      </aside>
      {analysis ? <TrackPlanAnalysisPanel analysis={analysis} selectedObjectId={selectedID}
        onSelectObject={(id) => { setSelectedFreeObjectID(""); setSelectedID(id); }} /> : null}
      {preview ? <TrackPlanChangePreviewPanel preview={preview} /> : null}
    </div>}
    <LayoutConfirmDialog action={pending} onClose={() => setPending(null)} />
    {reservationMaterial && selected ? <TrackPlanReservationDialog revisionId={revision.id}
      object={selected} material={reservationMaterial} onClose={() => setReservationMaterial(null)}
      onReserved={() => { setReservationMaterial(null); void refreshDerived(); }} /> : null}
    {flexEditorID && selected?.id === flexEditorID ? <FlexTrackEditorDialog
      object={selected.flexPath ? selected : { ...selected, flexPath: defaultFlexPath(selected) }}
      objects={objects} saving={saving} onClose={() => setFlexEditorID("")}
      onApply={async (path) => {
        if (await update(selected, selected.positionXMm, selected.positionYMm,
          selected.rotationDegrees, { flexPath: path, transitionPath: null })) setFlexEditorID("");
      }} /> : null}
    {transitionEditorID && selected?.id === transitionEditorID ? <TransitionCurveEditorDialog
      object={selected} saving={saving} onClose={() => setTransitionEditorID("")}
      onApply={async (path) => {
        if (await update(selected, selected.positionXMm, selected.positionYMm,
          selected.rotationDegrees, { flexPath: null, transitionPath: path })) setTransitionEditorID("");
      }} /> : null}
    {freeObjectDialog ? <FreePlanObjectDialog
      object={freeObjectDialog === "edit" ? selectedFreeObject : undefined}
      initialPosition={{ xMm: width / 2, yMm: height / 2 }} saving={saving}
      onClose={() => setFreeObjectDialog("")}
      onSubmit={async (input) => {
        if (freeObjectDialog === "edit" && selectedFreeObject) {
          const { expectedVersion: _expectedVersion, ...normalized } = input as typeof input & {
            expectedVersion: number;
          };
          if (await updateFreeObject(selectedFreeObject, normalized)) setFreeObjectDialog("");
          return;
        }
        await createFreeObject(input);
      }} /> : null}
  </section>;
}
