import type { CreateVehicleRequest, Vehicle, VehicleImageInput, VehicleSet } from "../../shared/api";
import { vehicleToForm } from "./vehicleTransforms";
import { emptyVehicle } from "./vehicleViewModel";

export type VehicleCreatePrefill = {
	kind: "set";
	shared: CreateVehicleRequest;
	members: CreateVehicleRequest[];
};

function reusableRemoteImages(vehicle: Vehicle): VehicleImageInput[] {
	return (vehicle.images || [])
		.map((image) => ({ image, reusableURL: /^https?:\/\//i.test(image.url) ? image.url : image.sourceUrl || "" }))
		.filter(({ reusableURL }) => /^https?:\/\//i.test(reusableURL))
		.map(({ image, reusableURL }, index) => ({
			url: reusableURL,
			title: image.title || "",
			sourceUrl: image.sourceUrl || reusableURL,
			isPrimary: Boolean(image.isPrimary),
			sortOrder: index
		}));
}

export function vehicleSetDuplicatePrefill(set: VehicleSet): VehicleCreatePrefill {
	const shared: CreateVehicleRequest = {
		...emptyVehicle,
		inventoryNumber: "",
		name: set.name,
		manufacturer: set.manufacturer,
		articleNumber: set.articleNumber || "",
		articleSourceUrl: set.articleSourceUrl || "",
		gauge: set.gauge,
		epoch: set.epoch || "",
		railwayCompany: set.railwayCompany || "",
		category: set.category || "",
		gattung: set.gattung || "",
		description: set.description || "",
		ean: set.ean || "",
		productionPeriod: set.productionPeriod || "",
		listPrice: set.listPrice || "",
		acquisitionType: set.acquisitionType || "",
		acquiredFrom: set.acquiredFrom || "",
		purchasePrice: set.purchasePrice || "",
		purchaseDate: set.purchaseDate || "",
		storageLocation: set.storageLocation || "",
		storageDetails: set.storageDetails || "",
		condition: set.condition || "",
		conditionDetails: set.conditionDetails || "",
		packaging: set.packaging || "",
		images: []
	};
	return {
		kind: "set",
		shared,
		members: set.members.map((member) => ({
			...vehicleToForm(member),
			inventoryNumber: "",
			images: reusableRemoteImages(member)
		}))
	};
}
