import { AlertTriangle, X } from "lucide-react";
import { KeyboardEvent, PointerEvent as ReactPointerEvent, useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";

import {
  api,
  type FlexTrackPath,
  type FlexTrackPreview,
  type PlanTrackObject,
  type TrackPoint
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { flexPreviewBounds, routePolylinePoints, snapFlexEnd } from "./trackPlannerGeometry";
import { sampleFlexPath } from "./flexTrackGeometry";

type FlexTrackForm = {
  endXMm: string;
  endYMm: string;
  endDirectionDegrees: string;
};

const focusableSelector = "button:not([disabled]), input:not([disabled]), [tabindex]:not([tabindex='-1'])";

function finiteForm(form: FlexTrackForm) {
  const values = [form.endXMm, form.endYMm, form.endDirectionDegrees].map((value) => Number(value));
  return values.every(Number.isFinite) && Math.hypot(values[0], values[1]) > 0;
}

export function FlexTrackEditorDialog({ object, objects = [], saving, onApply, onClose }: {
  object: PlanTrackObject;
  objects?: PlanTrackObject[];
  saving: boolean;
  onApply: (path: FlexTrackPath) => void | Promise<void>;
  onClose: () => void;
}) {
  const initial = object.flexPath!;
  const [form, setForm] = useState<FlexTrackForm>({
    endXMm: String(initial.endXMm),
    endYMm: String(initial.endYMm),
    endDirectionDegrees: String(initial.endDirectionDegrees)
  });
  const [preview, setPreview] = useState<FlexTrackPreview | null>(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const [draggingEnd, setDraggingEnd] = useState(false);
  const layerRef = useRef<HTMLDivElement | null>(null);
  const previewRef = useRef<SVGSVGElement | null>(null);
  const { language, t } = useI18n();
  const valid = finiteForm(form);
  const localPath = useMemo<FlexTrackPath>(() => ({
    ...initial,
    endXMm: Number(form.endXMm),
    endYMm: Number(form.endYMm),
    endDirectionDegrees: Number(form.endDirectionDegrees)
  }), [form, initial]);
  const localPoints = valid ? sampleFlexPath(localPath) : [];
  const previewPoints = preview?.effectiveGeometry.routes[0]?.points ?? [];
  const bounds = flexPreviewBounds(previewPoints.length > 0 ? previewPoints : localPoints);
  const numberFormat = new Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  });

  useEffect(() => {
    const first = layerRef.current?.querySelector<HTMLElement>(focusableSelector);
    first?.focus();
  }, []);

  const update = (field: keyof FlexTrackForm, value: string) => {
    setForm((current) => ({ ...current, [field]: value }));
    setPreview(null);
    setMessage("");
  };

  const moveEnd = (event: ReactPointerEvent<SVGSVGElement>) => {
    if (!draggingEnd || !valid) return;
    const rectangle = previewRef.current?.getBoundingClientRect();
    if (!rectangle || rectangle.width <= 0 || rectangle.height <= 0) return;
    const moved = {
      ...localPath,
      endXMm: bounds.minX + (event.clientX - rectangle.left) * bounds.width / rectangle.width,
      endYMm: bounds.minY + (event.clientY - rectangle.top) * bounds.height / rectangle.height
    };
    const snapped = snapFlexEnd(object, moved, objects).path;
    setForm((current) => ({
      ...current,
      endXMm: snapped.endXMm.toFixed(3),
      endYMm: snapped.endYMm.toFixed(3),
      endDirectionDegrees: snapped.endDirectionDegrees.toFixed(3)
    }));
    setPreview(null);
  };

  const suggest = async () => {
    if (!valid) return;
    setLoading(true);
    setMessage("");
    try {
      setPreview(await api.previewFlexTrackPath(object.id, {
        endXMm: Number(form.endXMm),
        endYMm: Number(form.endYMm),
        endDirectionDegrees: Number(form.endDirectionDegrees),
        expectedVersion: object.version
      }));
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      if (!saving && !loading) onClose();
      return;
    }
    if (event.key !== "Tab") return;
    const focusable = Array.from(layerRef.current?.querySelectorAll<HTMLElement>(focusableSelector) ?? []);
    if (focusable.length === 0) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if ((!event.shiftKey && event.target === last) || (event.shiftKey && event.target === first)) {
      event.preventDefault();
      (event.shiftKey ? last : first).focus();
    }
  };

  const renderRoute = (points: TrackPoint[], className: string) => points.length > 1
    ? <polyline className={className} points={routePolylinePoints(points)} /> : null;

  return createPortal(<div ref={layerRef} className="modal-layer flex-track-editor-layer" role="dialog"
    aria-modal="true" aria-label={t("layouts.flexEditor.title")} onKeyDown={handleKeyDown}>
    <section className="vehicle-modal flex-track-editor-dialog">
      <header className="modal-head">
        <div><h2>{t("layouts.flexEditor.title")}</h2>
          <p>Tillig {object.geometry.articleNumber} · {object.geometry.name}</p></div>
        <button type="button" className="icon-button" aria-label={t("common.close")}
          disabled={saving || loading} onClick={onClose}><X size={18} /></button>
      </header>
      <div className="modal-body flex-track-editor-body">
        <div className="flex-track-editor-fields">
          <AppNumberInput label={t("layouts.flexEditor.endX")} value={form.endXMm} step="0.1"
            disabled={saving || loading} onValueChange={(value) => update("endXMm", value)} />
          <AppNumberInput label={t("layouts.flexEditor.endY")} value={form.endYMm} step="0.1"
            disabled={saving || loading} onValueChange={(value) => update("endYMm", value)} />
          <AppNumberInput label={t("layouts.flexEditor.endDirection")} value={form.endDirectionDegrees}
            step="0.1" disabled={saving || loading}
            onValueChange={(value) => update("endDirectionDegrees", value)} />
        </div>
        <svg ref={previewRef} className="flex-track-preview" role="img"
          aria-label={t("layouts.flexEditor.preview")}
          viewBox={`${bounds.minX} ${bounds.minY} ${bounds.width} ${bounds.height}`}
          onPointerMove={moveEnd} onPointerUp={() => setDraggingEnd(false)}
          onPointerCancel={() => setDraggingEnd(false)}>
          {renderRoute(localPoints, "flex-track-preview-local")}
          {renderRoute(previewPoints, "flex-track-preview-server")}
          {valid ? <circle role="button" tabIndex={0} aria-label={t("layouts.flexEditor.dragEnd")}
            className="flex-track-end-handle" cx={localPath.endXMm} cy={localPath.endYMm} r="7"
            onPointerDown={(event) => {
              event.preventDefault();
              setDraggingEnd(true);
              event.currentTarget.ownerSVGElement?.setPointerCapture?.(event.pointerId);
            }} /> : null}
        </svg>
        {preview ? <div className="flex-track-preview-facts">
          <div><span>{t("layouts.flexEditor.length")}</span>
            <strong>{numberFormat.format(preview.effectiveLengthMm)} mm</strong></div>
          <div><span>{t("layouts.flexEditor.radius")}</span><strong>{preview.effectiveMinimumRadiusMm == null
            ? t("layouts.flexEditor.radiusInfinite")
            : `${numberFormat.format(preview.effectiveMinimumRadiusMm)} mm`}</strong></div>
          <div><span>{t("layouts.flexEditor.limit")}</span>
            <strong>{numberFormat.format(preview.radiusLimitMm)} mm</strong></div>
        </div> : null}
        {preview?.lengthExceeded ? <p className="flex-track-warning" role="alert"><AlertTriangle size={16} />
          {t("layouts.flexEditor.lengthExceeded", { length: numberFormat.format(preview.effectiveLengthMm),
            maximum: numberFormat.format(object.geometry.lengthMm) })}</p> : null}
        {preview?.radiusBelowLimit ? <p className="flex-track-warning" role="alert"><AlertTriangle size={16} />
          {t("layouts.flexEditor.radiusBelowLimit", {
            radius: numberFormat.format(preview.effectiveMinimumRadiusMm ?? 0),
            limit: numberFormat.format(preview.radiusLimitMm)
          })}</p> : null}
        {message ? <p className="form-message" role="alert">{message}</p> : null}
      </div>
      <footer className="modal-actions">
        <button type="button" className="secondary-button" disabled={saving || loading}
          onClick={onClose}>{t("common.cancel")}</button>
        <button type="button" className="secondary-button" disabled={saving || loading || !valid}
          onClick={() => void suggest()}>{loading ? t("layouts.flexEditor.suggesting")
            : t("layouts.flexEditor.suggest")}</button>
        <button type="button" className="primary-button" disabled={saving || loading || !preview?.applicable}
          onClick={() => preview && void onApply(preview.path)}>{t("layouts.flexEditor.apply")}</button>
      </footer>
    </section>
  </div>, document.body);
}
