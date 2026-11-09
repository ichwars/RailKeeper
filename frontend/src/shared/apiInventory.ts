import type {
  CreateVehicleRequest, CreateVehicleSetRequest, OverviewValuation, Vehicle, VehicleAttachment,
  VehicleAttachmentUpdateInput, VehicleCatalogBackfillReport, VehicleExternalMapping,
  VehicleExternalMappingInput, VehicleImageCreation, VehicleImageImportInput, VehicleSet, VehicleSetInput,
  VehicleSetMainImageMode
} from "./apiInventoryModels";
import type {
  ArticleSearchInput, ArticleSearchResponse, InventoryNumberScheme, InventoryNumberSchemeCreateInput,
  InventoryNumberSchemeInput, VehicleAttachmentImportInput
} from "./apiMasterDataModels";
import type { APIRequest } from "./apiTransport";

export function createInventoryAPI(request: APIRequest) {
  return {
    catalogBackfillReport: () => request<VehicleCatalogBackfillReport>("/catalog/legacy-backfill"),
    backfillLegacyCatalog: () => request<VehicleCatalogBackfillReport>("/catalog/legacy-backfill", { method: "POST" }),
    vehicles: (query = "") =>
      request<Vehicle[]>(`/vehicles${query ? `?q=${encodeURIComponent(query)}` : ""}`),
    overviewValuation: () => request<OverviewValuation>("/overview/valuation"),
    createVehicle: (input: CreateVehicleRequest & { externalMapping?: VehicleExternalMappingInput; creationId?: string }) =>
      request<Vehicle>("/vehicles", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    createVehicleSet: (input: CreateVehicleSetRequest) =>
      request<VehicleSet>("/vehicle-sets", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    vehicleSet: (id: string) => request<VehicleSet>(`/vehicle-sets/${encodeURIComponent(id)}`),
    updateVehicleSet: (id: string, input: VehicleSetInput) =>
      request<VehicleSet>(`/vehicle-sets/${encodeURIComponent(id)}`, {
        method: "PATCH",
        body: JSON.stringify(input)
      }),
    setVehicleSetMainImage: (id: string, input: { mode: VehicleSetMainImageMode; memberImageId?: string }) =>
      request<VehicleSet>(`/vehicle-sets/${encodeURIComponent(id)}/main-image`, {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    uploadVehicleSetImage: (id: string, file: File) => {
      const form = new FormData();
      form.append("file", file);
      return request<VehicleSet>(`/vehicle-sets/${encodeURIComponent(id)}/image`, {
        method: "POST",
        body: form
      }, { timeoutMs: 30000 });
    },
    deleteVehicleSetImage: (id: string) =>
      request<void>(`/vehicle-sets/${encodeURIComponent(id)}/image`, { method: "DELETE" }),
    vehicle: (id: string) => request<Vehicle>(`/vehicles/${encodeURIComponent(id)}`),
    updateVehicle: (id: string, input: CreateVehicleRequest) =>
      request<Vehicle>(`/vehicles/${encodeURIComponent(id)}`, {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    deleteVehicle: (id: string) =>
      request<void>(`/vehicles/${encodeURIComponent(id)}`, {
        method: "DELETE"
      }),
    uploadVehicleImage: (
      vehicleId: string, file: File, title = "", isPrimary = false, maintenanceId = "", expectedImagesRevision?: string
    ) => {
      const form = new FormData();
      form.append("file", file);
      form.append("title", title);
      form.append("isPrimary", String(isPrimary));
      form.append("maintenanceId", maintenanceId);
      if (expectedImagesRevision !== undefined) form.append("expectedImagesRevision", expectedImagesRevision);
      return request<VehicleImageCreation>(
        `/vehicles/${encodeURIComponent(vehicleId)}/images`,
        {
          method: "POST",
          body: form
        },
        { timeoutMs: 30000 }
      );
    },
    importVehicleImageFromUrl: (vehicleId: string, input: VehicleImageImportInput) =>
      request<VehicleImageCreation>(
        `/vehicles/${encodeURIComponent(vehicleId)}/images/import-url`,
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 30000 }
      ),
    deleteVehicleImage: (vehicleId: string, imageId: string, expectedRevision?: string) =>
      request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/images/${encodeURIComponent(imageId)}${expectedRevision === undefined ? "" : `?expectedRevision=${encodeURIComponent(expectedRevision)}`}`, {
        method: "DELETE"
      }),
    uploadVehicleAttachment: (vehicleId: string, file: File, category = "", description = "", maintenanceId = "") => {
      const form = new FormData();
      form.append("file", file);
      form.append("category", category);
      form.append("description", description);
      form.append("maintenanceId", maintenanceId);
      return request<VehicleAttachment>(
        `/vehicles/${encodeURIComponent(vehicleId)}/attachments`,
        {
          method: "POST",
          body: form
        },
        { timeoutMs: 30000 }
      );
    },
    updateVehicleAttachment: (vehicleId: string, attachmentId: string, input: VehicleAttachmentUpdateInput) =>
      request<VehicleAttachment>(
        `/vehicles/${encodeURIComponent(vehicleId)}/attachments/${encodeURIComponent(attachmentId)}`,
        {
          method: "PUT",
          body: JSON.stringify(input)
        }
      ),
    deleteVehicleAttachment: (vehicleId: string, attachmentId: string, expectedRevision?: string) =>
      request<void>(`/vehicles/${encodeURIComponent(vehicleId)}/attachments/${encodeURIComponent(attachmentId)}${expectedRevision === undefined ? "" : `?expectedRevision=${encodeURIComponent(expectedRevision)}`}`, {
        method: "DELETE"
      }),
    vehicleAttachmentDownloadUrl: (vehicleId: string, attachmentId: string) =>
      `/api/v1/vehicles/${encodeURIComponent(vehicleId)}/attachments/${encodeURIComponent(attachmentId)}/download`,
    importVehicleAttachmentFromUrl: (vehicleId: string, input: VehicleAttachmentImportInput) =>
      request<VehicleAttachment>(`/vehicles/${encodeURIComponent(vehicleId)}/attachments/import-url`, {
        method: "POST",
        body: JSON.stringify(input)
      }),
    inventoryNumberSchemes: () => request<InventoryNumberScheme[]>("/inventory-number-schemes"),
    createInventoryNumberScheme: (input: InventoryNumberSchemeCreateInput) =>
      request<InventoryNumberScheme>("/inventory-number-schemes", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    updateInventoryNumberScheme: (category: string, input: InventoryNumberSchemeInput) =>
      request<InventoryNumberScheme>(`/inventory-number-schemes/${encodeURIComponent(category)}`, {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    upsertVehicleExternalMapping: (vehicleId: string, input: VehicleExternalMappingInput) =>
      request<VehicleExternalMapping>(`/vehicles/${encodeURIComponent(vehicleId)}/external-mappings`, {
        method: "POST",
        body: JSON.stringify(input)
      }),
    articleSearch: (input: ArticleSearchInput) =>
      request<ArticleSearchResponse>(
        "/article-search",
        {
          method: "POST",
          body: JSON.stringify(input)
        },
        { timeoutMs: 15000 }
      )
  };
}
