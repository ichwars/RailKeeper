import type { Vehicle, VehicleExternalMappingInput } from "../../shared/api";
import {
  ecosVehicleDraftStorageKey,
  emptyVehicle,
  type ECoSVehicleDraftPayload
} from "../vehicles/vehicleViewModel";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";

export function digitalCenterExternalMapping(
  item: DigitalCenterWorkItem
): VehicleExternalMappingInput {
  const externalId = item.centerObjectId.trim();
  if (!/^\d+$/.test(externalId)) {
    throw new Error("invalid digital-center object id");
  }
  const address = item.center.decoderAddress ?? item.decoderAddress;
  return {
    provider: "ecos",
    externalId,
    externalName: item.center.name?.trim() || item.name.trim(),
    externalAddress: address > 0 ? String(address) : "",
    externalProtocol: item.center.protocol ?? item.protocol,
    syncStatus: "linked"
  };
}

export function buildDigitalCenterVehicleDraft(
  item: DigitalCenterWorkItem
): ECoSVehicleDraftPayload {
  const mapping = digitalCenterExternalMapping(item);
  const name = mapping.externalName || `ECoS ${mapping.externalId}`;
  return {
    source: "ecos",
    mode: "create",
    sourceSummary: {
      objectId: Number(mapping.externalId),
      name,
      address: mapping.externalAddress || "",
      protocol: mapping.externalProtocol || "",
      profile: ""
    },
    vehicle: {
      ...emptyVehicle,
      name,
      category: "Lokomotive",
      digital: true,
      digitalDecoderNumber: mapping.externalAddress || ""
    },
    importedKeys: ["name", "category", "digital", "digitalDecoderNumber"],
    externalMapping: mapping,
    cvValues: [],
    functionValues: [],
    unclearFields: ["manufacturer", "gauge", "gattung"],
    returnToDigitalCenters: {
      sessionId: item.sessionId,
      objectId: mapping.externalId
    }
  };
}

export function openDigitalCenterVehicleDraft(item: DigitalCenterWorkItem) {
  const draft = buildDigitalCenterVehicleDraft(item);
  window.sessionStorage.setItem(ecosVehicleDraftStorageKey, JSON.stringify(draft));
  window.history.pushState(null, "", "/vehicles?source=ecos");
  window.dispatchEvent(new PopStateEvent("popstate"));
}

export type DigitalCenterVehicleMatchReason = "mapping" | "address" | "name";

export function digitalCenterVehicleMatchReason(
  item: DigitalCenterWorkItem,
  vehicle: Vehicle
): DigitalCenterVehicleMatchReason | null {
  if (vehicle.externalMappings?.some((mapping) => (
    mapping.provider === "ecos" && mapping.externalId === item.centerObjectId
  ))) {
    return "mapping";
  }
  if (item.decoderAddress > 0 && vehicle.digitalDecoderNumber === String(item.decoderAddress)) {
    return "address";
  }
  const sourceName = comparableVehicleIdentity(item.center.name || item.name);
  const vehicleName = comparableVehicleIdentity(vehicle.name);
  const vehicleNumber = comparableVehicleIdentity(vehicle.vehicleNumber);
  if (sourceName && (
    vehicleName === sourceName ||
    vehicleNumber === sourceName ||
    (vehicleName.length > 0 && sourceName.includes(vehicleName)) ||
    (vehicleNumber.length > 0 && sourceName.includes(vehicleNumber))
  )) {
    return "name";
  }
  return null;
}

export function rankDigitalCenterVehicleCandidates(
  item: DigitalCenterWorkItem,
  vehicles: Vehicle[],
  query: string
) {
  const normalizedQuery = comparableVehicleIdentity(query);
  return vehicles
    .filter((vehicle) => !normalizedQuery || searchableVehicleIdentity(vehicle).includes(normalizedQuery))
    .map((vehicle, index) => ({ vehicle, index }))
    .sort((left, right) => {
      const rankDifference = matchRank(item, left.vehicle) - matchRank(item, right.vehicle);
      if (rankDifference !== 0) return rankDifference;
      const inventoryDifference = left.vehicle.inventoryNumber.localeCompare(
        right.vehicle.inventoryNumber,
        undefined,
        { numeric: true, sensitivity: "base" }
      );
      if (inventoryDifference !== 0) return inventoryDifference;
      const nameDifference = left.vehicle.name.localeCompare(right.vehicle.name, undefined, {
        sensitivity: "base"
      });
      return nameDifference || left.index - right.index;
    })
    .map(({ vehicle }) => vehicle);
}

function matchRank(item: DigitalCenterWorkItem, vehicle: Vehicle) {
  const reason = digitalCenterVehicleMatchReason(item, vehicle);
  if (reason === "mapping") return 0;
  if (reason === "address") return 1;
  if (reason === "name") return 2;
  return 3;
}

function searchableVehicleIdentity(vehicle: Vehicle) {
  return [
    vehicle.inventoryNumber,
    vehicle.name,
    vehicle.manufacturer,
    vehicle.articleNumber,
    vehicle.vehicleNumber,
    vehicle.digitalDecoderNumber
  ].map(comparableVehicleIdentity).join(" ");
}

function comparableVehicleIdentity(value?: string) {
  return (value || "").toLocaleLowerCase().replace(/[^a-z0-9äöüß]+/g, "");
}
