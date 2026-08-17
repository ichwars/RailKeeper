import { ChevronLeft, ChevronRight } from "lucide-react";
import { useEffect, useReducer, useState, type ComponentProps, type FormEventHandler } from "react";

import { api, type CreateVehicleRequest, type InventoryNumberScheme } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { VehicleCreateStepBasics } from "./VehicleCreateStepBasics";
import { VehicleCreateStepArticle } from "./VehicleCreateStepArticle";
import { VehicleCreateStepDetails } from "./VehicleCreateStepDetails";
import { VehicleCreateWizardShell } from "./VehicleCreateWizardShell";
import { VehicleModelTab } from "./VehicleModelTab";
import {
  clearVehicleCreateDraft,
  createVehicleCreateWizardState,
  loadVehicleCreateDraft,
  saveVehicleCreateDraft,
  vehicleCreateWizardReducer,
  type VehicleSetMemberDraft
} from "./vehicleCreateWizardState";
import type { VehicleCreatePrefill } from "./vehicleSetDuplicate";
import type { PendingArticleImage } from "./vehicleTransforms";
import type { ArticleSearchController } from "./useArticleSearchController";

type VehicleCreateWizardProps = {
  model: ComponentProps<typeof VehicleModelTab>;
  saving: boolean;
  message: string;
  onSubmitSingle: FormEventHandler<HTMLFormElement>;
  onSubmitSet: (members: VehicleSetMemberDraft[]) => Promise<void>;
  onClose: () => void;
  setCreationDisabled?: boolean;
  prefill?: VehicleCreatePrefill | null;
  articleSearchController: ArticleSearchController;
};

export function VehicleCreateWizard({
  model, saving, message, onSubmitSingle, onSubmitSet, onClose,
  setCreationDisabled = false, prefill, articleSearchController
}: VehicleCreateWizardProps) {
  const { t } = useI18n();
  const [wizard, dispatch] = useReducer(
    vehicleCreateWizardReducer,
    undefined,
    () => createVehicleCreateWizardState(model.form, prefill)
  );
  const [setScheme, setSetScheme] = useState<InventoryNumberScheme | null>(null);
  const [setSchemeLoading, setSetSchemeLoading] = useState(true);
  const [setSchemeError, setSetSchemeError] = useState("");
  const [draftResult, setDraftResult] = useState(() => loadVehicleCreateDraft());
  const [draftMessage, setDraftMessage] = useState("");
  const { kind, members, step, shared: form } = wizard;
  const { options, filteredGattungen, selectOptions, onUpdate, onUpdateCategory } = model;

  useEffect(() => dispatch({ type: "update-shared", patch: model.form }), [model.form]);

  useEffect(() => {
    let active = true;
    api.inventoryNumberSchemes()
      .then((schemes) => {
        if (!active) return;
        const scheme = schemes.find((item) => item.active && item.category.toLocaleLowerCase() === "set") || null;
        setSetScheme(scheme);
        setSetSchemeError(scheme ? "" : t("vehicles.wizard.setSchemeMissing"));
      })
      .catch(() => {
        if (active) setSetSchemeError(t("vehicles.wizard.setSchemeLoadFailed"));
      })
      .finally(() => {
        if (active) setSetSchemeLoading(false);
      });
    return () => { active = false; };
  }, [t]);

  const updateShared = (patch: Partial<CreateVehicleRequest>) => {
    dispatch({ type: "update-shared", patch });
    onUpdate(patch);
  };
  const updateCategory = (category: string) => {
    dispatch({ type: "update-shared", patch: { category } });
    onUpdateCategory(category);
  };
  const basicsComplete = Boolean(
    form.name.trim() && form.manufacturer.trim() && form.gauge.trim() && form.category?.trim()
  );
  const canContinueFromBasics = basicsComplete && (
    kind === "single" || (!setSchemeLoading && Boolean(setScheme))
  );

  const submit: FormEventHandler<HTMLFormElement> = (event) => {
    if (step !== "details") {
      event.preventDefault();
      if (step === "basics" && !canContinueFromBasics) return;
      dispatch({ type: "go-to-step", step: step === "basics" ? "article" : "details" });
      return;
    }
    if (kind === "set") {
      event.preventDefault();
      void onSubmitSet(members);
      return;
    }
    onSubmitSingle(event);
  };

  const back = () => {
    if (step === "basics") onClose();
    else dispatch({ type: "go-to-step", step: step === "details" ? "article" : "basics" });
  };
  const destination = step === "basics" ? t("vehicles.wizard.nextArticle")
    : step === "article" ? t("vehicles.wizard.nextDetails")
      : saving ? t("vehicles.saving")
        : kind === "set" ? t("vehicles.wizard.createSet") : t("vehicles.createAndContinue");
  const summaries = {
    basics: [kind === "set" ? t("vehicles.wizard.set") : t("vehicles.wizard.single"), form.manufacturer]
      .filter(Boolean).join(" · "),
    article: wizard.articleStage === "input" ? t("vehicles.wizard.articleInputSummary")
      : t(wizard.articleStage === "results"
        ? "vehicles.wizard.articleResultsSummary" : "vehicles.wizard.articleReviewSummary"),
    details: kind === "set" ? t("vehicles.set.memberCount", { count: members.length })
      : t("vehicles.wizard.singleDetailsSummary")
  };
  const footer = (
    <>
      {message && <p className="form-message">{message}</p>}
      {draftMessage && <p className="form-message">{draftMessage}</p>}
      <button type="button" className="secondary-button" onClick={() => {
        const result = saveVehicleCreateDraft(wizard);
        setDraftMessage(result.kind === "saved" ? t("vehicles.wizard.draftSaved") : t("vehicles.wizard.draftFailed"));
      }}>{t("vehicles.wizard.saveDraft")}</button>
      <button type="button" className="secondary-button" onClick={back}>
        {step === "basics" ? t("vehicles.cancel") : <><ChevronLeft size={16} />{t("vehicles.wizard.back")}</>}
      </button>
      <button type="submit" className="primary-button"
        disabled={saving || (step === "basics" && !canContinueFromBasics)}>
        {destination}{step !== "details" && <ChevronRight size={16} />}
      </button>
    </>
  );

  return (
    <VehicleCreateWizardShell step={step} summaries={summaries} onClose={onClose} onSubmit={submit} footer={footer}>
      {draftResult.kind === "loaded" && (
        <div className="vehicle-draft-notice" role="status">
          <p>{t("vehicles.wizard.draftFound")}</p>
          <button type="button" className="secondary-button" onClick={() => {
            dispatch({ type: "replace-state", state: draftResult.state });
            setDraftResult({ kind: "empty" });
          }}>{t("vehicles.wizard.resumeDraft")}</button>
          <button type="button" className="secondary-button" onClick={() => {
            clearVehicleCreateDraft(); setDraftResult({ kind: "empty" });
          }}>{t("vehicles.wizard.discardDraft")}</button>
        </div>
      )}
      {(draftResult.kind === "invalid" || draftResult.kind === "error") && (
        <div className="vehicle-draft-notice" role="status"><p>{t("vehicles.wizard.draftInvalid")}</p>
          <button type="button" className="secondary-button" onClick={() => {
            clearVehicleCreateDraft(); setDraftResult({ kind: "empty" });
          }}>{t("vehicles.wizard.cleanStart")}</button></div>
      )}
      {step === "basics" && (
        <VehicleCreateStepBasics state={wizard} dispatch={dispatch} options={options}
          filteredGattungen={filteredGattungen} selectOptions={selectOptions}
          setScheme={setScheme} setSchemeLoading={setSchemeLoading} setSchemeError={setSchemeError}
          setCreationDisabled={setCreationDisabled} onUpdateShared={updateShared} onUpdateCategory={updateCategory} />
      )}

      {step === "article" && (
        <VehicleCreateStepArticle state={wizard} dispatch={dispatch} controller={articleSearchController}
          onUpdateShared={updateShared} />
      )}

      {step === "details" && (
        <VehicleCreateStepDetails state={wizard} dispatch={dispatch} model={{
          ...model, form, onUpdate: updateShared, onUpdateCategory: updateCategory
        }} />
      )}
    </VehicleCreateWizardShell>
  );
}

