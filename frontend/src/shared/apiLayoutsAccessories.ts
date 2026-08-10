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
  maxGradePercent?: number | null;
  minimumTrackClearanceMm?: number | null;
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
  maxGradePercent?: number | null;
  minimumTrackClearanceMm?: number | null;
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
  widthMm?: number;
  heightMm?: number;
  archived?: boolean;
};

export type LayoutUnitUpdateInput = LayoutUnitInput & { expectedVersion: number };

export type LayoutUnitPortKind = "track" | "power" | "digital" | "feedback" | "accessory" | "other";

export type LayoutUnitPort = {
  id: string;
  layoutUnitId: string;
  name: string;
  kind: LayoutUnitPortKind;
  interfaceKey: string;
  xMm: number;
  yMm: number;
  directionDegrees: number;
  notes?: string;
  version: number;
  archived: boolean;
  createdAt: string;
  updatedAt: string;
};

export type LayoutUnitPortInput = {
  name: string;
  kind: LayoutUnitPortKind;
  interfaceKey: string;
  xMm: number;
  yMm: number;
  directionDegrees: number;
  notes?: string;
  archived?: boolean;
};

export type LayoutUnitPortUpdateInput = LayoutUnitPortInput & { expectedVersion: number };

export type LayoutTechnicalPositionKind =
  | "turnout"
  | "signal"
  | "feedback"
  | "decoder"
  | "lighting"
  | "power"
  | "sensor"
  | "other";

export type LayoutTechnicalPosition = {
  id: string;
  layoutUnitId: string;
  label: string;
  kind: LayoutTechnicalPositionKind;
  positionXMm: number;
  positionYMm: number;
  rotationDegrees: number;
  productId?: string;
  description?: string;
  version: number;
  archived: boolean;
  createdAt: string;
  updatedAt: string;
};

export type LayoutTechnicalPositionInput = {
  label: string;
  kind: LayoutTechnicalPositionKind;
  positionXMm: number;
  positionYMm: number;
  rotationDegrees?: number;
  productId?: string;
  description?: string;
  archived?: boolean;
};

export type LayoutTechnicalPositionUpdateInput = LayoutTechnicalPositionInput & { expectedVersion: number };

export type LayoutTwinStatus = "planned" | "reserved" | "installed" | "maintenance_due" | "defective";

export type LayoutTwinAllocation = {
  id: string;
  productId: string;
  inventoryNumber: string;
  manufacturer: string;
  articleNumber?: string;
  productName: string;
  quantity: number;
  reservationStatus?: AccessoryReservationStatus;
  installationCondition?: AccessoryCondition;
  placement?: string;
  digitalAddress?: string;
  decoderOutput?: string;
  connection?: string;
  wiringNotes?: string;
  note?: string;
};

export type LayoutTwinPosition = {
  id: string;
  layoutUnitId: string;
  label: string;
  kind: LayoutTechnicalPositionKind;
  localXMm: number;
  localYMm: number;
  localRotationDegrees: number;
  globalXMm: number;
  globalYMm: number;
  rotationDegrees: number;
  productId?: string;
  inventoryNumber?: string;
  manufacturer?: string;
  articleNumber?: string;
  productName?: string;
  description?: string;
  version: number;
  outsideOutline: boolean;
  statuses: LayoutTwinStatus[];
  reservations: LayoutTwinAllocation[];
  installations: LayoutTwinAllocation[];
};

export type LayoutTwinUnit = {
  id: string;
  name: string;
  kind: LayoutUnitKind;
  positionXMm: number;
  positionYMm: number;
  rotationDegrees: number;
  version: number;
  localOutline: Array<{ xMm: number; yMm: number }>;
  outline: Array<{ xMm: number; yMm: number }>;
  positions: LayoutTwinPosition[];
};

export type LayoutTwin = {
  layoutId: string;
  configurationId?: string;
  configurationName?: string;
  unitId?: string;
  bounds: { minXMm: number; minYMm: number; widthMm: number; heightMm: number };
  hasGeometry: boolean;
  units: LayoutTwinUnit[];
  warnings: Array<{ code: "outline_fallback" | "missing_geometry"; unitId?: string }>;
};

export type LayoutTwinSelection = { configurationId?: string; unitId?: string };
export type LayoutUnitOutline = {
  layoutUnitId: string;
  points: Array<{ xMm: number; yMm: number }>;
  version: number;
};
export type LayoutUnitOutlineUpdateInput = {
  points: Array<{ xMm: number; yMm: number }>;
  expectedVersion: number;
};

