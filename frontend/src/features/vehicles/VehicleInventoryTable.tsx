import { Fragment, useState, type ReactNode } from "react";
import {
  ArrowUpDown,
  ChevronDown,
  ChevronUp,
  Eye,
  ImageOff,
  Pencil,
  Trash2
} from "lucide-react";

import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { previewImageUrl, primaryImage, vehicleExhibitionEligible } from "./vehicleTransforms";
import {
  sortableVehicleColumn,
  vehicleColumnLabel,
  vehicleColumnText,
  type VehicleTableColumn
} from "./vehicleTableColumns";
import type { SortDirection, SortKey } from "./vehicleViewModel";
import { groupVehicleInventory } from "./vehicleSetGroups";
import { VehicleSetInventoryRow } from "./VehicleSetInventoryRow";

type VehicleInventoryTableProps = {
  vehicles: Vehicle[];
  columns: readonly VehicleTableColumn[];
  allVisibleSelected: boolean;
  selectedVehicleIDs: Set<string>;
  sort: { key: SortKey; direction: SortDirection };
  onToggleSort: (key: SortKey) => void;
  onToggleSelection: (vehicleID: string) => void;
	onToggleSetSelection?: (vehicleIDs: string[]) => void;
  onToggleAllVisibleSelection: () => void;
  onOpenDetail: (vehicle: Vehicle) => void;
	onOpenSet?: (setID: string) => void;
	onEditSet?: (setID: string) => void;
	onDuplicateSet?: (setID: string) => void;
  onOpenEdit?: (vehicle: Vehicle) => void;
  onDelete?: (vehicle: Vehicle) => void;
  onToggleExhibition: (vehicle: Vehicle, exhibition: boolean) => void;
  renderQuickMenu: (vehicle: Vehicle) => ReactNode;
};

