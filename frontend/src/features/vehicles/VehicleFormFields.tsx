import { QrCode } from "lucide-react";
import { CreateVehicleRequest } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import {
  acquiredFromOptions,
  acquisitionOptions,
  adapterOptions,
  couplingOptions,
  packagingOptions,
  powerPickupOptions,
  storageLocationOptions,
  vehicleConditionOptions,
  wheelsetOptions
} from "./vehicleOptions";

function renderStaticOptions(items: string[], emptyLabel = "Bitte wählen") {
  return (
    <>
      <option value="">{emptyLabel}</option>
      {items.map((item) => (
        <option key={item} value={item}>
          {item}
        </option>
      ))}
    </>
  );
}

export function RequiredLabel({ label, filled, showError = false }: { label: string; filled: boolean; showError?: boolean }) {
  const stateClass = filled ? "filled" : showError ? "missing" : "pending";
  return (
    <span className={`required-label ${stateClass}`}>
      {label}
      <span className={`required-dot ${stateClass}`} aria-hidden="true" />
    </span>
  );
}

export function VehicleDetailsFields({
  form,
  readonly,
  onOpenQr,
  canOpenQr,
  update,
  updateCouplingFront,
  updateCouplingSame
}: {
  form: CreateVehicleRequest;
  readonly: boolean;
  onOpenQr: () => void;
  canOpenQr: boolean;
  update: (patch: Partial<CreateVehicleRequest>) => void;
  updateCouplingFront: (couplingFront: string) => void;
  updateCouplingSame: (couplingSame: boolean) => void;
}) {
  const { t } = useI18n();
  const qrButtonTitle = canOpenQr ? t("vehicles.detail.qrShow") : t("vehicles.qr.missingInput");
  return (
    <>
      <div className="form-row four-columns">
        <label>
          {t("vehicle.field.lengthMm")}
          <input value={form.lengthMm || ""} onChange={(event) => update({ lengthMm: event.target.value })} disabled={readonly} inputMode="decimal" />
        </label>
        <label>
          {t("vehicle.field.weightG")}
          <input value={form.weightG || ""} onChange={(event) => update({ weightG: event.target.value })} disabled={readonly} inputMode="decimal" />
        </label>
        <label>
          {t("vehicle.field.color")}
          <input value={form.color || ""} onChange={(event) => update({ color: event.target.value })} disabled={readonly} />
        </label>
        <label>
          {t("vehicle.field.lettering")}
          <input value={form.lettering || ""} onChange={(event) => update({ lettering: event.target.value })} disabled={readonly} />
        </label>
      </div>

      <div className="form-row three-columns">
        <label>
          {t("vehicle.field.load")}
          <input value={form.load || ""} onChange={(event) => update({ load: event.target.value })} disabled={readonly} />
        </label>
        <label>
          {t("vehicle.field.interior")}
          <input value={form.interior || ""} onChange={(event) => update({ interior: event.target.value })} disabled={readonly} />
        </label>
        <label>
          {t("vehicle.field.axles")}
          <input value={form.axles || ""} onChange={(event) => update({ axles: event.target.value })} disabled={readonly} />
        </label>
      </div>

      <div className="form-row four-columns">
        <label>
          {t("vehicle.field.axleCount")}
          <input value={form.axleCount || ""} onChange={(event) => update({ axleCount: event.target.value })} disabled={readonly} inputMode="numeric" />
        </label>
        <label>
          {t("vehicle.field.tractionTireCount")}
          <input value={form.tractionTireCount || ""} onChange={(event) => update({ tractionTireCount: event.target.value })} disabled={readonly} inputMode="numeric" />
        </label>
        <label>
          {t("vehicle.field.wheelset")}
          <AppSelect value={form.wheelset || ""} onChange={(event) => update({ wheelset: event.target.value })} disabled={readonly}>
            {renderStaticOptions(wheelsetOptions, t("vehicles.select.placeholder"))}
          </AppSelect>
        </label>
        <label>
          {t("vehicle.field.powerPickup")}
          <AppSelect value={form.powerPickup || ""} onChange={(event) => update({ powerPickup: event.target.value })} disabled={readonly}>
            {renderStaticOptions(powerPickupOptions, t("vehicles.select.placeholder"))}
          </AppSelect>
        </label>
      </div>

      <div className="form-row details-coupling-row">
        <label>
          {t("vehicle.field.adapter")}
          <AppSelect value={form.adapter || ""} onChange={(event) => update({ adapter: event.target.value })} disabled={readonly}>
            {renderStaticOptions(adapterOptions, t("vehicles.select.placeholder"))}
          </AppSelect>
        </label>
        <label className="coupling-same-field">
          <span>{t("vehicles.detail.couplingSame")}</span>
          <span className="switch-field">
            <input type="checkbox" checked={Boolean(form.couplingSame)} onChange={(event) => updateCouplingSame(event.target.checked)} disabled={readonly} />
            <span />
          </span>
        </label>
        <label>
          {t("vehicle.field.couplingFront")}
          <AppSelect value={form.couplingFront || ""} onChange={(event) => updateCouplingFront(event.target.value)} disabled={readonly}>
            {renderStaticOptions(couplingOptions, t("vehicles.select.placeholder"))}
          </AppSelect>
        </label>
        <label>
          {t("vehicle.field.couplingRear")}
          <AppSelect value={form.couplingSame ? form.couplingFront || "" : form.couplingRear || ""} onChange={(event) => update({ couplingRear: event.target.value })} disabled={readonly || Boolean(form.couplingSame)}>
            {renderStaticOptions(couplingOptions, t("vehicles.select.placeholder"))}
          </AppSelect>
        </label>
      </div>

      <div className="form-row switch-description-row">
        <label>
          {t("vehicle.field.headlightsDescription")}
          <span className="inline-switch-input">
            <span className="switch-field" aria-label={t("vehicle.field.headlightsEnabled")}>
              <input type="checkbox" checked={Boolean(form.headlightsEnabled)} onChange={(event) => update({ headlightsEnabled: event.target.checked })} disabled={readonly} />
              <span />
            </span>
            <input value={form.headlightsDescription || ""} onChange={(event) => update({ headlightsDescription: event.target.value })} disabled={readonly || !form.headlightsEnabled} />
          </span>
        </label>
        <label>
          {t("vehicle.field.driveDescription")}
          <span className="inline-switch-input">
            <span className="switch-field" aria-label={t("vehicle.field.driveEnabled")}>
              <input type="checkbox" checked={Boolean(form.driveEnabled)} onChange={(event) => update({ driveEnabled: event.target.checked })} disabled={readonly} />
              <span />
            </span>
            <input value={form.driveDescription || ""} onChange={(event) => update({ driveDescription: event.target.value })} disabled={readonly || !form.driveEnabled} />
          </span>
        </label>
      </div>

      <div className="form-row switch-description-row">
        <label>
          {t("vehicle.field.lightingDescription")}
          <span className="inline-switch-input">
            <span className="switch-field" aria-label={t("vehicle.field.lightingEnabled")}>
              <input type="checkbox" checked={Boolean(form.lightingEnabled)} onChange={(event) => update({ lightingEnabled: event.target.checked })} disabled={readonly} />
              <span />
            </span>
            <input value={form.lightingDescription || ""} onChange={(event) => update({ lightingDescription: event.target.value })} disabled={readonly || !form.lightingEnabled} />
          </span>
        </label>
        <label>
          {t("vehicle.field.soundGeneratorDescription")}
          <span className="inline-switch-input">
            <span className="switch-field" aria-label={t("vehicle.field.soundGeneratorEnabled")}>
              <input type="checkbox" checked={Boolean(form.soundGeneratorEnabled)} onChange={(event) => update({ soundGeneratorEnabled: event.target.checked })} disabled={readonly} />
              <span />
            </span>
            <input value={form.soundGeneratorDescription || ""} onChange={(event) => update({ soundGeneratorDescription: event.target.value })} disabled={readonly || !form.soundGeneratorEnabled} />
          </span>
        </label>
      </div>

      <div className="form-row switch-description-row">
        <label>
          {t("vehicle.field.smokeGeneratorDescription")}
          <span className="inline-switch-input">
            <span className="switch-field" aria-label={t("vehicle.field.smokeGeneratorEnabled")}>
              <input type="checkbox" checked={Boolean(form.smokeGeneratorEnabled)} onChange={(event) => update({ smokeGeneratorEnabled: event.target.checked })} disabled={readonly} />
              <span />
            </span>
            <input value={form.smokeGeneratorDescription || ""} onChange={(event) => update({ smokeGeneratorDescription: event.target.value })} disabled={readonly || !form.smokeGeneratorEnabled} />
          </span>
        </label>
        <label className="qr-switch-field">
          <span>{t("vehicles.detail.qrCreate")}</span>
          <span className="qr-card-actions">
            <span className="switch-field">
              <input type="checkbox" checked={Boolean(form.qrCodeEnabled)} onChange={(event) => update({ qrCodeEnabled: event.target.checked })} disabled={readonly} />
              <span />
            </span>
            <button type="button" className="icon-button" onClick={onOpenQr} aria-label={qrButtonTitle} title={qrButtonTitle} disabled={!form.qrCodeEnabled || !canOpenQr}>
              <QrCode size={16} />
            </button>
          </span>
        </label>
      </div>
    </>
  );
}

