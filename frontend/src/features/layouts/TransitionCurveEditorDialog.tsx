import { AlertTriangle, X } from "lucide-react";
import { KeyboardEvent, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

import {
  api,
  type PlanTrackObject,
  type TrackPoint,
  type TransitionCurvePath,
  type TransitionCurvePreview,
  type TransitionDirection
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { AppSelect } from "../../shared/ui/AppSelect";
import { flexPreviewBounds, routePolylinePoints } from "./trackPlannerGeometry";
import { trackGeometryLabel } from "./trackGeometryLabel";

const focusableSelector = "button:not([disabled]), input:not([disabled]), [tabindex]:not([tabindex='-1'])";

export function TransitionCurveEditorDialog({ object, saving, onApply, onClose }: {
  object: PlanTrackObject;
  saving: boolean;
  onApply: (path: TransitionCurvePath) => void | Promise<void>;
  onClose: () => void;
}) {
  const initialLength = object.transitionPath?.lengthMm ?? Math.min(
    object.effectiveLengthMm || object.geometry.lengthMm,
    object.geometry.lengthMm
  );
  const initialRadius = object.transitionPath?.endRadiusMm ?? object.effectiveMinimumRadiusMm ??
    object.geometry.minimumRadiusMm ?? initialLength;
  const [lengthMM, setLengthMM] = useState(String(initialLength));
  const [endRadiusMM, setEndRadiusMM] = useState(String(initialRadius));
  const [direction, setDirection] = useState<TransitionDirection>(
    object.transitionPath?.direction ?? "left"
  );
  const [preview, setPreview] = useState<TransitionCurvePreview | null>(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const layerRef = useRef<HTMLDivElement | null>(null);
  const { language, t } = useI18n();
  const lengthValue = Number(lengthMM);
  const radiusValue = Number(endRadiusMM);
  const valid = lengthMM.trim() !== "" && endRadiusMM.trim() !== "" &&
    Number.isFinite(lengthValue) && lengthValue > 0 && Number.isFinite(radiusValue) && radiusValue > 0;
  const shownPoints = preview?.effectiveGeometry.routes[0]?.points ??
    object.effectiveGeometry.routes[0]?.points ?? [];
  const bounds = flexPreviewBounds(shownPoints);
  const numberFormat = new Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  });

  useEffect(() => {
    layerRef.current?.querySelector<HTMLElement>(focusableSelector)?.focus();
  }, []);

  const resetPreview = () => {
    setPreview(null);
    setMessage("");
  };

  const suggest = async () => {
    if (!valid) return;
    setLoading(true);
    setMessage("");
    try {
      setPreview(await api.previewTransitionCurvePath(object.id, {
        lengthMm: lengthValue,
        endRadiusMm: radiusValue,
        direction,
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

  const renderRoute = (points: TrackPoint[]) => points.length > 1
    ? <polyline className={preview ? "transition-curve-preview-server" : "transition-curve-preview-current"}
      points={routePolylinePoints(points)} /> : null;

  return createPortal(<div ref={layerRef} className="modal-layer transition-curve-editor-layer" role="dialog"
    aria-modal="true" aria-label={t("layouts.transitionEditor.title")} onKeyDown={handleKeyDown}>
    <section className="vehicle-modal transition-curve-editor-dialog">
      <header className="modal-head">
        <div><h2>{t("layouts.transitionEditor.title")}</h2>
          <p>{trackGeometryLabel(object.geometry)}</p></div>
        <button type="button" className="icon-button" aria-label={t("common.close")}
          disabled={saving || loading} onClick={onClose}><X size={18} /></button>
      </header>
      <div className="modal-body transition-curve-editor-body">
        <div className="transition-curve-editor-fields">
          <AppNumberInput label={t("layouts.transitionEditor.length")} value={lengthMM} min="0.1" step="0.1"
            disabled={saving || loading} onValueChange={(value) => { setLengthMM(value); resetPreview(); }} />
          <AppNumberInput label={t("layouts.transitionEditor.endRadius")} value={endRadiusMM}
            min="0.1" step="0.1" disabled={saving || loading}
            onValueChange={(value) => { setEndRadiusMM(value); resetPreview(); }} />
          <label className="app-field transition-curve-direction">
            <span className="app-field-label">{t("layouts.transitionEditor.direction")}</span>
            <AppSelect value={direction} disabled={saving || loading} onChange={(event) => {
              setDirection(event.target.value as TransitionDirection);
              resetPreview();
            }}>
              <option value="left">{t("layouts.transitionEditor.left")}</option>
              <option value="right">{t("layouts.transitionEditor.right")}</option>
            </AppSelect>
          </label>
        </div>
        <svg className="transition-curve-preview" role="img"
          aria-label={t("layouts.transitionEditor.preview")}
          viewBox={`${bounds.minX} ${bounds.minY} ${bounds.width} ${bounds.height}`}>
          {renderRoute(shownPoints)}
        </svg>
        {preview ? <div className="transition-curve-preview-facts">
          <div><span>{t("layouts.transitionEditor.effectiveLength")}</span>
            <strong>{numberFormat.format(preview.effectiveLengthMm)} mm</strong></div>
          <div><span>{t("layouts.transitionEditor.minimumRadius")}</span>
            <strong>{numberFormat.format(preview.effectiveMinimumRadiusMm ?? 0)} mm</strong></div>
          <div><span>{t("layouts.transitionEditor.limit")}</span>
            <strong>{numberFormat.format(preview.radiusLimitMm)} mm</strong></div>
        </div> : null}
        {preview?.lengthExceeded ? <p className="flex-track-warning" role="alert"><AlertTriangle size={16} />
          {t("layouts.transitionEditor.lengthExceeded", {
            length: numberFormat.format(preview.effectiveLengthMm),
            maximum: numberFormat.format(object.geometry.lengthMm)
          })}</p> : null}
        {preview?.radiusBelowLimit ? <p className="flex-track-warning" role="alert"><AlertTriangle size={16} />
          {t("layouts.transitionEditor.radiusBelowLimit", {
            radius: numberFormat.format(preview.effectiveMinimumRadiusMm ?? 0),
            limit: numberFormat.format(preview.radiusLimitMm)
          })}</p> : null}
        {message ? <p className="form-message" role="alert">{message}</p> : null}
      </div>
      <footer className="modal-actions">
        <button type="button" className="secondary-button" disabled={saving || loading}
          onClick={onClose}>{t("common.cancel")}</button>
        <button type="button" className="secondary-button" disabled={saving || loading || !valid}
          onClick={() => void suggest()}>{loading ? t("layouts.transitionEditor.suggesting")
            : t("layouts.transitionEditor.suggest")}</button>
        <button type="button" className="primary-button" disabled={saving || loading || !preview?.applicable}
          onClick={() => preview && void onApply(preview.path)}>{t("layouts.transitionEditor.apply")}</button>
      </footer>
    </section>
  </div>, document.body);
}
