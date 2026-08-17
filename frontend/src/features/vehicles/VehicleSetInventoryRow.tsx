import { useEffect, useRef } from "react";
import { ChevronDown, ChevronUp, Copy, Eye, ImageOff, Pencil } from "lucide-react";

import { useI18n } from "../../shared/i18n";
import type { VehicleInventorySetGroup } from "./vehicleSetGroups";
import type { VehicleTableColumn } from "./vehicleTableColumns";
import { previewImageUrl, primaryImage } from "./vehicleTransforms";

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
	const setPreviewImage = group.members.map((member) => primaryImage(member.images)).find(Boolean);

	useEffect(() => {
		if (checkboxRef.current) {
			checkboxRef.current.indeterminate = selectedCount > 0 && selectedCount < memberIDs.length;
		}
	}, [memberIDs.length, selectedCount]);

	const cell = (column: VehicleTableColumn) => {
		if (column === "type") {
			return (
				<div className="vehicle-set-type-cell">
					<button type="button" className="icon-button" onClick={onToggleCollapsed}
						aria-expanded={!collapsed} aria-label={collapsed ? t("vehicles.set.expand") : t("vehicles.set.collapse")}>
						{collapsed ? <ChevronDown size={15} /> : <ChevronUp size={15} />}
					</button>
					{setPreviewImage
						? <img className="inventory-thumb" src={previewImageUrl(setPreviewImage)} alt="" />
						: <div className="image-placeholder inventory-image-placeholder"
							aria-label={t("exhibition.noPreview")} title={t("exhibition.noPreview")}>
							<ImageOff size={18} strokeWidth={1.7} aria-hidden="true" />
						</div>}
				</div>
			);
		}
		if (column === "inventoryNumber") return <strong>{group.set.inventoryNumber}</strong>;
		if (column === "manufacturer") return group.set.manufacturer;
		if (column === "articleNumber") return group.set.articleNumber || "–";
		if (column === "gauge") return group.set.gauge;
		if (column === "epoch") return group.set.epoch || "–";
		if (column === "acquisitionType") return group.set.acquisitionType || "–";
		if (column === "purchaseDate") return group.set.purchaseDate || "–";
		if (column === "purchasePrice") return group.set.purchasePrice || "–";
		if (column === "condition") return group.set.condition || "–";
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
			</td>
			{columns.map((column) => (
				<td key={column} className={`vehicle-column-${column}`}>{cell(column)}</td>
			))}
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
