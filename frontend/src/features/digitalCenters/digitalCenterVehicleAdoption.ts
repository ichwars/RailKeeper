import type { VehicleExternalMappingInput } from "../../shared/api";
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
