import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { PackageCheck, X } from "lucide-react";

import {
  ApiError,
  api,
  type AccessoryAsset,
  type AccessoryStockSummary,
  type PlanTrackObject,
  type TrackMaterialStatus,
  type TrackPlanReservationBatch
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";

export function TrackPlanReservationDialog({ revisionId, object, material, onClose, onReserved }: {
  revisionId: string;
  object: PlanTrackObject;
  material: TrackMaterialStatus;
  onClose: () => void;
  onReserved: (batch: TrackPlanReservationBatch) => void;
}) {
  const { t } = useI18n();
  const genericError = t("layouts.error.generic");
  const [productID, setProductID] = useState(material.productIds[0] ?? "");
  const [stock, setStock] = useState<AccessoryStockSummary | null>(null);
  const [assets, setAssets] = useState<AccessoryAsset[]>([]);
  const [locationNames, setLocationNames] = useState<Record<string, string>>({});
  const [locationID, setLocationID] = useState("");
  const [assetID, setAssetID] = useState("");
  const [confirmed, setConfirmed] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    let active = true;
    if (!productID) { setLoading(false); return; }
    setLoading(true); setMessage(""); setLocationID(""); setAssetID("");
    void Promise.all([api.accessoryStock(productID), api.accessoryAssets(productID), api.storageLocations()])
      .then(([nextStock, nextAssets, locations]) => {
        if (!active) return;
        setStock(nextStock);
        setAssets(nextAssets.filter((asset) => asset.lifecycle === "stored" && asset.storageLocationId));
        setLocationNames(Object.fromEntries(locations.map((location) => [location.id, location.name])));
        const firstLocation = nextStock.locations.find((location) => location.quantity > 0)?.locationId ?? "";
        const firstAsset = nextAssets.find((asset) => asset.lifecycle === "stored" && asset.storageLocationId);
        if (nextStock.trackingMode === "individual" && firstAsset) {
          setAssetID(firstAsset.id);
          setLocationID(firstAsset.storageLocationId ?? "");
        } else {
          setLocationID(firstLocation);
        }
      })
      .catch((reason: unknown) => active && setMessage(reason instanceof Error ? reason.message : genericError))
      .finally(() => active && setLoading(false));
    return () => { active = false; };
  }, [productID, genericError]);

  const productLabels = useMemo(() => Object.fromEntries(material.productIds.map((id, index) => [
    id, material.inventoryNumbers[index] || `${material.manufacturer} ${material.articleNumber}`
  ])), [material]);
  const individual = stock?.trackingMode === "individual";
  const canReserve = confirmed && Boolean(productID && locationID) && (!individual || Boolean(assetID));

  const reserve = async () => {
    if (!canReserve) return;
    setSaving(true); setMessage("");
    try {
      const batch = await api.reserveTrackPlanMaterials(revisionId, {
        confirmed: true,
        items: [{
          trackObjectId: object.id, productId: productID, locationId: locationID,
          ...(assetID ? { assetId: assetID } : {}), expectedObjectVersion: object.version
        }]
      });
      onReserved(batch);
    } catch (reason) {
      setMessage(reason instanceof ApiError && reason.status === 409
        ? t("layouts.trackReservation.conflict")
        : reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally { setSaving(false); }
  };

  return createPortal(<div className="modal-layer track-reservation-layer" role="dialog" aria-modal="true"
    aria-label={t("layouts.trackReservation.title")}>
    <section className="panel track-reservation-dialog">
      <header><div><PackageCheck size={18} /><div><h2>{t("layouts.trackReservation.title")}</h2>
        <p>{material.manufacturer} {material.articleNumber} · {material.name}</p></div></div>
        <button type="button" className="icon-button" onClick={onClose} disabled={saving}
          aria-label={t("common.close")}><X size={17} /></button></header>
      <div className="track-reservation-object">
        <span>{t("layouts.trackReservation.object")}</span>
        <strong>{object.positionXMm.toFixed(1)} / {object.positionYMm.toFixed(1)} mm · {object.rotationDegrees}°</strong>
      </div>
      {material.productIds.length === 0 ? <p className="form-message">
        {t("layouts.trackReservation.noProduct")}</p> : <div className="form-grid two-columns">
        <label>{t("layouts.trackReservation.product")}<AppSelect value={productID}
          aria-label={t("layouts.trackReservation.product")} disabled={saving}
          onChange={(event) => setProductID(event.target.value)}>
          {material.productIds.map((id) => <option key={id} value={id}>{productLabels[id]}</option>)}
        </AppSelect></label>
        {individual ? <label>{t("layouts.trackReservation.item")}<AppSelect value={assetID}
          aria-label={t("layouts.trackReservation.item")} disabled={saving || loading}
          onChange={(event) => {
            setAssetID(event.target.value);
            setLocationID(assets.find((asset) => asset.id === event.target.value)?.storageLocationId ?? "");
          }}><option value="">{t("layouts.trackReservation.choose")}</option>
          {assets.map((asset) => <option key={asset.id} value={asset.id}>
            {asset.inventoryNumber || asset.serialNumber || asset.id} · {locationNames[asset.storageLocationId ?? ""]}
          </option>)}</AppSelect></label>
          : <label>{t("layouts.trackReservation.location")}<AppSelect value={locationID}
            aria-label={t("layouts.trackReservation.location")} disabled={saving || loading}
            onChange={(event) => setLocationID(event.target.value)}><option value="">{t("layouts.trackReservation.choose")}</option>
            {stock?.locations.filter((location) => location.quantity > 0).map((location) =>
              <option key={location.locationId} value={location.locationId}>
                {location.locationName} · {location.quantity} {t("layouts.trackReservation.pieces")}
              </option>)}</AppSelect></label>}
      </div>}
      <label className="track-reservation-confirm"><input type="checkbox" checked={confirmed} disabled={saving}
        onChange={(event) => setConfirmed(event.target.checked)} />
        <span>{t("layouts.trackReservation.confirm")}</span></label>
      {message ? <p className="form-message">{message}</p> : null}
      <footer><button type="button" className="secondary-button" onClick={onClose} disabled={saving}>
        {t("common.cancel")}</button>
        <button type="button" className="primary-button" onClick={() => void reserve()}
          disabled={!canReserve || saving || loading}>{saving ? t("common.saving")
            : t("layouts.trackReservation.reserve")}</button></footer>
    </section>
  </div>, document.body);
}
