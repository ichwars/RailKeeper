import type { components } from "../generated/api";

export type VehicleCatalogFields = components["schemas"]["VehicleCatalogFields"];
export type VehicleCatalogContext = components["schemas"]["VehicleCatalogContext"];
export type VehicleCatalogRevision = components["schemas"]["VehicleCatalogRevision"];
export type VehicleCatalogRevisionInput = components["schemas"]["VehicleCatalogRevisionInput"];
export type VehicleCatalogAdoptionInput = components["schemas"]["VehicleCatalogAdoptionInput"];

type APIRequest = <T>(path: string, init?: RequestInit) => Promise<T>;

export function createVehicleCatalogAPI(request: APIRequest) {
  const path = (targetType: "vehicle" | "set", id: string) => `/catalog/${targetType}/${encodeURIComponent(id)}`;
  return {
    catalogContext: (targetType: "vehicle" | "set", id: string) => request<VehicleCatalogContext>(path(targetType, id)),
    createCatalogRevision: (targetType: "vehicle" | "set", id: string, input: VehicleCatalogRevisionInput) =>
      request<VehicleCatalogContext>(`${path(targetType, id)}/revisions`, { method: "POST", body: JSON.stringify(input) }),
    adoptCatalogRevision: (targetType: "vehicle" | "set", id: string, input: VehicleCatalogAdoptionInput) =>
      request<VehicleCatalogContext>(`${path(targetType, id)}/adoption`, { method: "POST", body: JSON.stringify(input) })
  };
}
