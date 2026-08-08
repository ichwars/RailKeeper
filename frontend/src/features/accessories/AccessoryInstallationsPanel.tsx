import { useState } from "react";
import type { FormEvent } from "react";
import { Wrench } from "lucide-react";

import {
  api,
  type AccessoryAsset,
  type AccessoryCondition,
  type AccessoryInstallation,
  type AccessoryProduct,
  type AccessoryRemovalDisposition,
  type AccessoryReservation,
  type Layout,
  type LayoutUnit,
  type StorageLocation,
  type Vehicle
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AccessoryConfirmDialog, type AccessoryPendingAction } from "./AccessoryConfirmDialog";
import {
  AccessoryTargetFields,
  accessoryTargetInput,
  accessoryTargetLabel,
  resolveAccessoryTargetSelection,
  type AccessoryTargetSelection
} from "./AccessoryTargetFields";
import { activeStorageLocations, storageLocationPath } from "../../shared/storageLocations";

const conditions: AccessoryCondition[] = ["ready", "maintenance_due", "defective", "unknown"];

export function AccessoryInstallationsPanel({ product, reservations, installations, assets, locations, vehicles,
  layouts, units, canInstall, onChanged }: {
  product: AccessoryProduct | null;
  reservations: AccessoryReservation[];
  installations: AccessoryInstallation[];
  assets: AccessoryAsset[];
  locations: StorageLocation[];
  vehicles: Vehicle[];
  layouts: Layout[];
  units: LayoutUnit[];
  canInstall: boolean;
  onChanged: () => Promise<void>;
}) {
  const [reservationID, setReservationID] = useState("");
  const [locationID, setLocationID] = useState("");
  const [assetID, setAssetID] = useState("");
  const [quantity, setQuantity] = useState("1");
  const [condition, setCondition] = useState<AccessoryCondition>("ready");
  const [conditionDrafts, setConditionDrafts] = useState<Record<string, AccessoryCondition>>({});
  const [notes, setNotes] = useState("");
  const [target, setTarget] = useState<AccessoryTargetSelection>({ kind: "layout", id: "" });
  const [removalID, setRemovalID] = useState("");
  const [disposition, setDisposition] = useState<AccessoryRemovalDisposition>("stored");
  const [removalLocationID, setRemovalLocationID] = useState("");
  const [removalNotes, setRemovalNotes] = useState("");
  const [action, setAction] = useState<AccessoryPendingAction | null>(null);
  const { t } = useI18n();
  if (!product) return <section className="panel"><p>{t("accessories.selection.empty")}</p></section>;

  const activeReservations = reservations.filter((reservation) => reservation.status === "active");
  const activeLocations = activeStorageLocations(locations);
  const selectedReservation = activeReservations.find((reservation) => reservation.id === reservationID);
  const resolvedTarget = resolveAccessoryTargetSelection(target, vehicles, layouts, units);
  const targetInput = selectedReservation || accessoryTargetInput(resolvedTarget);
  const sourceLocations = selectedReservation
    ? locations.filter((location) => location.id === selectedReservation.locationId) : activeLocations;
  const effectiveLocationID = selectedReservation?.locationId ||
    (activeLocations.some((location) => location.id === locationID) ? locationID : activeLocations[0]?.id || "");
  const sourceAssets = selectedReservation?.assetId
    ? assets.filter((asset) => asset.id === selectedReservation.assetId)
    : assets.filter((asset) => asset.lifecycle === "stored");
  const effectiveAssetID = selectedReservation?.assetId ||
    (sourceAssets.some((asset) => asset.id === assetID) ? assetID : sourceAssets[0]?.id || "");
  const effectiveQuantity = selectedReservation?.quantity || Number(quantity);
  const isIndividual = product.trackingMode === "individual";
  const canSubmit = Boolean(targetInput && effectiveLocationID && (!isIndividual || effectiveAssetID));
  const effectiveRemovalLocationID = activeLocations.some((location) => location.id === removalLocationID)
    ? removalLocationID : activeLocations[0]?.id || "";

  const submitInstallation = (event: FormEvent) => {
    event.preventDefault();
    if (!targetInput || !canSubmit) return;
    const allocationTarget = accessoryTargetInput({
      kind: targetInput.vehicleId ? "vehicle" : targetInput.layoutId ? "layout" : "layoutUnit",
      id: targetInput.vehicleId || targetInput.layoutId || targetInput.layoutUnitId || ""
    });
    if (!allocationTarget) return;
    setAction({
      title: t("accessories.installations.confirmTitle"),
      body: t("accessories.installations.confirmBody", { product: product.name }),
      run: async () => {
        await api.createAccessoryInstallation({
          ...allocationTarget,
          reservationId: selectedReservation?.id,
          productId: product.id,
          assetId: isIndividual ? effectiveAssetID : undefined,
          sourceLocationId: effectiveLocationID,
          quantity: isIndividual ? 1 : effectiveQuantity,
          condition,
          notes: notes || undefined
        });
        setReservationID(""); setNotes("");
        await onChanged();
      }
    });
  };

  const updateCondition = (installation: AccessoryInstallation) => {
    const nextCondition = conditionDrafts[installation.id] || installation.condition;
    setAction({
      title: t("accessories.installations.conditionTitle"),
      body: t("accessories.installations.conditionBody"),
      run: async () => {
        await api.updateAccessoryInstallationCondition(installation.id, { condition: nextCondition });
        await onChanged();
      }
    });
  };

  const submitRemoval = (event: FormEvent) => {
    event.preventDefault();
    if (!removalID || (disposition === "stored" && !effectiveRemovalLocationID)) return;
    setAction({
      title: t("accessories.installations.removeTitle"),
      body: t("accessories.installations.removeBody"),
      run: async () => {
        const input = disposition === "stored"
          ? { disposition, storageLocationId: effectiveRemovalLocationID, notes: removalNotes || undefined }
          : { disposition, notes: removalNotes || undefined };
        await api.removeAccessoryInstallation(removalID, input);
        setRemovalID(""); setRemovalNotes(""); setDisposition("stored");
        await onChanged();
      }
    });
  };

  return <>
    <section className="panel accessory-stock-panel">
      <div className="panel-head"><Wrench size={17} aria-hidden="true" />
        <h2>{t("accessories.installations.title")}</h2>
      </div>
      <div className="accessory-work-grid">
        <div className="table-wrap"><table><thead><tr>
          <th>{t("accessories.field.target")}</th><th>{t("accessories.field.quantity")}</th>
          <th>{t("accessories.field.condition")}</th><th>{t("accessories.field.installedAt")}</th>
          <th><span className="sr-only">{t("accessories.field.actions")}</span></th>
        </tr></thead><tbody>{installations.map((installation) => <tr key={installation.id}>
          <td>{accessoryTargetLabel(installation, vehicles, layouts, units)}</td><td>{installation.quantity}</td>
          <td>{installation.removedAt || !canInstall ? t(`accessories.condition.${installation.condition}`)
            : <AppSelect value={conditionDrafts[installation.id] || installation.condition}
              aria-label={t("accessories.installations.conditionFor", { target: accessoryTargetLabel(
                installation, vehicles, layouts, units) })}
              onChange={(event) => setConditionDrafts((current) => ({ ...current,
                [installation.id]: event.target.value as AccessoryCondition }))}>
              {conditions.map((item) => <option key={item} value={item}>{t(`accessories.condition.${item}`)}</option>)}
            </AppSelect>}</td>
          <td>{new Date(installation.installedAt).toLocaleDateString()}
            {installation.removedAt ? ` · ${t("accessories.installations.removed")}` : ""}</td>
          <td>{canInstall && !installation.removedAt ? <div className="accessory-row-actions">
            <button type="button" className="text-button" onClick={() => updateCondition(installation)}>
              {t("accessories.installations.updateCondition")}</button>
            <button type="button" className="text-button" onClick={() => setRemovalID(installation.id)}>
              {t("accessories.installations.remove")}</button>
          </div> : null}</td>
        </tr>)}</tbody></table></div>
        {canInstall ? <div className="accessory-side-forms">
          <form className="accessory-form" onSubmit={submitInstallation}>
            <h3>{t("accessories.installations.create")}</h3>
            <label>{t("accessories.field.reservation")}<AppSelect value={reservationID}
              onChange={(event) => setReservationID(event.target.value)}>
              <option value="">{t("accessories.installations.withoutReservation")}</option>
              {activeReservations.map((reservation) => <option key={reservation.id} value={reservation.id}>
                {accessoryTargetLabel(reservation, vehicles, layouts, units)}</option>)}
            </AppSelect></label>
            {selectedReservation ? <p className="accessory-target-summary">{accessoryTargetLabel(
              selectedReservation, vehicles, layouts, units)}</p>
              : <AccessoryTargetFields target={resolvedTarget} vehicles={vehicles} layouts={layouts} units={units}
                onChange={setTarget} />}
            <label>{t("accessories.field.location")}<AppSelect value={effectiveLocationID}
              disabled={Boolean(selectedReservation)} onChange={(event) => setLocationID(event.target.value)}>
              {sourceLocations.map((location) => <option key={location.id} value={location.id}>
                {storageLocationPath(location, locations)}</option>)}
            </AppSelect></label>
            {isIndividual ? <label>{t("accessories.field.asset")}<AppSelect value={effectiveAssetID}
              disabled={Boolean(selectedReservation?.assetId)} onChange={(event) => setAssetID(event.target.value)}>
              {sourceAssets.map((asset) => <option key={asset.id} value={asset.id}>
                {asset.inventoryNumber || asset.serialNumber || asset.id}</option>)}
            </AppSelect></label> : <label>{t("accessories.field.quantity")}<input type="number" min="1"
              required disabled={Boolean(selectedReservation)} value={effectiveQuantity}
              onChange={(event) => setQuantity(event.target.value)} /></label>}
            <label>{t("accessories.field.condition")}<AppSelect value={condition}
              onChange={(event) => setCondition(event.target.value as AccessoryCondition)}>
              {conditions.map((item) => <option key={item} value={item}>{t(`accessories.condition.${item}`)}</option>)}
            </AppSelect></label>
            <label>{t("accessories.field.notes")}<textarea value={notes}
              onChange={(event) => setNotes(event.target.value)} /></label>
            <button type="submit" className="primary-button" disabled={!canSubmit}>
              {t("accessories.installations.save")}</button>
          </form>
          {removalID ? <form className="accessory-form accessory-removal-form" onSubmit={submitRemoval}>
            <h3>{t("accessories.installations.remove")}</h3>
            <label>{t("accessories.field.disposition")}<AppSelect value={disposition}
              onChange={(event) => setDisposition(event.target.value as AccessoryRemovalDisposition)}>
              {(["stored", "maintenance", "defective", "retired"] as const).map((item) =>
                <option key={item} value={item}>{t(`accessories.disposition.${item}`)}</option>)}
            </AppSelect></label>
            {disposition === "stored" ? <label>{t("accessories.field.location")}
              <AppSelect value={effectiveRemovalLocationID}
                onChange={(event) => setRemovalLocationID(event.target.value)}>
                {activeLocations.map((location) => <option key={location.id} value={location.id}>
                  {storageLocationPath(location, locations)}</option>)}
              </AppSelect></label> : null}
            <label>{t("accessories.field.notes")}<textarea value={removalNotes}
              onChange={(event) => setRemovalNotes(event.target.value)} /></label>
            <div className="accessory-form-actions">
              <button type="button" className="secondary-button" onClick={() => setRemovalID("")}>
                {t("common.cancel")}</button>
              <button type="submit" className="primary-button">{t("accessories.installations.remove")}</button>
            </div>
          </form> : null}
        </div> : null}
      </div>
    </section>
    {action ? <AccessoryConfirmDialog action={action} onClose={() => setAction(null)} /> : null}
  </>;
}
