import type { AllocationTarget, Layout, LayoutUnit, Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";

export type AccessoryTargetKind = "vehicle" | "layout" | "layoutUnit";
export type AccessoryTargetSelection = { kind: AccessoryTargetKind; id: string };

export function resolveAccessoryTargetSelection(target: AccessoryTargetSelection, vehicles: Vehicle[],
  layouts: Layout[], units: LayoutUnit[]): AccessoryTargetSelection {
  const ids = target.kind === "vehicle" ? vehicles.map((item) => item.id)
    : target.kind === "layout" ? layouts.filter((item) => !item.archived).map((item) => item.id)
      : units.filter((item) => !item.archived).map((item) => item.id);
  return ids.includes(target.id) ? target : { ...target, id: ids[0] || "" };
}

export function AccessoryTargetFields({ target, vehicles, layouts, units, onChange }: {
  target: AccessoryTargetSelection;
  vehicles: Vehicle[];
  layouts: Layout[];
  units: LayoutUnit[];
  onChange: (target: AccessoryTargetSelection) => void;
}) {
  const { t } = useI18n();
  const resolvedTarget = resolveAccessoryTargetSelection(target, vehicles, layouts, units);
  const options = target.kind === "vehicle" ? vehicles.map((vehicle) => ({
    id: vehicle.id,
    label: `${vehicle.inventoryNumber} · ${vehicle.name}`
  })) : target.kind === "layout" ? layouts.filter((layout) => !layout.archived)
    .map((layout) => ({ id: layout.id, label: layout.name }))
    : units.filter((unit) => !unit.archived).map((unit) => ({ id: unit.id, label: unit.name }));

  return <>
    <label>{t("accessories.field.targetType")}<AppSelect value={target.kind} onChange={(event) =>
      onChange({ kind: event.target.value as AccessoryTargetKind, id: "" })}>
      <option value="vehicle">{t("accessories.target.vehicle")}</option>
      <option value="layout">{t("accessories.target.layout")}</option>
      <option value="layoutUnit">{t("accessories.target.layoutUnit")}</option>
    </AppSelect></label>
    <label>{t("accessories.field.target")}<AppSelect value={resolvedTarget.id}
      onChange={(event) => onChange({ ...target, id: event.target.value })}>
      {options.map((option) => <option key={option.id} value={option.id}>{option.label}</option>)}
    </AppSelect></label>
  </>;
}

export function accessoryTargetInput(target: AccessoryTargetSelection): AllocationTarget | null {
  if (!target.id) return null;
  if (target.kind === "vehicle") return { vehicleId: target.id };
  if (target.kind === "layout") return { layoutId: target.id };
  return { layoutUnitId: target.id };
}

export function accessoryTargetLabel(target: { vehicleId?: string; layoutId?: string; layoutUnitId?: string },
  vehicles: Vehicle[], layouts: Layout[],
  units: LayoutUnit[]) {
  if (target.vehicleId) {
    const vehicle = vehicles.find((item) => item.id === target.vehicleId);
    return vehicle ? `${vehicle.inventoryNumber} · ${vehicle.name}` : target.vehicleId;
  }
  if (target.layoutId) return layouts.find((item) => item.id === target.layoutId)?.name || target.layoutId;
  return units.find((item) => item.id === target.layoutUnitId)?.name || target.layoutUnitId || "";
}
