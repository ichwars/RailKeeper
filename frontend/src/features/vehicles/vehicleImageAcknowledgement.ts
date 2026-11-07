import type { Vehicle, VehicleImage } from "../../shared/api";
import { reconcileImageDraft } from "./vehicleDraftRefresh";
import { uploadedImageToPending, type PendingArticleImage } from "./vehicleTransforms";

// This is a metadata baseline for reconciliation, not a source of revision tokens for writes.
export function acknowledgeVehicleImage(
  current: PendingArticleImage[], before: Vehicle, stored: VehicleImage, replacedId?: string
) {
  const images = (before.images || []).filter(image => image.id !== stored.id)
    .map(image => stored.isPrimary ? { ...image, isPrimary: false } : image);
  images.push(stored);
  images.sort((a, b) => Number(b.isPrimary) - Number(a.isPrimary) || a.sortOrder - b.sortOrder ||
    a.createdAt.localeCompare(b.createdAt));
  let pending = current;
  if (replacedId) {
    const primary = current.find(image => image.isPrimary)?.id;
    pending = current.map(image => image.id === replacedId
      ? { ...uploadedImageToPending(stored), isPrimary: primary ? primary === replacedId : stored.isPrimary } : image);
  }
  return { images: reconcileImageDraft(pending, before, { ...before, images }), baselineImages: images };
}
