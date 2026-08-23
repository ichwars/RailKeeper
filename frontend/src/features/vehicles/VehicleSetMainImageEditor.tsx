import { Check, ImageOff, Trash2, Upload } from "lucide-react";
import { useState, type ChangeEvent } from "react";

import { api, type VehicleImage, type VehicleSet } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

type VehicleSetMainImageEditorProps = {
	vehicleSet: VehicleSet;
	onChange: (vehicleSet: VehicleSet) => void;
	onError: (message: string) => void;
};

type MemberImage = {
	image: VehicleImage;
	memberName: string;
};

export function VehicleSetMainImageEditor({ vehicleSet, onChange, onError }: VehicleSetMainImageEditorProps) {
	const { t } = useI18n();
	const [busy, setBusy] = useState(false);
	const memberImages: MemberImage[] = vehicleSet.members.flatMap((member) =>
		(member.images ?? []).map((image) => ({ image, memberName: member.name }))
	);
	const sourceKey = vehicleSet.mainImage?.source ?? "empty";

	const run = async (action: () => Promise<VehicleSet>) => {
		setBusy(true);
		onError("");
		try {
			onChange(await action());
		} catch (reason) {
			onError(reason instanceof Error ? reason.message : t("vehicles.set.mainImage.failed"));
		} finally {
			setBusy(false);
		}
	};

	const upload = (event: ChangeEvent<HTMLInputElement>) => {
		const file = event.target.files?.[0];
		if (file) void run(() => api.uploadVehicleSetImage(vehicleSet.id, file));
		event.target.value = "";
	};

	const removeDedicated = async () => {
		setBusy(true);
		onError("");
		try {
			await api.deleteVehicleSetImage(vehicleSet.id);
			onChange(await api.vehicleSet(vehicleSet.id));
		} catch (reason) {
			onError(reason instanceof Error ? reason.message : t("vehicles.set.mainImage.failed"));
		} finally {
			setBusy(false);
		}
	};

	return (
		<section className="vehicle-set-main-image-section" aria-labelledby="vehicle-set-main-image-title">
			<div className="vehicle-set-main-image-heading">
				<div>
					<h3 id="vehicle-set-main-image-title">{t("vehicles.set.mainImage.title")}</h3>
					<p>{t("vehicles.set.mainImage.description")}</p>
				</div>
				{vehicleSet.mainImage && (
					<span className="vehicle-set-main-image-status">
						<Check size={14} aria-hidden="true" />{t("vehicles.set.mainImage.active")}
					</span>
				)}
			</div>

			<div className="vehicle-set-main-image-stage">
				{vehicleSet.mainImage ? (
					<img className="vehicle-set-main-image-preview"
						src={vehicleSet.mainImage.thumbnailUrl || vehicleSet.mainImage.url} alt={vehicleSet.name} />
				) : (
					<div className="image-placeholder"><ImageOff size={32} aria-hidden="true" /></div>
				)}
			</div>

			<div className="vehicle-set-main-image-meta">
				<div>
					<strong>{vehicleSet.name}</strong>
					<span>{t(`vehicles.set.mainImage.source.${sourceKey}`)}</span>
				</div>
				<div className="vehicle-set-main-image-actions">
					<label className="primary-button vehicle-set-image-upload-button">
						<Upload size={16} />{t("vehicles.set.mainImage.upload")}
						<input type="file" accept="image/jpeg,image/png,image/webp" onChange={upload} disabled={busy} />
					</label>
					<button type="button" className="secondary-button" disabled={busy}
						onClick={() => void run(() => api.setVehicleSetMainImage(vehicleSet.id, { mode: "automatic" }))}>
						{t("vehicles.set.mainImage.automatic")}
					</button>
					{vehicleSet.dedicatedImage && vehicleSet.mainImage?.source !== "dedicated" && (
						<button type="button" className="secondary-button" disabled={busy}
							onClick={() => void run(() => api.setVehicleSetMainImage(vehicleSet.id, { mode: "dedicated" }))}>
							{t("vehicles.set.mainImage.useDedicated")}
						</button>
					)}
					{vehicleSet.dedicatedImage && (
						<button type="button" className="danger-button" disabled={busy}
							onClick={() => void removeDedicated()}>
							<Trash2 size={16} />{t("vehicles.set.mainImage.removeDedicated")}
						</button>
					)}
				</div>
			</div>

			<div className="vehicle-set-member-image-section">
				<div className="vehicle-set-member-image-heading">
					<h4>{t("vehicles.set.mainImage.memberGallery")}</h4>
					<span>{t("vehicles.set.mainImage.availableCount", { count: memberImages.length })}</span>
				</div>
				{memberImages.length === 0 ? <p className="muted">{t("vehicles.set.mainImage.noMemberImages")}</p> : (
					<div className="vehicle-set-member-image-strip">
						{memberImages.map(({ image, memberName }) => {
							const selected = vehicleSet.mainImage?.imageId === image.id
								|| vehicleSet.selectedMemberImageId === image.id;
							return (
								<article key={image.id} className={selected ? "selected" : ""}>
									<div className="vehicle-set-member-image-preview">
										<img src={image.thumbnailUrl || image.url} alt={image.title || memberName} />
										{selected && <span aria-label={t("vehicles.set.mainImage.active")}>
											<Check size={13} aria-hidden="true" />
										</span>}
									</div>
									<div><strong>{memberName}</strong><small>{image.title || image.fileName}</small></div>
									<button type="button" className="secondary-button" disabled={busy}
										onClick={() => void run(() => api.setVehicleSetMainImage(vehicleSet.id, {
											mode: "member", memberImageId: image.id
										}))}>
										{t("vehicles.set.mainImage.use")}
									</button>
								</article>
							);
						})}
					</div>
				)}
			</div>
		</section>
	);
}
