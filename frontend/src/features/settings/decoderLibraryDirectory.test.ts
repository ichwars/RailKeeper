import { expect, it } from "vitest";
import { bundledDecoderInventory as inventory, bundledDecoderProfiles as profiles } from "../../shared/decoderLibraryModel";
import { directoryModels } from "./decoderLibraryDirectory";
const filter = { manufacturer: "", family: "", kind: "", coverage: "", search: "" };
it("scopes model and article search by manufacturer and family", () => {
  expect(directoryModels(inventory, profiles, { ...filter, manufacturer: "esu", search: "151" })).toHaveLength(58);
  expect(directoryModels(inventory, profiles, { ...filter, search: "CV51" }).map(m => m.name).sort()).toEqual([
    "LokPilot 5 micro DCC", "DH18A", "DH05C", "DH10C", "DH16A", "DH12A", "DH14B", "DH21A", "DH21B", "DH22A", "DH22B", "DH24A", "DHSP10A", "STANDARD+ V2", "N45", "SmartDecoder XP 5.1 PluX22", "IntelliDrive 76420", "10882 · PluX16-Decoder (NEM 658)"
  ].sort());
  expect(directoryModels(inventory, profiles, { ...filter, manufacturer: "zimo", kind: "sound" })).toHaveLength(84);
  expect(directoryModels(inventory, profiles, { ...filter, coverage: "reviewed" })).toHaveLength(19);
  expect(directoryModels(inventory, profiles, { ...filter, family: "dh-pd", coverage: "reviewed" })).toEqual([]);
});

it("keeps shared legacy article numbers searchable without collapsing hardware generations", () => {
  const matches = directoryModels(inventory, profiles, { ...filter, manufacturer: "lenz", search: "10104" });
  expect(matches.map(model => model.name)).toEqual(["LF100", "LF100XF", "LF101XF"]);
  expect(matches.every(model => !model.profileId && !model.identification)).toBe(true);
  expect(directoryModels(inventory, profiles, { ...filter, search: "MS501R" })).toEqual([]);
});


it("keeps documentary Roco rules searchable without adding identity evidence or generic registers", () => {
  const roco = { ...filter, manufacturer: "roco", coverage: "reviewed" };
  expect(directoryModels(inventory, profiles, { ...roco, search: "CV124" }).map(m => m.name))
    .toEqual(["10882 · PluX16-Decoder (NEM 658)"]);
  expect(directoryModels(inventory, profiles, { ...roco, search: "CV19" })).toEqual([]);
  expect(directoryModels(inventory, profiles, { ...roco, search: "CV152" })).toEqual([]);
});
