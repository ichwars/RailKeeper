import { KeyboardEvent, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { X } from "lucide-react";

import type {
  CreateFreePlanObjectInput,
  FreePlanObjectCategory,
  FreePlanObjectKind,
  PlanFreeObject,
  UpdateFreePlanObjectInput
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { AppSelect } from "../../shared/ui/AppSelect";

const focusableSelector = "button:not([disabled]), input:not([disabled]), [tabindex]:not([tabindex='-1'])";

export function FreePlanObjectDialog({ object, initialPosition = { xMm: 0, yMm: 0 }, saving, onSubmit,
  onClose }: {
  object?: PlanFreeObject;
  initialPosition?: { xMm: number; yMm: number };
  saving: boolean;
  onSubmit: (input: CreateFreePlanObjectInput | UpdateFreePlanObjectInput) => void | Promise<void>;
  onClose: () => void;
}) {
  const { t } = useI18n();
  const shape = object?.shape;
  const [name, setName] = useState(object?.name ?? "");
  const [category, setCategory] = useState<FreePlanObjectCategory>(object?.category ?? "structure");
  const [kind, setKind] = useState<FreePlanObjectKind>(shape?.kind ?? "rectangle");
  const [positionX, setPositionX] = useState(String(object?.positionXMm ?? initialPosition.xMm));
  const [positionY, setPositionY] = useState(String(object?.positionYMm ?? initialPosition.yMm));
  const [rotation, setRotation] = useState(String(object?.rotationDegrees ?? 0));
  const [width, setWidth] = useState(String(shape?.widthMm ?? 100));
  const [height, setHeight] = useState(String(shape?.heightMm ?? 50));
  const [endX, setEndX] = useState(String(shape?.endXMm ?? 100));
  const [endY, setEndY] = useState(String(shape?.endYMm ?? 0));
  const [text, setText] = useState(shape?.text ?? "");
  const [fontSize, setFontSize] = useState(String(shape?.fontSizeMm ?? 8));
  const layerRef = useRef<HTMLDivElement | null>(null);

  const commonNumbers = [positionX, positionY, rotation].map(Number);
  const validCommon = name.trim().length > 0 && name.trim().length <= 80 &&
    commonNumbers.every(Number.isFinite);
  const validShape = kind === "rectangle" || kind === "ellipse"
    ? positiveNumber(width) && positiveNumber(height)
    : kind === "line"
      ? finiteNumber(endX) && finiteNumber(endY) && Math.hypot(Number(endX), Number(endY)) > 0
      : text.trim().length > 0 && text.trim().length <= 120 && finiteNumber(fontSize) &&
        Number(fontSize) >= 2 && Number(fontSize) <= 50;
  const valid = validCommon && validShape;

  useEffect(() => {
    layerRef.current?.querySelector<HTMLElement>(focusableSelector)?.focus();
  }, []);

  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      if (!saving) onClose();
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

  const submit = () => {
    if (!valid) return;
    const input: CreateFreePlanObjectInput = {
      name: name.trim(), category, positionXMm: Number(positionX), positionYMm: Number(positionY),
      rotationDegrees: Number(rotation), shape: freePlanShape(kind, {
        width, height, endX, endY, text: text.trim(), fontSize
      })
    };
    void onSubmit(object ? { ...input, expectedVersion: object.version } : input);
  };

  return createPortal(<div ref={layerRef} className="modal-layer free-plan-object-layer" role="dialog"
    aria-modal="true" aria-label={t("layouts.freeObject.title")} onKeyDown={handleKeyDown}>
    <section className="vehicle-modal free-plan-object-dialog">
      <header className="modal-head">
        <h2>{object ? t("layouts.freeObject.editTitle") : t("layouts.freeObject.createTitle")}</h2>
        <button type="button" className="icon-button" aria-label={t("common.close")} disabled={saving}
          onClick={onClose}><X size={18} /></button>
      </header>
      <div className="modal-body free-plan-object-body">
        <div className="free-plan-object-common-fields">
          <label className="app-field"><span className="app-field-label">{t("layouts.freeObject.name")}</span>
            <input value={name} maxLength={80} disabled={saving}
              onChange={(event) => setName(event.target.value)} /></label>
          <label className="app-field"><span className="app-field-label">{t("layouts.freeObject.category")}</span>
            <AppSelect aria-label={t("layouts.freeObject.category")} value={category} disabled={saving}
              onChange={(event) => setCategory(event.target.value as FreePlanObjectCategory)}>
              {(["structure", "platform", "scenery", "annotation"] as const).map((value) =>
                <option key={value} value={value}>{t(`layouts.freeObject.category.${value}`)}</option>)}
            </AppSelect></label>
          <label className="app-field"><span className="app-field-label">{t("layouts.freeObject.kind")}</span>
            <AppSelect aria-label={t("layouts.freeObject.kind")} value={kind} disabled={saving}
              onChange={(event) => setKind(event.target.value as FreePlanObjectKind)}>
              {(["rectangle", "ellipse", "line", "label"] as const).map((value) =>
                <option key={value} value={value}>{t(`layouts.freeObject.kind.${value}`)}</option>)}
            </AppSelect></label>
          <AppNumberInput label={t("layouts.freeObject.positionX")} value={positionX} step="1"
            disabled={saving} onValueChange={setPositionX} />
          <AppNumberInput label={t("layouts.freeObject.positionY")} value={positionY} step="1"
            disabled={saving} onValueChange={setPositionY} />
          <AppNumberInput label={t("layouts.freeObject.rotation")} value={rotation} step="1"
            disabled={saving} onValueChange={setRotation} />
        </div>
        <div className="free-plan-object-shape-fields">
          {kind === "rectangle" || kind === "ellipse" ? <>
            <AppNumberInput label={t("layouts.freeObject.width")} value={width} min="0.1" step="1"
              disabled={saving} onValueChange={setWidth} />
            <AppNumberInput label={t("layouts.freeObject.height")} value={height} min="0.1" step="1"
              disabled={saving} onValueChange={setHeight} />
          </> : null}
          {kind === "line" ? <>
            <AppNumberInput label={t("layouts.freeObject.endX")} value={endX} step="1"
              disabled={saving} onValueChange={setEndX} />
            <AppNumberInput label={t("layouts.freeObject.endY")} value={endY} step="1"
              disabled={saving} onValueChange={setEndY} />
          </> : null}
          {kind === "label" ? <>
            <label className="app-field"><span className="app-field-label">{t("layouts.freeObject.text")}</span>
              <input value={text} maxLength={120} disabled={saving}
                onChange={(event) => setText(event.target.value)} /></label>
            <AppNumberInput label={t("layouts.freeObject.fontSize")} value={fontSize} min="2" max="50"
              step="0.5" disabled={saving} onValueChange={setFontSize} />
          </> : null}
        </div>
      </div>
      <footer className="modal-actions">
        <button type="button" className="secondary-button" disabled={saving}
          onClick={onClose}>{t("common.cancel")}</button>
        <button type="button" className="primary-button" disabled={saving || !valid}
          onClick={submit}>{t("layouts.freeObject.save")}</button>
      </footer>
    </section>
  </div>, document.body);
}

function finiteNumber(value: string) {
  return value.trim() !== "" && Number.isFinite(Number(value));
}

function positiveNumber(value: string) {
  return finiteNumber(value) && Number(value) > 0;
}

function freePlanShape(kind: FreePlanObjectKind, values: {
  width: string; height: string; endX: string; endY: string; text: string; fontSize: string;
}) {
  if (kind === "rectangle" || kind === "ellipse") {
    return { schemaVersion: 1 as const, kind, widthMm: Number(values.width), heightMm: Number(values.height) };
  }
  if (kind === "line") {
    return { schemaVersion: 1 as const, kind, endXMm: Number(values.endX), endYMm: Number(values.endY) };
  }
  return { schemaVersion: 1 as const, kind, text: values.text, fontSizeMm: Number(values.fontSize) };
}
