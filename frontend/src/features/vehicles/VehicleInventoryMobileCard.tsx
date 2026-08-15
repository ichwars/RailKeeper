import type { ReactNode } from "react";
import { ChevronDown, ChevronUp, Eye, Pencil } from "lucide-react";

import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { previewImageUrl, primaryImage } from "./vehicleTransforms";
import {
  vehicleColumnLabel,
  vehicleColumnText,
  type VehicleTableColumn
} from "./vehicleTableColumns";

const mobileSummaryColumns = new Set<VehicleTableColumn>([
  "inventoryNumber",
  "name",
  "manufacturer",
  "articleNumber",
  "gauge",
  "epoch"
]);

type VehicleInventoryMobileCardProps = {
  vehicle: Vehicle;
  columns: readonly VehicleTableColumn[];
  expanded: boolean;
  onToggleExpanded: () => void;
  onOpenDetail: (vehicle: Vehicle) => void;
  onOpenEdit: (vehicle: Vehicle) => void;
  renderQuickMenu: (vehicle: Vehicle) => ReactNode;
};

export function VehicleInventoryMobileCard({
  vehicle,
  columns,
  expanded,
  onToggleExpanded,
  onOpenDetail,
  onOpenEdit,
  renderQuickMenu
}: VehicleInventoryMobileCardProps) {
  const { language, t } = useI18n();
  const image = primaryImage(vehicle.images);
  const showsImage = columns.includes("image");
  const shows = (column: VehicleTableColumn) => columns.includes(column);
  const text = (column: VehicleTableColumn) =>
    vehicleColumnText(vehicle, column, language, t);
  const detailColumns = columns.filter((column) => (
    column !== "image" && !mobileSummaryColumns.has(column)
  ));
  const makerLine = [
    shows("manufacturer") ? text("manufacturer") : "",
    shows("articleNumber") ? text("articleNumber") : ""
  ].filter(Boolean).join(" · ");
  const metaLine = [
    shows("gauge") ? text("gauge") : "",
    shows("epoch") ? text("epoch") : ""
  ].filter(Boolean).join(" · ");
  const detailsID = `vehicle-mobile-details-${vehicle.id}`;

  return (
    <article
      className={`inventory-mobile-item vehicle-mobile-item${showsImage ? "" : " no-image"}`}
    >
      <button
        type="button"
        className="vehicle-mobile-toggle"
        onClick={onToggleExpanded}
        aria-expanded={expanded}
        aria-controls={detailsID}
        aria-label={t(expanded ? "vehicles.mobile.collapse" : "vehicles.mobile.expand", {
          inventoryNumber: vehicle.inventoryNumber
        })}
      >
        {showsImage ? (
          <span className="vehicle-mobile-image" aria-hidden="true">
            {image
              ? <img src={previewImageUrl(image)} alt="" />
              : (
                  <span className="image-placeholder">
                    {t("exhibition.noPreview")}
                  </span>
                )}
          </span>
        ) : null}
        <span className="vehicle-mobile-summary">
          {shows("inventoryNumber") ? <small>{text("inventoryNumber")}</small> : null}
          {shows("name") ? <strong>{text("name")}</strong> : null}
          {makerLine ? <span>{makerLine}</span> : null}
          {metaLine ? <em>{metaLine}</em> : null}
        </span>
        {expanded
          ? <ChevronUp size={18} aria-hidden="true" />
          : <ChevronDown size={18} aria-hidden="true" />}
      </button>

      {expanded && detailColumns.length > 0 ? (
        <dl className="vehicle-mobile-fields" id={detailsID}>
          {detailColumns.map((column) => (
            <div key={column}>
              <dt>{vehicleColumnLabel(column, t)}</dt>
              <dd>{text(column)}</dd>
            </div>
          ))}
        </dl>
      ) : null}

      <div className="inventory-mobile-actions">
        <button
          type="button"
          className="icon-button"
          onClick={() => onOpenDetail(vehicle)}
          aria-label={t("vehicles.view")}
          title={t("vehicles.view")}
        >
          <Eye size={16} aria-hidden="true" />
        </button>
        <button
          type="button"
          className="icon-button"
          onClick={() => onOpenEdit(vehicle)}
          aria-label={t("vehicles.edit")}
          title={t("vehicles.edit")}
        >
          <Pencil size={16} aria-hidden="true" />
        </button>
        {renderQuickMenu(vehicle)}
      </div>
    </article>
  );
}