export type ConfigurationUnit = {
  unitId: string;
  planRevisionId?: string;
  positionXMm: number;
  positionYMm: number;
  rotationDegrees: number;
  sortOrder: number;
};

export type ConfigurationUnitInput = {
  unitId: string;
  planRevisionId?: string;
  positionXMm?: number;
  positionYMm?: number;
  rotationDegrees?: number;
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
  units: ConfigurationUnitInput[];
};

export type LayoutConfigurationUpdateInput = LayoutConfigurationInput & { expectedVersion: number };

export type ModulePortConnection = {
  unitAId: string;
  unitAName: string;
  portAId: string;
  portAName: string;
  unitBId: string;
  unitBName: string;
  portBId: string;
  portBName: string;
};

export type ModulePortIssue = {
  code: "open_port" | "incompatible_port";
  unitIds: string[];
  unitNames: string[];
  portIds: string[];
  portNames: string[];
};

export type ModulePortAnalysis = {
  connections: ModulePortConnection[];
  issues: ModulePortIssue[];
};

export type ConfigurationUnitSnapPreviewInput = {
  unitId: string;
  positionXMm: number;
  positionYMm: number;
  rotationDegrees: number;
};

export type ModulePortSnapResult = {
  snapped: boolean;
  pose: { positionXMm: number; positionYMm: number; rotationDegrees: number };
  movingPortId?: string;
  targetUnitId?: string;
  targetPortId?: string;
  distanceMm?: number;
};

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

export type TrackGeometryKind = "straight" | "curve" | "turnout" | "crossing";
export type TrackGeometryStatus = "draft" | "verified" | "retired";
export type TrackPoint = { xMm: number; yMm: number };
export type TrackPort = TrackPoint & { id: string; directionDegrees: number };
export type TrackRoute = { id: string; points: TrackPoint[] };
export type TrackGeometry = { schemaVersion: number; ports: TrackPort[]; routes: TrackRoute[] };
export type TrackGeometryDefinition = {
  id: string;
  libraryId: string;
  articleNumber: string;
  name: string;
  kind: TrackGeometryKind;
  lengthMm: number;
  geometry: TrackGeometry;
  sourceUrl: string;
  status: TrackGeometryStatus;
  createdAt: string;
};
export type PlanTrackObject = {
  id: string;
  lineageId: string;
  revisionId: string;
  geometryId: string;
  geometry: TrackGeometryDefinition;
  positionXMm: number;
  positionYMm: number;
  rotationDegrees: number;
  elevationStartMm: number;
  elevationEndMm: number;
  version: number;
  createdAt: string;
  updatedAt: string;
};
export type TrackPlan = {
  revisionId: string;
  status: PlanRevisionStatus;
  objects: PlanTrackObject[];
};
export type TrackPlanConnection = {
  objectAId: string;
  portAId: string;
  objectBId: string;
  portBId: string;
};
export type TrackPlanIssueCode = "open_end" | "incompatible_connection" | "overlap" | "broken_geometry"
  | "elevation_mismatch" | "grade_limit_exceeded" | "insufficient_clearance";
