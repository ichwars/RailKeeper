import type {
  AccessoryArticleListResult,
  Vehicle,
  VehicleMaintenance
} from "../../shared/api";

export const overviewMetricProfileKey = "railkeeper.overview.metrics";

export const overviewMetricIDs = [
  "vehicles",
  "accessories",
  "inventoryValue",
  "maintenance",
  "digitalized",
  "dataQuality",
  "vehicleListValue",
  "vehiclePurchaseValue",
  "accessoryListValue",
  "accessoryPurchaseValue"
] as const;

export type OverviewMetricID = typeof overviewMetricIDs[number];

export const defaultOverviewMetrics: OverviewMetricID[] = [
  "vehicles",
  "accessories",
  "inventoryValue",
  "maintenance"
];

export const maxOverviewMetrics = 6;

export function overviewMetricLimitForWidth(width: number) {
  if (width <= 0) return 4;
  const cardMinimum = 250;
  const gap = 12;
  return Math.max(2, Math.min(maxOverviewMetrics, Math.floor((width + gap) / (cardMinimum + gap))));
}

export type OverviewMetricPreference = {
  active: OverviewMetricID[];
  order: OverviewMetricID[];
};

export type ScheduledMaintenance = {
  vehicle: Vehicle;
  entry: VehicleMaintenance;
  days: number;
};

export type OverviewStats = {
  digital: number;
  analog: number;
  withImages: number;
  withArticleNumbers: number;
  withDecoderAddresses: number;
  digitalWithoutAddress: number;
  documented: number;
  due: number;
  upcoming: number;
  nextMaintenance: ScheduledMaintenance[];
  recentVehicles: Vehicle[];
  categories: Array<[string, number]>;
  manufacturers: Array<[string, number]>;
};

export type OverviewTrendPoint = {
  label: string;
  vehicles: number;
  accessories: number;
};

function isMetricID(value: unknown): value is OverviewMetricID {
  return typeof value === "string" && overviewMetricIDs.includes(value as OverviewMetricID);
}

export function normalizeMetricPreference(value: unknown): OverviewMetricPreference | null {
  if (!value || typeof value !== "object") return null;
  const candidate = value as Partial<OverviewMetricPreference>;
  if (!Array.isArray(candidate.active) || !Array.isArray(candidate.order)) return null;

  const active = [...new Set(candidate.active.filter(isMetricID))].slice(0, maxOverviewMetrics);
  const storedOrder = [...new Set(candidate.order.filter(isMetricID))];
  const order = [
    ...storedOrder,
    ...overviewMetricIDs.filter((metric) => !storedOrder.includes(metric))
  ];
  if (active.length === 0) return null;
  return { active, order };
}

export function parseMetricPreference(value: string | null): OverviewMetricPreference | null {
  if (!value) return null;
  try {
    return normalizeMetricPreference(JSON.parse(value));
  } catch {
    return null;
  }
}

export function overviewMetricStorageKey(username: string) {
  return `${overviewMetricProfileKey}:${username || "local"}`;
}

export function readMetricPreference(username: string): OverviewMetricPreference {
  return parseMetricPreference(window.localStorage.getItem(overviewMetricStorageKey(username))) || {
    active: defaultOverviewMetrics,
    order: [...overviewMetricIDs]
  };
}

export function persistMetricPreference(username: string, preference: OverviewMetricPreference) {
  window.localStorage.setItem(overviewMetricStorageKey(username), JSON.stringify(preference));
}

function dateDistance(entry: VehicleMaintenance) {
  if (!entry.dueDate || entry.status === "erledigt") return null;
  const now = new Date();
  const due = new Date(`${entry.dueDate}T00:00:00`);
  return Math.ceil((due.getTime() - now.getTime()) / 86_400_000);
}

function topEntries(values: string[], limit = 5) {
  const counts = new Map<string, number>();
  for (const value of values.filter(Boolean)) {
    counts.set(value, (counts.get(value) || 0) + 1);
  }
  return [...counts.entries()]
    .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
    .slice(0, limit);
}

