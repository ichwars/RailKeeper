export type DataTransferArea = "vehicles" | "accessories" | "exhibitionLists";
export type DataTransferDirection = "import" | "export";
export type DataTransferFormat = "csv" | "railkeeper-json";
export type DataTransferJobState =
  | "draft"
  | "reading"
  | "review_required"
  | "ready"
  | "running"
  | "completed"
  | "completed_with_warnings"
  | "failed"
  | "cancelled";
export type DataTransferJobStage =
  | "created"
  | "snapshot"
  | "preview"
  | "review"
  | "completed"
  | "failed"
  | "cancelled";
export type DataTransferIssueSeverity = "warning" | "error";
export type DataTransferIssueResolution = "replace" | "merge" | "copy" | "skip" | "use_existing" | "create" | "link";
export type DataTransferProposedAction = "create" | "replace" | "use_existing" | "copy";
export type DataTransferProposedResolution = DataTransferIssueResolution | "replace_or_copy";
export type DataTransferOptions = Record<string, unknown>;
export type DataTransferCSVMappingOrigin = "alias" | "profile" | "manual" | "ignored" | "unmapped";

export type DataTransferCSVColumnMapping = {
  index: number;
  sourceHeader: string;
  normalizedHeader: string;
  targetField: string;
  origin: DataTransferCSVMappingOrigin;
};

export type DataTransferCSVMappingInput = {
  columns: DataTransferCSVColumnMapping[];
  saveToProfile: boolean;
};

export type VehicleTransferField = {
  key: string;
  labelDE: string;
  labelEN: string;
  kind: "string" | "integer" | "boolean";
  aliases?: string[];
};

export type DataTransferProfile = {
  id: string;
  name: string;
  direction: DataTransferDirection;
  format: DataTransferFormat;
  areas: DataTransferArea[];
  options: DataTransferOptions;
  enabled: boolean;
  createdByUserId: string;
  lastUsedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type DataTransferProfileInput = {
  name: string;
  direction: DataTransferDirection;
  format: DataTransferFormat;
  areas: DataTransferArea[];
  options?: DataTransferOptions;
};

export type DataTransferJobInput = {
  profileId: string;
};

export type DataTransferJob = {
  id: string;
  profileId: string;
  profileName: string;
  direction: DataTransferDirection;
  format: DataTransferFormat;
  areas: DataTransferArea[];
  options: DataTransferOptions;
  state: DataTransferJobState;
  stage: DataTransferJobStage;
  sourceName: string;
  sourceSha256: string;
  packageVersion: number;
  revision: number;
  totalRecords: number;
  readyRecords: number;
  warningRecords: number;
  errorRecords: number;
  preview: Record<string, unknown>;
  createdByUserId: string;
  confirmedByUserId: string;
  confirmedAt: string;
  completedAt: string;
  resultMessage: string;
  createdAt: string;
  updatedAt: string;
};

export type DataTransferJobFilter = {
  profileId?: string;
  direction?: DataTransferDirection;
  states?: DataTransferJobState[];
  limit?: number;
};

export type DataTransferArtifact = {
  id: string;
  jobId: string;
  relativePath: string;
  displayName: string;
  mimeType: string;
  sizeBytes: number;
  sha256: string;
  deletedAt: string;
  createdAt: string;
};

export type DataTransferExportResult = {
  job: DataTransferJob;
  artifact: DataTransferArtifact;
  openFolderAvailable: boolean;
};

export type DataTransferIssue = {
  id: string;
  jobId: string;
  area: DataTransferArea;
  recordKey: string;
  rowNumber: number | null;
  field: string;
  severity: DataTransferIssueSeverity;
  code: string;
  message: string;
  proposedResolution: DataTransferProposedResolution | "";
  selectedResolution: DataTransferIssueResolution | "";
  createdAt: string;
  updatedAt: string;
};

export type DataTransferPreviewRecord = {
  area: DataTransferArea;
  recordKey: string;
  rowNumber?: number;
  classification: "ready" | "warning" | "error";
  proposedAction: DataTransferProposedAction;
  targetId?: string;
  targetUpdatedAt?: string;
  targetFingerprint?: string;
  data: Record<string, unknown>;
};

export type DataTransferVehicleSetMember = {
  vehicleId: string;
  vehicleInventoryNumber: string;
  position: number;
  label?: string;
};

export type DataTransferVehicleSet = {
  id?: string;
  inventoryNumber: string;
  name: string;
  manufacturer: string;
  articleNumber: string;
  articleSourceUrl: string;
  gauge: string;
  epoch: string;
  railwayCompany: string;
  category: string;
  gattung: string;
  description: string;
  ean: string;
  productionPeriod: string;
  listPrice: string;
  acquisitionType: string;
  acquiredFrom: string;
  purchasePrice: string;
  purchaseDate: string;
  storageLocation: string;
  storageDetails: string;
  condition: string;
  conditionDetails: string;
  packaging: string;
  members: DataTransferVehicleSetMember[];
  createdAt?: string;
  updatedAt?: string;
};

export type DataTransferVehicleSetDiagnostic = {
  rowNumber: number;
  field: string;
  code: string;
};

export type DataTransferVehicleSetPreview = {
  recordKey: string;
  classification: "ready" | "warning" | "error";
  proposedAction: DataTransferProposedAction;
  targetId?: string;
  targetUpdatedAt?: string;
  targetFingerprint?: string;
  memberRecordKeys: string[];
  rowNumbers?: number[];
  diagnostics?: DataTransferVehicleSetDiagnostic[];
  data: DataTransferVehicleSet;
};

export type DataTransferPreview = {
  job: DataTransferJob;
  records: DataTransferPreviewRecord[];
  issues: DataTransferIssue[];
  totalRecords: number;
  readyRecords: number;
  warningRecords: number;
  errorRecords: number;
  csvMapping?: DataTransferCSVColumnMapping[];
  vehicleFields?: VehicleTransferField[];
  vehicleSets?: DataTransferVehicleSetPreview[];
};

export type DataTransferSummary = {
  openJobs: number;
  selectedRecords: number;
  lastExportAt: string;
  artifactCount: number;
  artifactBytes: number;
  openFolderAvailable: boolean;
  artifactDirectory: string;
};

export type DataTransferJobDetails = {
  job: DataTransferJob;
  issues: DataTransferIssue[];
  artifacts: DataTransferArtifact[];
};

export const emptyDataTransferSummary: DataTransferSummary = {
  openJobs: 0,
  selectedRecords: 0,
  lastExportAt: "",
  artifactCount: 0,
  artifactBytes: 0,
  openFolderAvailable: false,
  artifactDirectory: ""
};
