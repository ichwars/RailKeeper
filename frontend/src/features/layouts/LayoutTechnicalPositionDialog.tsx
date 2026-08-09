import { X } from "lucide-react";
import { FormEvent, KeyboardEvent, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

import type {
  AccessoryArticleListItem,
  LayoutTechnicalPosition,
  LayoutTechnicalPositionInput,
  LayoutTechnicalPositionKind,
  LayoutUnit
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppTextArea } from "../../shared/ui/AppTextArea";
import { AppTextInput } from "../../shared/ui/AppTextInput";

type PositionForm = {
  label: string;
  kind: LayoutTechnicalPositionKind;
  positionXMm: string;
  positionYMm: string;
  rotationDegrees: string;
  productId: string;
  description: string;
  archived: boolean;
};

const positionKinds: LayoutTechnicalPositionKind[] = [
  "turnout", "signal", "feedback", "decoder", "lighting", "power", "sensor", "other"
];

const focusableSelector = [
  "button:not([disabled])", "input:not([disabled])", "textarea:not([disabled])", "[tabindex]:not([tabindex='-1'])"
].join(",");

function formValue(position?: LayoutTechnicalPosition): PositionForm {
  return {
    label: position?.label || "",
    kind: position?.kind || "turnout",
    positionXMm: String(position?.positionXMm ?? 0),
    positionYMm: String(position?.positionYMm ?? 0),
    rotationDegrees: String(position?.rotationDegrees ?? 0),
    productId: position?.productId || "",
    description: position?.description || "",
    archived: Boolean(position?.archived)
  };
}

export function LayoutTechnicalPositionDialog({ unit, position, products, saving, message, conflict,
  returnFocusTo, onSubmit, onReloadConflict, onClose }: {
  unit: LayoutUnit;
  position?: LayoutTechnicalPosition;
  products: AccessoryArticleListItem[];
  saving: boolean;
  message: string;
  conflict: boolean;
  returnFocusTo?: HTMLElement | null;
  onSubmit: (input: LayoutTechnicalPositionInput, expectedVersion?: number) => void | Promise<void>;
  onReloadConflict?: () => Promise<number | void>;
  onClose: () => void;
}) {
  const [form, setForm] = useState(() => formValue(position));
  const [expectedVersion, setExpectedVersion] = useState(position?.version);
  const layerRef = useRef<HTMLDivElement | null>(null);
  const labelRef = useRef<HTMLInputElement | null>(null);
  const { t } = useI18n();
  const title = t(position ? "layouts.technology.edit" : "layouts.technology.create");

  useEffect(() => {
    labelRef.current?.focus();
    return () => {
      if (returnFocusTo?.isConnected) returnFocusTo.focus();
    };
  }, [returnFocusTo]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    const positionXMm = Number(form.positionXMm);
    const positionYMm = Number(form.positionYMm);
    const rotationDegrees = Number(form.rotationDegrees);
    if (!form.label.trim() || !Number.isFinite(positionXMm) || !Number.isFinite(positionYMm) ||
      !Number.isFinite(rotationDegrees)) return;
    void onSubmit({
      label: form.label.trim(), kind: form.kind, positionXMm, positionYMm, rotationDegrees,
      productId: form.productId || undefined, description: form.description.trim() || undefined,
      archived: form.archived
    }, expectedVersion);
  };

  const reloadConflict = async () => {
    const version = await onReloadConflict?.();
    if (version) setExpectedVersion(version);
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      if (!saving) onClose();
      return;
    }
    if (event.key !== "Tab") return;
    const focusable = Array.from(layerRef.current?.querySelectorAll<HTMLElement>(focusableSelector) || []);
    if (focusable.length === 0) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if ((!event.shiftKey && event.target === last) || (event.shiftKey && event.target === first)) {
      event.preventDefault();
      (event.shiftKey ? last : first).focus();
    }
  };

  return createPortal(
    <div ref={layerRef} className="modal-layer layout-form-layer" role="dialog" aria-modal="true"
      aria-label={title} onKeyDown={handleKeyDown}>
      <form className="vehicle-modal layout-form-dialog layout-position-dialog" onSubmit={submit}>
        <header className="modal-head">
          <div><h2>{title}</h2><p>{unit.name}</p></div>
          <button type="button" className="icon-button" aria-label={t("common.close")} title={t("common.close")}
            disabled={saving} onClick={onClose}><X size={18} /></button>
        </header>
        <div className="modal-body layout-form-dialog-body">
          <AppTextInput ref={labelRef} label={t("layouts.field.name")} required value={form.label}
            disabled={saving} onChange={(event) => setForm({ ...form, label: event.target.value })} />
          <label className="app-field"><span className="app-field-label">{t("layouts.technology.kind")}</span>
            <AppSelect aria-label={t("layouts.technology.kind")} value={form.kind} disabled={saving}
              onChange={(event) => setForm({ ...form, kind: event.target.value as LayoutTechnicalPositionKind })}>
              {positionKinds.map((kind) => <option key={kind} value={kind}>
                {t(`layouts.positionKind.${kind}`)}</option>)}
            </AppSelect>
          </label>
          <div className="layout-position-coordinate-fields">
            <AppNumberInput label={t("layouts.technology.positionX")} required step="0.1"
              value={form.positionXMm} disabled={saving}
              onValueChange={(value) => setForm({ ...form, positionXMm: value })} />
            <AppNumberInput label={t("layouts.technology.positionY")} required step="0.1"
              value={form.positionYMm} disabled={saving}
              onValueChange={(value) => setForm({ ...form, positionYMm: value })} />
            <AppNumberInput label={t("layouts.technology.rotation")} required step="0.1"
              value={form.rotationDegrees} disabled={saving}
              onValueChange={(value) => setForm({ ...form, rotationDegrees: value })} />
          </div>
          <label className="app-field"><span className="app-field-label">{t("layouts.technology.article")}</span>
            <AppSelect aria-label={t("layouts.technology.article")} value={form.productId} disabled={saving}
              onChange={(event) => setForm({ ...form, productId: event.target.value })}>
              <option value="">{t("layouts.technology.noArticle")}</option>
              {products.map((product) => <option key={product.id} value={product.id}>
                {[product.inventoryNumber, product.manufacturer, product.articleNumber, product.name]
                  .filter(Boolean).join(" · ")}
              </option>)}
            </AppSelect>
          </label>
          <AppTextArea label={t("layouts.field.description")} value={form.description} disabled={saving}
            onChange={(event) => setForm({ ...form, description: event.target.value })} />
          {position ? <AppCheckbox label={t("layouts.field.archived")} checked={form.archived}
            disabled={saving} onChange={(event) => setForm({ ...form, archived: event.target.checked })} /> : null}
        </div>
        <footer className="modal-actions">
          {message ? <div className={conflict ? "layout-conflict layout-dialog-message" : "form-message"}
            role="alert"><span>{message}</span>{conflict && onReloadConflict ? <button type="button"
              className="secondary-button compact-action" disabled={saving} onClick={() => void reloadConflict()}>
              {t("layouts.conflict.reload")}</button> : null}</div> : null}
          <button type="button" className="secondary-button" disabled={saving}
            onClick={onClose}>{t("common.cancel")}</button>
          <button type="submit" className="primary-button" disabled={saving || !form.label.trim()}>
            {saving ? t("common.saving") : t("layouts.technology.save")}
          </button>
        </footer>
      </form>
    </div>,
    document.body
  );
}
