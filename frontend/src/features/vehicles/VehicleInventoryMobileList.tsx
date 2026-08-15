import type { ReactNode } from "react";
import { Pencil, Trash2 } from "lucide-react";

import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { previewImageUrl, primaryImage } from "./vehicleTransforms";
import {
  vehicleColumnLabel,
  vehicleColumnText,
  type VehicleTableColumn
} from "./vehicleTableColumns";

type VehicleInventoryMobileListProps = {
  vehicles: Vehicle[];
  columns: readonly VehicleTableColumn[];
  onOpenDetail: (vehicle: Vehicle) => void;
  onOpenEdit: (vehicle: Vehicle) => void;
  onDelete: (vehicle: Vehicle) => void;
  renderQuickMenu: (vehicle: Vehicle) => ReactNode;
};

export function VehicleInventoryMobileList({
  vehicles,
  columns,
  onOpenDetail,
  onOpenEdit,
  onDelete,
  renderQuickMenu
}: VehicleInventoryMobileListProps) {
  const { language, t } = useI18n();
  const dataColumns = columns.filter((column) => column !== "image");

  return (
    <div className="inventory-mobile-list" aria-label={t("vehicles.mobileList")}>
      {vehicles.map((vehicle) => {
        const image = primaryImage(vehicle.images);
        const showsImage = columns.includes("image");
        return (
          <article
            key={vehicle.id}
            className={`inventory-mobile-item vehicle-mobile-item${showsImage ? "" : " no-image"}`}
          >
            {showsImage ? (
              <button
                type="button"
                className="inventory-mobile-media"
                onClick={() => onOpenDetail(vehicle)}
                aria-label={`${vehicle.inventoryNumber} ${t("vehicles.view")}`}
              >
                {image
                  ? <img src={previewImageUrl(image)} alt="" />
                  : <div className="image-placeholder">{t("exhibition.noPreview")}</div>}
              </button>
            ) : null}
            <button
              type="button"
              className="inventory-mobile-main vehicle-mobile-main"
              onClick={() => onOpenDetail(vehicle)}
              aria-label={`${vehicle.inventoryNumber} ${vehicle.name}`}
            >
              <dl className="vehicle-mobile-fields">
                {dataColumns.map((column) => (
                  <div key={column}>
                    <dt>{vehicleColumnLabel(column, t)}</dt>
                    <dd>{vehicleColumnText(vehicle, column, language, t)}</dd>
                  </div>
                ))}
              </dl>
            </button>
            <div className="inventory-mobile-actions">
              <button type="button" className="icon-button" onClick={() => onOpenEdit(vehicle)} aria-label={t("vehicles.edit")} title={t("vehicles.edit")}>
                <Pencil size={16} aria-hidden="true" />
              </button>
              <button type="button" className="icon-button danger" onClick={() => onDelete(vehicle)} aria-label={t("vehicles.delete")} title={t("vehicles.delete")}>
                <Trash2 size={16} aria-hidden="true" />
              </button>
              {renderQuickMenu(vehicle)}
            </div>
          </article>
        );
      })}
    </div>
  );
}
