import { useEffect, useRef, useState } from "react";
import type { FormEvent } from "react";
import { CalendarClock } from "lucide-react";

import {
  api,
  type AccessoryAsset,
  type AccessoryArticle,
  type AccessoryReservation,
  type AccessoryTechnicalPlacement,
  type Layout,
  type LayoutUnit,
  type StorageLocation,
  type Vehicle
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { AccessoryConfirmDialog, type AccessoryPendingAction } from "./AccessoryConfirmDialog";
import {
  AccessoryTargetFields,
  accessoryTargetInput,
  accessoryTargetLabel,
  resolveAccessoryTargetSelection,
  type AccessoryTargetSelection
} from "./AccessoryTargetFields";
import { activeStorageLocations, storageLocationPath } from "../../shared/storageLocations";
import { AccessoryTechnicalFields, emptyTechnicalPlacement } from "./AccessoryTechnicalFields";

export function AccessoryReservationsPanel({ article, reservations, assets, locations, vehicles, layouts, units,
  canReserve, onChanged, onDirtyChange }: {
  article: AccessoryArticle;
  reservations: AccessoryReservation[];
  assets: AccessoryAsset[];
  locations: StorageLocation[];
  vehicles: Vehicle[];
  layouts: Layout[];
  units: LayoutUnit[];
  canReserve: boolean;
  onChanged: () => Promise<void>;
  onDirtyChange: (dirty: boolean) => void;
}) {
  const [locationID, setLocationID] = useState("");
  const [assetID, setAssetID] = useState("");
  const [allocationMode, setAllocationMode] = useState<"quantity" | "asset">("quantity");
  const [quantity, setQuantity] = useState("1");
  const [note, setNote] = useState("");
  const [target, setTarget] = useState<AccessoryTargetSelection>({ kind: "layout", id: "" });
  const [technical, setTechnical] = useState<AccessoryTechnicalPlacement>(emptyTechnicalPlacement);
  const [action, setAction] = useState<AccessoryPendingAction | null>(null);
  const headingRef = useRef<HTMLHeadingElement | null>(null);
  const { t } = useI18n();
  const activeLocations = activeStorageLocations(locations);
  const effectiveLocationID = activeLocations.some((location) => location.id === locationID)
    ? locationID : activeLocations[0]?.id || "";
  const availableAssets = assets.filter((asset) => asset.lifecycle === "stored");
  const effectiveAssetID = availableAssets.some((asset) => asset.id === assetID)
    ? assetID : availableAssets[0]?.id || "";
  const resolvedTarget = resolveAccessoryTargetSelection(target, vehicles, layouts, units);
  const targetInput = accessoryTargetInput(resolvedTarget);
  const isHybrid = article.inventoryStrategy === "quantity_later_individual";
  const isIndividual = article.inventoryStrategy === "individual" || (isHybrid && allocationMode === "asset");
  const canSubmit = Boolean(effectiveLocationID && targetInput && (!isIndividual || effectiveAssetID));
  const dirty = Boolean(locationID || assetID || note || target.id || Object.values(technical).some(Boolean)) ||
    quantity !== "1" || allocationMode !== "quantity";

  useEffect(() => onDirtyChange(dirty), [dirty, onDirtyChange]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    if (!targetInput || !canSubmit) return;
    setAction({
      title: t("accessories.reservations.confirmTitle"),
      body: t("accessories.reservations.confirmBody", { product: article.name }),
      run: async () => {
        await api.createAccessoryReservation({
          ...targetInput,
          ...technical,
          productId: article.id,
          ...(isIndividual ? { assetId: effectiveAssetID } : {}),
          locationId: effectiveLocationID,
          quantity: isIndividual ? 1 : Number(quantity),
          note: note || undefined
        });
        setLocationID(""); setAssetID(""); setAllocationMode("quantity"); setQuantity("1"); setNote("");
        setTarget({ kind: "layout", id: "" });
        setTechnical(emptyTechnicalPlacement());
      },
      afterSuccess: onChanged
    });
  };

  const cancel = (reservation: AccessoryReservation) => setAction({
    title: t("accessories.reservations.cancelTitle"),
    body: t("accessories.reservations.cancelBody"),
    successReturnFocusRef: headingRef,
    run: async () => { await api.cancelAccessoryReservation(reservation.id); },
    afterSuccess: onChanged
  });

  return <>
    <section className="panel accessory-stock-panel">
      <div className="panel-head"><CalendarClock size={17} aria-hidden="true" />
        <h2 ref={headingRef} tabIndex={-1}>{t("accessories.reservations.title")}</h2>
      </div>
      <div className="accessory-work-grid">
        <div className="table-wrap"><table><thead><tr>
          <th>{t("accessories.field.target")}</th><th>{t("accessories.field.quantity")}</th>
          <th>{t("accessories.field.status")}</th><th><span className="sr-only">{t("accessories.field.actions")}</span></th>
        </tr></thead><tbody>{reservations.map((reservation) => <tr key={reservation.id}>
          <td>{accessoryTargetLabel(reservation, vehicles, layouts, units)}</td>
          <td>{reservation.quantity}</td><td>{t(`accessories.reservationStatus.${reservation.status}`)}</td>
          <td>{canReserve && reservation.status === "active" ? <button type="button" className="text-button"
            onClick={() => cancel(reservation)}>{t("accessories.reservations.cancel")}</button> : null}</td>
        </tr>)}</tbody></table></div>
        {canReserve ? <form className="accessory-form" onSubmit={submit}>
          <h3>{t("accessories.reservations.create")}</h3>
          <AccessoryTargetFields target={resolvedTarget} vehicles={vehicles} layouts={layouts} units={units}
            onChange={setTarget} />
          <AccessoryTechnicalFields value={technical} onChange={setTechnical} />
          {isHybrid ? <label>{t("accessories.field.allocationSource")}<AppSelect value={allocationMode}
            aria-label={t("accessories.field.allocationSource")}
            onChange={(event) => setAllocationMode(event.target.value as "quantity" | "asset")}>
            <option value="quantity">{t("accessories.allocationSource.quantity")}</option>
            <option value="asset" disabled={availableAssets.length === 0}>
              {t("accessories.allocationSource.asset")}</option>
          </AppSelect></label> : null}
          <label>{t("accessories.field.location")}<AppSelect value={effectiveLocationID}
            onChange={(event) => setLocationID(event.target.value)}>
            {activeLocations.map((location) => <option key={location.id} value={location.id}>
              {storageLocationPath(location, locations)}</option>)}
          </AppSelect></label>
          {isIndividual ? <label>{t("accessories.field.asset")}<AppSelect value={effectiveAssetID}
            onChange={(event) => setAssetID(event.target.value)}>
            {availableAssets.map((asset) => <option key={asset.id} value={asset.id}>
              {asset.inventoryNumber || asset.serialNumber || asset.id}</option>)}
          </AppSelect></label> : <AppNumberInput label={t("accessories.field.quantity")} min="1" required
            value={quantity} onValueChange={setQuantity} />}
          <label>{t("accessories.field.notes")}<textarea value={note}
            onChange={(event) => setNote(event.target.value)} /></label>
          <button type="submit" className="primary-button" disabled={!canSubmit}>
            {t("accessories.reservations.save")}
          </button>
        </form> : null}
      </div>
    </section>
    {action ? <AccessoryConfirmDialog action={action} onClose={() => setAction(null)} /> : null}
  </>;
}