export type TrackPlanIssue = {
  code: TrackPlanIssueCode;
  severity: "warning" | "error";
  objectIds: string[];
  portIds?: string[];
  elevationDifferenceMm?: number;
  gradePercent?: number;
  gradeLimitPercent?: number;
  clearanceMm?: number;
  clearanceLimitMm?: number;
  intersectionXMm?: number;
  intersectionYMm?: number;
};
export type TrackBOMLine = {
  geometryId: string;
  libraryId: string;
  articleNumber: string;
  name: string;
  quantity: number;
};
export type TrackGrade = {
  objectId: string;
  elevationStartMm: number;
  elevationEndMm: number;
  lengthMm: number;
  gradePercent: number;
};
export type TrackMaterialStatus = {
  geometryId: string;
  manufacturer: string;
  articleNumber: string;
  name: string;
  requiredQuantity: number;
  productIds: string[];
  inventoryNumbers: string[];
  physicalQuantity: number;
  reservedQuantity: number;
  availableQuantity: number;
  missingQuantity: number;
};
export type TrackPlanAnalysis = {
  revisionId: string;
  status: PlanRevisionStatus;
  connections: TrackPlanConnection[];
  issues: TrackPlanIssue[];
  bom: TrackBOMLine[];
  grades: TrackGrade[];
  materials: TrackMaterialStatus[];
  reservations: TrackPlanObjectReservation[];
};
export type TrackPlanObjectChangeType = "added" | "removed" | "changed";
export type TrackPlanObjectChange = {
  type: TrackPlanObjectChangeType;
  lineageId: string;
  before?: PlanTrackObject;
  after?: PlanTrackObject;
};
export type TrackPlanMaterialDelta = {
  geometryId: string;
  libraryId: string;
  articleNumber: string;
  name: string;
  baseQuantity: number;
  currentQuantity: number;
  delta: number;
};
export type TrackPlanIssueChange = {
  code: TrackPlanIssueCode;
  severity: "warning" | "error";
  lineageIds: string[];
  portIds?: string[];
};
export type TrackPlanChangePreview = {
  revisionId: string;
  baseRevisionId: string;
  objectChanges: TrackPlanObjectChange[];
  materialDeltas: TrackPlanMaterialDelta[];
  issues: { added: TrackPlanIssueChange[]; resolved: TrackPlanIssueChange[] };
  affectedConfigurations: { id: string; name: string }[];
};
export type TrackPlanReservationInput = {
  trackObjectId: string;
  productId: string;
  locationId: string;
  assetId?: string;
  expectedObjectVersion: number;
};
export type ReserveTrackPlanMaterialsInput = {
  confirmed: boolean;
  items: TrackPlanReservationInput[];
};
export type TrackPlanObjectReservation = {
  trackObjectId: string;
  reservation: AccessoryReservation;
};
export type TrackPlanReservationBatch = {
  revisionId: string;
  reservations: TrackPlanObjectReservation[];
  materials: TrackMaterialStatus[];
};
export type CreatePlanTrackObjectInput = {
  geometryId: string;
  positionXMm: number;
  positionYMm: number;
  rotationDegrees: number;
  elevationStartMm: number;
  elevationEndMm: number;
};
export type UpdatePlanTrackObjectInput = {
  positionXMm: number;
  positionYMm: number;
  rotationDegrees: number;
  elevationStartMm: number;
  elevationEndMm: number;
  expectedVersion: number;
};

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

export type AccessoryDocumentUpdateInput = {
  category: AccessoryDocumentCategory;
  description?: string;
  isPrimary?: boolean;
};

export type AllocationTarget =
  | { vehicleId: string; layoutId?: never; layoutUnitId?: never }
  | { vehicleId?: never; layoutId: string; layoutUnitId?: never }
  | { vehicleId?: never; layoutId?: never; layoutUnitId: string };

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
  technicalPositionId?: string;
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
  technicalPositionId?: string;
  quantity: number;
  note?: string;
};

