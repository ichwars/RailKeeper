import type { AppView } from "./App";

const appViews = [
  "overview",
  "vehicles",
  "accessories",
  "exhibition",
  "importExport",
  "digitalCenters",
  "settings"
] as const;

function isAppView(value: string): value is AppView {
  return appViews.some((view) => view === value);
}

export function availableStartView(value: string): AppView {
  if (value === "inventory") return "vehicles";
  if (!isAppView(value)) return "overview";
  return value;
}
