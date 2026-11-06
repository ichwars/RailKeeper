import { useI18n } from "../../shared/i18n";
import type { ModalTab } from "./vehicleViewModel";
import type { VehicleSavedAction } from "./useVehicleEditorController";
import type { DraftRefreshConflict } from "./vehicleDraftRefresh";

type Props = {
  tab: ModalTab;
  saved: VehicleSavedAction | null;
  conflicts: DraftRefreshConflict[];
  onResolve: (choice: "draft" | "saved") => void;
};

export function VehicleSaveState({ tab, saved, conflicts, onResolve }: Props) {
  const { t } = useI18n();
  const timing = tab === "speedCurve" ? "read" : tab === "uploads" ? "media" : tab === "cv" ? "cv"
    : tab === "maintenance" || tab === "spareParts" ? "side" : "draft";
  const display = (value: unknown) => typeof value === "boolean" ? t(value ? "common.yes" : "common.no")
    : value === undefined || value === null || value === "" ? t("vehicles.saveState.empty") : String(value);
  return (
    <aside className="vehicle-save-state" aria-label={t("vehicles.saveState.title")}>
      <p>{t(`vehicles.saveState.${timing}`)}</p>
      {saved && <p className="vehicle-save-result" role="status">
        {t("vehicles.saveState.saved", { area: t(`vehicles.saveState.area.${saved}`) })}
      </p>}
      {conflicts.length > 0 && <div className="vehicle-refresh-conflicts" role="alert">
        <strong>{t("vehicles.saveState.conflict")}</strong>
        <div className="table-wrap">
          <table>
            <thead><tr><th>{t("vehicles.saveState.field")}</th><th>{t("vehicles.saveState.local")}</th>
              <th>{t("vehicles.saveState.stored")}</th></tr></thead>
            <tbody>{conflicts.map(conflict => <tr key={conflict.field}>
              <th scope="row">{t(`vehicle.field.${conflict.field}`)}</th>
              <td>{display(conflict.local)}</td><td>{display(conflict.saved)}</td>
            </tr>)}</tbody>
          </table>
        </div>
        <div className="vehicle-refresh-actions">
          <button type="button" className="secondary-button" onClick={() => onResolve("saved")}>
            {t("vehicles.saveState.useStored")}
          </button>
          <button type="button" className="secondary-button" onClick={() => onResolve("draft")}>
            {t("vehicles.saveState.keepDraft")}
          </button>
        </div>
      </div>}
    </aside>
  );
}
