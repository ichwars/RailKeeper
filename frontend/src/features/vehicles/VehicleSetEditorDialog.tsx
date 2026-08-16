import { useEffect, useState, type FormEvent } from "react";
import { X } from "lucide-react";

import { api, type VehicleSet, type VehicleSetInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

type VehicleSetEditorDialogProps = {
	setId: string;
	onClose: () => void;
	onUpdated: (set: VehicleSet) => void;
};

const blankInput: VehicleSetInput = {
	name: "", manufacturer: "", articleNumber: "", articleSourceUrl: "", gauge: "", epoch: "",
	railwayCompany: "", category: "", gattung: "", description: "", ean: "", productionPeriod: "",
	listPrice: "", acquisitionType: "", acquiredFrom: "", purchasePrice: "", purchaseDate: "",
	storageLocation: "", storageDetails: "", condition: "", conditionDetails: "", packaging: ""
};

function inputFromSet(set: VehicleSet): VehicleSetInput {
	return Object.fromEntries(Object.keys(blankInput).map((key) => [
		key,
		set[key as keyof VehicleSet] || ""
	])) as VehicleSetInput;
}

const fields: Array<{ key: keyof VehicleSetInput; label: string; required?: boolean }> = [
	{ key: "name", label: "vehicle.field.name", required: true },
	{ key: "manufacturer", label: "vehicle.field.manufacturer", required: true },
	{ key: "articleNumber", label: "vehicle.field.articleNumber" },
	{ key: "articleSourceUrl", label: "vehicle.field.articleSourceUrl" },
	{ key: "gauge", label: "vehicle.field.gauge", required: true },
	{ key: "epoch", label: "vehicle.field.epoch" },
	{ key: "railwayCompany", label: "vehicle.field.railwayCompany" },
	{ key: "category", label: "vehicle.field.category", required: true },
	{ key: "gattung", label: "vehicle.field.gattung", required: true },
	{ key: "ean", label: "vehicle.field.ean" },
	{ key: "productionPeriod", label: "vehicle.field.productionPeriod" },
	{ key: "listPrice", label: "vehicle.field.listPrice" },
	{ key: "acquisitionType", label: "vehicle.field.acquisitionType" },
	{ key: "acquiredFrom", label: "vehicle.field.acquiredFrom" },
	{ key: "purchasePrice", label: "vehicle.field.purchasePrice" },
	{ key: "purchaseDate", label: "vehicle.field.purchaseDate" },
	{ key: "storageLocation", label: "vehicle.field.storageLocation" },
	{ key: "storageDetails", label: "vehicle.field.storageDetails" },
	{ key: "condition", label: "vehicle.field.condition" },
	{ key: "conditionDetails", label: "vehicle.field.conditionDetails" },
	{ key: "packaging", label: "vehicle.field.packaging" },
	{ key: "description", label: "vehicle.field.description" }
];

export function VehicleSetEditorDialog({ setId, onClose, onUpdated }: VehicleSetEditorDialogProps) {
	const { t } = useI18n();
	const [form, setForm] = useState<VehicleSetInput>(blankInput);
	const [loading, setLoading] = useState(true);
	const [saving, setSaving] = useState(false);
	const [error, setError] = useState("");

	useEffect(() => {
		let active = true;
		api.vehicleSet(setId)
			.then((set) => { if (active) setForm(inputFromSet(set)); })
			.catch((reason: Error) => { if (active) setError(reason.message); })
			.finally(() => { if (active) setLoading(false); });
		return () => { active = false; };
	}, [setId]);

	const submit = async (event: FormEvent<HTMLFormElement>) => {
		event.preventDefault();
		setSaving(true);
		setError("");
		try {
			const updated = await api.updateVehicleSet(setId, form);
			onUpdated(updated);
		} catch (reason) {
			setError(reason instanceof Error ? reason.message : t("vehicles.set.updateFailed"));
		} finally {
			setSaving(false);
		}
	};

	return (
		<div className="modal-layer" role="dialog" aria-modal="true" aria-label={t("vehicles.set.editTitle")}>
			<form className="vehicle-modal vehicle-set-editor-dialog" onSubmit={submit}>
				<header className="modal-head">
					<h2>{t("vehicles.set.editTitle")}</h2>
					<button type="button" className="icon-button" onClick={onClose} aria-label={t("vehicles.close")}><X size={18} /></button>
				</header>
				<div className="modal-body vehicle-form">
					{loading && <p>{t("vehicles.set.loading")}</p>}
					{error && <p className="error-text" role="alert">{error}</p>}
					{!loading && (
						<div className="form-grid">
							{fields.map((field) => (
								<label key={field.key}>
									{t(field.label)}
									<input
										value={form[field.key] || ""}
										required={field.required}
										type={field.key === "purchaseDate" ? "date" : "text"}
										onChange={(event) => setForm((current) => ({ ...current, [field.key]: event.target.value }))}
									/>
								</label>
							))}
						</div>
					)}
				</div>
				<footer className="modal-actions">
					<button type="button" onClick={onClose}>{t("common.cancel")}</button>
					<button type="submit" className="primary" disabled={loading || saving}>{t("common.save")}</button>
				</footer>
			</form>
		</div>
	);
}
