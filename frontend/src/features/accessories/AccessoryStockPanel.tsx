import { useState, type FormEvent } from "react";
import { ArrowRightLeft, Boxes, PackageCheck } from "lucide-react";

import {
  api,
  type AccessoryArticle,
  type AccessoryAsset,
  type AccessoryAssetInput,
  type AccessoryStockMovement,
  type AccessoryStockSummary,
  type StorageLocation
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { activeStorageLocations, storageLocationPath } from "../../shared/storageLocations";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppTextInput } from "../../shared/ui/AppTextInput";
import { AccessoryConfirmDialog, type AccessoryPendingAction } from "./AccessoryConfirmDialog";

export function AccessoryStockPanel({ article, stock, movements, assets, locations, canEdit, onChanged }: {
  article: AccessoryArticle;
  stock: AccessoryStockSummary | null;
  movements: AccessoryStockMovement[];
  assets: AccessoryAsset[];
  locations: StorageLocation[];
  canEdit: boolean;
  onChanged: () => Promise<void>;
}) {
  const [delta, setDelta] = useState("1");
  const [locationId, setLocationId] = useState("");
  const [transferToId, setTransferToId] = useState("");
  const [transferQuantity, setTransferQuantity] = useState("1");
  const [transferNote, setTransferNote] = useState("");
  const [asset, setAsset] = useState<AccessoryAssetInput>({ condition: "ready", lifecycle: "stored" });
  const [action, setAction] = useState<AccessoryPendingAction | null>(null);
  const { t } = useI18n();
  const activeLocations = activeStorageLocations(locations);
  const effectiveLocationId = activeLocations.some((location) => location.id === locationId)
    ? locationId : activeLocations[0]?.id || "";
  const targetLocations = activeLocations.filter((location) => location.id !== effectiveLocationId);
  const effectiveTransferToId = targetLocations.some((location) => location.id === transferToId)
    ? transferToId : targetLocations[0]?.id || "";
  const supportsQuantity = article.inventoryStrategy !== "individual";
  const supportsAssets = article.inventoryStrategy !== "quantity";

  const confirm = (title: string, body: string, run: () => Promise<void>) => setAction({ title, body, run });

  const submitAdjustment = (event: FormEvent) => {
    event.preventDefault();
    confirm(t("accessories.stock.confirmTitle"), t("accessories.stock.confirmBody", {
      count: Number(delta), product: article.name
    }), async () => {
      await api.adjustAccessoryStock(article.id, { locationId: effectiveLocationId, delta: Number(delta) });
      setDelta("1");
      await onChanged();
    });
  };

  const submitTransfer = (event: FormEvent) => {
    event.preventDefault();
    confirm(t("accessories.editor.stock.transferConfirm"), t("accessories.editor.stock.transferConfirmBody"), async () => {
      await api.transferAccessoryStock(article.id, {
        fromLocationId: effectiveLocationId,
        toLocationId: effectiveTransferToId,
        quantity: Number(transferQuantity),
        note: transferNote.trim() || undefined
      });
      setTransferQuantity("1");
      setTransferNote("");
      await onChanged();
    });
  };

  const submitAsset = (event: FormEvent) => {
    event.preventDefault();
    const run = article.inventoryStrategy === "quantity_later_individual"
      ? () => api.individualizeAccessoryProduct(article.id, { locationId: effectiveLocationId, asset })
      : () => api.createAccessoryAsset(article.id, { ...asset, storageLocationId: effectiveLocationId });
    confirm(t("accessories.assets.confirmTitle"), t("accessories.assets.confirmBody", { product: article.name }), async () => {
      await run();
      setAsset({ condition: "ready", lifecycle: "stored" });
      await onChanged();
    });
  };

  return <>
    <div className="article-stock-sections">
      <section className="article-editor-section">
        <div className="panel-head"><Boxes size={17} aria-hidden="true" /><h3>{t("accessories.stock.title")}</h3></div>
        <div className="table-wrap"><table><thead><tr>
          <th>{t("accessories.field.location")}</th><th>{t("accessories.field.quantity")}</th>
        </tr></thead><tbody>{(stock?.locations || []).map((level) => <tr key={level.locationId}>
          <td>{level.locationName}</td><td>{level.quantity}</td>
        </tr>)}</tbody></table></div>
        {canEdit && supportsQuantity ? <div className="article-stock-commands">
          <form className="accessory-form" onSubmit={submitAdjustment}>
            <h4>{t("accessories.stock.adjust")}</h4>
            <LocationSelect label={t("accessories.field.location")} value={effectiveLocationId}
              locations={activeLocations} allLocations={locations} onChange={setLocationId} />
            <AppNumberInput label={t("accessories.field.delta")} required value={delta}
              onValueChange={setDelta} />
            <button type="submit" className="primary-button" disabled={!effectiveLocationId || !Number(delta)}>
              {t("accessories.stock.book")}
            </button>
          </form>
          <form className="accessory-form" onSubmit={submitTransfer}>
            <h4><ArrowRightLeft size={15} aria-hidden="true" /> {t("accessories.editor.stock.transfer")}</h4>
            <LocationSelect label={t("accessories.editor.stock.fromLocation")} value={effectiveLocationId}
              locations={activeLocations} allLocations={locations} onChange={setLocationId} />
            <LocationSelect label={t("accessories.editor.stock.toLocation")} value={effectiveTransferToId}
              locations={targetLocations} allLocations={locations} onChange={setTransferToId} />
            <AppNumberInput label={t("accessories.field.quantity")} min="1" required value={transferQuantity}
              onValueChange={setTransferQuantity} />
            <AppTextInput label={t("accessories.field.notes")} value={transferNote}
              onChange={(event) => setTransferNote(event.target.value)} />
            <button type="submit" className="primary-button"
              disabled={!effectiveLocationId || !effectiveTransferToId || Number(transferQuantity) <= 0}>
              {t("accessories.editor.stock.transfer")}
            </button>
          </form>
        </div> : null}
      </section>

      {supportsAssets ? <section className="article-editor-section">
        <div className="panel-head"><PackageCheck size={17} aria-hidden="true" /><h3>{t("accessories.assets.title")}</h3></div>
        <div className="table-wrap"><table><thead><tr>
          <th>{t("accessories.field.inventoryNumber")}</th><th>{t("accessories.field.condition")}</th>
          <th>{t("accessories.field.lifecycle")}</th>
        </tr></thead><tbody>{assets.map((item) => <tr key={item.id}>
          <td>{item.inventoryNumber || item.serialNumber || "-"}</td>
          <td>{t(`accessories.condition.${item.condition}`)}</td><td>{t(`accessories.lifecycle.${item.lifecycle}`)}</td>
        </tr>)}</tbody></table></div>
        {canEdit ? <form className="accessory-form article-asset-form" onSubmit={submitAsset}>
          <h4>{article.inventoryStrategy === "quantity_later_individual"
            ? t("accessories.editor.stock.individualize") : t("accessories.assets.create")}</h4>
          <AppTextInput label={t("accessories.field.inventoryNumber")} value={asset.inventoryNumber || ""}
            onChange={(event) => setAsset((current) => ({ ...current, inventoryNumber: event.target.value }))} />
          <AppTextInput label={t("accessories.field.serialNumber")} value={asset.serialNumber || ""}
            onChange={(event) => setAsset((current) => ({ ...current, serialNumber: event.target.value }))} />
          <LocationSelect label={t("accessories.field.location")} value={effectiveLocationId}
            locations={activeLocations} allLocations={locations} onChange={setLocationId} />
          <button type="submit" className="primary-button" disabled={!effectiveLocationId}>
            {t("accessories.assets.save")}
          </button>
        </form> : null}
      </section> : null}

      <section className="article-editor-section">
        <h3>{t("accessories.editor.stock.journal")}</h3>
        <div className="table-wrap"><table><thead><tr>
          <th>{t("accessories.editor.stock.date")}</th><th>{t("accessories.editor.stock.movement")}</th>
          <th>{t("accessories.field.quantity")}</th><th>{t("accessories.field.notes")}</th>
        </tr></thead><tbody>{movements.map((movement) => <tr key={movement.id}>
          <td>{new Date(movement.createdAt).toLocaleDateString()}</td>
          <td>{t(`accessories.editor.movement.${movement.movementType}`)}</td>
          <td>{movement.quantity}</td><td>{movement.note || "-"}</td>
        </tr>)}</tbody></table></div>
      </section>
    </div>
    {action ? <AccessoryConfirmDialog action={action} onClose={() => setAction(null)} /> : null}
  </>;
}

function LocationSelect({ label, value, locations, allLocations, onChange }: {
  label: string;
  value: string;
  locations: StorageLocation[];
  allLocations: StorageLocation[];
  onChange: (value: string) => void;
}) {
  return <label className="app-field">
    <span className="app-field-label">{label}</span>
    <AppSelect value={value} aria-label={label} onChange={(event) => onChange(event.target.value)}>
      {locations.map((location) => <option key={location.id} value={location.id}>
        {storageLocationPath(location, allLocations)}
      </option>)}
    </AppSelect>
  </label>;
}
