export type AccessoryTrackingMode = "quantity" | "individual";

export type AccessoryInventoryStrategy = "quantity" | "individual" | "quantity_later_individual";

export type AccessoryArticleType =
  | "track"
  | "signal"
  | "decoder"
  | "electrical_control"
  | "building_equipment"
  | "landscape_consumable"
  | "lighting"
  | "other";

export type AccessoryManufacturerStatus = "announced" | "available" | "discontinued" | "unknown";

export type AccessoryArticleStatus =
  | "available"
  | "reserved"
  | "installed"
  | "maintenance_due"
  | "defective"
  | "archived";

export type AccessoryArticleSort =
  | "article"
  | "image"
  | "inventoryNumber"
  | "manufacturer"
  | "articleNumber"
  | "name"
  | "type"
  | "gauge"
  | "stock"
  | "storage"
  | "updatedAt";

export type AccessorySortDirection = "asc" | "desc";

export type AccessoryCondition = "ready" | "maintenance_due" | "defective" | "unknown";

export type AccessoryLifecycle = "stored" | "reserved" | "installed" | "maintenance" | "retired";

export type AccessoryReservationStatus = "active" | "fulfilled" | "cancelled";

export type AccessoryRemovalDisposition = "stored" | "maintenance" | "defective" | "retired";

export type AccessoryAttributeValue =
  | { key: string; kind: "text"; textValue: string }
  | { key: string; kind: "number"; numberValue: number; unit?: string }
  | { key: string; kind: "boolean"; booleanValue: boolean }
  | { key: string; kind: "date"; dateValue: string }
  | { key: string; kind: "single_select"; optionValues: [string] }
  | { key: string; kind: "multi_select"; optionValues: string[] };

export type AccessoryArticle = {
  id: string;
  inventoryNumber: string;
  manufacturer: string;
  articleNumber?: string;
  name: string;
  category: string;
  trackingMode: AccessoryTrackingMode;
  description?: string;
  ean?: string;
  manufacturerStatus: AccessoryManufacturerStatus;
  articleType: AccessoryArticleType;
  subtype: string;
  gauges: string[];
  scale?: string;
  listPrice?: string;
  packageQuantity: number;
  stockUnit: string;
  minimumStock: number;
  inventoryStrategy: AccessoryInventoryStrategy;
  manufacturerUrl?: string;
  productUrl?: string;
  alternativeNumbers: string[];
  keywords: string[];
  compatibilityNotes?: string;
  internalNotes?: string;
  archived: boolean;
  attributes: AccessoryAttributeValue[];
  primaryImageUrl?: string;
  createdAt: string;
  updatedAt: string;
};

export type AccessoryArticleWriteInput = {
  manufacturer: string;
  articleNumber?: string;
  name: string;
  category?: string;
  trackingMode?: AccessoryTrackingMode;
  description?: string;
  ean?: string;
  manufacturerStatus?: AccessoryManufacturerStatus;
  articleType: AccessoryArticleType;
  subtype: string;
  gauges?: string[];
  scale?: string;
  listPrice: string;
  packageQuantity: number;
  stockUnit: string;
  minimumStock?: number;
  inventoryStrategy: AccessoryInventoryStrategy;
  manufacturerUrl?: string;
  productUrl?: string;
  alternativeNumbers?: string[];
  keywords?: string[];
  compatibilityNotes?: string;
  internalNotes?: string;
  archived?: boolean;
  attributes?: AccessoryAttributeValue[];
};

export type AccessoryArticleListQuery = {
  query?: string;
  articleTypes?: AccessoryArticleType[];
  manufacturer?: string;
  gauges?: string[];
  statuses?: AccessoryArticleStatus[];
  locationId?: string;
  sort?: AccessoryArticleSort;
  direction?: AccessorySortDirection;
};

export type AccessoryArticleListItem = {
  id: string;
  inventoryNumber: string;
  primaryImageUrl?: string;
  manufacturer: string;
  articleNumber: string;
  name: string;
  articleType: AccessoryArticleType;
  subtype: string;
  gauges: string[];
  listPrice?: string;
  inventoryStrategy: AccessoryInventoryStrategy;
  archived: boolean;
  owned: number;
  available: number;
  reserved: number;
  installed: number;
  locationNames: string[];
  hasUsageHistory: boolean;
  careHintCount: number;
  updatedAt: string;
  attributes: AccessoryAttributeValue[];
};

export type AccessoryOverviewMetrics = {
  articleCount: number;
  articleTypeCount: number;
  available: number;
  locationCount: number;
  reserved: number;
  installed: number;
  careHintCount: number;
};

export type AccessoryArticleFilterOptions = {
  manufacturers: string[];
  articleTypes: AccessoryArticleType[];
  gauges: string[];
  storageLocations: Array<{ id: string; name: string }>;
};

export type AccessoryArticleListResult = {
  items: AccessoryArticleListItem[];
  metrics: AccessoryOverviewMetrics;
  filters: AccessoryArticleFilterOptions;
};

