import { useMemo, useRef, useState } from "react";
import { TriangleAlert, X } from "lucide-react";

import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppTextInput } from "../../shared/ui/AppTextInput";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";
import {
  digitalCenterVehicleMatchReason,
  rankDigitalCenterVehicleCandidates,
  type DigitalCenterVehicleAdoptionProvider
} from "./digitalCenterVehicleAdoption";

type VehicleAssignmentDialogProps = {
  item: DigitalCenterWorkItem;
  provider: DigitalCenterVehicleAdoptionProvider;
  vehicles: Vehicle[];
  selectedVehicleId: string;
  loading: boolean;
  saving: boolean;
  error: string;
  onSelect: (vehicleId: string) => void;
  onAssign: (vehicleId: string) => void | Promise<void>;
  onClose: () => void;
};

export function VehicleAssignmentDialog({
  item,
  provider,
  vehicles,
  selectedVehicleId,
  loading,
  saving,
  error,
  onSelect,
  onAssign,
  onClose
}: VehicleAssignmentDialogProps) {
  const { t } = useI18n();
  const [query, setQuery] = useState("");
  const searchRef = useRef<HTMLInputElement>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(onClose, searchRef);
  const candidates = useMemo(
    () => rankDigitalCenterVehicleCandidates(item, vehicles, query, provider),
    [item, provider, query, vehicles]
  );

  return <>
    <span ref={anchorRef} aria-hidden="true" />
    <div ref={layerRef} className="digital-assignment-layer" role="dialog" aria-modal="true"
      aria-label={t("digitalCenters.assignment.dialogLabel", { name: item.name })} onKeyDown={onKeyDown}>
      <section className="digital-assignment-dialog">
        <header>
          <div>
            <p className="eyebrow">{t("digitalCenters.assignment.eyebrow")}</p>
            <h2>{t("digitalCenters.assignment.title")}</h2>
            <p>{t("digitalCenters.assignment.subtitle", { name: item.name })}</p>
          </div>
          <button type="button" className="digital-center-icon-button"
            aria-label={t("digitalCenters.assignment.close")} onClick={onClose}>
            <X size={19} aria-hidden="true" />
          </button>
        </header>

        <div className="digital-assignment-body">
          <AppTextInput ref={searchRef} label={t("digitalCenters.assignment.search")}
            value={query} onChange={(event) => setQuery(event.target.value)}
            placeholder={t("digitalCenters.assignment.searchPlaceholder")} />

          {loading ? <p className="digital-assignment-state">{t("digitalCenters.assignment.loading")}</p> :
            candidates.length === 0 ? <p className="digital-assignment-state">
              {t("digitalCenters.assignment.noResults")}
            </p> : <div className="digital-assignment-candidates">
              {candidates.map((vehicle) => {
                const reason = digitalCenterVehicleMatchReason(item, vehicle, provider);
                return <label key={vehicle.id} className="digital-assignment-candidate">
                  <input type="radio" name="digital-center-vehicle" value={vehicle.id}
                    checked={selectedVehicleId === vehicle.id}
                    onChange={() => onSelect(vehicle.id)} />
                  <span className="digital-assignment-candidate-copy">
                    <strong>{vehicle.inventoryNumber} · {vehicle.name}</strong>
                    <small>{vehicle.manufacturer} · {vehicle.articleNumber || "–"}</small>
                    <span>{t("digitalCenters.assignment.address", {
                      value: vehicle.digitalDecoderNumber || "–"
                    })}</span>
                  </span>
                  {reason && <span className="digital-assignment-reason">
                    {t(`digitalCenters.assignment.match.${reason}`)}
                  </span>}
                </label>;
              })}
            </div>}

          {error && <p className="digital-write-result error" role="alert">
            <TriangleAlert size={17} aria-hidden="true" />{error}
          </p>}
        </div>

        <footer>
          <button type="button" className="digital-center-button"
            disabled={!selectedVehicleId || loading || saving}
            onClick={() => void onAssign(selectedVehicleId)}>
            {saving ? t("digitalCenters.assignment.saving") : t("digitalCenters.assignment.confirm")}
          </button>
          <button type="button" className="digital-center-button" onClick={onClose}>
            {t("digitalCenters.assignment.cancel")}
          </button>
        </footer>
      </section>
    </div>
  </>;
}
