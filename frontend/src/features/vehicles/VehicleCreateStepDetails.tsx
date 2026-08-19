import type { ComponentProps } from "react";

import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { RequiredLabel, VehicleOwnershipFields } from "./VehicleFormFields";
import { VehicleModelTab } from "./VehicleModelTab";
import { VehicleSetDetailsTabs } from "./VehicleSetDetailsTabs";
import type { VehicleCreateWizardAction, VehicleCreateWizardState } from "./vehicleCreateWizardState";
import { gattungenForCategory } from "./vehicleViewModel";

export function VehicleCreateStepDetails({ state, dispatch, model }: {
  state: VehicleCreateWizardState;
  dispatch: (action: VehicleCreateWizardAction) => void;
  model: ComponentProps<typeof VehicleModelTab>;
}) {
  const { t } = useI18n();
  const updateShared = (patch: Partial<VehicleCreateWizardState["shared"]>) => {
    dispatch({ type: "update-shared", patch });
    model.onUpdate(patch);
  };

  if (state.kind === "single") {
    return <div className="vehicle-wizard-page vehicle-wizard-final">
      <VehicleModelTab {...model} form={state.shared} onUpdate={updateShared} />
    </div>;
  }

  const setPanel = (
    <div className="vehicle-create-detail-groups vehicle-form">
      <div className="vehicle-inherited-summary">
        <strong>{state.shared.name}</strong>
        <span>{[state.shared.manufacturer, state.shared.articleNumber, state.shared.gauge].filter(Boolean).join(" · ")}</span>
        <small>{t("vehicles.inventoryNumberAuto")}</small>
      </div>
      <div className="form-row">
        <label>
          <RequiredLabel label={t("vehicle.field.gattung")} filled={Boolean(state.shared.gattung?.trim())} />
          <AppSelect value={state.shared.gattung || ""}
            onChange={(event) => updateShared({ gattung: event.target.value })} required>
            {model.selectOptions(model.filteredGattungen, state.shared.gattung || "", t("vehicles.select.placeholder"))}
          </AppSelect>
        </label>
      </div>
      <details open><summary>{t("vehicles.wizard.acquisitionAndStock")}</summary>
        <VehicleOwnershipFields
          form={state.shared}
          readonly={false}
          hideAdditionalInfo
          update={updateShared}
        />
      </details>
    </div>
  );
  const memberPanel = (index: number) => {
    const member = state.members[index];
    if (!member) return null;
    const updateMember = (patch: Partial<typeof member.form>) => dispatch({ type: "update-member", index, patch });
    const memberOverrides = Object.fromEntries(
      (member.overriddenFields || []).map((field) => [field, member.form[field]])
    ) as Partial<typeof member.form>;
    const memberForm = {
      ...state.shared,
      ...memberOverrides,
      inventoryNumber: "",
      name: member.form.name,
      vehicleNumber: member.form.vehicleNumber
    };
    return (
      <div className="vehicle-create-detail-groups vehicle-form">
        <div className="vehicle-inherited-summary">
          <strong>{t("vehicles.wizard.inheritedData")}</strong>
          <span>{[
            state.shared.manufacturer, state.shared.articleNumber, state.shared.gauge, state.shared.epoch,
            state.shared.category, state.shared.gattung
          ].filter(Boolean).join(" · ")}</span>
          <small>{t("vehicles.inventoryNumberAuto")}</small>
        </div>
        <VehicleModelTab {...model}
          form={memberForm}
          onUpdate={updateMember}
          onUpdateCategory={(category) => updateMember({ category, gattung: "" })}
          filteredGattungen={gattungenForCategory(model.options, memberForm.category)}
          onUpdateCouplingFront={(couplingFront) => updateMember({
            couplingFront,
            ...(memberForm.couplingSame ? { couplingRear: couplingFront } : {})
          })}
          onUpdateCouplingSame={(couplingSame) => updateMember({
            couplingSame,
            ...(couplingSame ? { couplingRear: memberForm.couplingFront || "" } : {})
          })}
          onOpenQr={() => undefined}
          canOpenQr={false}
          hideInventoryNumber
          hideArticleSearch
          hideAcquisitionDetails
        />
      </div>
    );
  };
  return <div className="vehicle-wizard-page vehicle-wizard-final">
    <VehicleSetDetailsTabs state={state} dispatch={dispatch} setPanel={setPanel} memberPanel={memberPanel} />
  </div>;
}
