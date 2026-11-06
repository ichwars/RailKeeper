import { api, ApiError, type Vehicle, type VehicleImage } from "../../shared/api";
import { validImagesRevision } from "./vehicleUploadBatch";
import type { PendingArticleImage } from "./vehicleTransforms";

const isRemotePendingImage = (image: PendingArticleImage) => !image.persisted && /^https?:\/\//i.test(image.url);

export function partitionVehicleImages(pendingImages: PendingArticleImage[], previous?: Vehicle | null) {
  const remoteImages = pendingImages.filter(isRemotePendingImage);
  const localImages = pendingImages.filter((image) => !isRemotePendingImage(image));
  const stored = previous?.images || [];
  const unchangedOrder = localImages.length === stored.length && localImages.every((image, index) =>
    image.persisted && image.id === stored[index]?.id);
  const images = localImages.map((image, index) => ({
    id: image.persisted ? image.id : undefined,
    url: image.url,
    title: image.title,
    sourceUrl: image.source,
    maintenanceId: image.maintenanceId || "",
    isPrimary: Boolean(image.isPrimary),
    sortOrder: unchangedOrder ? stored[index].sortOrder : index
  }));
  return { images, remoteImages };
}

export async function importPendingVehicleImages(
  vehicle: Vehicle, remoteImages: PendingArticleImage[], offset: number,
  onImported?: (pending: PendingArticleImage, stored: VehicleImage) => void
) {
  let imagesRevision = vehicle.imagesRevision;
  if (remoteImages.length && !validImagesRevision(imagesRevision)) {
    throw new ApiError("Image comparison version missing", "image_revision_missing", 409);
  }
  for (const [imageIndex, image] of remoteImages.entries()) {
    const stored = await api.importVehicleImageFromUrl(vehicle.id, {
      url: image.url,
      title: image.title || "",
      sourceUrl: image.source || image.url,
      maintenanceId: image.maintenanceId || "",
      isPrimary: Boolean(image.isPrimary),
      sortOrder: offset + imageIndex,
      expectedImagesRevision: imagesRevision
    });
    onImported?.(image, stored);
    if (!validImagesRevision(stored.imagesRevision)) {
      throw new ApiError("Image saved, comparison version missing", "image_revision_missing", 409);
    }
    imagesRevision = stored.imagesRevision;
  }
  if (remoteImages.length > 0) vehicle = await api.vehicle(vehicle.id);

  return vehicle;
}
