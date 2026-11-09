import type {
  VehicleForm, VehicleFunctionDraftChange, VehicleFunctionDraftInput
} from "../features/vehicles/vehicleFormModel";
import type { components } from "../generated/api";

export type OverviewValuation = {
  vehicleListValue: string;
  vehiclePurchaseValue: string;
  accessoryListValue: string;
  accessoryPurchaseCost: string;
  excludedForeignCurrencyPurchases: number;
};

export type VehicleSetSummary = components["schemas"]["VehicleSetSummary"];

export type VehicleSetMainImageMode = components["schemas"]["VehicleSetMainImage"]["source"];

export type VehicleSetMainImage = components["schemas"]["VehicleSetMainImage"];

export type VehicleSetImage = components["schemas"]["VehicleSetImage"];

export type ECoSDecoderProfileEvidence = components["schemas"]["ECoSDecoderProfileEvidence"];

export type Vehicle = components["schemas"]["Vehicle"];

export type VehicleCatalogAssignment = components["schemas"]["VehicleCatalogAssignment"];

export type VehicleCatalogBackfillReport = components["schemas"]["VehicleCatalogBackfillReport"];

export type VehicleSetInput = Pick<CreateVehicleRequest,
  | "name"
  | "manufacturer"
  | "articleNumber"
  | "articleSourceUrl"
  | "gauge"
  | "epoch"
  | "railwayCompany"
  | "category"
  | "gattung"
  | "description"
  | "ean"
  | "productionPeriod"
  | "listPrice"
  | "acquisitionType"
  | "acquiredFrom"
  | "purchasePrice"
  | "purchaseDate"
  | "storageLocation"
  | "storageDetails"
  | "condition"
  | "conditionDetails"
  | "packaging"
>;

export type CreateVehicleSetRequest = {
  creationId?: string;
  set: VehicleSetInput;
  members: CreateVehicleRequest[];
};

export type VehicleSet = components["schemas"]["VehicleSet"];

export type VehicleExternalMapping = components["schemas"]["VehicleExternalMapping"];

export type VehicleExternalMappingInput = components["schemas"]["VehicleExternalMappingInput"];

export type VehicleImage = components["schemas"]["VehicleImage"];

export type VehicleImageCreation = components["schemas"]["VehicleImageCreation"];

export type VehicleImageInput = components["schemas"]["VehicleImageInput"];

export type VehicleImageImportInput = components["schemas"]["VehicleImageImportInput"];

export type VehicleAttachment = components["schemas"]["VehicleAttachment"];

export type VehicleAttachmentUpdateInput = components["schemas"]["VehicleAttachmentUpdateRequest"];

export type VehicleMaintenance = components["schemas"]["VehicleMaintenance"];

export type VehicleMaintenanceInput = components["schemas"]["VehicleMaintenanceInput"];

export type VehicleSparePart = components["schemas"]["VehicleSparePart"];

export type VehicleSparePartInput = components["schemas"]["VehicleSparePartInput"];

export type ECoSFunctionValue = { index: number; description: number; momentary: boolean };

export type ECoSFunctionSnapshot = { objectId: number; host: string; readAt: string; functions: ECoSFunctionValue[] };

export type VehicleFunction = components["schemas"]["VehicleFunction"];

export type VehicleFunctionInput = VehicleFunctionDraftInput;

export type VehicleFunctionChange = VehicleFunctionDraftChange;

export type VehicleCVValue = components["schemas"]["VehicleCVValue"];

export type VehicleCVValueHistory = components["schemas"]["VehicleCVValueHistory"];

export type VehicleCVValueInput = components["schemas"]["VehicleCVValueInput"];

export type VehicleCVChange = components["schemas"]["VehicleCVChange"];

export type VehicleCVFile = components["schemas"]["VehicleCVFile"];

export type VehicleCVFilePreview = components["schemas"]["VehicleCVFilePreview"];

export type CreateVehicleRequest = VehicleForm;

export type VehicleCVWorkspaceComparison = Required<Pick<VehicleCVChange, "workspaceId" | "expectedWorkspaceRevision">>;

export type VehicleCVDeletionScope = VehicleCVWorkspaceComparison;
