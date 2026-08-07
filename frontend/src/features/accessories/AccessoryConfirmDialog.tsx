import { useState } from "react";

import { useI18n } from "../../shared/i18n";

export type AccessoryPendingAction = {
  title: string;
  body: string;
  run: () => Promise<void>;
};

export function AccessoryConfirmDialog({ action, onClose }: {
  action: AccessoryPendingAction | null;
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
      setError(reason instanceof Error ? reason.message : t("accessories.error.generic"));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="confirm-layer accessory-confirm-layer" role="dialog" aria-modal="true" aria-label={action.title}>
      <section className="panel accessory-confirm-dialog">
        <h2>{action.title}</h2>
        <p>{action.body}</p>
        {error ? <p className="form-message">{error}</p> : null}
        <div className="accessory-form-actions">
          <button type="button" className="secondary-button" onClick={onClose} disabled={busy}>
            {t("common.cancel")}
          </button>
          <button type="button" className="primary-button" onClick={confirm} disabled={busy}>
            {busy ? t("common.saving") : t("common.confirm")}
          </button>
        </div>
      </section>
    </div>
  );
}
