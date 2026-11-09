import type { AccessoryAllocationSummary, AccessoryArticle, AccessoryArticleListQuery, AccessoryArticleListResult, AccessoryArticleWriteInput, AccessoryAsset, AccessoryAssetInput, AccessoryDocument, AccessoryDocumentImportURLInput, AccessoryDocumentUpdateInput, AccessoryDocumentUploadInput, AccessoryDuplicateCheckInput, AccessoryDuplicateCheckResult, AccessoryIndividualizationInput, AccessoryInstallation, AccessoryInstallationConditionInput, AccessoryInstallationInput, AccessoryInstallationRemovalInput, AccessoryPurchase, AccessoryPurchaseInput, AccessoryReservation, AccessoryReservationInput, AccessoryStockAdjustmentInput, AccessoryStockMovement, AccessoryStockSummary, AccessoryStockTransferInput, AccessoryUsageHistory, StorageLocation, StorageLocationInput } from "./apiAccessoryModels";
import type { APIRequest } from "./apiTransport";
export type * from "./apiAccessoryModels";

export function createAccessoriesAPI(request: APIRequest) {
  const productQuery = (productId?: string) => productId ? `?productId=${encodeURIComponent(productId)}` : "";
  return {
    accessoryArticles: (filters: AccessoryArticleListQuery = {}) =>
      request<AccessoryArticleListResult>(`/accessory-products${articleQuery(filters)}`),
    createAccessoryArticle: (input: AccessoryArticleWriteInput) =>
      request<AccessoryArticle>("/accessory-products", json("POST", input)),
    accessoryArticle: (id: string) =>
      request<AccessoryArticle>(`/accessory-products/${encodeURIComponent(id)}`),
    updateAccessoryArticle: (id: string, input: AccessoryArticleWriteInput) =>
      request<AccessoryArticle>(`/accessory-products/${encodeURIComponent(id)}`, json("PUT", input)),
    checkAccessoryArticleDuplicates: (input: AccessoryDuplicateCheckInput) =>
      request<AccessoryDuplicateCheckResult>("/accessory-products/duplicate-check", json("POST", input)),
    archiveAccessoryProduct: (id: string) =>
      request<AccessoryArticle>(`/accessory-products/${encodeURIComponent(id)}/archive`, { method: "POST" }),
    restoreAccessoryProduct: (id: string) =>
      request<AccessoryArticle>(`/accessory-products/${encodeURIComponent(id)}/restore`, { method: "POST" }),
    deleteAccessoryProduct: (id: string) =>
      request<void>(`/accessory-products/${encodeURIComponent(id)}`, { method: "DELETE" }),
    storageLocations: () => request<StorageLocation[]>("/storage-locations"),
    createStorageLocation: (input: StorageLocationInput) =>
      request<StorageLocation>("/storage-locations", json("POST", input)),
    updateStorageLocation: (id: string, input: StorageLocationInput) =>
      request<StorageLocation>(`/storage-locations/${encodeURIComponent(id)}`, json("PUT", input)),
    accessoryStock: (productId: string) =>
      request<AccessoryStockSummary>(`/accessory-products/${encodeURIComponent(productId)}/stock`),
    adjustAccessoryStock: (productId: string, input: AccessoryStockAdjustmentInput) =>
      request<AccessoryStockSummary>(
        `/accessory-products/${encodeURIComponent(productId)}/stock-adjustments`, json("POST", input)
      ),
    accessoryStockMovements: (productId: string) =>
      request<AccessoryStockMovement[]>(
        `/accessory-products/${encodeURIComponent(productId)}/stock-movements`
      ),
    transferAccessoryStock: (productId: string, input: AccessoryStockTransferInput) =>
      request<AccessoryStockSummary>(
        `/accessory-products/${encodeURIComponent(productId)}/stock-transfers`, json("POST", input)
      ),
    accessoryPurchases: (productId: string) =>
      request<AccessoryPurchase[]>(`/accessory-products/${encodeURIComponent(productId)}/purchases`),
    createAccessoryPurchase: (productId: string, input: AccessoryPurchaseInput) =>
      request<AccessoryPurchase>(
        `/accessory-products/${encodeURIComponent(productId)}/purchases`, json("POST", input)
      ),
    accessoryAssets: (productId: string) =>
      request<AccessoryAsset[]>(`/accessory-products/${encodeURIComponent(productId)}/assets`),
    createAccessoryAsset: (productId: string, input: AccessoryAssetInput) =>
      request<AccessoryAsset>(`/accessory-products/${encodeURIComponent(productId)}/assets`, json("POST", input)),
    individualizeAccessoryProduct: (productId: string, input: AccessoryIndividualizationInput) =>
      request<AccessoryAsset>(
        `/accessory-products/${encodeURIComponent(productId)}/individualizations`, json("POST", input)
      ),
    updateAccessoryAsset: (id: string, input: AccessoryAssetInput) =>
      request<AccessoryAsset>(`/accessory-assets/${encodeURIComponent(id)}`, json("PUT", input)),
    accessoryDocuments: (productId: string) =>
      request<AccessoryDocument[]>(`/accessory-products/${encodeURIComponent(productId)}/documents`),
    uploadAccessoryDocument: (productId: string, input: AccessoryDocumentUploadInput) =>
      request<AccessoryDocument>(
        `/accessory-products/${encodeURIComponent(productId)}/documents`,
        { method: "POST", body: accessoryDocumentForm(input) }
      ),
    importAccessoryDocumentFromUrl: (productId: string, input: AccessoryDocumentImportURLInput) =>
      request<AccessoryDocument>(
        `/accessory-products/${encodeURIComponent(productId)}/documents/import-url`,
        json("POST", input)
      ),
    accessoryDocument: (productId: string, documentId: string) =>
      request<AccessoryDocument>(accessoryDocumentPath(productId, documentId)),
    updateAccessoryDocument: (
      productId: string,
      documentId: string,
      input: AccessoryDocumentUpdateInput
    ) => request<AccessoryDocument>(accessoryDocumentPath(productId, documentId), json("PUT", input)),
    deleteAccessoryDocument: (productId: string, documentId: string) =>
      request<void>(accessoryDocumentPath(productId, documentId), { method: "DELETE" }),
    accessoryDocumentDownloadPath: (productId: string, documentId: string) =>
      `/api/v1/accessory-products/${encodeURIComponent(productId)}/documents/${encodeURIComponent(documentId)}/download`,
    accessoryAllocationSummary: (productId: string) =>
      request<AccessoryAllocationSummary>(
        `/accessory-products/${encodeURIComponent(productId)}/allocation-summary`
      ),
    accessoryUsageHistory: (productId: string) =>
      request<AccessoryUsageHistory>(
        `/accessory-products/${encodeURIComponent(productId)}/usage-history`
      ),
    accessoryReservations: (productId?: string) =>
      request<AccessoryReservation[]>(`/accessory-reservations${productQuery(productId)}`),
    createAccessoryReservation: (input: AccessoryReservationInput) =>
      request<AccessoryReservation>("/accessory-reservations", json("POST", input)),
    cancelAccessoryReservation: (id: string) =>
      request<AccessoryReservation>(`/accessory-reservations/${encodeURIComponent(id)}/cancel`, { method: "POST" }),
    accessoryInstallations: (productId?: string) =>
      request<AccessoryInstallation[]>(`/accessory-installations${productQuery(productId)}`),
    createAccessoryInstallation: (input: AccessoryInstallationInput) =>
      request<AccessoryInstallation>("/accessory-installations", json("POST", input)),
    removeAccessoryInstallation: (id: string, input: AccessoryInstallationRemovalInput) =>
      request<AccessoryInstallation>(`/accessory-installations/${encodeURIComponent(id)}/remove`, json("POST", input)),
    updateAccessoryInstallationCondition: (id: string, input: AccessoryInstallationConditionInput) =>
      request<AccessoryInstallation>(
        `/accessory-installations/${encodeURIComponent(id)}/condition`, json("PUT", input)
      )
  };
}

