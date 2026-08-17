import { Barcode, PackageSearch, Search } from "lucide-react";

import type { CreateVehicleRequest } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { VehicleCreateArticleResults } from "./VehicleCreateArticleResults";
import { VehicleCreateArticleReview } from "./VehicleCreateArticleReview";
import type { VehicleCreateWizardAction, VehicleCreateWizardState } from "./vehicleCreateWizardState";
import type { ArticleSearchController } from "./useArticleSearchController";

export function VehicleCreateStepArticle({ state, dispatch, controller, onUpdateShared }: {
  state: VehicleCreateWizardState;
  dispatch: (action: VehicleCreateWizardAction) => void;
  controller: ArticleSearchController;
  onUpdateShared: (patch: Partial<CreateVehicleRequest>) => void;
}) {
  const { t } = useI18n();
  const response = controller.state.response;
  const selectedIndex = state.selectedResultIndex ?? 0;
  const selectedResult = response?.results[selectedIndex];
  const continueToDetails = () => dispatch({ type: "go-to-step", step: "details" });

  if (state.articleStage === "review" && selectedResult) {
    return <div className="vehicle-wizard-page"><VehicleCreateArticleReview result={selectedResult}
      resultIndex={selectedIndex} current={state.shared} controller={controller}
      memberCount={state.kind === "set" ? state.members.length : 0}
      imageOwners={state.articleImageOwners || {}}
      onAssignImage={(imageURL, memberIndex) => dispatch({
        type: "assign-article-image", imageURL, memberIndex
      })} /></div>;
  }
  if (state.articleStage === "results" && response) {
    return <div className="vehicle-wizard-page"><VehicleCreateArticleResults response={response}
      onRevise={() => dispatch({ type: "set-article-stage", stage: "input" })}
      onSelect={(index) => dispatch({ type: "select-article-result", index })} /></div>;
  }

  return (
    <div className="vehicle-wizard-page">
      <section className="vehicle-wizard-section vehicle-form">
        <div className="vehicle-wizard-section-head">
          <div><span>02</span><h3>{t("vehicles.wizard.articleData")}</h3></div>
          <small>{t("vehicles.wizard.optional")}</small>
        </div>
        <p className="vehicle-wizard-intro">{t("vehicles.wizard.articleIntro")}</p>
        <div className="vehicle-lookup-actions">
          <button type="button" className="vehicle-lookup-card" onClick={() => {
            controller.commands.openBarcode();
          }}><Barcode size={22} /><span><strong>{t("vehicles.articleSearch.barcode")}</strong>
            <small>{t("vehicles.wizard.barcodeHint")}</small></span></button>
          <button type="button" className="vehicle-lookup-card" onClick={() => {
            dispatch({ type: "set-article-stage", stage: "results" });
            controller.commands.run(state.shared);
          }}><PackageSearch size={22} /><span><strong>{t("vehicles.articleSearch.search")}</strong>
            <small>{t("vehicles.wizard.webHint")}</small></span></button>
        </div>
        {controller.state.barcodeOpen && (
          <div className="vehicle-create-barcode-input">
            <label>{t("vehicles.articleSearch.barcodeTitle")}
              <input value={controller.state.barcodeValue} autoFocus
                onChange={(event) => controller.setters.setBarcodeValue(event.target.value)} />
            </label>
            <button type="button" className="primary-button" onClick={() => {
              const ean = controller.state.barcodeValue.trim();
              if (!ean) return;
              onUpdateShared({ ean });
              controller.setters.setBarcodeOpen(false);
              dispatch({ type: "set-article-stage", stage: "results" });
              controller.commands.run({ ...state.shared, ean }, { fields: { ean } });
            }}>{t("vehicles.articleSearch.barcode")}</button>
          </div>
        )}
        <div className="form-row three-columns">
          <label>{t("vehicle.field.articleNumber")}<input value={state.shared.articleNumber || ""}
            onChange={(event) => onUpdateShared({ articleNumber: event.target.value })} /></label>
          <label>EAN<input value={state.shared.ean || ""}
            onChange={(event) => onUpdateShared({ ean: event.target.value })} /></label>
          <label>{t("vehicle.field.productionPeriod")}<input value={state.shared.productionPeriod || ""}
            onChange={(event) => onUpdateShared({ productionPeriod: event.target.value })} /></label>
        </div>
        {controller.state.loading && <p className="empty-state compact">{t("vehicles.articleSearch.loading")}</p>}
        {controller.state.error && <p className="form-message">{controller.state.error}</p>}
        <button type="button" className="vehicle-manual-note" onClick={continueToDetails}>
          <Search size={17} /><span>{t("vehicles.wizard.manualHint")}</span>
        </button>
      </section>
    </div>
  );
}
