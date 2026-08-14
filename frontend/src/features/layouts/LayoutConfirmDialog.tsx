import { KeyboardEvent, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

import { useI18n } from "../../shared/i18n";

export type LayoutPendingAction = {
  title: string;
  body: string;
  run: () => void | Promise<void>;
  cancelLabel?: string;
  confirmLabel?: string;
  dangerous?: boolean;
};

const focusableSelector = "button:not([disabled]), input:not([disabled]), textarea:not([disabled]), " +
  "[tabindex]:not([tabindex='-1'])";

export function LayoutConfirmDialog({ action, onClose }: {
  action: LayoutPendingAction | null;
  onClose: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const layerRef = useRef<HTMLDivElement | null>(null);
  const anchorRef = useRef<HTMLSpanElement | null>(null);
  const cancelRef = useRef<HTMLButtonElement | null>(null);
  const invokerRef = useRef<HTMLElement | null>(null);
  const { t } = useI18n();
  const isOpen = Boolean(action);

  useEffect(() => {
    if (!isOpen) return;
    invokerRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const parentDialog = anchorRef.current?.closest<HTMLElement>("[role='dialog']") || null;
    const previousAriaHidden = parentDialog?.getAttribute("aria-hidden") ?? null;
    const previouslyInert = parentDialog?.hasAttribute("inert") ?? false;
    parentDialog?.setAttribute("aria-hidden", "true");
    parentDialog?.setAttribute("inert", "");
    cancelRef.current?.focus();
    return () => {
      if (parentDialog) {
        if (previousAriaHidden === null) parentDialog.removeAttribute("aria-hidden");
        else parentDialog.setAttribute("aria-hidden", previousAriaHidden);
        if (!previouslyInert) parentDialog.removeAttribute("inert");
      }
      if (invokerRef.current?.isConnected) invokerRef.current.focus();
    };
  }, [isOpen]);

  if (!action) return <span ref={anchorRef} hidden aria-hidden="true" />;

  const confirm = () => {
    setBusy(true);
    setError("");
    try {
      const result = action.run();
      if (result && typeof result.then === "function") {
        void result.then(() => {
          setBusy(false);
          onClose();
        }).catch((reason: unknown) => {
          setError(reason instanceof Error ? reason.message : t("layouts.error.generic"));
          setBusy(false);
        });
        return;
      }
      setBusy(false);
      onClose();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("layouts.error.generic"));
      setBusy(false);
    }
  };

  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    event.stopPropagation();
    if (event.key === "Escape") {
      event.preventDefault();
      if (!busy) onClose();
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

  const dialog = <div ref={layerRef} className="confirm-layer layout-confirm-layer" role="dialog" aria-modal="true"
    aria-label={action.title} onKeyDown={onKeyDown}>
    <section className="panel layout-confirm-dialog">
      <h2>{action.title}</h2>
      <p>{action.body}</p>
      {error ? <p className="form-message">{error}</p> : null}
      <div className="layout-form-actions">
        <button ref={cancelRef} type="button" className="secondary-button" onClick={onClose} disabled={busy}>
          {action.cancelLabel || t("common.cancel")}
        </button>
        <button type="button" className={action.dangerous ? "danger-button" : "primary-button"}
          onClick={confirm} disabled={busy}>
          {busy ? t("common.saving") : action.confirmLabel || t("common.confirm")}
        </button>
      </div>
    </section>
  </div>;
  return <><span ref={anchorRef} hidden aria-hidden="true" />{createPortal(dialog, document.body)}</>;
}