function articleQuery(filters: AccessoryArticleListQuery): string {
  const query = new URLSearchParams();
  if (filters.query !== undefined) query.set("query", filters.query);
  filters.articleTypes?.forEach((value) => query.append("articleType", value));
  if (filters.manufacturer !== undefined) query.set("manufacturer", filters.manufacturer);
  filters.gauges?.forEach((value) => query.append("gauge", value));
  filters.statuses?.forEach((value) => query.append("status", value));
  if (filters.locationId !== undefined) query.set("locationId", filters.locationId);
  if (filters.sort !== undefined) query.set("sort", filters.sort);
  if (filters.direction !== undefined) query.set("direction", filters.direction);
  const encoded = query.toString();
  return encoded ? `?${encoded}` : "";
}

function accessoryDocumentPath(productId: string, documentId: string): string {
  return `/accessory-products/${encodeURIComponent(productId)}/documents/${encodeURIComponent(documentId)}`;
}

function accessoryDocumentForm(input: AccessoryDocumentUploadInput): FormData {
  const form = new FormData();
  form.append("file", input.file);
  form.append("category", input.category);
  if (input.description !== undefined) form.append("description", input.description);
  if (input.isPrimary !== undefined) form.append("isPrimary", String(input.isPrimary));
  return form;
}

function json(method: "POST" | "PUT", body: object): RequestInit {
  return { method, body: JSON.stringify(body) };
}
