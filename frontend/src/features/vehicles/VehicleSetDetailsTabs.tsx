import { Plus } from "lucide-react";
import { useRef, type KeyboardEvent, type ReactNode } from "react";

import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import type { VehicleCreateWizardAction, VehicleCreateWizardState } from "./vehicleCreateWizardState";

const tabId = (value: VehicleCreateWizardState["activeDetailsTab"]) => `vehicle-create-tab-${value}`;
const panelId = (value: VehicleCreateWizardState["activeDetailsTab"]) => `vehicle-create-panel-${value}`;

export function VehicleSetDetailsTabs({ state, dispatch, setPanel, memberPanel }: {
  state: VehicleCreateWizardState;
  dispatch: (action: VehicleCreateWizardAction) => void;
  setPanel: ReactNode;
  memberPanel: (index: number) => ReactNode;
}) {
  const { t } = useI18n();
  const tabs: VehicleCreateWizardState["activeDetailsTab"][] = [
    "set",
    ...state.members.map((_, index) => `member:${index}` as const)
  ];
  const refs = useRef<Array<HTMLButtonElement | null>>([]);
  const select = (tab: VehicleCreateWizardState["activeDetailsTab"]) => {
    dispatch({ type: "set-active-details-tab", tab });
  };
  const onKeyDown = (event: KeyboardEvent<HTMLButtonElement>, index: number) => {
    if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") return;
    event.preventDefault();
    const offset = event.key === "ArrowRight" ? 1 : -1;
    const nextIndex = (index + offset + tabs.length) % tabs.length;
    select(tabs[nextIndex]);
    refs.current[nextIndex]?.focus();
  };
  const label = (tab: VehicleCreateWizardState["activeDetailsTab"]) => tab === "set"
    ? t("vehicles.wizard.setAcquisition")
    : t("vehicles.wizard.memberLabel", { count: Number(tab.split(":")[1]) + 1 });
  const memberIndex = state.activeDetailsTab === "set" ? -1 : Number(state.activeDetailsTab.split(":")[1]);

  return (
    <div className="vehicle-set-details-tabs">
      <div className="vehicle-set-details-tabbar" role="tablist">
        {tabs.map((tab, index) => (
          <button key={tab} ref={(node) => { refs.current[index] = node; }} type="button" role="tab"
            id={tabId(tab)} aria-controls={panelId(tab)} aria-selected={state.activeDetailsTab === tab}
            tabIndex={state.activeDetailsTab === tab ? 0 : -1} onClick={() => select(tab)}
            onKeyDown={(event) => onKeyDown(event, index)}>{label(tab)}</button>
        ))}
        <button type="button" className="vehicle-add-member" onClick={() => dispatch({ type: "add-member" })}>
          <Plus size={15} />{t("vehicles.wizard.addMember")}
        </button>
      </div>
      <label className="vehicle-set-details-select">{t("vehicles.wizard.detailsFor")}
        <AppSelect value={state.activeDetailsTab}
          onChange={(event) => select(event.target.value as VehicleCreateWizardState["activeDetailsTab"])}>
          {tabs.map((tab) => <option key={tab} value={tab}>{label(tab)}</option>)}
        </AppSelect>
      </label>
      <div role="tabpanel" id={panelId(state.activeDetailsTab)} aria-labelledby={tabId(state.activeDetailsTab)}>
        {state.activeDetailsTab === "set" ? setPanel : memberPanel(memberIndex)}
      </div>
    </div>
  );
}
