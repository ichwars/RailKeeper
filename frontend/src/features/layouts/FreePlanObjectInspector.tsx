import { Pencil, RotateCcw, RotateCw, Trash2 } from "lucide-react";

import type { PlanFreeObject } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export function FreePlanObjectInspector({ object, editable, saving, onEdit, onRotate, onDelete }: {
  object: PlanFreeObject;
  editable: boolean;
  saving: boolean;
  onEdit: () => void;
  onRotate: (degrees: number) => void;
  onDelete: () => void;
}) {
  const { language, t } = useI18n();
  const format = (value: number) => new Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB", {
    minimumFractionDigits: 2, maximumFractionDigits: 2
  }).format(value);
  return <div className="free-plan-object-inspector">
    <h5>{object.name}</h5>
    <dl>
      <div><dt>{t("layouts.freeObject.category")}</dt>
        <dd>{t(`layouts.freeObject.category.${object.category}`)}</dd></div>
      <div><dt>{t("layouts.freeObject.kind")}</dt>
        <dd>{t(`layouts.freeObject.kind.${object.shape.kind}`)}</dd></div>
      <div><dt>{t("layouts.trackPlanner.position")}</dt>
        <dd>{format(object.positionXMm)} / {format(object.positionYMm)} mm</dd></div>
      <div><dt>{t("layouts.trackPlanner.rotation")}</dt><dd>{format(object.rotationDegrees)}°</dd></div>
      <div><dt>{t("layouts.freeObject.dimensions")}</dt><dd>{shapeFacts(object, format)}</dd></div>
    </dl>
    {editable ? <div className="track-planner-actions">
      <button type="button" className="secondary-button compact-action" disabled={saving}
        onClick={onEdit}><Pencil size={14} />{t("layouts.freeObject.edit")}</button>
      <button type="button" className="secondary-button compact-action" disabled={saving}
        onClick={() => onRotate(-15)}><RotateCcw size={14} />-15°</button>
      <button type="button" className="secondary-button compact-action" disabled={saving}
        onClick={() => onRotate(15)}><RotateCw size={14} />+15°</button>
      <button type="button" className="danger-button compact-action" disabled={saving}
        onClick={onDelete}><Trash2 size={14} />{t("layouts.freeObject.delete")}</button>
    </div> : null}
  </div>;
}

function shapeFacts(object: PlanFreeObject, format: (value: number) => string) {
  const shape = object.shape;
  if (shape.kind === "rectangle" || shape.kind === "ellipse") {
    return `${format(shape.widthMm ?? 0)} × ${format(shape.heightMm ?? 0)} mm`;
  }
  if (shape.kind === "line") {
    return `${format(shape.endXMm ?? 0)} / ${format(shape.endYMm ?? 0)} mm`;
  }
  return `${shape.text ?? ""} · ${format(shape.fontSizeMm ?? 0)} mm`;
}
