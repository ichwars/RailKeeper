import { api, ApiError, type Vehicle, type VehicleImage } from "../../shared/api";
import { validImagesRevision } from "./vehicleUploadBatch";
import type { PendingArticleImage } from "./vehicleTransforms";

const isRemotePendingImage = (image: PendingArticleImage) => !image.persisted && /^https?:\/\//i.test(image.url);

export function partitionVehicleImages(pendingImages: PendingArticleImage[], previous?: Vehicle | null) {
  // Imported images use zero as the append sentinel. Reserve positive positions when a
  // draft mixes existing images with imports, so the first position is explicit too.
  const positions = new Map(pendingImages.map((image, index) => [image.id, index + 1]));
  const remoteImages = pendingImages.filter(isRemotePendingImage).map(image => ({ image, sortOrder: positions.get(image.id)! }));
  const localImages = pendingImages.filter((image) => !isRemotePendingImage(image));
  const stored = [...(previous?.images || [])].sort((a, b) => a.sortOrder - b.sortOrder);
  const unchangedOrder = !remoteImages.length && localImages.length === stored.length && localImages.every((image, index) =>
    image.persisted && image.id === stored[index]?.id);
  const images = localImages.map((image, index) => ({
    id: image.persisted ? image.id : undefined,
    url: image.url,
    title: image.title,
    sourceUrl: image.source,
    maintenanceId: image.maintenanceId || "",
    isPrimary: Boolean(image.isPrimary),
    sortOrder: remoteImages.length ? positions.get(image.id)! : unchangedOrder ? stored[index].sortOrder : index
  }));
  return { images, remoteImages };
}

export async function importPendingVehicleImages(
  vehicle: Vehicle, remoteImages: { image: PendingArticleImage; sortOrder: number }[],
  onImported?: (pending: PendingArticleImage, stored: VehicleImage) => void,
  isCurrent: () => boolean = () => true
) {
  let imagesRevision = vehicle.imagesRevision;
  if (remoteImages.length && !validImagesRevision(imagesRevision)) {
    throw new ApiError("Image comparison version missing", "image_revision_missing", 409);
  }
  for (const { image, sortOrder } of remoteImages) {
    if (!isCurrent()) return vehicle;
    const stored = await api.importVehicleImageFromUrl(vehicle.id, {
      url: image.url,
      title: image.title || "",
      sourceUrl: image.source || image.url,
      maintenanceId: image.maintenanceId || "",
      isPrimary: Boolean(image.isPrimary),
      sortOrder,
      expectedImagesRevision: imagesRevision
    });
    if (!isCurrent()) return vehicle;
    onImported?.(image, stored);
    if (!validImagesRevision(stored.imagesRevision)) {
      throw new ApiError("Image saved, comparison version missing", "image_revision_missing", 409);
    }
    imagesRevision = stored.imagesRevision;
  }
  if (remoteImages.length > 0) vehicle = await api.vehicle(vehicle.id);

  return vehicle;
}
