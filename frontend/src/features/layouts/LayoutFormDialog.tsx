import { X } from "lucide-react";
import { FormEvent, KeyboardEvent, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

import type { LayoutInput, LayoutKind } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppTextArea } from "../../shared/ui/AppTextArea";
import { AppTextInput } from "../../shared/ui/AppTextInput";
import { LayoutConfirmDialog } from "./LayoutConfirmDialog";

export type LayoutFormMode = "create" | "edit";

export type LayoutFormValue = Required<Pick<LayoutInput, "name" | "kind" | "gauge" | "scale">> & {
  description: string;
  archived: boolean;
};

export type LayoutFormDialogProps = {
  mode: LayoutFormMode;
  initialValue: LayoutFormValue;
  saving: boolean;
  message: string;
  conflict: boolean;
  returnFocusTo?: HTMLElement | null;
  onSubmit: (value: LayoutFormValue) => void | Promise<void>;
  onReloadConflict?: () => LayoutFormValue | void | Promise<LayoutFormValue | void>;
  onClose: () => void;
};

const focusableSelector = [
  "button:not([disabled])",
  "input:not([disabled])",
  "textarea:not([disabled])",
  "[tabindex]:not([tabindex='-1'])"
].join(",");

export function LayoutFormDialog({ mode, initialValue, saving, message, conflict, returnFocusTo,
  onSubmit, onReloadConflict, onClose }: LayoutFormDialogProps) {
  const [form, setForm] = useState(initialValue);
  const [discardOpen, setDiscardOpen] = useState(false);
  const layerRef = useRef<HTMLDivElement | null>(null);
  const nameRef = useRef<HTMLInputElement | null>(null);
  const { t } = useI18n();
  const initialSnapshot = JSON.stringify(initialValue);
  const dirty = JSON.stringify(form) !== initialSnapshot;
  const title = t(mode === "create" ? "layouts.create.title" : "layouts.edit.title");

  useEffect(() => {
    setForm(initialValue);
    setDiscardOpen(false);
  }, [initialSnapshot]);

  useEffect(() => {
    nameRef.current?.focus();
    return () => {
      if (returnFocusTo?.isConnected) returnFocusTo.focus();
    };
  }, [returnFocusTo]);

  const requestClose = () => {
    if (saving) return;
    if (dirty) setDiscardOpen(true);
    else onClose();
  };

  const submit = (event: FormEvent) => {
    event.preventDefault();
    if (!form.name.trim() || !form.gauge.trim() || !form.scale.trim()) return;
    void onSubmit({
      ...form,
      name: form.name.trim(),
      gauge: form.gauge.trim(),
      scale: form.scale.trim(),
      description: form.description.trim()
    });
  };

  const reloadConflict = async () => {
    const nextValue = await onReloadConflict?.();
    if (nextValue) setForm(nextValue);
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      requestClose();
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

  const dialog = (
    <div ref={layerRef} className="modal-layer layout-form-layer" role="dialog" aria-modal="true"
      aria-label={title} onKeyDown={handleKeyDown}>
      <form className="vehicle-modal layout-form-dialog" onSubmit={submit}>
        <header className="modal-head">
          <h2>{title}</h2>
          <button type="button" className="icon-button" aria-label={t("common.close")} title={t("common.close")}
            disabled={saving} onClick={requestClose}><X size={18} /></button>
        </header>
        <div className="modal-body layout-form-dialog-body">
          <AppTextInput ref={nameRef} label={t("layouts.field.name")} required value={form.name}
            disabled={saving} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} />
          <label className="app-field">
            <span className="app-field-label">{t("layouts.field.kind")}</span>
            <AppSelect value={form.kind} aria-label={t("layouts.field.kind")} disabled={saving}
              onChange={(event) => setForm((current) => ({
                ...current, kind: event.target.value as LayoutKind
              }))}>
              <option value="private">{t("layouts.kind.private")}</option>
              <option value="club">{t("layouts.kind.club")}</option>
            </AppSelect>
          </label>
          <div className="layout-inline-fields">
            <AppTextInput label={t("layouts.field.gauge")} required value={form.gauge} disabled={saving}
              onChange={(event) => setForm((current) => ({ ...current, gauge: event.target.value }))} />
            <AppTextInput label={t("layouts.field.scale")} required value={form.scale} disabled={saving}
              onChange={(event) => setForm((current) => ({ ...current, scale: event.target.value }))} />
          </div>
          <AppTextArea label={t("layouts.field.description")} value={form.description} disabled={saving}
            onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))} />
          {mode === "edit" ? <AppCheckbox label={t("layouts.field.archived")} checked={form.archived}
            disabled={saving} onChange={(event) => setForm((current) => ({
              ...current, archived: event.target.checked
            }))} /> : null}
        </div>
        <footer className="modal-actions">
          {message ? <div className={conflict ? "layout-conflict layout-dialog-message" : "form-message"}
            role="alert"><span>{message}</span>{conflict && onReloadConflict ? <button type="button"
              className="secondary-button compact-action" disabled={saving}
              onClick={() => void reloadConflict()}>{t("layouts.conflict.reload")}</button> : null}</div> : null}
          <button type="button" className="secondary-button" disabled={saving}
            onClick={requestClose}>{t("common.cancel")}</button>
          <button type="submit" className="primary-button" disabled={saving || !form.name.trim() ||
            !form.gauge.trim() || !form.scale.trim()}>
            {saving ? t("common.saving") : t(mode === "create" ? "layouts.create.save" : "layouts.edit.save")}
          </button>
        </footer>
        <LayoutConfirmDialog action={discardOpen ? {
          title: t("layouts.dialog.discardTitle"),
          body: t("layouts.dialog.discardBody"),
          confirmLabel: t("layouts.dialog.discardAction"),
          dangerous: true,
          run: onClose
        } : null} onClose={() => setDiscardOpen(false)} />
      </form>
    </div>
  );

  return createPortal(dialog, document.body);
}
