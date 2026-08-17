import { Layers3, TrainFront } from "lucide-react";
import type { ReactNode } from "react";

import type { CreateVehicleRequest, InventoryNumberScheme, MasterDataEntry } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { RequiredLabel } from "./VehicleFormFields";
import type { MasterDataOptions } from "./vehicleViewModel";
import {
  maximumVehicleSetMembers,
  type VehicleCreateWizardAction,
  type VehicleCreateWizardState
} from "./vehicleCreateWizardState";

type VehicleCreateStepBasicsProps = {
  state: VehicleCreateWizardState;
  dispatch: (action: VehicleCreateWizardAction) => void;
  options: MasterDataOptions;
  filteredGattungen: MasterDataEntry[];
  selectOptions: (entries: MasterDataEntry[], currentValue: string, emptyLabel?: string) => ReactNode;
  setScheme: InventoryNumberScheme | null;
  setSchemeLoading: boolean;
  setSchemeError: string;
  setCreationDisabled: boolean;
  onUpdateShared: (patch: Partial<CreateVehicleRequest>) => void;
  onUpdateCategory: (category: string) => void;
};

export function inventoryNumberPreview(scheme: InventoryNumberScheme) {
  return `${scheme.prefix}-${String(scheme.nextNumber).padStart(scheme.padding, "0")}`;
}

export function VehicleCreateStepBasics({
  state,
  dispatch,
  options,
  selectOptions,
  setScheme,
  setSchemeLoading,
  setSchemeError,
  setCreationDisabled,
  onUpdateShared,
  onUpdateCategory
}: VehicleCreateStepBasicsProps) {
  const { t } = useI18n();
  const { kind, shared: form } = state;

  return (
    <div className="vehicle-wizard-page vehicle-create-basics">
      <section className="vehicle-wizard-section vehicle-form">
        <div className="vehicle-wizard-section-head">
          <div><span>01</span><h3>{t("vehicles.wizard.basicData")}</h3></div>
          <small>{t("vehicles.wizard.requiredHint")}</small>
        </div>

        <div className="vehicle-kind-grid" role="radiogroup" aria-label={t("vehicles.wizard.kind")}>
          <button type="button" className={kind === "single" ? "vehicle-kind-card selected" : "vehicle-kind-card"}
            onClick={() => dispatch({ type: "set-kind", kind: "single" })}
            role="radio" aria-checked={kind === "single"}>
            <TrainFront size={24} />
            <span><strong>{t("vehicles.wizard.single")}</strong><small>{t("vehicles.wizard.singleHint")}</small></span>
          </button>
          <button type="button" className={kind === "set" ? "vehicle-kind-card selected" : "vehicle-kind-card"}
            onClick={() => dispatch({ type: "set-kind", kind: "set" })}
            disabled={setCreationDisabled} title={setCreationDisabled ? t("vehicles.wizard.setDisabledEcos") : undefined}
            role="radio" aria-checked={kind === "set"}>
            <Layers3 size={24} />
            <span><strong>{t("vehicles.wizard.set")}</strong><small>{t("vehicles.wizard.setHint")}</small></span>
          </button>
        </div>

        {setSchemeError && <p className="form-message vehicle-set-scheme-error" role="status">{setSchemeError}</p>}
        {kind === "set" && setScheme && (
          <p className="vehicle-set-number-preview">
            <span>{t("vehicles.wizard.setNumber")}</span>
            <strong>{inventoryNumberPreview(setScheme)} <small>({t("vehicles.wizard.provisional")})</small></strong>
          </p>
        )}
        {kind === "set" && setSchemeLoading && <p className="muted">{t("vehicles.wizard.setSchemeLoading")}</p>}

        <div className="form-row">
          <label>
            <RequiredLabel label={t("vehicle.field.manufacturer")} filled={Boolean(form.manufacturer.trim())} />
            <AppSelect value={form.manufacturer} onChange={(event) => onUpdateShared({ manufacturer: event.target.value })}
              required>{selectOptions(options.manufacturers, form.manufacturer, t("vehicles.select.placeholder"))}</AppSelect>
          </label>
          <label>
            <RequiredLabel label={t("vehicle.field.name")} filled={Boolean(form.name.trim())} />
            <input value={form.name} onChange={(event) => onUpdateShared({ name: event.target.value })} required autoFocus />
          </label>
        </div>
        <div className="form-row">
          <label>{t("vehicle.field.articleNumber")}
            <input value={form.articleNumber || ""}
              onChange={(event) => onUpdateShared({ articleNumber: event.target.value })} />
          </label>
          <label>
            <RequiredLabel label={t("vehicle.field.gauge")} filled={Boolean(form.gauge.trim())} />
            <AppSelect value={form.gauge} onChange={(event) => onUpdateShared({ gauge: event.target.value })} required>
              {selectOptions(options.gauges, form.gauge, t("vehicles.select.placeholder"))}
            </AppSelect>
          </label>
        </div>
        <div className="form-row">
          <label>
            <RequiredLabel label={t("vehicle.field.category")} filled={Boolean(form.category?.trim())} />
            <AppSelect value={form.category || ""} onChange={(event) => onUpdateCategory(event.target.value)} required>
              {selectOptions(options.categories, form.category || "", t("vehicles.select.placeholder"))}
            </AppSelect>
          </label>
          {kind === "set" && (
            <label>{t("vehicles.wizard.memberCount")}
              <input type="number" min={2} max={maximumVehicleSetMembers} value={state.members.length}
                onChange={(event) => dispatch({ type: "set-member-count", count: Number(event.target.value) })} />
            </label>
          )}
        </div>

        {state.pendingMemberReduction && (
          <div className="vehicle-member-reduction" role="alert">
            <p>{t("vehicles.wizard.memberReductionWarning", {
              count: state.pendingMemberReduction.populatedIndexes.length
            })}</p>
            <button type="button" className="secondary-button"
              onClick={() => dispatch({ type: "cancel-member-reduction" })}>{t("common.cancel")}</button>
            <button type="button" className="danger-button"
              onClick={() => dispatch({ type: "confirm-member-reduction" })}>{t("vehicles.wizard.confirmReduction")}</button>
          </div>
        )}
      </section>
    </div>
  );
}
