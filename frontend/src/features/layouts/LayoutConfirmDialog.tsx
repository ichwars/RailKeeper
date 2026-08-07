import { useState } from "react";

import { useI18n } from "../../shared/i18n";

export type LayoutPendingAction = { title: string; body: string; run: () => Promise<void> };

export function LayoutConfirmDialog({ action, onClose }: {
  action: LayoutPendingAction | null;
  onClose: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const { t } = useI18n();
  if (!action) return null;

  const confirm = async () => {
    setBusy(true);
    setError("");
    try {
      await action.run();
      onClose();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally {
      setBusy(false);
    }
  };

  return <div className="confirm-layer layout-confirm-layer" role="dialog" aria-modal="true" aria-label={action.title}>
    <section className="panel layout-confirm-dialog">
      <h2>{action.title}</h2>
      <p>{action.body}</p>
      {error ? <p className="form-message">{error}</p> : null}
      <div className="layout-form-actions">
        <button type="button" className="secondary-button" onClick={onClose} disabled={busy}>{t("common.cancel")}</button>
        <button type="button" className="primary-button" onClick={confirm} disabled={busy}>
          {busy ? t("common.saving") : t("common.confirm")}
        </button>
      </div>
    </section>
  </div>;
}
