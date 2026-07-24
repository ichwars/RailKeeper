import { useRef, useState } from "react";

import { api, type CreateVehicleRequest, type Vehicle, type VehicleFunction, type VehicleFunctionInput } from "../../shared/api";
import { functionKeys, functionMappingsFromImport, isValidFunctionMapping } from "./cvImport";
import { functionsToEditState, type FunctionEditState } from "./vehicleTransforms";
import { emptyFunctionEdit } from "./vehicleViewModel";

type UseVehicleFunctionsControllerOptions = {
  selected: Vehicle | null;
  form: CreateVehicleRequest;
  setSaving: (saving: boolean) => void;
  onMessage: (message: string) => void;
  refreshSelectedVehicle: (vehicleId?: string) => Promise<void>;
};

export function useVehicleFunctionsController({
  selected,
  form,
  setSaving,
  onMessage,
  refreshSelectedVehicle
}: UseVehicleFunctionsControllerOptions) {
  const [edits, setEdits] = useState<FunctionEditState>({});
  const [showConfiguredOnly, setShowConfiguredOnly] = useState(false);
  const importInputRef = useRef<HTMLInputElement | null>(null);

  const functionEdit = (functionKey: string) => edits[functionKey] || emptyFunctionEdit(functionKey);

  const updateFunctionEdit = (functionKey: string, patch: Partial<VehicleFunctionInput>) => {
    setEdits((current) => ({
      ...current,
      [functionKey]: {
        ...emptyFunctionEdit(functionKey),
        ...current[functionKey],
        ...patch
      }
    }));
  };

  const reset = () => setEdits({});
  const loadDetail = (functions: VehicleFunction[] | undefined) => setEdits(functionsToEditState(functions));
  const mergeDetail = (functions: VehicleFunction[] | undefined, overrides: FunctionEditState) => {
    setEdits({ ...functionsToEditState(functions), ...overrides });
  };

  const save = (functionKey: string) => {
    if (!selected) return;
    const edit = functionEdit(functionKey);

    if (!edit.persisted && !edit.name?.trim() && !edit.symbolKey && !edit.notes?.trim()) {
      onMessage(`${functionKey}: Bitte Funktionsname, Symbol oder Notiz eintragen.`);
      return;
    }

    setSaving(true);
    onMessage("");
    api.updateVehicleFunction(selected.id, functionKey, {
      name: edit.name || "",
      symbolKey: edit.symbolKey || "",
      functionType: edit.functionType || "standard",
      mode: edit.mode || "dauer",
      directionDependent: Boolean(edit.directionDependent),
      notes: edit.notes || ""
    })
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  const remove = (functionKey: string) => {
    if (!selected) return;
    setSaving(true);
    onMessage("");
    api.deleteVehicleFunction(selected.id, functionKey)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  const configuredKeys = functionKeys.filter((functionKey) => {
    const edit = functionEdit(functionKey);
    return Boolean(edit.persisted || edit.name || edit.symbolKey || edit.notes);
  });
  const visibleKeys = showConfiguredOnly ? configuredKeys : functionKeys;
  const summary = {
    configured: configuredKeys.length,
    sound: configuredKeys.filter((functionKey) => functionEdit(functionKey).functionType === "sound").length,
    light: configuredKeys.filter((functionKey) => functionEdit(functionKey).functionType === "licht").length
  };

  const exportValues = () => {
    if (!selected) return;
    const functionMappings = configuredKeys.map((functionKey) => {
      const edit = functionEdit(functionKey);
      return {
        functionKey,
        name: edit.name || "",
        symbolKey: edit.symbolKey || "",
        functionType: edit.functionType || "standard",
        mode: edit.mode || "dauer",
        directionDependent: Boolean(edit.directionDependent),
        notes: edit.notes || ""
      };
    });
    const payload = {
      vehicle: {
        inventoryNumber: selected.inventoryNumber,
        name: selected.name,
        decoder: form.digitalDecoderNumber || form.dtDecoderNumber || ""
      },
      functions: functionMappings
    };
    const blob = new Blob([JSON.stringify(payload, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = `${selected.inventoryNumber || "railkeeper"}-funktionen.json`;
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const importValues = (files: FileList | null) => {
    if (!selected || !files || files.length === 0) return;
    const [file] = Array.from(files);
    setSaving(true);
    onMessage("");
    file.text()
      .then(functionMappingsFromImport)
      .then(async (rows) => {
        const valid = rows.filter(isValidFunctionMapping);
        if (valid.length === 0) throw new Error("Keine gültigen Funktionszuordnungen gefunden.");

        for (const row of valid) {
          await api.updateVehicleFunction(selected.id, row.functionKey, {
            name: row.name || "",
            symbolKey: row.symbolKey || "",
            functionType: row.functionType || "standard",
            mode: row.mode || "dauer",
            directionDependent: Boolean(row.directionDependent),
            notes: row.notes || ""
          });
        }
      })
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => onMessage(error.message))
      .finally(() => {
        setSaving(false);
        if (importInputRef.current) importInputRef.current.value = "";
      });
  };

  return {
    state: { edits, showConfiguredOnly, configuredKeys, visibleKeys, summary },
    refs: { importInputRef },
    setters: { setEdits, setShowConfiguredOnly },
    commands: { functionEdit, updateFunctionEdit, reset, loadDetail, mergeDetail, save, remove, exportValues, importValues }
  };
}
