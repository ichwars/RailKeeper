import { X } from "lucide-react";
import { useI18n } from "../../shared/i18n";
import type { SettingsConfirmDialog } from "./settingsModel";

type SettingsConfirmationDialogProps = {
  confirmDialog: SettingsConfirmDialog;
  onCancel: () => void;
  onConfirm: () => void;
};

export function SettingsConfirmationDialog({ confirmDialog, onCancel, onConfirm }: SettingsConfirmationDialogProps) {
  const { t } = useI18n();
  return (
    <div className="confirm-layer" role="dialog" aria-modal="true" aria-label={confirmDialog.title}>
      <section className="confirm-card">
        <div className="panel-head form-head">
          <h2>{confirmDialog.title}</h2>
          <button type="button" className="icon-button" onClick={onCancel} aria-label={t("vehicles.close")}>
            <X size={17} />
          </button>
        </div>
        {confirmDialog.body && <p>{confirmDialog.body}</p>}
        <div className="confirm-actions">
          <button type="button" className="secondary-button" onClick={onCancel}>
            {t("vehicles.cancel")}
          </button>
          <button type="button" className={confirmDialog.danger ? "danger-button" : "primary-button"} onClick={onConfirm}>
            {confirmDialog.confirmLabel}
          </button>
        </div>
      </section>
    </div>
  );
}
