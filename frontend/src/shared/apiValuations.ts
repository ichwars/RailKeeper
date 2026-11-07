import type { components } from "../generated/api";

export type VehicleValuation = components["schemas"]["VehicleValuation"];
export type VehicleValuationInput = components["schemas"]["VehicleValuationInput"];
export type VehicleValuationHistory = components["schemas"]["VehicleValuationHistory"];
export type VehicleValuationSummary = components["schemas"]["VehicleValuationSummary"];
export type VehicleValuationExport = components["schemas"]["VehicleValuationExport"];

type APIRequest = <T>(path: string, init?: RequestInit) => Promise<T>;

export function createValuationsAPI(request: APIRequest) {
  return {
    vehicleValuations: (vehicleID: string, offset = 0) => request<VehicleValuationHistory>(
      `/vehicles/${encodeURIComponent(vehicleID)}/valuations?offset=${offset}`),
    recordVehicleValuation: (vehicleID: string, input: VehicleValuationInput) => request<VehicleValuation>(
      `/vehicles/${encodeURIComponent(vehicleID)}/valuations`, { method: "POST", body: JSON.stringify(input) }),
    vehicleValuationsExport: (vehicleID: string) => request<VehicleValuationExport>(
      `/vehicles/${encodeURIComponent(vehicleID)}/valuations/export`),
    vehicleValuationsExportURL: (vehicleID: string) => `/api/v1/vehicles/${encodeURIComponent(vehicleID)}/valuations/export`,
    valuationSummary: (asOf: string) => request<VehicleValuationSummary>(`/valuations/summary?asOf=${encodeURIComponent(asOf)}`)
  };
}
