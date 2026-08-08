import { useState } from "react";
import type { FormEvent } from "react";
import { Boxes, PackageCheck } from "lucide-react";

import {
  api,
  type AccessoryAsset,
  type AccessoryAssetInput,
  type AccessoryProduct,
  type AccessoryStockSummary,
  type StorageLocation
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AccessoryConfirmDialog, type AccessoryPendingAction } from "./AccessoryConfirmDialog";
import { activeStorageLocations, storageLocationPath } from "../settings/storageLocations";

export function AccessoryStockPanel({ mode, product, stock, assets, locations, canEdit, onChanged }: {
  mode: "stock" | "assets";
  product: AccessoryProduct | null;
  stock: AccessoryStockSummary | null;
  assets: AccessoryAsset[];
  locations: StorageLocation[];
  canEdit: boolean;
  onChanged: () => Promise<void>;
}) {
  const [delta, setDelta] = useState("1");
  const [locationID, setLocationID] = useState("");
  const [asset, setAsset] = useState<AccessoryAssetInput>({ condition: "ready", lifecycle: "stored" });
  const [action, setAction] = useState<AccessoryPendingAction | null>(null);
  const { t } = useI18n();
  if (!product) return <section className="panel"><p>{t("accessories.selection.empty")}</p></section>;
  const activeLocations = activeStorageLocations(locations);
  const effectiveLocationID = activeLocations.some((location) => location.id === locationID)
    ? locationID : activeLocations[0]?.id || "";

  const submitStock = (event: FormEvent) => {
    event.preventDefault();
    setAction({
      title: t("accessories.stock.confirmTitle"),
      body: t("accessories.stock.confirmBody", { count: Number(delta), product: product.name }),
      run: async () => {
        await api.adjustAccessoryStock(product.id, { locationId: effectiveLocationID, delta: Number(delta) });
        setDelta("1");
        await onChanged();
      }
    });
  };

  const submitAsset = (event: FormEvent) => {
    event.preventDefault();
    setAction({
      title: t("accessories.assets.confirmTitle"),
      body: t("accessories.assets.confirmBody", { product: product.name }),
      run: async () => {
        await api.createAccessoryAsset(product.id, { ...asset, storageLocationId: effectiveLocationID });
        setAsset({ condition: "ready", lifecycle: "stored" });
        await onChanged();
      }
    });
  };

  if (mode === "stock") {
    return <>
      <section className="panel accessory-stock-panel">
        <div className="panel-head"><Boxes size={17} aria-hidden="true" /><h2>{t("accessories.stock.title")}</h2></div>
        {product.trackingMode !== "quantity" ? <p>{t("accessories.stock.individualHint")}</p> : (
          <div className="accessory-work-grid">
            <div className="table-wrap"><table><thead><tr><th>{t("accessories.field.location")}</th><th>{t("accessories.field.quantity")}</th></tr></thead>
              <tbody>{(stock?.locations || []).map((level) => <tr key={level.locationId}><td>{level.locationName}</td><td>{level.quantity}</td></tr>)}</tbody>
            </table></div>
            {canEdit ? <form className="accessory-form" onSubmit={submitStock}>
              <h3>{t("accessories.stock.adjust")}</h3>
              <label>{t("accessories.field.location")}<AppSelect value={effectiveLocationID} onChange={(event) => setLocationID(event.target.value)}>
                {activeLocations.map((location) => <option key={location.id} value={location.id}>
                  {storageLocationPath(location, locations)}</option>)}
              </AppSelect></label>
              <label>{t("accessories.field.delta")}<input type="number" required value={delta}
                onChange={(event) => setDelta(event.target.value)} /></label>
              <button type="submit" className="primary-button" disabled={!effectiveLocationID}>
                {t("accessories.stock.book")}
              </button>
            </form> : null}
          </div>
        )}
      </section>
      {action ? <AccessoryConfirmDialog action={action} onClose={() => setAction(null)} /> : null}
    </>;
  }

  return <>
    <section className="panel accessory-stock-panel">
      <div className="panel-head"><PackageCheck size={17} aria-hidden="true" /><h2>{t("accessories.assets.title")}</h2></div>
      {product.trackingMode !== "individual" ? <p>{t("accessories.assets.quantityHint")}</p> : (
        <div className="accessory-work-grid">
          <div className="table-wrap"><table><thead><tr><th>{t("accessories.field.inventoryNumber")}</th><th>{t("accessories.field.condition")}</th><th>{t("accessories.field.lifecycle")}</th></tr></thead>
            <tbody>{assets.map((item) => <tr key={item.id}><td>{item.inventoryNumber || item.serialNumber || "-"}</td><td>{t(`accessories.condition.${item.condition}`)}</td><td>{t(`accessories.lifecycle.${item.lifecycle}`)}</td></tr>)}</tbody>
          </table></div>
          {canEdit ? <form className="accessory-form" onSubmit={submitAsset}>
            <h3>{t("accessories.assets.create")}</h3>
            <label>{t("accessories.field.inventoryNumber")}<input value={asset.inventoryNumber || ""}
              onChange={(event) => setAsset((current) => ({ ...current, inventoryNumber: event.target.value }))} /></label>
            <label>{t("accessories.field.serialNumber")}<input value={asset.serialNumber || ""}
              onChange={(event) => setAsset((current) => ({ ...current, serialNumber: event.target.value }))} /></label>
            <label>{t("accessories.field.location")}<AppSelect value={effectiveLocationID} onChange={(event) => setLocationID(event.target.value)}>
              {activeLocations.map((location) => <option key={location.id} value={location.id}>
                {storageLocationPath(location, locations)}</option>)}
            </AppSelect></label>
            <label>{t("accessories.field.condition")}<AppSelect value={asset.condition || "ready"}
              onChange={(event) => setAsset((current) => ({ ...current,
                condition: event.target.value as AccessoryAssetInput["condition"] }))}>
              {(["ready", "maintenance_due", "defective", "unknown"] as const).map((condition) =>
                <option key={condition} value={condition}>{t(`accessories.condition.${condition}`)}</option>)}
            </AppSelect></label>
            <label>{t("accessories.field.lifecycle")}<AppSelect value={asset.lifecycle || "stored"}
              onChange={(event) => setAsset((current) => ({ ...current,
                lifecycle: event.target.value as AccessoryAssetInput["lifecycle"] }))}>
              {(["stored", "maintenance", "retired"] as const).map((lifecycle) =>
                <option key={lifecycle} value={lifecycle}>{t(`accessories.lifecycle.${lifecycle}`)}</option>)}
            </AppSelect></label>
            <label>{t("accessories.field.purchaseDate")}<input type="date" value={asset.purchaseDate || ""}
              onChange={(event) => setAsset((current) => ({ ...current, purchaseDate: event.target.value }))} /></label>
            <label>{t("accessories.field.purchasePrice")}<input inputMode="decimal" value={asset.purchasePrice || ""}
              onChange={(event) => setAsset((current) => ({ ...current, purchasePrice: event.target.value }))} /></label>
            <label>{t("accessories.field.warrantyUntil")}<input type="date" value={asset.warrantyUntil || ""}
              onChange={(event) => setAsset((current) => ({ ...current, warrantyUntil: event.target.value }))} /></label>
            <label>{t("accessories.field.notes")}<textarea value={asset.notes || ""}
              onChange={(event) => setAsset((current) => ({ ...current, notes: event.target.value }))} /></label>
            <button type="submit" className="primary-button" disabled={!effectiveLocationID}>
              {t("accessories.assets.save")}
            </button>
          </form> : null}
        </div>
      )}
    </section>
    {action ? <AccessoryConfirmDialog action={action} onClose={() => setAction(null)} /> : null}
  </>;
}
