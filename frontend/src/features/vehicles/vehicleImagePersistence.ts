import { api, type Vehicle } from "../../shared/api";
import type { PendingArticleImage } from "./vehicleTransforms";

const isRemotePendingImage = (image: PendingArticleImage) => !image.persisted && /^https?:\/\//i.test(image.url);

export function partitionVehicleImages(pendingImages: PendingArticleImage[]) {
  const remoteImages = pendingImages.filter(isRemotePendingImage);
  const images = pendingImages.filter((image) => !isRemotePendingImage(image)).map((image, index) => ({
    id: image.persisted ? image.id : undefined,
    url: image.url,
    title: image.title,
    sourceUrl: image.source,
    maintenanceId: image.maintenanceId || "",
    isPrimary: Boolean(image.isPrimary),
    sortOrder: index
  }));
  return { images, remoteImages };
}

export async function importPendingVehicleImages(vehicle: Vehicle, remoteImages: PendingArticleImage[], offset: number) {
  for (const [imageIndex, image] of remoteImages.entries()) {
    await api.importVehicleImageFromUrl(vehicle.id, {
      url: image.url,
      title: image.title || "",
      sourceUrl: image.source || image.url,
      maintenanceId: image.maintenanceId || "",
      isPrimary: Boolean(image.isPrimary),
      sortOrder: offset + imageIndex
    });
  }
  if (remoteImages.length > 0) vehicle = await api.vehicle(vehicle.id);

  return vehicle;
}
