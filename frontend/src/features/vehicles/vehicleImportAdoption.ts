import { api, type Vehicle } from "../../shared/api";
import { cvValueKey } from "./cvImport";
import type { ECoSVehicleDraftPayload } from "./vehicleViewModel";

type VehicleImportCallbacks = {
  markSaved: (draft: ECoSVehicleDraftPayload, vehicleID: string) => void;
  clear: () => void;
  returnToSession: (draft: ECoSVehicleDraftPayload) => void;
  reloadVehicles: () => void;
  closeEditor: () => void;
};

export async function adoptVehicleImportDraft(
  vehicle: Vehicle,
  savedDraft: ECoSVehicleDraftPayload,
  callbacks: VehicleImportCallbacks
) {
  if (!vehicle.externalMappings?.some(mapping => mapping.provider === savedDraft.externalMapping.provider
    && mapping.externalId === savedDraft.externalMapping.externalId && mapping.vehicleId === vehicle.id)) {
    await api.upsertVehicleExternalMapping(vehicle.id, savedDraft.externalMapping);
  }
  if (savedDraft.source === "ecos" && savedDraft.returnToDigitalCenters && savedDraft.cvValues.length) {
    await api.importDigitalCenterCVs(savedDraft.returnToDigitalCenters.sessionId,
      savedDraft.sourceWorkItemId || "", vehicle.id);
  } else {
    const detailBeforeECoSValues = await api.vehicle(vehicle.id);
    for (const cvValue of savedDraft.cvValues) {
      const existing = (detailBeforeECoSValues.cvValues || []).find((entry) => (
        cvValueKey(entry) === cvValueKey(cvValue)
      ));
      if (existing) await api.updateVehicleCVValue(vehicle.id, existing.id, cvValue);
      else await api.createVehicleCVValue(vehicle.id, cvValue);
    }
  }
  callbacks.markSaved(savedDraft, vehicle.id);
  vehicle = await api.vehicle(vehicle.id);
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
