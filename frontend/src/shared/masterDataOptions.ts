import type { MasterDataEntry } from "./api";

export type MasterDataOption = {
  id: string;
  value: string;
  label: string;
  active: boolean;
};

export function masterDataOptions(
  entries: readonly MasterDataEntry[],
  currentValues: readonly string[],
  persistedValue: (entry: MasterDataEntry) => string
): MasterDataOption[] {
  const current = new Set(currentValues.filter(Boolean));
  const options = entries
    .filter((entry) => entry.active || current.has(persistedValue(entry)))
    .map((entry) => ({
      id: entry.id,
      value: persistedValue(entry),
      label: entry.label,
      active: entry.active
    }));
  for (const value of current) {
    if (!options.some((option) => option.value === value)) {
      options.push({ id: `legacy:${value}`, value, label: value, active: false });
    }
  }
  return options;
}
