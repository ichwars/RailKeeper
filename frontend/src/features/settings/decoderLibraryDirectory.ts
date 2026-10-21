import type { DecoderInventory, DecoderLibraryProfile, DecoderModel } from "../../shared/decoderLibraryModel";
import { decoderCVMatches } from "../../shared/decoderLibrary";

export type DirectoryFilter = { manufacturer: string; family: string; kind: string; coverage: string; search: string };

export function directoryModels(inventory: DecoderInventory, profiles: DecoderLibraryProfile[], filter: DirectoryFilter) {
  const query = filter.search.trim().toLocaleLowerCase();
  const cv = query.match(/^cv\s*(\d+)$/);
  return inventory.models.filter(model => {
    const family = inventory.families.find(item => item.id === model.familyId);
    const maker = inventory.manufacturers.find(item => item.id === family?.manufacturerId);
    const profile = profiles.find(item => item.id === (model.profileId || model.documentationProfileId));
    if (filter.manufacturer && maker?.id !== filter.manufacturer || filter.family && family?.id !== filter.family ||
      filter.kind && model.kind !== filter.kind || filter.coverage === "reviewed" && !profile ||
      filter.coverage === "basic" && profile) return false;
    if (cv) {
      const foundation = profile?.rules?.modelDefinitionsOnly || profile?.rules?.roco10882 || profile?.rules?.uhlenbrock76420 || profile?.rules?.piko56500
        ? [] : profiles.find(item => item.id === "dcc-basics")?.definitions || [];
      const definitions = [...(profile?.definitions || []), ...foundation];
      return definitions.some(d => decoderCVMatches(d, Number(cv[1])));
    }
    return [maker?.name, maker?.cv8, family?.name, model.name, ...model.articleNumbers, ...model.interfaces]
      .join(" ").toLocaleLowerCase().includes(query);
  });
}

export function modelContext(inventory: DecoderInventory, model: DecoderModel) {
  const family = inventory.families.find(item => item.id === model.familyId);
  const maker = inventory.manufacturers.find(item => item.id === family?.manufacturerId);
  return { family, maker };
}
