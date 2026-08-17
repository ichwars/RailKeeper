import type { ComponentProps } from "react";

import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { RequiredLabel } from "./VehicleFormFields";
import { VehicleModelTab } from "./VehicleModelTab";
import { VehicleSetDetailsTabs } from "./VehicleSetDetailsTabs";
import type { VehicleCreateWizardAction, VehicleCreateWizardState } from "./vehicleCreateWizardState";

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
        <div className="form-row">
          <label>{t("vehicle.field.acquisitionType")}<input value={state.shared.acquisitionType || ""}
            onChange={(event) => updateShared({ acquisitionType: event.target.value })} /></label>
          <label>{t("vehicle.field.acquiredFrom")}<input value={state.shared.acquiredFrom || ""}
            onChange={(event) => updateShared({ acquiredFrom: event.target.value })} /></label>
        </div>
        <div className="form-row">
          <label>{t("vehicle.field.purchasePrice")}<input value={state.shared.purchasePrice || ""}
            onChange={(event) => updateShared({ purchasePrice: event.target.value })} /></label>
          <label>{t("vehicle.field.purchaseDate")}<input type="date" value={state.shared.purchaseDate || ""}
            onChange={(event) => updateShared({ purchaseDate: event.target.value })} /></label>
        </div>
        <div className="form-row">
          <label>{t("vehicle.field.storageLocation")}<input value={state.shared.storageLocation || ""}
            onChange={(event) => updateShared({ storageLocation: event.target.value })} /></label>
          <label>{t("vehicle.field.condition")}<input value={state.shared.condition || ""}
            onChange={(event) => updateShared({ condition: event.target.value })} /></label>
        </div>
      </details>
    </div>
  );
  const memberPanel = (index: number) => {
    const member = state.members[index];
    if (!member) return null;
    const updateMember = (patch: Partial<typeof member.form>) => dispatch({ type: "update-member", index, patch });
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
        <details open><summary>{t("vehicles.wizard.identification")}</summary>
          <div className="form-row">
            <label>{t("vehicle.field.name")}<input value={member.form.name}
              onChange={(event) => updateMember({ name: event.target.value })} /></label>
            <label>{t("vehicle.field.vehicleNumber")}<input value={member.form.vehicleNumber || ""}
              onChange={(event) => updateMember({ vehicleNumber: event.target.value })} /></label>
          </div>
        </details>
        <details><summary>{t("vehicles.wizard.technicalAndDecoder")}</summary>
          <div className="form-row">
            <label>{t("vehicle.field.lengthMm")}<input value={member.form.lengthMm || ""}
              onChange={(event) => updateMember({ lengthMm: event.target.value })} /></label>
            <label>{t("vehicle.field.decoderType")}<input value={member.form.decoderType || ""}
              onChange={(event) => updateMember({ decoderType: event.target.value })} /></label>
          </div>
        </details>
        <details><summary>{t("vehicles.wizard.notesAndQr")}</summary>
          <label>{t("vehicle.field.additionalInfo")}<textarea value={member.form.additionalInfo || ""}
            onChange={(event) => updateMember({ additionalInfo: event.target.value })} /></label>
        </details>
      </div>
    );
  };
  return <div className="vehicle-wizard-page vehicle-wizard-final">
    <VehicleSetDetailsTabs state={state} dispatch={dispatch} setPanel={setPanel} memberPanel={memberPanel} />
  </div>;
}
