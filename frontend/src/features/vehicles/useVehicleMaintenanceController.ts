import { useState } from "react";

import {
  api,
  type Vehicle,
  type VehicleMaintenance,
  type VehicleMaintenanceInput
} from "../../shared/api";
import { todayISODate } from "./vehicleMaintenance";
import { emptyMaintenanceForm } from "./vehicleViewModel";

type UseVehicleMaintenanceControllerOptions = {
  selected: Vehicle | null;
  setSaving: (saving: boolean) => void;
  onMessage: (message: string) => void;
  refreshSelectedVehicle: (vehicleId?: string) => Promise<void>;
};

export function useVehicleMaintenanceController({
  selected,
  setSaving,
  onMessage,
  refreshSelectedVehicle
}: UseVehicleMaintenanceControllerOptions) {
  const [form, setForm] = useState<VehicleMaintenanceInput>(emptyMaintenanceForm);
  const [editingId, setEditingId] = useState<string | null>(null);

  const updateForm = (patch: Partial<VehicleMaintenanceInput>) => {
    setForm((current) => ({ ...current, ...patch }));
  };

  const resetForm = () => {
    setForm(emptyMaintenanceForm);
    setEditingId(null);
  };

  const edit = (entry: VehicleMaintenance) => {
    setForm({
      kind: entry.kind || "Wartung",
      status: entry.status || "geplant",
      conditionRating: entry.conditionRating || "",
      dueDate: entry.dueDate || "",
      completedAt: entry.completedAt || "",
      cost: entry.cost || "",
      notes: entry.notes || ""
    });
    setEditingId(entry.id);
  };

  const save = () => {
    if (!selected) return;
    setSaving(true);
    onMessage("");

    const payload: VehicleMaintenanceInput = {
      ...form,
      status: form.status === "f?llig" ? "faellig" : form.status,
      cost: form.cost?.trim().replace(/\s*?$/, "") || "",
      completedAt: form.status === "erledigt" && !form.completedAt ? todayISODate() : form.completedAt
    };
    const action = editingId
      ? api.updateVehicleMaintenance(selected.id, editingId, payload)
      : api.createVehicleMaintenance(selected.id, payload);

    action
      .then(() => refreshSelectedVehicle(selected.id))
      .then(resetForm)
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  const complete = (entry: VehicleMaintenance) => {
    if (!selected) return;
    setSaving(true);
    onMessage("");
    api.updateVehicleMaintenance(selected.id, entry.id, {
      kind: entry.kind,
      status: "erledigt",
      conditionRating: entry.conditionRating || "",
      dueDate: entry.dueDate || "",
      completedAt: entry.completedAt || todayISODate(),
      cost: entry.cost || "",
      notes: entry.notes || ""
    })
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  const remove = (entry: VehicleMaintenance) => {
    if (!selected) return;
    setSaving(true);
    onMessage("");
    api.deleteVehicleMaintenance(selected.id, entry.id)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  return {
    state: { form, editingId },
    setters: { setForm, setEditingId },
    commands: { updateForm, resetForm, edit, save, complete, remove }
  };
}
