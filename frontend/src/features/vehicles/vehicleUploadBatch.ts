import { ApiError, type Vehicle } from "../../shared/api";

export const maxVehicleUploadFiles = 20;
export type VehicleUploadKind = "image" | "attachment" | "decoder";
export type VehicleUploadStatus = "pending" | "uploading" | "saved" | "failed" | "unknown" | "skipped";
export type VehicleUploadItem = {
  key: string;
  file: File;
  status: VehicleUploadStatus;
  storedId?: string;
  error?: string;
};
export type VehicleUploadMetadata = {
  category?: string;
  description?: string;
  maintenanceId?: string;
  decoderProfile?: string;
  isPrimary?: boolean;
  imagesRevision?: string;
};
export type VehicleUploadBatch = {
  id: string;
  kind: VehicleUploadKind;
  vehicleId: string;
  vehicleLabel: string;
  metadata: VehicleUploadMetadata;
  items: VehicleUploadItem[];
  phase: "confirm" | "uploading" | "result";
  needsReview: boolean;
  notice?: string;
};

let nextBatch = 0;

export function createVehicleUploadBatch(
  kind: VehicleUploadKind, vehicle: Vehicle, files: File[], metadata: VehicleUploadMetadata = {}
): VehicleUploadBatch {
  if (!files.length || files.length > maxVehicleUploadFiles) throw new RangeError("upload batch size");
  const id = `media-upload-${++nextBatch}`;
  return {
    id, kind, vehicleId: vehicle.id, vehicleLabel: `${vehicle.inventoryNumber} · ${vehicle.name}`,
    metadata: { ...metadata }, phase: "confirm", needsReview: false,
    items: files.map((file, index) => ({ key: `${id}-${index}`, file, status: "pending" }))
  };
}

export const retryableUploadItem = (item: VehicleUploadItem) => item.status === "pending" || item.status === "failed";
export const validImagesRevision = (value: unknown): value is string => typeof value === "string" && /^[0-9a-f]{64}$/.test(value);

export function vehicleUploadCounts(batch: VehicleUploadBatch) {
  const counts = { total: batch.items.length, pending: 0, uploading: 0, saved: 0, failed: 0, unknown: 0, skipped: 0 };
  for (const item of batch.items) counts[item.status]++;
  return counts;
}

// A missing response can follow a committed write. Only definite rejections are safe to retry.
export function vehicleUploadFailure(error: unknown): "failed" | "unknown" {
  return error instanceof ApiError && error.status >= 400 && error.status < 500 && ![408, 499].includes(error.status)
    ? "failed" : "unknown";
}
