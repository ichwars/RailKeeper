import type { AllocationTarget, Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";

export type AccessoryTargetSelection = { kind: "vehicle"; id: string };
export function resolveAccessoryTargetSelection(target: AccessoryTargetSelection, vehicles: Vehicle[]): AccessoryTargetSelection {
  return vehicles.some((vehicle) => vehicle.id === target.id) ? target : { kind: "vehicle", id: vehicles[0]?.id || "" };
}
export function AccessoryTargetFields({ target, vehicles, onChange }: {
  target: AccessoryTargetSelection; vehicles: Vehicle[]; onChange: (target: AccessoryTargetSelection) => void;
}) {
  const { t } = useI18n();
  const resolved = resolveAccessoryTargetSelection(target, vehicles);
  return <label>{t("accessories.target.vehicle")}<AppSelect value={resolved.id}
    onChange={(event) => onChange({ kind: "vehicle", id: event.target.value })}>
    {vehicles.map((vehicle) => <option key={vehicle.id} value={vehicle.id}>{vehicle.inventoryNumber} · {vehicle.name}</option>)}
  </AppSelect></label>;
}
export function accessoryTargetInput(target: AccessoryTargetSelection): AllocationTarget | null {
  return target.id ? { vehicleId: target.id } : null;
}
export function accessoryTargetLabel(target: { vehicleId?: string }, vehicles: Vehicle[]) {
  const vehicle = vehicles.find((item) => item.id === target.vehicleId);
  return vehicle ? `${vehicle.inventoryNumber} · ${vehicle.name}` : target.vehicleId || "";
}
