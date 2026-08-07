export type LayoutKind = "private" | "club";
export type LayoutUnitKind = "baseboard" | "module" | "segment" | "area";
export type PlanRevisionStatus = "draft" | "review" | "published" | "archived";

export type Layout = {
  id: string;
  name: string;
  kind: LayoutKind;
  gauge: string;
  scale: string;
  description?: string;
  version: number;
  archived: boolean;
  createdAt: string;
  updatedAt: string;
};

export type LayoutInput = {
  name: string;
  kind: LayoutKind;
  gauge: string;
  scale: string;
  description?: string;
  archived?: boolean;
};

export type LayoutUpdateInput = LayoutInput & { expectedVersion: number };

export type LayoutUnit = {
  id: string;
  layoutId: string;
  name: string;
  kind: LayoutUnitKind;
  ownerLabel?: string;
  widthMm: number;
  heightMm: number;
  version: number;
  archived: boolean;
  createdAt: string;
  updatedAt: string;
};

export type LayoutUnitInput = {
  name: string;
  kind: LayoutUnitKind;
  ownerLabel?: string;
  widthMm: number;
  heightMm: number;
  archived?: boolean;
};

export type LayoutUnitUpdateInput = LayoutUnitInput & { expectedVersion: number };

export type ConfigurationUnit = {
  unitId: string;
  planRevisionId?: string;
  positionXMm: number;
  positionYMm: number;
  rotationDegrees: number;
  sortOrder: number;
};

export type LayoutConfiguration = {
  id: string;
  layoutId: string;
  name: string;
  description?: string;
  version: number;
  archived: boolean;
  units: ConfigurationUnit[];
  createdAt: string;
  updatedAt: string;
};

export type LayoutConfigurationInput = {
  name: string;
  description?: string;
  archived?: boolean;
  units: ConfigurationUnit[];
};

export type LayoutConfigurationUpdateInput = LayoutConfigurationInput & { expectedVersion: number };