export function VehicleInventoryTable({
  vehicles,
  columns,
  allVisibleSelected,
  selectedVehicleIDs,
  sort,
  onToggleSort,
  onToggleSelection,
	onToggleSetSelection,
  onToggleAllVisibleSelection,
  onOpenDetail,
	onOpenSet,
	onEditSet,
	onDuplicateSet,
  onOpenEdit,
  onDelete,
  onToggleExhibition,
  renderQuickMenu
}: VehicleInventoryTableProps) {
  const { language, t } = useI18n();
  const [collapsedSetIDs, setCollapsedSetIDs] = useState<Set<string>>(() => new Set());
  const groupedVehicles = groupVehicleInventory(vehicles, sort);

  const header = (column: VehicleTableColumn) => {
    const label = vehicleColumnLabel(column, t);
    if (!sortableVehicleColumn(column)) return label;
    return (
      <button
        type="button"
        className={`sort-button ${sort.key === column ? "active" : ""}`}
        onClick={() => onToggleSort(column)}
        title={t("common.sort", { label })}
      >
        {label}
        {sort.key === column
          ? sort.direction === "asc"
            ? <ChevronUp size={14} aria-hidden="true" />
            : <ChevronDown size={14} aria-hidden="true" />
          : <ArrowUpDown size={13} aria-hidden="true" />}
      </button>
    );
  };

	const cell = (vehicle: Vehicle, column: VehicleTableColumn, setMember = false) => {
		if (column === "type") {
			const image = primaryImage(vehicle.images);
			return (
				<div className="vehicle-member-type-cell">
					{image
						? <img className="inventory-thumb" src={previewImageUrl(image)} alt="" />
						: <div className="image-placeholder inventory-image-placeholder"
							aria-label={t("exhibition.noPreview")} title={t("exhibition.noPreview")}>
							<ImageOff size={18} strokeWidth={1.7} aria-hidden="true" />
						</div>}
				</div>
			);
		}
		if (column === "inventoryNumber" && setMember) {
			return <span className="vehicle-member-inventory">
				<strong>{vehicle.inventoryNumber}</strong>
				{vehicle.vehicleNumber && <small>{vehicle.vehicleNumber}</small>}
			</span>;
		}
    if (column === "image") {
      const image = primaryImage(vehicle.images);
      return image
        ? <img className="inventory-thumb" src={previewImageUrl(image)} alt="" />
        : <div className="image-placeholder inventory-image-placeholder"
          aria-label={t("exhibition.noPreview")} title={t("exhibition.noPreview")}>
          <ImageOff size={18} strokeWidth={1.7} aria-hidden="true" />
        </div>;
    }
    if (column === "name") {
      return (
        <button type="button" className="inventory-name-link" onClick={() => onOpenDetail(vehicle)}>
          {vehicleColumnText(vehicle, column, language, t)}
        </button>
      );
    }
    if (column === "exhibition") {
      return (
        <label
          className={vehicle.exhibition ? "inventory-inline-switch active" : "inventory-inline-switch"}
          title={vehicle.exhibition || vehicleExhibitionEligible(vehicle)
            ? t("vehicles.exhibition.toggle")
            : t("vehicles.exhibition.requiresDecoder")}
        >
          <input
            type="checkbox"
            checked={Boolean(vehicle.exhibition)}
            disabled={!vehicle.exhibition && !vehicleExhibitionEligible(vehicle)}
            onChange={(event) => onToggleExhibition(vehicle, event.target.checked)}
            aria-label={t("vehicles.exhibition.toggle")}
          />
          <span aria-hidden="true" />
          <em>{vehicleColumnText(vehicle, column, language, t)}</em>
        </label>
      );
    }
    return vehicleColumnText(vehicle, column, language, t);
  };

  const vehicleRow = (vehicle: Vehicle, setMember = false, lastSetMember = false) => (
    <tr
      key={vehicle.id}
      className={[
        selectedVehicleIDs.has(vehicle.id) ? "selected-row" : "",
        setMember ? "vehicle-set-child-row" : "",
				lastSetMember ? "vehicle-set-child-row-last" : ""
      ].filter(Boolean).join(" ")}
    >
      <td className="select-cell">
        <label className="table-select-field" title={t("vehicles.report.selectVehicle")}>
          <input
            type="checkbox"
            checked={selectedVehicleIDs.has(vehicle.id)}
            onChange={() => onToggleSelection(vehicle.id)}
            aria-label={`${vehicle.inventoryNumber} ${t("vehicles.report.selectVehicle")}`}
          />
        </label>
      </td>
			{columns.map((column) => (
				<td key={column} className={`vehicle-column-${column}`}>{cell(vehicle, column, setMember)}</td>
			))}
      <td className="actions-cell">
        <div className="table-actions">
          <button type="button" className="icon-button" onClick={() => onOpenDetail(vehicle)} aria-label={t("exhibition.view")} title={t("exhibition.view")}>
            <Eye size={16} aria-hidden="true" />
          </button>
          {onOpenEdit ? (
            <button type="button" className="icon-button" onClick={() => onOpenEdit(vehicle)} aria-label={t("vehicles.edit")} title={t("vehicles.edit")}>
              <Pencil size={16} aria-hidden="true" />
            </button>
          ) : null}
          {onDelete ? (
            <button type="button" className="icon-button danger" onClick={() => onDelete(vehicle)} aria-label={t("vehicles.delete")} title={t("vehicles.delete")}>
              <Trash2 size={16} aria-hidden="true" />
            </button>
          ) : null}
          {renderQuickMenu(vehicle)}
        </div>
      </td>
    </tr>
  );

  return (
    <div className="table-wrap">
      <table className="inventory-table vehicle-inventory-table">
        <thead>
          <tr>
            <th className="select-cell">
              <label className="table-select-field" title={t("vehicles.report.selectAll")}>
                <input
                  type="checkbox"
                  checked={allVisibleSelected}
                  onChange={onToggleAllVisibleSelection}
                  aria-label={t("vehicles.report.selectAll")}
                  disabled={vehicles.length === 0}
                />
              </label>
            </th>
            {columns.map((column) => (
              <th key={column} className={`vehicle-column-${column}`}>{header(column)}</th>
            ))}
            <th className="actions-cell">{t("vehicles.actions")}</th>
          </tr>
        </thead>
        <tbody>
          {groupedVehicles.map((group) => group.kind === "single" ? vehicleRow(group.vehicle) : (
            <Fragment key={group.id}>
							<VehicleSetInventoryRow
								group={group}
								columns={columns}
								collapsed={collapsedSetIDs.has(group.id)}
								selectedVehicleIDs={selectedVehicleIDs}
								onToggleCollapsed={() => setCollapsedSetIDs((current) => {
									const next = new Set(current);
									if (next.has(group.id)) next.delete(group.id);
									else next.add(group.id);
									return next;
								})}
								onToggleSelection={(ids) => onToggleSetSelection?.(ids)}
								onOpen={(id) => onOpenSet?.(id)}
								onEdit={onEditSet}
								onDuplicate={onDuplicateSet}
							/>
              {!collapsedSetIDs.has(group.id) && group.members.map((vehicle, index) => (
								vehicleRow(vehicle, true, index === group.members.length - 1)
							))}
            </Fragment>
          ))}
        </tbody>
      </table>
    </div>
  );
}
