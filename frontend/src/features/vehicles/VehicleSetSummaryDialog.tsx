import { useEffect, useState } from "react";
import { Copy, Pencil, X } from "lucide-react";

import { api, type VehicleSet } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { vehicleSetDuplicatePrefill, type VehicleCreatePrefill } from "./vehicleSetDuplicate";

type VehicleSetDialogProps = {
	setId: string;
	canEdit: boolean;
	onClose: () => void;
	onUpdated: (set: VehicleSet) => void;
	onDuplicate: (prefill: VehicleCreatePrefill) => void;
	onEdit?: (setId: string) => void;
	onOpenVehicle?: (vehicleId: string) => void;
};

export function VehicleSetSummaryDialog({
	setId,
	canEdit,
	onClose,
	onDuplicate,
	onEdit,
	onOpenVehicle
}: VehicleSetDialogProps) {
	const { t } = useI18n();
	const [vehicleSet, setVehicleSet] = useState<VehicleSet | null>(null);
	const [error, setError] = useState("");

	useEffect(() => {
		let active = true;
		setVehicleSet(null);
		setError("");
		api.vehicleSet(setId)
			.then((loaded) => {
				if (active) setVehicleSet(loaded);
			})
			.catch((reason: Error) => {
				if (active) setError(reason.message);
			});
		return () => { active = false; };
	}, [setId]);

	return (
		<div className="modal-layer" role="dialog" aria-modal="true" aria-label={t("vehicles.set.viewTitle")}>
			<section className="vehicle-modal vehicle-set-summary-dialog">
				<header className="modal-head">
					<div><h2>{t("vehicles.set.viewTitle")}</h2></div>
					<button type="button" className="icon-button" onClick={onClose} aria-label={t("vehicles.close")}>
						<X size={18} />
					</button>
				</header>
				<div className="modal-body vehicle-set-dialog-body">
					{!vehicleSet && !error && <p>{t("vehicles.set.loading")}</p>}
					{error && <p className="error-text" role="alert">{error}</p>}
					{vehicleSet && (
						<>
						<dl className="detail-grid">
							<div><dt>{t("vehicle.field.inventoryNumber")}</dt><dd>{vehicleSet.inventoryNumber}</dd></div>
							<div><dt>{t("vehicle.field.name")}</dt><dd>{vehicleSet.name}</dd></div>
							<div><dt>{t("vehicle.field.manufacturer")}</dt><dd>{vehicleSet.manufacturer}</dd></div>
							<div><dt>{t("vehicle.field.articleNumber")}</dt><dd>{vehicleSet.articleNumber || "–"}</dd></div>
							<div><dt>{t("vehicle.field.gauge")}</dt><dd>{vehicleSet.gauge}</dd></div>
							<div><dt>{t("vehicle.field.epoch")}</dt><dd>{vehicleSet.epoch || "–"}</dd></div>
							<div><dt>{t("vehicle.field.purchaseDate")}</dt><dd>{vehicleSet.purchaseDate || "–"}</dd></div>
							<div><dt>{t("vehicle.field.purchasePrice")}</dt><dd>{vehicleSet.purchasePrice || "–"}</dd></div>
						</dl>
						<h3>{t("vehicles.set.members")}</h3>
						<ul className="vehicle-set-member-list">
							{vehicleSet.members.map((member) => (
								<li key={member.id}>
									<button type="button" className="text-button vehicle-set-member-button" onClick={() => onOpenVehicle?.(member.id)}>
										<span>{member.name}</span>
										<small>{member.inventoryNumber}</small>
									</button>
									<span>{[member.category, member.gattung].filter(Boolean).join(" · ") || "–"}</span>
								</li>
							))}
						</ul>
						</>
					)}
				</div>
				{vehicleSet && (
					<footer className="modal-actions">
							{canEdit && onEdit && (
								<button type="button" className="secondary-button" onClick={() => onEdit(vehicleSet.id)}><Pencil size={16} />{t("common.edit")}</button>
							)}
							{canEdit && (
								<button type="button" className="secondary-button" onClick={() => onDuplicate(vehicleSetDuplicatePrefill(vehicleSet))}>
									<Copy size={16} />{t("common.duplicate")}
								</button>
							)}
							<button type="button" className="primary-button" onClick={onClose}>{t("common.close")}</button>
					</footer>
				)}
			</section>
		</div>
	);
}
