import { useEffect, useRef } from "react";
import { ChevronDown, ChevronUp, Copy, Eye, ImageOff, Pencil } from "lucide-react";

import { useI18n } from "../../shared/i18n";
import type { VehicleInventorySetGroup } from "./vehicleSetGroups";
import type { VehicleTableColumn } from "./vehicleTableColumns";

type VehicleSetInventoryRowProps = {
	group: VehicleInventorySetGroup;
	columns: readonly VehicleTableColumn[];
	collapsed: boolean;
	selectedVehicleIDs: ReadonlySet<string>;
	onToggleCollapsed: () => void;
	onToggleSelection: (vehicleIDs: string[]) => void;
	onOpen: (setId: string) => void;
	onEdit?: (setId: string) => void;
	onDuplicate?: (setId: string) => void;
};

const sharedSetColumns = new Set<VehicleTableColumn>([
	"manufacturer", "articleNumber", "name", "gauge", "epoch", "railwayCompany", "category", "gattung",
	"ean", "productionPeriod", "listPrice", "acquisitionType", "acquiredFrom", "purchasePrice", "purchaseDate",
	"storageLocation", "condition", "packaging"
]);

export function VehicleSetInventoryRow({
	group,
	columns,
	collapsed,
	selectedVehicleIDs,
	onToggleCollapsed,
	onToggleSelection,
	onOpen,
	onEdit,
	onDuplicate
}: VehicleSetInventoryRowProps) {
	const { t } = useI18n();
	const checkboxRef = useRef<HTMLInputElement | null>(null);
	const memberIDs = group.members.map((member) => member.id);
	const selectedCount = memberIDs.filter((id) => selectedVehicleIDs.has(id)).length;
	const setPreviewImage = group.set.mainImage;

	useEffect(() => {
		if (checkboxRef.current) {
			checkboxRef.current.indeterminate = selectedCount > 0 && selectedCount < memberIDs.length;
		}
	}, [memberIDs.length, selectedCount]);

	const cell = (column: VehicleTableColumn) => {
		if (column === "type" || column === "image") {
			return (
				<div className="vehicle-set-type-cell">
					{setPreviewImage
						? <img className="inventory-thumb" src={setPreviewImage.thumbnailUrl || setPreviewImage.url} alt="" />
						: <div className="image-placeholder inventory-image-placeholder"
							aria-label={t("exhibition.noPreview")} title={t("exhibition.noPreview")}>
							<ImageOff size={18} strokeWidth={1.7} aria-hidden="true" />
						</div>}
				</div>
			);
		}
		if (column === "inventoryNumber") return <strong>{group.set.inventoryNumber}</strong>;
		if (column === "name") {
			return (
				<div className="vehicle-set-name-cell">
					<button type="button" className="inventory-name-link" onClick={() => onOpen(group.id)}>{group.set.name}</button>
					<small>{t("vehicles.set.memberCount", { count: group.totalMemberCount })}
						{group.visibleMemberCount !== group.totalMemberCount
							? ` · ${t("vehicles.set.visibleCount", { count: group.visibleMemberCount })}` : ""}</small>
					{(group.set.purchaseDate || group.set.purchasePrice) && (
						<small>{[group.set.purchaseDate, group.set.purchasePrice].filter(Boolean).join(" · ")}</small>
					)}
				</div>
			);
		}
		if (column === "category" || column === "gattung") {
			const values = [...new Set(group.members.map((member) => member[column]).filter(Boolean))];
			if (values.length > 1) return t("vehicles.set.mixed");
			return values[0] || "–";
		}
		if (sharedSetColumns.has(column)) {
			const setValue = (group.set as unknown as Record<string, unknown>)[column];
			const memberValue = (group.members[0] as unknown as Record<string, unknown> | undefined)?.[column];
			const value = setValue ?? memberValue;
			return typeof value === "string" || typeof value === "number" ? String(value) || "–" : "–";
		}
		return null;
	};

	return (
		<tr className="vehicle-set-inventory-row">
			<td className="select-cell">
				<label className="table-select-field" title={t("vehicles.report.selectVehicle")}>
					<input ref={checkboxRef} type="checkbox" checked={memberIDs.length > 0 && selectedCount === memberIDs.length}
						onChange={() => onToggleSelection(memberIDs)}
						aria-label={`${group.set.inventoryNumber} ${t("vehicles.report.selectVehicle")}`} />
				</label>
				<button type="button" className="icon-button vehicle-set-hierarchy-toggle" onClick={onToggleCollapsed}
					aria-expanded={!collapsed} aria-label={collapsed ? t("vehicles.set.expand") : t("vehicles.set.collapse")}>
					{collapsed ? <ChevronDown size={15} /> : <ChevronUp size={15} />}
				</button>
			</td>
			{columns.map((column) => (
				<td key={column} className={`vehicle-column-${column}`}>{cell(column)}</td>
			))}
			<td className="table-fill-cell" aria-hidden="true" />
			<td className="actions-cell">
				<div className="table-actions">
					<button type="button" className="icon-button" onClick={() => onOpen(group.id)} aria-label={t("exhibition.view")}><Eye size={16} /></button>
					{onEdit && <button type="button" className="icon-button" onClick={() => onEdit(group.id)} aria-label={t("common.edit")}><Pencil size={16} /></button>}
					{onDuplicate && <button type="button" className="icon-button" onClick={() => onDuplicate(group.id)} aria-label={t("common.duplicate")}><Copy size={16} /></button>}
				</div>
			</td>
		</tr>
	);
}
