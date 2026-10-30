import { expect, it } from "vitest";
import type { MasterDataEntry } from "../../shared/api";
import { bundledDecoderInventory, bundledDecoderProfiles } from "../../shared/decoderLibraryModel";
import { translate } from "../../shared/i18n";
import { buildDigitalCenterVehicleDraft } from "../digitalCenters/digitalCenterVehicleAdoption";
import type { DigitalCenterWorkItem } from "../digitalCenters/digitalCenterModel";
import { supplementECoSFunctionDraft } from "./ecosFunctionDraft";

const item: DigitalCenterWorkItem = {
  id: "item", sessionId: "read-session", centerObjectId: "1002", vehicleId: "", name: "Neue Lok",
  decoderAddress: 3, protocol: "DCC", compareStatus: "new", stationStatus: "read",
  center: { profile: "LokPilot 5 micro DCC", cvs: [{ number: 8, value: 151 }], functionSnapshot: {
    objectId: 1002, host: "ecos.test", readAt: "2026-10-29T08:00:00Z", functions: [
      { index: 0, description: 3, momentary: false }, { index: 2, description: 7431, momentary: true }
    ] } }, railkeeper: {}, proposed: {}, conflicts: [], createdAt: "", updatedAt: ""
};
const symbol: MasterDataEntry = { id: "light", key: "light", type: "symbols", label: "Licht",
  active: true, sortOrder: 0, createdAt: "", updatedAt: "", metadata: { ecosCode: 3 } };
const library = { schemaVersion: 1, profiles: bundledDecoderProfiles, inventory: bundledDecoderInventory };
const t = (key: string, args?: Record<string, string | number>) => translate("de", key, args);

it("prepares supplied keys in the normal new-vehicle draft with exact symbols and preserved raw evidence", () => {
  const draft = buildDigitalCenterVehicleDraft(item, "ecos"), before = structuredClone(draft);
  const prepared = supplementECoSFunctionDraft(draft, [symbol], library, t);
  expect(prepared.functionValues).toEqual([
    expect.objectContaining({ functionKey: "F0", name: "Licht", symbolKey: "light", mode: "dauer" }),
    expect.objectContaining({ functionKey: "F2", mode: "moment" })
  ]);
  expect(prepared.functionValues[1].symbolKey).toBeUndefined();
  expect(prepared.functionValues[1].notes).toContain("funcexists=7431");
  expect(draft).toEqual(before);
  expect(prepared.ecosFunctions).toEqual(item.center.functionSnapshot);
});

it("does not change existing draft functions or fill other keys from guessed decoder defaults", () => {
  const draft = buildDigitalCenterVehicleDraft(item, "ecos");
  const manual = { functionKey: "F0", name: "Mein Licht", symbolKey: "hand", mode: "moment", notes: "Eigene Notiz" };
  draft.functionValues = [manual];
  draft.vehicle.decoderType = "DH18A";
  const prepared = supplementECoSFunctionDraft(draft, [symbol], library, t);
  expect(prepared.functionValues[0]).toEqual(manual);
  expect(prepared.functionValues).toHaveLength(2);
  expect(prepared.vehicle.decoderType).toBe("DH18A");
  expect(prepared.functionValues[1]).not.toHaveProperty("functionType");
});

it("keeps observed modes without inventing unknown symbols if library data is unavailable", () => {
  const prepared = supplementECoSFunctionDraft(buildDigitalCenterVehicleDraft(item, "ecos"), [], null, t);
  expect(prepared.functionValues).toHaveLength(2);
  for (const value of prepared.functionValues) expect(value.symbolKey).toBeUndefined();
});

it("rejects mismatched source objects and duplicate or invalid keys, and never modifies update drafts", () => {
  const draft = buildDigitalCenterVehicleDraft(item, "ecos");
  draft.ecosFunctions!.objectId = 1003;
  expect(supplementECoSFunctionDraft(draft, [symbol], library, t)).toBe(draft);
  draft.ecosFunctions!.objectId = 1002;
  draft.ecosFunctions!.functions.push({ ...draft.ecosFunctions!.functions[0] });
  expect(supplementECoSFunctionDraft(draft, [symbol], library, t)).toBe(draft);
  draft.ecosFunctions!.functions.pop();
  draft.mode = "update";
  expect(supplementECoSFunctionDraft(draft, [symbol], library, t)).toBe(draft);
});

it("does not treat a quarantined decoder-identity reading as a missing identity that permits profile fallback", () => {
  const draft = buildDigitalCenterVehicleDraft(item, "ecos");
  draft.cvSourceConflicts = [{ number: 8 }];
  draft.cvValues = [{ cvNumber: 257, indexCV31: 1, indexCV32: 1, value: 0, decoderProfile: "ECoS" }];
  const prepared = supplementECoSFunctionDraft(draft, [symbol], library, t);
  expect(prepared.functionValues[0]).toMatchObject({ functionKey: "F0", symbolKey: "light", mode: "dauer" });
  expect(prepared.functionValues[0].inverted).toBeUndefined();
  expect(prepared.functionValues[0].functionType).toBeUndefined();
});
