export function validVehicleCreationIdentity(value: unknown): value is string {
  return typeof value === "string" && /^[0-9a-f]{32}$/.test(value);
}

export { newRandomIdentity as newVehicleCreationIdentity } from "../../shared/randomIdentity";