export type AccessoryInstallation = AllocationTarget & AccessoryTechnicalPlacement & {
  id: string;
  productId: string;
  assetId?: string;
  sourceLocationId: string;
  technicalPositionId?: string;
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
  technicalPositionId?: string;
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
  layoutId?: string;
  layoutUnitId?: string;
  technicalPositionId?: string;
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

type RequestOptions = { retries?: number; timeoutMs?: number };
type APIRequest = <T>(path: string, init?: RequestInit, options?: RequestOptions) => Promise<T>;

export function createLayoutsAccessoriesAPI(request: APIRequest) {
  const productQuery = (productId?: string) => productId ? `?productId=${encodeURIComponent(productId)}` : "";
  const twinQuery = (selection: LayoutTwinSelection) => {
    const query = new URLSearchParams();
    if (selection.configurationId) query.set("configurationId", selection.configurationId);
    if (selection.unitId) query.set("unitId", selection.unitId);
    const encoded = query.toString();
    return encoded ? `?${encoded}` : "";
  };
  return {
    layouts: () => request<Layout[]>("/layouts"),
    createLayout: (input: LayoutInput) => request<Layout>("/layouts", json("POST", input)),
    layout: (id: string) => request<Layout>(`/layouts/${encodeURIComponent(id)}`),
    layoutTwin: (id: string, selection: LayoutTwinSelection = {}) =>
      request<LayoutTwin>(`/layouts/${encodeURIComponent(id)}/twin${twinQuery(selection)}`),
    updateLayout: (id: string, input: LayoutUpdateInput) =>
      request<Layout>(`/layouts/${encodeURIComponent(id)}`, json("PUT", input)),
    layoutUnits: (layoutId: string) => request<LayoutUnit[]>(`/layouts/${encodeURIComponent(layoutId)}/units`),
    createLayoutUnit: (layoutId: string, input: LayoutUnitInput) =>
      request<LayoutUnit>(`/layouts/${encodeURIComponent(layoutId)}/units`, json("POST", input)),
    updateLayoutUnit: (id: string, input: LayoutUnitUpdateInput) =>
      request<LayoutUnit>(`/layout-units/${encodeURIComponent(id)}`, json("PUT", input)),
    layoutUnitPorts: (unitId: string) =>
      request<LayoutUnitPort[]>(`/layout-units/${encodeURIComponent(unitId)}/ports`),
    createLayoutUnitPort: (unitId: string, input: LayoutUnitPortInput) =>
      request<LayoutUnitPort>(`/layout-units/${encodeURIComponent(unitId)}/ports`, json("POST", input)),
    updateLayoutUnitPort: (id: string, input: LayoutUnitPortUpdateInput) =>
      request<LayoutUnitPort>(`/layout-unit-ports/${encodeURIComponent(id)}`, json("PUT", input)),
    updateLayoutUnitOutline: (id: string, input: LayoutUnitOutlineUpdateInput) =>
      request<LayoutUnitOutline>(`/layout-units/${encodeURIComponent(id)}/outline`, json("PUT", input)),
    layoutTechnicalPositions: (unitId: string) =>
      request<LayoutTechnicalPosition[]>(`/layout-units/${encodeURIComponent(unitId)}/technical-positions`),
    createLayoutTechnicalPosition: (unitId: string, input: LayoutTechnicalPositionInput) =>
      request<LayoutTechnicalPosition>(
        `/layout-units/${encodeURIComponent(unitId)}/technical-positions`,
        json("POST", input)
      ),
    updateLayoutTechnicalPosition: (id: string, input: LayoutTechnicalPositionUpdateInput) =>
      request<LayoutTechnicalPosition>(
        `/layout-technical-positions/${encodeURIComponent(id)}`,
        json("PUT", input)
      ),
    layoutConfigurations: (layoutId: string) =>
      request<LayoutConfiguration[]>(`/layouts/${encodeURIComponent(layoutId)}/configurations`),
    createLayoutConfiguration: (layoutId: string, input: LayoutConfigurationInput) =>
      request<LayoutConfiguration>(`/layouts/${encodeURIComponent(layoutId)}/configurations`, json("POST", input)),
    updateLayoutConfiguration: (id: string, input: LayoutConfigurationUpdateInput) =>
      request<LayoutConfiguration>(`/layout-configurations/${encodeURIComponent(id)}`, json("PUT", input)),
    layoutConfigurationPortAnalysis: (id: string) =>
      request<ModulePortAnalysis>(`/layout-configurations/${encodeURIComponent(id)}/port-analysis`),
    previewLayoutConfigurationUnitSnap: (id: string, input: ConfigurationUnitSnapPreviewInput) =>
      request<ModulePortSnapResult>(
        `/layout-configurations/${encodeURIComponent(id)}/unit-snap-preview`, json("POST", input)
      ),
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
    trackGeometries: (gauge: string) =>
      request<TrackGeometryDefinition[]>(`/track-geometries?gauge=${encodeURIComponent(gauge)}`),
    trackPlan: (revisionId: string) =>
      request<TrackPlan>(`/plan-revisions/${encodeURIComponent(revisionId)}/track-plan`),
    trackPlanAnalysis: (revisionId: string) =>
      request<TrackPlanAnalysis>(`/plan-revisions/${encodeURIComponent(revisionId)}/track-analysis`),
    trackPlanChangePreview: (revisionId: string) =>
      request<TrackPlanChangePreview>(
        `/plan-revisions/${encodeURIComponent(revisionId)}/track-change-preview`
      ),
    reserveTrackPlanMaterials: (revisionId: string, input: ReserveTrackPlanMaterialsInput) =>
      request<TrackPlanReservationBatch>(
        `/plan-revisions/${encodeURIComponent(revisionId)}/track-reservations`, json("POST", input)
      ),
    createPlanTrackObject: (revisionId: string, input: CreatePlanTrackObjectInput) =>
      request<PlanTrackObject>(
        `/plan-revisions/${encodeURIComponent(revisionId)}/track-objects`, json("POST", input)
      ),
    updatePlanTrackObject: (id: string, input: UpdatePlanTrackObjectInput) =>
      request<PlanTrackObject>(`/plan-track-objects/${encodeURIComponent(id)}`, json("PUT", input)),
    deletePlanTrackObject: (id: string, expectedVersion: number) =>
      request<void>(
        `/plan-track-objects/${encodeURIComponent(id)}?expectedVersion=${encodeURIComponent(expectedVersion)}`,
        { method: "DELETE" }
      ),
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