export type AccessoryDuplicateCheckInput = {
  manufacturer: string;
  articleNumber: string;
  excludeId?: string;
};

export type AccessoryDuplicateCandidate = {
  id: string;
  manufacturer: string;
  articleNumber: string;
  name: string;
  articleType: AccessoryArticleType;
  subtype: string;
};

export type AccessoryDuplicateCheckResult = { candidates: AccessoryDuplicateCandidate[] };

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

export type AccessoryStockMovementType =
  | "purchase"
  | "adjustment"
  | "transfer_in"
  | "transfer_out"
  | "individualization"
  | "installation"
  | "removal";

export type AccessoryStockMovement = {
  id: string;
  productId: string;
  locationId: string;
  movementType: AccessoryStockMovementType;
  quantity: number;
  sourceType?: string;
  sourceId?: string;
  actor?: string;
  note?: string;
  createdAt: string;
};

export type AccessoryStockTransferInput = {
  fromLocationId: string;
  toLocationId: string;
  quantity: number;
  note?: string;
};

export type AccessoryAsset = {
  id: string;
  productId: string;
  purchaseId?: string;
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

export type AccessoryIndividualizationInput = {
  locationId: string;
  asset: AccessoryAssetInput;
};

export type AccessoryPurchase = {
  id: string;
  productId: string;
  storageLocationId?: string;
  quantity: number;
  purchasedAt: string;
  supplier?: string;
  unitPrice?: string;
  currency?: string;
  invoiceNumber?: string;
  warrantyUntil?: string;
  bookToStock: boolean;
  notes?: string;
  createdBy?: string;
  createdAt: string;
  updatedAt: string;
};

export type AccessoryPurchaseInput = {
  purchasedAt: string;
  supplier?: string;
  quantity: number;
  unitPrice?: string;
  currency?: string;
  invoiceNumber?: string;
  warrantyUntil?: string;
  storageLocationId?: string;
  bookToStock?: boolean;
  notes?: string;
};

export type AccessoryDocumentCategory =
  | "invoice"
  | "delivery_note"
  | "manual"
  | "data_sheet"
  | "floor_plan"
  | "image"
  | "other";

export type AccessoryDocument = {
  id: string;
  productId: string;
  fileName: string;
  originalName: string;
  description?: string;
  category: AccessoryDocumentCategory;
  mimeType: string;
  sizeBytes: number;
  isPrimary: boolean;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
};

export type AccessoryDocumentUploadInput = {
  file: File;
  category: AccessoryDocumentCategory;
  description?: string;
  isPrimary?: boolean;
};

export type AccessoryDocumentImportURLInput = {
  url: string;
  idempotencyKey: string;
  title?: string;
  description?: string;
  isPrimary?: boolean;
};

export type AccessoryDocumentUpdateInput = {
  category: AccessoryDocumentCategory;
  description?: string;
  isPrimary?: boolean;
};

export type AllocationTarget = { vehicleId: string };

export type AccessoryTechnicalPlacement = {
  placement?: string;
  digitalAddress?: string;
  decoderOutput?: string;
  connection?: string;
  wiringNotes?: string;
};

export type AccessoryReservation = AllocationTarget & AccessoryTechnicalPlacement & {
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

export type AccessoryReservationInput = AllocationTarget & AccessoryTechnicalPlacement & {
  productId: string;
  assetId?: string;
  locationId: string;
  quantity: number;
  note?: string;
};

export type AccessoryInstallation = AllocationTarget & AccessoryTechnicalPlacement & {
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

export type AccessoryInstallationInput = AllocationTarget & AccessoryTechnicalPlacement & {
  reservationId?: string;
  productId: string;
  assetId?: string;
  sourceLocationId: string;
  quantity: number;
  condition?: AccessoryCondition;
  notes?: string;
};

export type AccessoryInstallationRemovalInput =
  | { disposition: "stored"; storageLocationId: string; notes?: string }
  | {
    disposition: Exclude<AccessoryRemovalDisposition, "stored">;
    storageLocationId?: never;
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

export type AccessoryUsageEventType = "reservation" | "installation" | "condition_changed" | "removal";

type AccessoryUsageEventBase = {
  id: string;
  productId: string;
  reservationId?: string;
  installationId?: string;
  assetId?: string;
  locationId?: string;
  quantity: number;
  vehicleId?: string;
  placement?: string;
  digitalAddress?: string;
  decoderOutput?: string;
  connection?: string;
  wiringNotes?: string;
  actor?: string;
  occurredAt: string;
};

export type AccessoryUsageEvent = AccessoryUsageEventBase & (
  | { type: "reservation"; status?: AccessoryReservationStatus }
  | { type: "installation"; condition?: AccessoryCondition }
  | { type: "condition_changed"; previousCondition?: AccessoryCondition; condition?: AccessoryCondition }
  | { type: "removal"; removalDisposition?: AccessoryRemovalDisposition }
);

export type AccessoryUsageHistory = {
  productId: string;
  events: AccessoryUsageEvent[];
};
