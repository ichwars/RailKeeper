import type {
  VehicleCVChange, VehicleCVDeletionScope, VehicleCVFile, VehicleCVFilePreview, VehicleCVValue,
  VehicleCVValueInput, VehicleCVWorkspaceComparison
} from "./apiInventoryModels";
import type { APIRequest } from "./apiTransport";

function cvComparisonQuery(expectedRevision?: string, scope?: VehicleCVWorkspaceComparison) {
  const query = new URLSearchParams();
  if (expectedRevision !== undefined) query.set("expectedRevision", expectedRevision);
  if (scope !== undefined) {
    query.set("workspaceId", scope.workspaceId);
    query.set("expectedWorkspaceRevision", scope.expectedWorkspaceRevision);
  }
  return query.size ? `?${query}` : "";
}

export function createVehicleCVAPI(request: APIRequest) {
  return {
    vehicleCVValues: (vehicleId: string) =>
      request<VehicleCVValue[]>(`/vehicles/${encodeURIComponent(vehicleId)}/cv-values`),
    importVehicleCVValues: (vehicleId: string, values: VehicleCVValueInput[]) =>
      request<{ imported: number }>(`/vehicles/${encodeURIComponent(vehicleId)}/cv-values/import`, {
        method: "POST", body: JSON.stringify({ values })
      }, { timeoutMs: 30000 }),
    applyVehicleCVChanges: (vehicleId: string, changes: VehicleCVChange[]) =>
      request<{ changed: number }>(`/vehicles/${encodeURIComponent(vehicleId)}/cv-values/changes`, {
        method: "POST", body: JSON.stringify({ changes })
      }, { timeoutMs: 30000 }),
    createVehicleCVValue: (vehicleId: string, input: VehicleCVValueInput) =>
      request<VehicleCVValue>(`/vehicles/${encodeURIComponent(vehicleId)}/cv-values`, {
        method: "POST",
        body: JSON.stringify(input)
      }),
    updateVehicleCVValue: (vehicleId: string, cvValueId: string, input: VehicleCVValueInput, sourceScope?: VehicleCVWorkspaceComparison) =>
      request<VehicleCVValue>(`/vehicles/${encodeURIComponent(vehicleId)}/cv-values/${encodeURIComponent(cvValueId)}${sourceScope ? cvComparisonQuery(input.expectedRevision, sourceScope) : ""}`, {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    deleteVehicleCVValue: (vehicleId: string, cvValueId: string, expectedRevision?: string, scope?: VehicleCVDeletionScope) =>
      request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/cv-values/${encodeURIComponent(cvValueId)}${cvComparisonQuery(expectedRevision, scope)}`, {
        method: "DELETE"
      }),
    uploadVehicleCVFile: (vehicleId: string, file: File, decoderProfile = "", description = "",
      workspaceId?: string, expectedWorkspaceRevision?: string) => {
      const form = new FormData();
      form.append("file", file);
      form.append("decoderProfile", decoderProfile);
      form.append("description", description);
      if (workspaceId !== undefined) form.append("workspaceId", workspaceId);
      if (expectedWorkspaceRevision !== undefined) form.append("expectedWorkspaceRevision", expectedWorkspaceRevision);
      return request<VehicleCVFile>(
        `/vehicles/${encodeURIComponent(vehicleId)}/cv-files`,
        {
          method: "POST",
          body: form
        },
        { timeoutMs: 30000 }
      );
    },
    previewVehicleCVFile: (file: File) => {
      const form = new FormData();
      form.append("file", file);
      return request<VehicleCVFilePreview>(
        "/cv-files/preview",
        {
          method: "POST",
          body: form
        },
        { timeoutMs: 30000 }
      );
    },
    deleteVehicleCVFile: (vehicleId: string, cvFileId: string, expectedRevision?: string, scope?: VehicleCVDeletionScope) =>
      request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/cv-files/${encodeURIComponent(cvFileId)}${cvComparisonQuery(expectedRevision, scope)}`, {
        method: "DELETE"
      }),
    vehicleCVFileDownloadUrl: (vehicleId: string, cvFileId: string) =>
      `/api/v1/vehicles/${encodeURIComponent(vehicleId)}/cv-files/${encodeURIComponent(cvFileId)}/download`
  };
}