export function buildOverviewStats(vehicles: Vehicle[]): OverviewStats {
  const digital = vehicles.filter((vehicle) => vehicle.digital).length;
  const withImages = vehicles.filter((vehicle) => (vehicle.images || []).length > 0).length;
  const withArticleNumbers = vehicles.filter((vehicle) => vehicle.articleNumber?.trim()).length;
  const withDecoderAddresses = vehicles.filter((vehicle) => !vehicle.digital ||
    vehicle.digitalDecoderNumber?.trim() || vehicle.dtDecoderNumber?.trim() ||
    vehicle.externalMappings?.some((mapping) => mapping.externalAddress?.trim())).length;
  const digitalWithoutAddress = vehicles.filter((vehicle) => vehicle.digital &&
    !vehicle.digitalDecoderNumber?.trim() && !vehicle.dtDecoderNumber?.trim() &&
    !vehicle.externalMappings?.some((mapping) => mapping.externalAddress?.trim())).length;
  const documented = vehicles.filter((vehicle) =>
    vehicle.articleNumber?.trim() && (vehicle.images || []).length > 0 &&
    (!vehicle.digital || vehicle.digitalDecoderNumber?.trim() || vehicle.dtDecoderNumber?.trim() ||
      vehicle.externalMappings?.some((mapping) => mapping.externalAddress?.trim()))).length;
  const scheduled = vehicles.flatMap((vehicle) => (vehicle.maintenance || [])
    .map((entry) => ({ vehicle, entry, days: dateDistance(entry) }))
    .filter((item): item is ScheduledMaintenance => item.days !== null));

  return {
    digital,
    analog: vehicles.length - digital,
    withImages,
    withArticleNumbers,
    withDecoderAddresses,
    digitalWithoutAddress,
    documented,
    due: scheduled.filter((item) => item.days <= 0).length,
    upcoming: scheduled.filter((item) => item.days > 0 && item.days <= 30).length,
    nextMaintenance: [...scheduled].sort((a, b) => a.days - b.days).slice(0, 3),
    recentVehicles: [...vehicles]
      .sort((a, b) => Date.parse(b.updatedAt) - Date.parse(a.updatedAt))
      .slice(0, 3),
    categories: topEntries(vehicles.map((vehicle) => vehicle.category || ""), 4),
    manufacturers: topEntries(vehicles.map((vehicle) => vehicle.manufacturer || ""), 5)
  };
}

export function percentage(value: number, total: number) {
  return total > 0 ? Math.round((value / total) * 100) : 0;
}

export function accessoryCount(accessories: AccessoryArticleListResult | null) {
  return accessories?.metrics.articleCount || 0;
}

export function buildOverviewTrend(
  vehicles: Vehicle[],
  accessories: AccessoryArticleListResult | null,
  months: number,
  language: string
): OverviewTrendPoint[] {
  const today = new Date();
  const accessoryTotal = accessoryCount(accessories);
  return Array.from({ length: months }, (_, index) => {
    const monthOffset = index - months + 1;
    const monthStart = new Date(today.getFullYear(), today.getMonth() + monthOffset, 1);
    const monthEnd = new Date(today.getFullYear(), today.getMonth() + monthOffset + 1, 0, 23, 59, 59);
    return {
      label: new Intl.DateTimeFormat(language === "en" ? "en-GB" : "de-DE", {
        month: "short",
        year: months <= 6 ? undefined : "2-digit"
      }).format(monthStart),
      vehicles: vehicles.filter((vehicle) => Date.parse(vehicle.createdAt) <= monthEnd.getTime()).length,
      accessories: accessoryTotal
    };
  });
}

export function primaryVehicleImage(vehicle: Vehicle) {
  const image = (vehicle.images || []).find((candidate) => candidate.isPrimary) || vehicle.images?.[0];
  return image?.thumbnailUrl || image?.url || "";
}

export function safeNumber(value: string | undefined) {
  if (!value || !/^\d+(?:\.\d{1,2})?$/.test(value)) return 0;
  return Number(value);
}
