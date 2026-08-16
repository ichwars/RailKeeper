import { Archive, ArchiveRestore, Pencil, Trash2 } from "lucide-react";

import type { MasterDataEntry } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export type MasterDataLifecycleActionsProps = {
  entry: MasterDataEntry;
  displayLabel: string;
  disabled?: boolean;
  onEdit: () => void;
  onDeactivate: () => void;
  onReactivate: () => void;
  onDelete: () => void;
};

export function MasterDataLifecycleActions({
  entry,
  displayLabel,
  disabled = false,
  onEdit,
  onDeactivate,
  onReactivate,
  onDelete
}: MasterDataLifecycleActionsProps) {
  const { t } = useI18n();
  const actionLabel = (label: string) => `${displayLabel} ${label.toLocaleLowerCase()}`;
  const editLabel = t("vehicles.edit");
  const deactivateLabel = t("settings.master.deactivate");
  const reactivateLabel = t("settings.master.reactivate");
  const deleteLabel = t("settings.master.deletePermanently");

  return (
    <div className="table-actions master-data-lifecycle-actions">
      <button type="button" className="icon-button" onClick={onEdit} disabled={disabled}
        aria-label={actionLabel(editLabel)} title={editLabel}>
        <Pencil size={16} aria-hidden="true" />
      </button>
      {entry.active && entry.capabilities?.canDeactivate && (
        <button type="button" className="icon-button" onClick={onDeactivate} disabled={disabled}
          aria-label={actionLabel(deactivateLabel)} title={deactivateLabel}>
          <Archive size={16} aria-hidden="true" />
        </button>
      )}
      {!entry.active && entry.capabilities?.canReactivate && (
        <button type="button" className="icon-button" onClick={onReactivate} disabled={disabled}
          aria-label={actionLabel(reactivateLabel)} title={reactivateLabel}>
          <ArchiveRestore size={16} aria-hidden="true" />
        </button>
      )}
      {entry.capabilities?.canDelete && (
        <button type="button" className="icon-button danger" onClick={onDelete} disabled={disabled}
          aria-label={actionLabel(deleteLabel)} title={deleteLabel}>
          <Trash2 size={16} aria-hidden="true" />
        </button>
      )}
    </div>
  );
}