export type PlanRevision = {
  id: string;
  variantId: string;
  revisionNumber: number;
  status: PlanRevisionStatus;
  baseRevisionId?: string;
  version: number;
  createdBy: string;
  publishedBy?: string;
  publishedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type PlanVariant = {
  id: string;
  layoutUnitId: string;
  name: string;
  description?: string;
  archived: boolean;
  revisions: PlanRevision[];
  createdAt: string;
  updatedAt: string;
};

export type PlanVariantInput = { name: string; description?: string };
export type PlanRevisionInput = { baseRevisionId?: string };

export type AccessoryTrackingMode = "quantity" | "individual";
export type AccessoryCondition = "ready" | "maintenance_due" | "defective" | "unknown";
export type AccessoryLifecycle = "stored" | "reserved" | "installed" | "maintenance" | "retired";
export type AccessoryReservationStatus = "active" | "fulfilled" | "cancelled";
export type AccessoryRemovalDisposition = "stored" | "maintenance" | "defective" | "retired";

export type AccessoryProduct = {
  id: string;
  manufacturer: string;
  articleNumber?: string;
  name: string;
  category: string;
  trackingMode: AccessoryTrackingMode;
  description?: string;
  createdAt: string;
  updatedAt: string;
};

export type AccessoryProductInput = {
  manufacturer: string;
  articleNumber?: string;
  name: string;
  category: string;
  trackingMode: AccessoryTrackingMode;
  description?: string;
};

export type StorageLocation = {
  id: string;
  parentId?: string;
  name: string;
  description?: string;
  archived: boolean;
  createdAt: string;
  updatedAt: string;
};

export type StorageLocationInput = {
  parentId?: string;
  name: string;
  description?: string;
  archived?: boolean;
};

export type AccessoryStockLevel = {
  locationId: string;
  locationName: string;
  quantity: number;
  updatedAt: string;
};

export type AccessoryStockSummary = {
  productId: string;
  trackingMode: AccessoryTrackingMode;
  totalQuantity: number;
  locations: AccessoryStockLevel[];
};

export type AccessoryStockAdjustmentInput = { locationId: string; delta: number };

export type AccessoryAsset = {
  id: string;
  productId: string;
  inventoryNumber?: string;
  serialNumber?: string;
  condition: AccessoryCondition;
  lifecycle: AccessoryLifecycle;
  storageLocationId?: string;
  purchaseDate?: string;
  purchasePrice?: string;
  warrantyUntil?: string;
  notes?: string;
  createdAt: string;
  updatedAt: string;
};

export type AccessoryAssetInput = {
  inventoryNumber?: string;
  serialNumber?: string;
  condition?: AccessoryCondition;
  lifecycle?: Exclude<AccessoryLifecycle, "reserved" | "installed">;
  storageLocationId?: string;
  purchaseDate?: string;
  purchasePrice?: string;
  warrantyUntil?: string;
  notes?: string;
};

export type AllocationTarget =
  | { vehicleId: string; layoutId?: never; layoutUnitId?: never }
  | { vehicleId?: never; layoutId: string; layoutUnitId?: never }
  | { vehicleId?: never; layoutId?: never; layoutUnitId: string };

export type AccessoryReservation = AllocationTarget & {
  id: string;
  productId: string;
  assetId?: string;
  locationId: string;
  quantity: number;
  status: AccessoryReservationStatus;
  note?: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
};

export type AccessoryReservationInput = AllocationTarget & {
  productId: string;
  assetId?: string;
  locationId: string;
  quantity: number;
  note?: string;
};

export type AccessoryInstallation = AllocationTarget & {
  id: string;
  productId: string;
  assetId?: string;
  sourceLocationId: string;
  quantity: number;
  condition: AccessoryCondition;
  installedBy: string;
  installedAt: string;
  removedBy?: string;
  removedAt?: string;
  removalDisposition?: AccessoryRemovalDisposition;
  notes?: string;
  removalNotes?: string;
};

export type AccessoryInstallationInput = AllocationTarget & {
  reservationId?: string;
  productId: string;
  assetId?: string;
  sourceLocationId: string;
  quantity: number;
  condition?: AccessoryCondition;
  notes?: string;
};

export type AccessoryInstallationRemovalInput = {
  disposition: AccessoryRemovalDisposition;
  storageLocationId?: string;
  notes?: string;
};

export type AccessoryInstallationConditionInput = { condition: AccessoryCondition };

export type AccessoryAllocationSummary = {
  productId: string;
  owned: number;
  stored: number;
  reserved: number;
  installed: number;
  available: number;
  missing: number;
};

type RequestOptions = { retries?: number; timeoutMs?: number };
type APIRequest = <T>(path: string, init?: RequestInit, options?: RequestOptions) => Promise<T>;

export function createLayoutsAccessoriesAPI(request: APIRequest) {
  const productQuery = (productId?: string) => productId ? `?productId=${encodeURIComponent(productId)}` : "";
  return {
    layouts: () => request<Layout[]>("/layouts"),
    createLayout: (input: LayoutInput) => request<Layout>("/layouts", json("POST", input)),
    layout: (id: string) => request<Layout>(`/layouts/${encodeURIComponent(id)}`),
    updateLayout: (id: string, input: LayoutUpdateInput) =>
      request<Layout>(`/layouts/${encodeURIComponent(id)}`, json("PUT", input)),
    layoutUnits: (layoutId: string) => request<LayoutUnit[]>(`/layouts/${encodeURIComponent(layoutId)}/units`),
    createLayoutUnit: (layoutId: string, input: LayoutUnitInput) =>
      request<LayoutUnit>(`/layouts/${encodeURIComponent(layoutId)}/units`, json("POST", input)),
    updateLayoutUnit: (id: string, input: LayoutUnitUpdateInput) =>
      request<LayoutUnit>(`/layout-units/${encodeURIComponent(id)}`, json("PUT", input)),
    layoutConfigurations: (layoutId: string) =>
      request<LayoutConfiguration[]>(`/layouts/${encodeURIComponent(layoutId)}/configurations`),
    createLayoutConfiguration: (layoutId: string, input: LayoutConfigurationInput) =>
      request<LayoutConfiguration>(`/layouts/${encodeURIComponent(layoutId)}/configurations`, json("POST", input)),
    updateLayoutConfiguration: (id: string, input: LayoutConfigurationUpdateInput) =>
      request<LayoutConfiguration>(`/layout-configurations/${encodeURIComponent(id)}`, json("PUT", input)),
    planVariants: (unitId: string) =>
      request<PlanVariant[]>(`/layout-units/${encodeURIComponent(unitId)}/plan-variants`),
    createPlanVariant: (unitId: string, input: PlanVariantInput) =>
      request<PlanVariant>(`/layout-units/${encodeURIComponent(unitId)}/plan-variants`, json("POST", input)),
    createPlanRevision: (variantId: string, input: PlanRevisionInput) =>
      request<PlanRevision>(`/plan-variants/${encodeURIComponent(variantId)}/revisions`, json("POST", input)),
    submitPlanRevision: (id: string, expectedVersion: number) =>
      request<PlanRevision>(`/plan-revisions/${encodeURIComponent(id)}/submit`, json("POST", { expectedVersion })),
    publishPlanRevision: (id: string, expectedVersion: number) =>
      request<PlanRevision>(`/plan-revisions/${encodeURIComponent(id)}/publish`, json("POST", { expectedVersion })),
    accessoryProducts: (query = "") =>
      request<AccessoryProduct[]>(`/accessory-products${query ? `?query=${encodeURIComponent(query)}` : ""}`),
    createAccessoryProduct: (input: AccessoryProductInput) =>
      request<AccessoryProduct>("/accessory-products", json("POST", input)),
    accessoryProduct: (id: string) => request<AccessoryProduct>(`/accessory-products/${encodeURIComponent(id)}`),
    updateAccessoryProduct: (id: string, input: AccessoryProductInput) =>
      request<AccessoryProduct>(`/accessory-products/${encodeURIComponent(id)}`, json("PUT", input)),
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
    accessoryAssets: (productId: string) =>
      request<AccessoryAsset[]>(`/accessory-products/${encodeURIComponent(productId)}/assets`),
    createAccessoryAsset: (productId: string, input: AccessoryAssetInput) =>
      request<AccessoryAsset>(`/accessory-products/${encodeURIComponent(productId)}/assets`, json("POST", input)),
    updateAccessoryAsset: (id: string, input: AccessoryAssetInput) =>
      request<AccessoryAsset>(`/accessory-assets/${encodeURIComponent(id)}`, json("PUT", input)),
    accessoryAllocationSummary: (productId: string) =>
      request<AccessoryAllocationSummary>(
        `/accessory-products/${encodeURIComponent(productId)}/allocation-summary`
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

function json(method: "POST" | "PUT", body: object): RequestInit {
  return { method, body: JSON.stringify(body) };
}