export function VehicleOwnershipFields({
  form,
  readonly,
  update
}: {
  form: CreateVehicleRequest;
  readonly: boolean;
  update: (patch: Partial<CreateVehicleRequest>) => void;
}) {
  const { t } = useI18n();
  return (
    <>
      <div className="form-row four-columns">
        <label>
          {t("vehicle.field.acquisitionType")}
          <AppSelect value={form.acquisitionType || ""} onChange={(event) => update({ acquisitionType: event.target.value })} disabled={readonly}>
            {renderStaticOptions(acquisitionOptions, t("vehicles.select.placeholder"))}
          </AppSelect>
        </label>
        <label>
          {t("vehicle.field.acquiredFrom")}
          <AppSelect value={form.acquiredFrom || ""} onChange={(event) => update({ acquiredFrom: event.target.value })} disabled={readonly}>
            {renderStaticOptions(acquiredFromOptions, t("vehicles.select.placeholder"))}
          </AppSelect>
        </label>
        <label>
          {t("vehicle.field.purchasePrice")}
          <input value={form.purchasePrice || ""} onChange={(event) => update({ purchasePrice: event.target.value })} disabled={readonly} inputMode="decimal" />
        </label>
        <label>
          {t("vehicle.field.purchaseDate")}
          <input type="date" value={form.purchaseDate || ""} onChange={(event) => update({ purchaseDate: event.target.value })} disabled={readonly} />
        </label>
      </div>

      <div className="form-row">
        <label>
          {t("vehicle.field.storageLocation")}
          <AppSelect value={form.storageLocation || ""} onChange={(event) => update({ storageLocation: event.target.value })} disabled={readonly}>
            {renderStaticOptions(storageLocationOptions, t("vehicles.select.placeholder"))}
          </AppSelect>
        </label>
        <label>
          {t("vehicle.field.storageDetails")}
          <input value={form.storageDetails || ""} onChange={(event) => update({ storageDetails: event.target.value })} disabled={readonly} />
        </label>
      </div>

      <div className="form-row three-columns">
        <label>
          {t("vehicle.field.condition")}
          <AppSelect value={form.condition || ""} onChange={(event) => update({ condition: event.target.value })} disabled={readonly}>
            {renderStaticOptions(vehicleConditionOptions, t("vehicles.select.placeholder"))}
          </AppSelect>
        </label>
        <label>
          {t("vehicle.field.conditionDetails")}
          <input value={form.conditionDetails || ""} onChange={(event) => update({ conditionDetails: event.target.value })} disabled={readonly} />
        </label>
        <label>
          {t("vehicle.field.packaging")}
          <AppSelect value={form.packaging || ""} onChange={(event) => update({ packaging: event.target.value })} disabled={readonly}>
            {renderStaticOptions(packagingOptions, t("vehicles.select.placeholder"))}
          </AppSelect>
        </label>
      </div>

      <label>
        {t("vehicle.field.additionalInfo")}
        <textarea value={form.additionalInfo || ""} onChange={(event) => update({ additionalInfo: event.target.value })} disabled={readonly} rows={5} />
      </label>
    </>
  );
}
