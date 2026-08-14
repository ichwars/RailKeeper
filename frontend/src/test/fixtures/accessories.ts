import type { AccessoryArticleListItem, MasterDataEntry } from "../../shared/api";

export function accessoryArticleFixture(
  overrides: Partial<AccessoryArticleListItem> = {}
): AccessoryArticleListItem {
  return {
    id: "article-1",
    inventoryNumber: "RK-ART-000001",
    manufacturer: "Tillig",
    articleNumber: "83101",
    name: "Gerades Modellgleis",
    articleType: "track",
    subtype: "straight",
    gauges: ["TT"],
    inventoryStrategy: "quantity",
    archived: false,
    owned: 18,
    available: 12,
    reserved: 4,
    installed: 2,
    locationNames: ["Werkstatt"],
    hasUsageHistory: true,
    careHintCount: 0,
    updatedAt: "2026-08-08T10:00:00Z",
    attributes: [],
    ...overrides
  };
}

const timestamp = "2026-08-08T10:00:00Z";

export const accessoryArticleTypes: MasterDataEntry[] = [{
  id: "article-type-track",
  type: "article_type",
  key: "track",
  label: "Gleis",
  active: true,
  sortOrder: 10,
  metadata: {},
  createdAt: timestamp,
  updatedAt: timestamp
}];

export const accessorySubtypes: MasterDataEntry[] = [{
  id: "article-subtype-track-straight",
  type: "accessory_subtype",
  key: "track:straight",
  label: "Gerade",
  active: true,
  sortOrder: 10,
  metadata: {},
  createdAt: timestamp,
  updatedAt: timestamp
}];
