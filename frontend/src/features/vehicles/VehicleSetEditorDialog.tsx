import { useEffect, useState, type FormEvent, type ReactNode } from "react";
import { X } from "lucide-react";

import { api, type CreateVehicleRequest, type MasterDataEntry, type VehicleSet } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { VehicleOwnershipFields } from "./VehicleFormFields";
import { VehicleSetMainImageEditor } from "./VehicleSetMainImageEditor";
import { vehicleSetInputFromForm } from "./VehicleCreateWizard";
import { emptyOptions, emptyVehicle, gattungenForCategory, type MasterDataOptions } from "./vehicleViewModel";

type VehicleSetEditorDialogProps = {
	setId: string;
	options?: MasterDataOptions;
	onClose: () => void;
	onUpdated: (set: VehicleSet) => void;
	onImageChanged?: (set: VehicleSet) => void;
};

type SetEditorTab = "general" | "upload";

function formFromSet(set: VehicleSet): CreateVehicleRequest {
	return {
		...emptyVehicle,
		...set,
		inventoryNumber: ""
	};
}

function selectOptions(entries: MasterDataEntry[], currentValue: string, emptyLabel: string): ReactNode {
	const hasCurrentValue = entries.some((entry) => entry.label === currentValue);
	return (
		<>
			<option value="">{emptyLabel}</option>
			{currentValue && !hasCurrentValue && <option value={currentValue}>{currentValue}</option>}
			{entries.map((entry) => (
				<option key={entry.id} value={entry.label} disabled={!entry.active}>
					{entry.label}
				</option>
			))}
		</>
	);
}

