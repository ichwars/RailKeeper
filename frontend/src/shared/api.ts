import { createAccessoriesAPI } from "./apiAccessories";
import { createAuthenticationAPI } from "./apiAuthentication";
import { createBugReportAPI } from "./apiBugReports";
import { createDataTransferAPI } from "./apiDataTransfer";
import { createDecoderConfigurationsAPI } from "./apiDecoderConfigurations";
import { createDigitalCentersAPI } from "./apiDigitalCenters";
import { createExhibitionAPI } from "./apiExhibition";
import { createInventoryAPI } from "./apiInventory";
import { createMasterDataAPI } from "./apiMasterData";
import { createSystemAPI } from "./apiSystem";
import { request } from "./apiTransport";
import { createValuationsAPI } from "./apiValuations";
import { createVehicleCVAPI } from "./apiVehicleCV";
import { createVehicleCatalogAPI } from "./apiVehicleCatalog";
import { createWorkshopAPI } from "./apiWorkshop";

export type * from "../features/digitalCenters/digitalCenterModel";
export type * from "../features/importExport/dataTransferModel";
export * from "./apiAccessories";
export type * from "./apiAuthModels";
export * from "./apiBugReports";
export type * from "./apiDecoderConfigurations";
export type * from "./apiDigitalCenterModels";
export type * from "./apiExhibitionModels";
export type * from "./apiInventoryModels";
export type * from "./apiMasterDataModels";
export type * from "./apiSystemModels";
export { ApiError,captureSessionGuard } from "./apiTransport";
export type * from "./apiValuations";
export type * from "./apiVehicleCatalog";

export const api = {
  ...createValuationsAPI(request),
  ...createDecoderConfigurationsAPI(request),
  ...createVehicleCatalogAPI(request),
  ...createBugReportAPI(request),
  ...createAccessoriesAPI(request),
  ...createInventoryAPI(request),
  ...createDataTransferAPI(request),
  ...createDigitalCentersAPI(request),
  ...createAuthenticationAPI(request),
  ...createSystemAPI(request),
  ...createExhibitionAPI(request),
  ...createWorkshopAPI(request),
  ...createVehicleCVAPI(request),
  ...createMasterDataAPI(request),
};
