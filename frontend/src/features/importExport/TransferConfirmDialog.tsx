import { useEffect, useRef, useState, type KeyboardEvent } from "react";
import { createPortal } from "react-dom";

export type TransferPendingAction = {
  body: string;
  confirmLabel: string;
  dangerous?: boolean;
  errorMessage: string;
  run: () => void | Promise<void>;
  title: string;
};

const focusableSelector = "button:not([disabled]), [tabindex]:not([tabindex='-1'])";

export function TransferConfirmDialog({ action, cancelLabel, onClose }: {
  action: TransferPendingAction | null;
  cancelLabel: string;
  onClose: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const layerRef = useRef<HTMLDivElement | null>(null);
  const anchorRef = useRef<HTMLSpanElement | null>(null);
  const cancelRef = useRef<HTMLButtonElement | null>(null);
  const invokerRef = useRef<HTMLElement | null>(null);
  const isOpen = Boolean(action);

  useEffect(() => {
    if (!isOpen) return;
    setError("");
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

  async function confirm() {
    setBusy(true);
    setError("");
    try {
      await action?.run();
      onClose();
    } catch {
      setError(action?.errorMessage || "");
    } finally {
      setBusy(false);
    }
  }

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
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
  }

  return <>
    <span ref={anchorRef} hidden aria-hidden="true" />
    {createPortal(
      <div ref={layerRef} className="confirm-layer data-transfer-confirm-layer" role="dialog" aria-modal="true"
        aria-label={action.title} onKeyDown={onKeyDown}>
        <section className="panel data-transfer-confirm-dialog">
          <h2>{action.title}</h2>
          <p>{action.body}</p>
          {error ? <p className="form-message error" role="alert">{error}</p> : null}
          <div className="data-transfer-confirm-actions">
            <button ref={cancelRef} type="button" className="secondary-button" disabled={busy} onClick={onClose}>
              {cancelLabel}
            </button>
            <button type="button" className={action.dangerous ? "danger-button" : "primary-button"}
              disabled={busy} onClick={() => void confirm()}>
              {action.confirmLabel}
            </button>
          </div>
        </section>
      </div>,
      document.body
    )}
  </>;
}
