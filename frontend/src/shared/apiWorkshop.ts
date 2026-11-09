import type {
  VehicleFunction, VehicleFunctionInput, VehicleMaintenance, VehicleMaintenanceInput, VehicleSparePart,
  VehicleSparePartInput
} from "./apiInventoryModels";
import type { ArticleSearchSparePart } from "./apiMasterDataModels";
import type { APIRequest } from "./apiTransport";

export function createWorkshopAPI(request: APIRequest) {
  return {
    vehicleMaintenance: (vehicleId: string) =>
      request<VehicleMaintenance[]>(`/vehicles/${encodeURIComponent(vehicleId)}/maintenance`),
    createVehicleMaintenance: (vehicleId: string, input: VehicleMaintenanceInput) =>
      request<VehicleMaintenance>(`/vehicles/${encodeURIComponent(vehicleId)}/maintenance`, {
        method: "POST",
        body: JSON.stringify(input)
      }),
    updateVehicleMaintenance: (vehicleId: string, maintenanceId: string, input: VehicleMaintenanceInput) =>
      request<VehicleMaintenance>(
        `/vehicles/${encodeURIComponent(vehicleId)}/maintenance/${encodeURIComponent(maintenanceId)}`,
        {
          method: "PUT",
          body: JSON.stringify(input)
        }
      ),
    deleteVehicleMaintenance: (vehicleId: string, maintenanceId: string) =>
      request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/maintenance/${encodeURIComponent(maintenanceId)}`, {
        method: "DELETE"
      }),
    vehicleSpareParts: (vehicleId: string) =>
      request<VehicleSparePart[]>(`/vehicles/${encodeURIComponent(vehicleId)}/spare-parts`),
    vehicleSparePartSuggestions: (vehicleId: string, attachmentId = "") =>
      request<ArticleSearchSparePart[]>(
        `/vehicles/${encodeURIComponent(vehicleId)}/spare-parts/suggestions${attachmentId ? `?attachmentId=${encodeURIComponent(attachmentId)}` : ""}`,
        undefined,
        { timeoutMs: 15000 }
      ),
    createVehicleSparePart: (vehicleId: string, input: VehicleSparePartInput) =>
      request<VehicleSparePart>(`/vehicles/${encodeURIComponent(vehicleId)}/spare-parts`, {
        method: "POST",
        body: JSON.stringify(input)
      }),
    updateVehicleSparePart: (vehicleId: string, sparePartId: string, input: VehicleSparePartInput) =>
      request<VehicleSparePart>(
        `/vehicles/${encodeURIComponent(vehicleId)}/spare-parts/${encodeURIComponent(sparePartId)}`,
        {
          method: "PUT",
          body: JSON.stringify(input)
        }
      ),
    deleteVehicleSparePart: (vehicleId: string, sparePartId: string) =>
      request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/spare-parts/${encodeURIComponent(sparePartId)}`, {
        method: "DELETE"
      }),
    vehicleFunctions: (vehicleId: string) =>
      request<VehicleFunction[]>(`/vehicles/${encodeURIComponent(vehicleId)}/functions`),
    updateVehicleFunction: (vehicleId: string, functionKey: string, input: VehicleFunctionInput) =>
      request<VehicleFunction>(`/vehicles/${encodeURIComponent(vehicleId)}/functions/${encodeURIComponent(functionKey)}`, {
        method: "PUT",
        body: JSON.stringify({ ...input, flagsVersion: input.flagsVersion ?? 2 })
      }),
    deleteVehicleFunction: (vehicleId: string, functionKey: string) =>
      request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/functions/${encodeURIComponent(functionKey)}`, {
        method: "DELETE"
      })
  };
}