export function vehicleSetInputFromForm(form: CreateVehicleRequest) {
  return {
    name: form.name, manufacturer: form.manufacturer, articleNumber: form.articleNumber,
    articleSourceUrl: form.articleSourceUrl, gauge: form.gauge, epoch: form.epoch,
    railwayCompany: form.railwayCompany, category: form.category, gattung: form.gattung,
    description: form.description, ean: form.ean, productionPeriod: form.productionPeriod,
    listPrice: form.listPrice, acquisitionType: form.acquisitionType, acquiredFrom: form.acquiredFrom,
    purchasePrice: form.purchasePrice, purchaseDate: form.purchaseDate, storageLocation: form.storageLocation,
    storageDetails: form.storageDetails, condition: form.condition, conditionDetails: form.conditionDetails,
    packaging: form.packaging
  };
}

export function vehicleSetMembersFromForm(
  form: CreateVehicleRequest,
  members: VehicleSetMemberDraft[],
  images: PendingArticleImage[]
): CreateVehicleRequest[] {
  const firstMemberImages = images.map((image, index) => ({
    id: image.persisted ? image.id : undefined,
    url: image.url,
    title: image.title,
    sourceUrl: image.source,
    maintenanceId: image.maintenanceId || "",
    isPrimary: Boolean(image.isPrimary),
    sortOrder: index
  }));
  return members.map((member, index) => ({
    ...form,
    ...member.form,
    inventoryNumber: "",
    manufacturer: form.manufacturer,
    articleNumber: form.articleNumber,
    articleSourceUrl: form.articleSourceUrl,
    gauge: form.gauge,
    epoch: form.epoch,
    railwayCompany: form.railwayCompany,
    category: form.category,
    gattung: form.gattung,
    name: member.form.name.trim() || `${form.name} (${index + 1})`,
    vehicleNumber: member.form.vehicleNumber,
    images: [...(member.form.images || []), ...(index === 0 ? firstMemberImages : [])]
  }));
}
