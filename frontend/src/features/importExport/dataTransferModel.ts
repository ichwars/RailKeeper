import type { components, paths } from "../../generated/api";

export type DataTransferArea = components["schemas"]["TransferArea"];
export type DataTransferDirection = components["schemas"]["TransferDirection"];
export type DataTransferFormat = components["schemas"]["TransferFormat"];
export type DataTransferJobState = components["schemas"]["TransferJobState"];
export type DataTransferJobStage = components["schemas"]["TransferJobStage"];
export type DataTransferIssueSeverity = components["schemas"]["TransferIssueSeverity"];
export type DataTransferIssueResolution = components["schemas"]["TransferIssueResolution"];
export type DataTransferProposedAction = components["schemas"]["TransferProposedAction"];
export type DataTransferProposedResolution = components["schemas"]["TransferProposedResolution"];
export type DataTransferOptions = components["schemas"]["DataTransferProfile"]["options"];
export type DataTransferCSVMappingOrigin = components["schemas"]["DataTransferCSVColumnMapping"]["origin"];

export type DataTransferCSVColumnMapping = components["schemas"]["DataTransferCSVColumnMapping"];

export type DataTransferCSVMappingInput = components["schemas"]["DataTransferCSVMappingInput"];

// Stored previews from older versions can omit aliases; keep their existing rendering path.
export type VehicleTransferField = Omit<components["schemas"]["VehicleTransferField"], "aliases"> & {
  aliases?: string[];
};

export type DataTransferProfile = components["schemas"]["DataTransferProfile"];

export type DataTransferProfileInput = components["schemas"]["DataTransferProfileInput"];

export type DataTransferJobInput = components["schemas"]["DataTransferImportJobInput"];

export type DataTransferJob = components["schemas"]["DataTransferJob"];

export type DataTransferJobFilter = NonNullable<
  paths["/data-transfer/jobs"]["get"]["parameters"]["query"]
>;

export type DataTransferArtifact = components["schemas"]["DataTransferArtifact"];

export type DataTransferExportResult = components["schemas"]["DataTransferExportResult"];

export type DataTransferIssue = components["schemas"]["DataTransferIssue"];

export type DataTransferPreviewRecord = components["schemas"]["DataTransferPreviewRecord"];

export type DataTransferVehicleSetMember = components["schemas"]["DataTransferVehicleSetMember"];

export type DataTransferVehicleSet = components["schemas"]["DataTransferVehicleSet"];

export type DataTransferVehicleSetDiagnostic = components["schemas"]["DataTransferVehicleSetDiagnostic"];

export type DataTransferVehicleSetPreview = components["schemas"]["DataTransferVehicleSetPreview"];

export type DataTransferPreview = Omit<components["schemas"]["DataTransferPreview"], "vehicleFields"> & {
  vehicleFields?: VehicleTransferField[];
};

export type DataTransferSummary = components["schemas"]["DataTransferSummary"];

export type DataTransferJobDetails = components["schemas"]["DataTransferJobDetails"];

export const emptyDataTransferSummary: DataTransferSummary = {
  openJobs: 0,
  selectedRecords: 0,
  lastExportAt: "",
  artifactCount: 0,
  artifactBytes: 0,
  openFolderAvailable: false,
  artifactDirectory: ""
};
