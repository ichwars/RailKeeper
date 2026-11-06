import type { components } from "../../generated/api";

// Form controls also hold incomplete values before the existing save validation runs.
export type VehicleFunctionDraftInput = Omit<
  components["schemas"]["VehicleFunctionInput"], "functionType" | "mode" | "flagsVersion"
> & {
  functionType?: string;
  mode?: string;
  flagsVersion?: number;
};

export type VehicleFunctionDraftChange = Omit<
  components["schemas"]["VehicleFunctionChange"], "value" | "delete"
> & {
  value?: VehicleFunctionDraftInput;
  delete?: boolean;
};

export type VehicleForm = Omit<
  components["schemas"]["CreateVehicleRequest"], "maximumSpeedKmh" | "functionChanges"
> & {
  maximumSpeedKmh?: number;
  functionChanges?: VehicleFunctionDraftChange[];
};
