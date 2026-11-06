import { useState } from "react";
import type { Vehicle } from "../../shared/api";
import { vehicleImagesToPending, type PendingArticleImage } from "./vehicleTransforms";

export function useVehicleImageDraftController() {
  const [pendingImages, setPendingImages] = useState<PendingArticleImage[]>([]);
  const [previewImage, setPreviewImage] = useState<PendingArticleImage | null>(null);
  const reset = (clearPreview: boolean) => {
    setPendingImages([]);
    if (clearPreview) setPreviewImage(null);
  };
  const loadDetail = (detail: Vehicle) => setPendingImages(vehicleImagesToPending(detail));

  const addImages = (images: PendingArticleImage[]) => {
    setPendingImages((current) => {
      const existing = new Set(current.map((image) => image.url));
      const next = [...current, ...images.filter((image) => !existing.has(image.url))];
      if (!next.some((image) => image.isPrimary) && next.length > 0) {
        next[0] = { ...next[0], isPrimary: true };
      }
      return next;
    });
  };

  const setPrimaryPendingImage = (id: string) => {
    setPendingImages((current) => current.map((image) => ({ ...image, isPrimary: image.id === id })));
  };

  const updatePendingImageTitle = (id: string, title: string) => {
    setPendingImages((current) => current.map((image) => (image.id === id ? { ...image, title } : image)));
  };

  const updatePendingImageMaintenance = (id: string, maintenanceId: string) => {
    setPendingImages((current) => current.map((image) => (
      image.id === id ? { ...image, maintenanceId } : image
    )));
  };

  const movePendingImage = (id: string, direction: -1 | 1) => {
    setPendingImages((current) => {
      const index = current.findIndex((image) => image.id === id);
      const target = index + direction;
      if (index < 0 || target < 0 || target >= current.length) return current;
      const next = [...current];
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  };

  const removeFromState = (id: string) => {
    setPendingImages((current) => {
      const next = current.filter((entry) => entry.id !== id);
      if (next.length > 0 && !next.some((entry) => entry.isPrimary)) {
        next[0] = { ...next[0], isPrimary: true };
      }
      return next;
    });
  };

  return {
    state: { pendingImages, previewImage },
    setters: { setPendingImages, setPreviewImage },
    commands: {
      reset, loadDetail, addImages, setPrimaryPendingImage, updatePendingImageTitle,
      updatePendingImageMaintenance, movePendingImage, removeFromState
    }
  };
}
