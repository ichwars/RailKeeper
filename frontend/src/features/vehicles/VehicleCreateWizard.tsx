import {
  Barcode,
  ChevronLeft,
  ChevronRight,
  Layers3,
  PackageSearch,
  Plus,
  Search,
  TrainFront,
  Trash2,
  X
} from "lucide-react";
import { useState, type ComponentProps, type FormEventHandler } from "react";

import type { CreateVehicleRequest } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { RequiredLabel } from "./VehicleFormFields";
import { VehicleModelTab } from "./VehicleModelTab";

export type VehicleSetMemberDraft = {
  inventoryNumber: string;
  name: string;
  vehicleNumber: string;
};

type VehicleCreateWizardProps = {
  model: ComponentProps<typeof VehicleModelTab>;
  saving: boolean;
  message: string;
  onSubmitSingle: FormEventHandler<HTMLFormElement>;
  onSubmitSet: (members: VehicleSetMemberDraft[]) => Promise<void>;
  onClose: () => void;
};

type CreationKind = "single" | "set";

const initialMembers: VehicleSetMemberDraft[] = [
  { inventoryNumber: "", name: "", vehicleNumber: "" },
  { inventoryNumber: "", name: "", vehicleNumber: "" }
];

export function VehicleCreateWizard({
  model,
  saving,
  message,
  onSubmitSingle,
  onSubmitSet,
  onClose
}: VehicleCreateWizardProps) {
  const { t } = useI18n();
  const [step, setStep] = useState(1);
  const [kind, setKind] = useState<CreationKind>("single");
  const [members, setMembers] = useState<VehicleSetMemberDraft[]>(initialMembers);
  const { form, options, filteredGattungen, selectOptions, onUpdate, onUpdateCategory } = model;
  const basicsComplete = Boolean(
    form.name.trim() && form.manufacturer.trim() && form.gauge.trim() && form.category?.trim() && form.gattung?.trim()
  );

  const updateMember = (index: number, patch: Partial<VehicleSetMemberDraft>) => {
    setMembers((current) => current.map((member, memberIndex) => (
      memberIndex === index ? { ...member, ...patch } : member
    )));
  };

  const removeMember = (index: number) => {
    setMembers((current) => current.length > 2 ? current.filter((_, memberIndex) => memberIndex !== index) : current);
  };

  const submit: FormEventHandler<HTMLFormElement> = (event) => {
    if (step < 3) {
      event.preventDefault();
      if (step === 1 && !basicsComplete) return;
      setStep((current) => Math.min(3, current + 1));
      return;
    }
    if (kind === "set") {
      event.preventDefault();
      void onSubmitSet(members);
      return;
    }
    onSubmitSingle(event);
  };

  return (
    <form className={`vehicle-modal vehicle-create-wizard${kind === "set" ? " is-set" : ""}`} onSubmit={submit}>
      <header className="modal-head vehicle-create-head">
        <div>
          <h2>{t("vehicles.wizard.title")}</h2>
          <p>{t("vehicles.wizard.subtitle")}</p>
        </div>
        <button
          type="button"
          className="icon-button"
          onClick={onClose}
          aria-label={t("vehicles.close")}
          title={t("vehicles.close")}
        >
          <X size={18} />
        </button>
      </header>

      <ol className="vehicle-wizard-steps" aria-label={t("vehicles.wizard.progress") }>
        {[1, 2, 3].map((item) => (
          <li key={item} className={item === step ? "active" : item < step ? "done" : ""}>
            <span>{item}</span>
            <strong>{t(`vehicles.wizard.step${item}`)}</strong>
          </li>
        ))}
      </ol>

      <div className="modal-body vehicle-wizard-body">
        {step === 1 && (
          <div className="vehicle-wizard-page">
            <div className="vehicle-kind-grid" role="radiogroup" aria-label={t("vehicles.wizard.kind") }>
              <button
                type="button"
                className={kind === "single" ? "vehicle-kind-card selected" : "vehicle-kind-card"}
                onClick={() => setKind("single")}
                role="radio"
                aria-checked={kind === "single"}
              >
                <TrainFront size={24} />
                <span>
                  <strong>{t("vehicles.wizard.single")}</strong>
                  <small>{t("vehicles.wizard.singleHint")}</small>
                </span>
              </button>
              <button
                type="button"
                className={kind === "set" ? "vehicle-kind-card selected" : "vehicle-kind-card"}
                onClick={() => setKind("set")}
                role="radio"
                aria-checked={kind === "set"}
              >
                <Layers3 size={24} />
                <span><strong>{t("vehicles.wizard.set")}</strong><small>{t("vehicles.wizard.setHint")}</small></span>
              </button>
            </div>

            <section className="vehicle-wizard-section vehicle-form">
              <div className="vehicle-wizard-section-head">
                <div><span>01</span><h3>{t("vehicles.wizard.basicData")}</h3></div>
                <small>{t("vehicles.wizard.requiredHint")}</small>
              </div>
              <div className="form-row">
                <label>
                  <RequiredLabel
                    label={kind === "set" ? t("vehicles.wizard.setName") : t("vehicle.field.name")}
                    filled={Boolean(form.name.trim())}
                  />
                  <input
                    value={form.name}
                    onChange={(event) => onUpdate({ name: event.target.value })}
                    required
                    autoFocus
                  />
                </label>
                <label>
                  <RequiredLabel label={t("vehicle.field.manufacturer")} filled={Boolean(form.manufacturer.trim())} />
                  <AppSelect
                    value={form.manufacturer}
                    onChange={(event) => onUpdate({ manufacturer: event.target.value })}
                    required
                  >
                    {selectOptions(options.manufacturers, form.manufacturer, t("vehicles.select.placeholder"))}
                  </AppSelect>
                </label>
              </div>
              <div className="form-row">
                <label>
                  <RequiredLabel label={t("vehicle.field.gauge")} filled={Boolean(form.gauge.trim())} />
                  <AppSelect value={form.gauge} onChange={(event) => onUpdate({ gauge: event.target.value })} required>
                    {selectOptions(options.gauges, form.gauge, t("vehicles.select.placeholder"))}
                  </AppSelect>
                </label>
                <label>
                  {t("vehicle.field.epoch")}
                  <AppSelect value={form.epoch || ""} onChange={(event) => onUpdate({ epoch: event.target.value })}>
                    {selectOptions(options.epochs, form.epoch || "")}
                  </AppSelect>
                </label>
              </div>
              <div className="form-row">
                <label>
                  <RequiredLabel label={t("vehicle.field.category")} filled={Boolean(form.category?.trim())} />
                  <AppSelect
                    value={form.category || ""}
                    onChange={(event) => onUpdateCategory(event.target.value)}
                    required
                  >
                    {selectOptions(options.categories, form.category || "", t("vehicles.select.placeholder"))}
                  </AppSelect>
                </label>
                <label>
                  <RequiredLabel label={t("vehicle.field.gattung")} filled={Boolean(form.gattung?.trim())} />
                  <AppSelect
                    value={form.gattung || ""}
                    onChange={(event) => onUpdate({ gattung: event.target.value })}
                    disabled={!filteredGattungen.length}
                    required
                  >
                    {selectOptions(filteredGattungen, form.gattung || "", t("vehicles.select.placeholder"))}
                  </AppSelect>
                </label>
              </div>
            </section>
          </div>
        )}

        {step === 2 && (
          <div className="vehicle-wizard-page">
            <section className="vehicle-wizard-section vehicle-form">
              <div className="vehicle-wizard-section-head">
                <div><span>02</span><h3>{t("vehicles.wizard.articleData")}</h3></div>
                <small>{t("vehicles.wizard.optional")}</small>
              </div>
              <p className="vehicle-wizard-intro">{t("vehicles.wizard.articleIntro")}</p>
              <div className="vehicle-lookup-actions">
                <button
                  type="button"
                  className="vehicle-lookup-card"
                  onClick={model.onOpenBarcodeSearch}
                  disabled={model.articleSearchLoading}
                >
                  <Barcode size={22} />
                  <span>
                    <strong>{t("vehicles.articleSearch.barcode")}</strong>
                    <small>{t("vehicles.wizard.barcodeHint")}</small>
                  </span>
                </button>
                <button
                  type="button"
                  className="vehicle-lookup-card"
                  onClick={model.onRunArticleSearch}
                  disabled={model.articleSearchLoading || !model.canRunArticleSearch}
                >
                  <PackageSearch size={22} />
                  <span>
                    <strong>{t("vehicles.articleSearch.search")}</strong>
                    <small>{t("vehicles.wizard.webHint")}</small>
                  </span>
                </button>
              </div>
              <div className="form-row three-columns">
                <label>{t("vehicle.field.articleNumber")}<input value={form.articleNumber || ""} onChange={(event) => onUpdate({ articleNumber: event.target.value })} /></label>
                <label>EAN<input value={form.ean || ""} onChange={(event) => onUpdate({ ean: event.target.value })} /></label>
                <label>{t("vehicle.field.productionPeriod")}<input value={form.productionPeriod || ""} onChange={(event) => onUpdate({ productionPeriod: event.target.value })} /></label>
              </div>
              <label>{t("vehicle.field.articleSourceUrl")}<input type="url" value={form.articleSourceUrl || ""} onChange={(event) => onUpdate({ articleSourceUrl: event.target.value })} /></label>
              <div className="vehicle-manual-note">
                <Search size={17} />
                <span>{t("vehicles.wizard.manualHint")}</span>
              </div>
            </section>
          </div>
        )}

        {step === 3 && (
          <div className="vehicle-wizard-page vehicle-wizard-final">
            {kind === "set" && (
              <section className="vehicle-wizard-section vehicle-form">
                <div className="vehicle-wizard-section-head">
                  <div><span>03</span><h3>{t("vehicles.wizard.members")}</h3></div>
                  <button
                    type="button"
                    className="secondary-button"
                    onClick={() => setMembers((current) => [
                      ...current,
                      { inventoryNumber: "", name: "", vehicleNumber: "" }
                    ])}
                  >
                    <Plus size={15} />{t("vehicles.wizard.addMember")}
                  </button>
                </div>
                <div className="vehicle-set-members">
                  {members.map((member, index) => (
                    <div className="vehicle-set-member" key={index}>
                      <span className="vehicle-set-member-index">{index + 1}</span>
                      <label>{t("vehicle.field.name")}<input value={member.name} onChange={(event) => updateMember(index, { name: event.target.value })} placeholder={`${form.name} (${index + 1})`} /></label>
                      <label>{t("vehicle.field.inventoryNumber")}<input value={member.inventoryNumber} onChange={(event) => updateMember(index, { inventoryNumber: event.target.value })} placeholder={t("vehicles.inventoryNumberAuto")} /></label>
                      <label>{t("vehicle.field.vehicleNumber")}<input value={member.vehicleNumber} onChange={(event) => updateMember(index, { vehicleNumber: event.target.value })} /></label>
                      <button type="button" className="icon-button" onClick={() => removeMember(index)} disabled={members.length <= 2} aria-label={t("vehicles.wizard.removeMember")}><Trash2 size={16} /></button>
                    </div>
                  ))}
                </div>
              </section>
            )}
            <section className="vehicle-wizard-section">
              <div className="vehicle-wizard-section-head">
                <div><span>{kind === "set" ? "04" : "03"}</span><h3>{t("vehicles.wizard.remainingData")}</h3></div>
                <small>{t("vehicles.wizard.reviewHint")}</small>
              </div>
              <VehicleModelTab {...model} hideInventoryNumber={kind === "set"} />
            </section>
          </div>
        )}
      </div>

      <footer className="modal-actions vehicle-wizard-actions">
        {message && <p className="form-message">{message}</p>}
        <button type="button" className="secondary-button" onClick={step === 1 ? onClose : () => setStep((current) => current - 1)}>
          {step === 1 ? t("vehicles.cancel") : <><ChevronLeft size={16} />{t("vehicles.wizard.back")}</>}
        </button>
        <button type="submit" className="primary-button" disabled={saving || (step === 1 && !basicsComplete)}>
          {step < 3 ? <>{t("vehicles.wizard.next")}<ChevronRight size={16} /></> : saving ? t("vehicles.saving") : kind === "set" ? t("vehicles.wizard.createSet") : t("vehicles.createAndContinue")}
        </button>
      </footer>
    </form>
  );
}

export function vehicleSetInputFromForm(form: CreateVehicleRequest) {
  return {
    name: form.name,
    manufacturer: form.manufacturer,
    articleNumber: form.articleNumber,
    articleSourceUrl: form.articleSourceUrl,
    gauge: form.gauge,
    epoch: form.epoch,
    railwayCompany: form.railwayCompany,
    category: form.category,
    gattung: form.gattung,
    description: form.description,
    ean: form.ean,
    productionPeriod: form.productionPeriod,
    listPrice: form.listPrice,
    acquisitionType: form.acquisitionType,
    acquiredFrom: form.acquiredFrom,
    purchasePrice: form.purchasePrice,
    purchaseDate: form.purchaseDate,
    storageLocation: form.storageLocation,
    storageDetails: form.storageDetails,
    condition: form.condition,
    conditionDetails: form.conditionDetails,
    packaging: form.packaging
  };
}