export function VehicleSetEditorDialog({
	setId,
	options = emptyOptions,
	onClose,
	onUpdated,
	onImageChanged
}: VehicleSetEditorDialogProps) {
	const { t } = useI18n();
	const [form, setForm] = useState<CreateVehicleRequest>(emptyVehicle);
	const [vehicleSet, setVehicleSet] = useState<VehicleSet | null>(null);
	const [loading, setLoading] = useState(true);
	const [saving, setSaving] = useState(false);
	const [error, setError] = useState("");
	const [activeTab, setActiveTab] = useState<SetEditorTab>("general");
	const filteredGattungen = gattungenForCategory(options, form.category);
	const update = (patch: Partial<CreateVehicleRequest>) => {
		setForm((current) => ({ ...current, ...patch }));
	};

	useEffect(() => {
		let active = true;
		api.vehicleSet(setId)
			.then((set) => { if (active) { setVehicleSet(set); setForm(formFromSet(set)); } })
			.catch((reason: Error) => { if (active) setError(reason.message); })
			.finally(() => { if (active) setLoading(false); });
		return () => { active = false; };
	}, [setId]);

	const submit = async (event: FormEvent<HTMLFormElement>) => {
		event.preventDefault();
		setSaving(true);
		setError("");
		try {
			const updated = await api.updateVehicleSet(setId, vehicleSetInputFromForm(form));
			setVehicleSet(updated);
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
					<button type="button" className="icon-button" onClick={onClose} aria-label={t("vehicles.close")}>
						<X size={18} />
					</button>
				</header>
				<nav className="modal-tabs vehicle-set-editor-tabs" role="tablist"
					aria-label={t("vehicles.set.tabs.label")}>
					<button id="vehicle-set-tab-general" type="button" role="tab"
						className={activeTab === "general" ? "active" : ""}
						aria-selected={activeTab === "general"} aria-controls="vehicle-set-panel-general"
						onClick={() => setActiveTab("general")}>{t("vehicles.set.tabs.general")}</button>
					<button id="vehicle-set-tab-upload" type="button" role="tab"
						className={activeTab === "upload" ? "active" : ""}
						aria-selected={activeTab === "upload"} aria-controls="vehicle-set-panel-upload"
						onClick={() => setActiveTab("upload")}>{t("vehicles.set.tabs.upload")}</button>
				</nav>
				<div className="modal-body vehicle-set-dialog-body">
					{loading && <p>{t("vehicles.set.loading")}</p>}
					{error && <p className="error-text" role="alert">{error}</p>}
					{!loading && activeTab === "general" && (
						<div id="vehicle-set-panel-general" role="tabpanel" aria-labelledby="vehicle-set-tab-general"
							className="vehicle-create-detail-groups vehicle-form">
							<details open>
								<summary>{t("vehicles.wizard.basicData")}</summary>
								<div className="form-row">
									<label>{t("vehicle.field.name")}<input value={form.name}
										onChange={(event) => update({ name: event.target.value })} required /></label>
									<label>{t("vehicle.field.manufacturer")}<AppSelect value={form.manufacturer}
										onChange={(event) => update({ manufacturer: event.target.value })} required>
										{selectOptions(options.manufacturers, form.manufacturer, t("vehicles.select.placeholder"))}
									</AppSelect></label>
								</div>
								<div className="form-row">
									<label>{t("vehicle.field.articleNumber")}<input value={form.articleNumber || ""}
										onChange={(event) => update({ articleNumber: event.target.value })} /></label>
									<label>{t("vehicle.field.articleSourceUrl")}<input value={form.articleSourceUrl || ""}
										onChange={(event) => update({ articleSourceUrl: event.target.value })} /></label>
								</div>
								<div className="form-row three-columns">
									<label>{t("vehicle.field.gauge")}<AppSelect value={form.gauge}
										onChange={(event) => update({ gauge: event.target.value })} required>
										{selectOptions(options.gauges, form.gauge, t("vehicles.select.placeholder"))}
									</AppSelect></label>
									<label>{t("vehicle.field.epoch")}<AppSelect value={form.epoch || ""}
										onChange={(event) => update({ epoch: event.target.value })}>
										{selectOptions(options.epochs, form.epoch || "", t("vehicles.select.placeholder"))}
									</AppSelect></label>
									<label>{t("vehicle.field.railwayCompany")}<AppSelect value={form.railwayCompany || ""}
										onChange={(event) => update({ railwayCompany: event.target.value })}>
										{selectOptions(options.railwayCompanies, form.railwayCompany || "", t("vehicles.select.placeholder"))}
									</AppSelect></label>
								</div>
								<div className="form-row">
									<label>{t("vehicle.field.category")}<AppSelect value={form.category || ""}
										onChange={(event) => update({ category: event.target.value, gattung: "" })} required>
										{selectOptions(options.categories, form.category || "", t("vehicles.select.placeholder"))}
									</AppSelect></label>
									<label>{t("vehicle.field.gattung")}<AppSelect value={form.gattung || ""}
										onChange={(event) => update({ gattung: event.target.value })} required>
										{selectOptions(filteredGattungen, form.gattung || "", t("vehicles.select.placeholder"))}
									</AppSelect></label>
								</div>
								<div className="form-row three-columns">
									<label>{t("vehicle.field.ean")}<input value={form.ean || ""}
										onChange={(event) => update({ ean: event.target.value })} /></label>
									<label>{t("vehicle.field.productionPeriod")}<input value={form.productionPeriod || ""}
										onChange={(event) => update({ productionPeriod: event.target.value })} /></label>
									<label>{t("vehicle.field.listPrice")}<input value={form.listPrice || ""}
										onChange={(event) => update({ listPrice: event.target.value })} /></label>
								</div>
								<label>{t("vehicle.field.description")}<textarea value={form.description || ""}
									onChange={(event) => update({ description: event.target.value })} rows={3} /></label>
							</details>
							<details open>
								<summary>{t("vehicles.wizard.acquisitionAndStock")}</summary>
								<VehicleOwnershipFields form={form} readonly={false} hideAdditionalInfo update={update} />
							</details>
						</div>
					)}
					{!loading && activeTab === "upload" && vehicleSet && (
						<div id="vehicle-set-panel-upload" role="tabpanel" aria-labelledby="vehicle-set-tab-upload">
							<VehicleSetMainImageEditor vehicleSet={vehicleSet}
								onChange={(updated) => {
									setVehicleSet(updated);
									onImageChanged?.(updated);
								}} onError={setError} />
						</div>
					)}
				</div>
				<footer className="modal-actions">
					{activeTab === "general" ? (
						<>
							<button type="button" className="secondary-button" onClick={onClose}>{t("common.cancel")}</button>
							<button type="submit" className="primary-button"
								disabled={loading || saving}>{t("common.save")}</button>
						</>
					) : (
						<button type="button" className="secondary-button" onClick={onClose}>{t("common.close")}</button>
					)}
				</footer>
			</form>
		</div>
	);
}
