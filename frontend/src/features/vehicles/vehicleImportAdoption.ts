import { api, ApiError, type Vehicle } from "../../shared/api";
import { buildCVImportPreview } from "./cvImport";
import { guardedCVImportValues } from "./cvImportGuard";
import type { ECoSVehicleDraftPayload } from "./vehicleViewModel";

type VehicleImportCallbacks = {
  markSaved: (draft: ECoSVehicleDraftPayload, vehicleID: string) => void;
  clear: () => void;
  returnToSession: (draft: ECoSVehicleDraftPayload) => void;
  reloadVehicles: () => void;
  closeEditor: () => void;
  isCurrent?: () => boolean;
};

export async function adoptVehicleImportDraft(
  vehicle: Vehicle,
  savedDraft: ECoSVehicleDraftPayload,
  callbacks: VehicleImportCallbacks
) {
  const isCurrent = callbacks.isCurrent || (() => true);
  const stopped = () => ({ vehicle, returnedToSource: false });
  if (!isCurrent()) return stopped();
  if (!vehicle.externalMappings?.some(mapping => mapping.provider === savedDraft.externalMapping.provider
    && mapping.externalId === savedDraft.externalMapping.externalId && mapping.vehicleId === vehicle.id)) {
    await api.upsertVehicleExternalMapping(vehicle.id, savedDraft.externalMapping);
    if (!isCurrent()) return stopped();
  }
  if (savedDraft.source === "ecos" && savedDraft.returnToDigitalCenters && savedDraft.cvValues.length) {
    await api.importDigitalCenterCVs(savedDraft.returnToDigitalCenters.sessionId,
      savedDraft.sourceWorkItemId || "", vehicle.id);
  } else if (savedDraft.cvValues.length) {
    const current = await api.vehicleCVValues(vehicle.id);
    if (!isCurrent()) return stopped();
    const preview = buildCVImportPreview(savedDraft.sourceSummary.name, savedDraft.cvValues, current);
    if (preview.rows.some(row => row.status === "invalid")) {
      throw new ApiError("The source CV selection is invalid. Review it before importing.", "cv_import_invalid", 400);
    }
    if (preview.rows.some(row => row.selected)) {
      const values = guardedCVImportValues(preview);
      let result: { imported: number };
      try { result = await api.importVehicleCVValues(vehicle.id, values); }
      catch (error) {
        if (!(error instanceof ApiError) || error.status >= 500 || [408, 499].includes(error.status)) {
          throw new ApiError("The CV import result is uncertain. Compare current values before retrying.", "cv_import_unknown", 503);
        }
        throw error;
      }
      if (result.imported !== values.length) {
        throw new ApiError("The CV import result is uncertain. Compare current values before retrying.", "cv_import_unknown", 503);
      }
    }
  }
  if (!isCurrent()) return stopped();
  callbacks.markSaved(savedDraft, vehicle.id);
  vehicle = await api.vehicle(vehicle.id);
  if (!isCurrent()) return stopped();
  const returnsToSource = Boolean(savedDraft.returnToEcos || savedDraft.returnToDigitalCenters);
  callbacks.clear();
  if (returnsToSource) {
    callbacks.reloadVehicles();
    callbacks.closeEditor();
    callbacks.returnToSession(savedDraft);
    return { vehicle, returnedToSource: true };
  }
  return { vehicle, returnedToSource: false };
}
