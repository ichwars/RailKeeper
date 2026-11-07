import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import type { MediaDraftConflict } from "./mediaDraftConflicts";

export function VehicleMediaConflicts({ conflicts, saving, onResolve, vehicle }: {
  vehicle: Vehicle | null;
  conflicts: MediaDraftConflict[];
  saving: boolean;
  onResolve: (conflict: MediaDraftConflict, choice: "draft" | "saved") => void;
}) {
  const { t } = useI18n();
  if (!conflicts.length) return null;
  const display = (value: unknown, field: string) => {
    if (field === "maintenanceId" && typeof value === "string" && value) {
      const entry = vehicle?.maintenance?.find(item => item.id === value);
      return entry ? `${entry.kind} · ${entry.dueDate || entry.createdAt.slice(0, 10)}` : t("vehicles.mediaConflict.missingMaintenance");
    }
    return typeof value === "boolean" ? t(value ? "common.yes" : "common.no")
    : value === undefined || value === null || value === "" ? t("vehicles.saveState.empty") : String(value);
  };
  return <section className="vehicle-media-conflicts vehicle-refresh-conflicts" role="alert"
    aria-label={t("vehicles.mediaConflict.title")}>
    <h3>{t("vehicles.mediaConflict.title")}</h3>
    <p>{t("vehicles.mediaConflict.resolve")}</p>
    <p className="vehicle-media-conflict-count">{t("vehicles.mediaConflict.count", { count: conflicts.length })}</p>
    <div className="vehicle-media-conflict-list">
      {conflicts.map(conflict => <article key={`${conflict.kind}:${conflict.id}`}>
        <strong>{t(`vehicles.mediaConflict.${conflict.kind}`)}{conflict.label ? `: ${conflict.label}` : ""}</strong>
        {conflict.deleted && <p>{t("vehicles.mediaConflict.deleted")}</p>}
        <dl>{conflict.fields.map(field => <div key={field.field}>
          <dt>{t(`vehicles.mediaConflict.field.${field.field}`)}</dt>
          <dd><span>{t("vehicles.saveState.local")}</span><span>{display(field.local, field.field)}</span></dd>
          <dd><span>{t("vehicles.saveState.stored")}</span><span>{display(field.saved, field.field)}</span></dd>
        </div>)}</dl>
        <div className="vehicle-refresh-actions">
          <button type="button" className="secondary-button" disabled={saving} onClick={() => onResolve(conflict, "saved")}>
            {t(conflict.deleted ? "vehicles.mediaConflict.discardDeleted" : "vehicles.saveState.useStored")}
          </button>
          {!conflict.deleted && <button type="button" className="secondary-button" disabled={saving}
            onClick={() => onResolve(conflict, "draft")}>{t("vehicles.saveState.keepDraft")}</button>}
        </div>
      </article>)}
    </div>
  </section>;
}
