import { Fragment, useState, type CSSProperties, type ReactNode } from "react";
import {
  ArrowUpDown,
  ChevronDown,
  ChevronUp,
  CircleOff,
  Eye,
  ImageOff,
  Pencil,
  Trash2
} from "lucide-react";

import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import {
  tableColumnWidth,
  tableMinimumWidth,
  type TableColumnWidths
} from "../../shared/tableColumnLayout";
import { TableColumnResizeHandle } from "../../shared/ui/TableColumnResizeHandle";
import { previewImageUrl, primaryImage, vehicleExhibitionEligible } from "./vehicleTransforms";
import {
  sortableVehicleColumn,
  vehicleColumnLabel,
  vehicleColumnText,
  vehicleTableColumnWidthDefinitions,
  type VehicleTableColumn
} from "./vehicleTableColumns";
import type { SortDirection, SortKey } from "./vehicleViewModel";
import { groupVehicleInventory } from "./vehicleSetGroups";
import { VehicleSetInventoryRow } from "./VehicleSetInventoryRow";

type VehicleInventoryTableProps = {
  vehicles: Vehicle[];
  columns: readonly VehicleTableColumn[];
  columnWidths?: TableColumnWidths<VehicleTableColumn>;
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
  onPreviewColumnWidth?: (column: VehicleTableColumn, width: number) => void;
  onCommitColumnWidth?: (column: VehicleTableColumn, width: number) => void;
  renderQuickMenu: (vehicle: Vehicle) => ReactNode;
};

export function VehicleInventoryTable({
  vehicles,
  columns,
  columnWidths = {},
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
  onPreviewColumnWidth,
  onCommitColumnWidth,
  renderQuickMenu
}: VehicleInventoryTableProps) {
  const { language, t } = useI18n();
  const [expandedSetIDs, setExpandedSetIDs] = useState<Set<string>>(() => new Set());
  const groupedVehicles = groupVehicleInventory(vehicles, sort);
  const columnLayout = { columns: [...columns], widths: columnWidths };
  const minimumWidth = tableMinimumWidth(
    columnLayout,
    vehicleTableColumnWidthDefinitions,
    64 + 122
  );
  const growColumn = columns.includes("name") ? "name" : columns.at(-1);

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
      if (!vehicle.exhibition && !vehicleExhibitionEligible(vehicle)) {
        return (
          <span
            className="inventory-exhibition-unavailable"
            aria-label={t("vehicles.exhibition.requiresDecoder")}
          >
            <CircleOff size={14} aria-hidden="true" />
            <span>{t("vehicles.exhibition.unavailable")}</span>
          </span>
        );
      }
      return (
        <label
          className={vehicle.exhibition ? "inventory-inline-switch active" : "inventory-inline-switch"}
          title={t("vehicles.exhibition.toggle")}
        >
          <input
            type="checkbox"
            checked={Boolean(vehicle.exhibition)}
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
      <table
        className="inventory-table vehicle-inventory-table"
        style={{
          "--vehicle-data-column-count": Math.max(columns.length, 1),
          "--vehicle-table-min-width": `${minimumWidth}px`
        } as CSSProperties}
      >
        <colgroup>
          <col className="select-cell" style={{ width: 64, minWidth: 64, maxWidth: 64 }} />
          {columns.map((column) => <col
            key={column}
            data-column={column}
            style={{ width: column === growColumn
              ? `calc(${tableColumnWidth(columnLayout, column, vehicleTableColumnWidthDefinitions)}px + ` +
                `max(0px, 100% - ${minimumWidth}px))`
              : tableColumnWidth(columnLayout, column, vehicleTableColumnWidthDefinitions) }}
          />)}
          <col className="actions-cell" style={{ width: 122, minWidth: 122, maxWidth: 122 }} />
        </colgroup>
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
            {columns.map((column) => {
              const width = tableColumnWidth(
                columnLayout,
                column,
                vehicleTableColumnWidthDefinitions
              );
              const definition = vehicleTableColumnWidthDefinitions[column];
              const label = vehicleColumnLabel(column, t);
              return (
                <th key={column} className={`vehicle-column-${column}`}>
                  {header(column)}
                  {onPreviewColumnWidth && onCommitColumnWidth ? <TableColumnResizeHandle
                    label={t("common.resizeColumn", { label })}
                    width={width}
                    minWidth={definition.minWidth}
                    maxWidth={definition.maxWidth}
                    defaultWidth={definition.defaultWidth}
                    onPreview={(next) => onPreviewColumnWidth(column, next)}
                    onCommit={(next) => onCommitColumnWidth(column, next)}
                  /> : null}
                </th>
              );
            })}
            <th className="actions-cell">{t("vehicles.actions")}</th>
          </tr>
        </thead>
        <tbody>
          {groupedVehicles.map((group) => group.kind === "single" ? vehicleRow(group.vehicle) : (
            <Fragment key={group.id}>
							<VehicleSetInventoryRow
								group={group}
								columns={columns}
								collapsed={!expandedSetIDs.has(group.id)}
								selectedVehicleIDs={selectedVehicleIDs}
								onToggleCollapsed={() => setExpandedSetIDs((current) => {
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
              {expandedSetIDs.has(group.id) && group.members.map((vehicle, index) => (
								vehicleRow(vehicle, true, index === group.members.length - 1)
							))}
            </Fragment>
          ))}
        </tbody>
      </table>
    </div>
  );
}
