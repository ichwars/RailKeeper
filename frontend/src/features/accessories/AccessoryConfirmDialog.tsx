import { useEffect, useRef, useState, type KeyboardEvent, type ReactNode } from "react";

import { useI18n } from "../../shared/i18n";

export type AccessoryPendingAction = {
  title: string;
  body: ReactNode;
  run: () => void | Promise<void>;
  afterSuccess?: () => void | Promise<void>;
  cancelLabel?: string;
  confirmLabel?: string;
  dangerous?: boolean;
};

const focusableSelector = "button:not([disabled]), input:not([disabled]), textarea:not([disabled]), " +
  "a[href], [tabindex]:not([tabindex='-1'])";

export function AccessoryConfirmDialog({ action, onClose }: {
  action: AccessoryPendingAction | null;
  onClose: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const layerRef = useRef<HTMLDivElement | null>(null);
  const cancelRef = useRef<HTMLButtonElement | null>(null);
  const invokerRef = useRef<HTMLElement | null>(null);
  const { t } = useI18n();
  const isOpen = Boolean(action);

  useEffect(() => {
    if (!isOpen) return;
    setError("");
    invokerRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    cancelRef.current?.focus();
    return () => invokerRef.current?.focus();
  }, [isOpen]);

  if (!action) return null;

  const confirm = async () => {
    setBusy(true);
    setError("");
    try {
      await action.run();
      onClose();
      try {
        await action.afterSuccess?.();
      } catch {
        // The owning resource controller exposes refresh failures and retry state.
      }
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("accessories.error.generic"));
    } finally {
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

  return (
    <div ref={layerRef} className="confirm-layer accessory-confirm-layer" role="dialog" aria-modal="true"
      aria-label={action.title} onKeyDown={onKeyDown}>
      <section className="panel accessory-confirm-dialog">
        <h2>{action.title}</h2>
        {typeof action.body === "string" ? <p>{action.body}</p> : <div>{action.body}</div>}
        {error ? <p className="form-message" role="alert">{error}</p> : null}
        <div className="accessory-form-actions">
          <button ref={cancelRef} type="button" className="secondary-button" onClick={onClose} disabled={busy}>
            {action.cancelLabel || t("common.cancel")}
          </button>
          <button type="button" className={action.dangerous ? "danger-button" : "primary-button"}
            onClick={confirm} disabled={busy}>
            {busy ? t("common.saving") : action.confirmLabel || t("common.confirm")}
          </button>
        </div>
      </section>
    </div>
  );
}
