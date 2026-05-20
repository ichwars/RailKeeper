import { FileText, Image, Pencil, Save, Trash2, Wrench } from "lucide-react";
import { Vehicle, VehicleMaintenance, VehicleMaintenanceInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import {
  formatMaintenanceCost,
  maintenanceIsDue,
  maintenanceStatusClass,
  normalizeMaintenanceStatus
} from "./vehicleMaintenance";
import { conditionRatings, maintenanceKinds, maintenanceStatuses } from "./vehicleOptions";
import { formatDate } from "./vehicleFormat";
import { PendingArticleImage } from "./vehicleTransforms";

export function VehicleMaintenanceTab({
  selected,
  pendingArticleImages,
  readonly,
  saving,
  maintenanceForm,
  editingMaintenanceID,
  maintenanceSummary,
  onUpdateMaintenanceForm,
  onResetMaintenanceForm,
  onSaveMaintenance,
  onCompleteMaintenance,
  onEditMaintenance,
  onDeleteMaintenance
}: {
  selected: Vehicle | null;
  pendingArticleImages: PendingArticleImage[];
  readonly: boolean;
  saving: boolean;
  maintenanceForm: VehicleMaintenanceInput;
  editingMaintenanceID: string | null;
  maintenanceSummary: {
    due: number;
    planned: number;
    done: number;
  };
  onUpdateMaintenanceForm: (patch: Partial<VehicleMaintenanceInput>) => void;
  onResetMaintenanceForm: () => void;
  onSaveMaintenance: () => void;
  onCompleteMaintenance: (entry: VehicleMaintenance) => void;
  onEditMaintenance: (entry: VehicleMaintenance) => void;
  onDeleteMaintenance: (entry: VehicleMaintenance) => void;
}) {
  const { t } = useI18n();

  return (
    <section className="maintenance-tab">
      <section className="maintenance-editor">
        <div className="upload-head">
          <div>
            <h3>{t("vehicles.maintenance.title")}</h3>
            <p>{t("vehicles.maintenance.subtitle")}</p>
          </div>
          <Wrench size={22} aria-hidden="true" />
        </div>
        {!selected && <p className="empty-state compact">{t("vehicles.maintenance.emptyUntilSave")}</p>}
        {selected && (
          <>
            <div className="maintenance-summary">
              <div>
                <span>{t("vehicles.maintenance.due")}</span>
                <strong>{maintenanceSummary.due}</strong>
              </div>
              <div>
                <span>{t("vehicles.maintenance.plannedOpen")}</span>
                <strong>{maintenanceSummary.planned}</strong>
              </div>
              <div>
                <span>{t("vehicles.maintenance.done")}</span>
                <strong>{maintenanceSummary.done}</strong>
              </div>
            </div>
            <div className="maintenance-form">
              <label>
                {t("vehicles.maintenance.kind")}
                <select value={maintenanceForm.kind} onChange={(event) => onUpdateMaintenanceForm({ kind: event.target.value })} disabled={readonly || saving}>
                  {maintenanceKinds.map((kind) => (
                    <option key={kind} value={kind}>{t(`vehicles.maintenance.kind.${kind}`)}</option>
                  ))}
                </select>
              </label>
              <label>
                {t("vehicles.maintenance.status")}
                <select value={maintenanceForm.status} onChange={(event) => onUpdateMaintenanceForm({ status: event.target.value })} disabled={readonly || saving}>
                  {maintenanceStatuses.map((status) => (
                    <option key={status.value} value={status.value}>{t(`vehicles.maintenance.status.${status.value}`)}</option>
                  ))}
                </select>
              </label>
              <label>
                {t("vehicles.maintenance.condition")}
                <select value={maintenanceForm.conditionRating || ""} onChange={(event) => onUpdateMaintenanceForm({ conditionRating: event.target.value })} disabled={readonly || saving}>
                  <option value="">{t("vehicles.select.placeholder")}</option>
                  {conditionRatings.map((rating) => (
                    <option key={rating} value={rating}>{rating}</option>
                  ))}
                </select>
              </label>
              <label>
                {t("vehicles.maintenance.dueDate")}
                <input type="date" value={maintenanceForm.dueDate || ""} onChange={(event) => onUpdateMaintenanceForm({ dueDate: event.target.value })} disabled={readonly || saving} />
              </label>
              <label>
                {t("vehicles.maintenance.completedAt")}
                <input type="date" value={maintenanceForm.completedAt || ""} onChange={(event) => onUpdateMaintenanceForm({ completedAt: event.target.value })} disabled={readonly || saving} />
              </label>
              <label>
                {t("vehicles.maintenance.cost")}
                <input value={maintenanceForm.cost || ""} onChange={(event) => onUpdateMaintenanceForm({ cost: event.target.value })} disabled={readonly || saving} inputMode="decimal" placeholder="0,00" />
              </label>
              <label className="maintenance-notes">
                {t("vehicles.maintenance.notes")}
                <textarea value={maintenanceForm.notes || ""} onChange={(event) => onUpdateMaintenanceForm({ notes: event.target.value })} disabled={readonly || saving} rows={4} />
              </label>
            </div>
            <div className="maintenance-actions">
              {editingMaintenanceID && (
                <button type="button" className="secondary-button" onClick={onResetMaintenanceForm} disabled={readonly || saving}>
                  {t("vehicles.cancel")}
                </button>
              )}
              <button type="button" className="primary-button" onClick={onSaveMaintenance} disabled={readonly || saving}>
                <Save size={15} aria-hidden="true" />
                {editingMaintenanceID ? t("vehicles.maintenance.saveEntry") : t("vehicles.maintenance.addEntry")}
              </button>
            </div>
          </>
        )}
      </section>

      <section className="maintenance-list">
        {selected && (!selected.maintenance || selected.maintenance.length === 0) && (
          <p className="empty-state compact">{t("vehicles.maintenance.empty")}</p>
        )}
        {selected?.maintenance?.map((entry) => {
          const linkedImages = pendingArticleImages.filter((image) => image.maintenanceId === entry.id).length;
          const linkedAttachments = (selected.attachments || []).filter((attachment) => attachment.maintenanceId === entry.id).length;
          return (
            <article key={entry.id} className={maintenanceIsDue(entry) ? "maintenance-card due" : "maintenance-card"}>
              <div className="maintenance-card-head">
                <div>
                  <strong>{entry.kind}</strong>
                  <span>{entry.notes || t("vehicles.maintenance.noNote")}</span>
                </div>
                <span className={`maintenance-badge ${maintenanceStatusClass(entry.status)}`}>{t(`vehicles.maintenance.status.${normalizeMaintenanceStatus(entry.status)}`)}</span>
              </div>
              <dl className="maintenance-meta">
                <div>
                  <dt>{t("vehicles.maintenance.due")}</dt>
                  <dd>{formatDate(entry.dueDate)}</dd>
                </div>
                <div>
                  <dt>{t("vehicles.maintenance.completedAt")}</dt>
                  <dd>{formatDate(entry.completedAt)}</dd>
                </div>
                <div>
                  <dt>{t("vehicles.maintenance.condition")}</dt>
                  <dd>{entry.conditionRating || "-"}</dd>
                </div>
                <div>
                  <dt>{t("vehicles.maintenance.cost")}</dt>
                  <dd>{formatMaintenanceCost(entry.cost)}</dd>
                </div>
              </dl>
              {(linkedImages > 0 || linkedAttachments > 0) && (
                <div className="maintenance-linked-media" aria-label={t("vehicles.maintenance.linkedMedia")}>
                  {linkedImages > 0 && (
                    <span><Image size={14} aria-hidden="true" /> {linkedImages} {t("vehicles.maintenance.images")}</span>
                  )}
                  {linkedAttachments > 0 && (
                    <span><FileText size={14} aria-hidden="true" /> {linkedAttachments} {t("vehicles.maintenance.attachments")}</span>
                  )}
                </div>
              )}
              <div className="maintenance-card-actions">
                {entry.status !== "erledigt" && (
                  <button type="button" className="secondary-button" onClick={() => onCompleteMaintenance(entry)} disabled={readonly || saving}>
                    {t("vehicles.maintenance.done")}
                  </button>
                )}
                <button type="button" className="icon-button" onClick={() => onEditMaintenance(entry)} disabled={readonly || saving} aria-label={t("vehicles.maintenance.edit")} title={t("vehicles.maintenance.edit")}>
                  <Pencil size={15} />
                </button>
                <button type="button" className="icon-button danger" onClick={() => onDeleteMaintenance(entry)} disabled={readonly || saving} aria-label={t("vehicles.maintenance.delete")} title={t("vehicles.maintenance.delete")}>
                  <Trash2 size={15} />
                </button>
              </div>
            </article>
          );
        })}
      </section>
    </section>
  );
}
