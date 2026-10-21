import { useRef, useState } from "react";

import { type CreateVehicleRequest, type Vehicle, type VehicleFunction, type VehicleFunctionInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { functionKeys, functionMappingsFromImport, isValidFunctionMapping } from "./cvImport";
import { functionDraftChanges, functionHasContent, functionInput } from "./functionDraft";
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
  onMessage
}: UseVehicleFunctionsControllerOptions) {
  const { t } = useI18n();
  const [edits, setEdits] = useState<FunctionEditState>({});
  const [showConfiguredOnly, setShowConfiguredOnly] = useState(true);
  const loadedVehicle = useRef<string | null>(null);
  const importInputRef = useRef<HTMLInputElement | null>(null);

  const functionEdit = (functionKey: string) => edits[functionKey] || emptyFunctionEdit(functionKey);

  const updateFunctionEdit = (functionKey: string, patch: Partial<VehicleFunctionInput>) => {
    setEdits((current) => ({
      ...current,
      [functionKey]: {
        ...emptyFunctionEdit(functionKey),
        ...current[functionKey],
        ...patch, dirty: true,
        ...(patch.mode !== undefined ? { modeAssigned: true } : {})
      }
    }));
  };

  const reset = () => { loadedVehicle.current = null; setEdits({}); setShowConfiguredOnly(true); };
  const acknowledgeSaved = (functions: VehicleFunction[] | undefined, vehicleID?: string) => {
    if (vehicleID) loadedVehicle.current = vehicleID;
    setEdits(functionsToEditState(functions));
  };
  const loadDetail = (functions: VehicleFunction[] | undefined, vehicleID?: string) => {
    const sameVehicle = Boolean(vehicleID && vehicleID === loadedVehicle.current);
    loadedVehicle.current = vehicleID || null;
    setEdits(current => ({ ...functionsToEditState(functions),
      ...(sameVehicle ? Object.fromEntries(Object.entries(current).filter(([, edit]) => edit.dirty)) : {}) }));
  };
  const mergeDetail = (functions: VehicleFunction[] | undefined, overrides: FunctionEditState) => {
    const merged = functionsToEditState(functions);
    for (const [key, edit] of Object.entries(overrides)) {
      const existing = merged[key];
      merged[key] = { ...edit, persisted: existing?.persisted, revision: existing?.revision, dirty: true,
        modeAssigned: true, directionDependent: edit.directionDependent ?? existing?.directionDependent ?? null,
        inverted: edit.inverted ?? existing?.inverted ?? null,
        legacyDirectionDependent: existing?.legacyDirectionDependent ?? edit.legacyDirectionDependent };
    }
    setEdits(merged);
  };
  const remove = (functionKey: string) => setEdits(current => {
    const next = { ...current }, edit = next[functionKey];
    if (!edit) return current;
    if (edit.persisted) next[functionKey] = { ...edit, deleted: !edit.deleted, dirty: true };
    else delete next[functionKey];
    return next;
  });

  const configuredKeys = functionKeys.filter((functionKey) => {
    const edit = edits[functionKey];
    return Boolean(edit && !edit.deleted && functionHasContent(edit));
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
      return { functionKey, ...functionInput(functionEdit(functionKey)) };
    });
    const payload = {
      flagsVersion: 2,
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
    if (!files?.length) return;
    const [file] = Array.from(files);
    onMessage("");
    return file.text().then(text => {
      const rows = functionMappingsFromImport(text);
      if (!rows.length || rows.some(row => !isValidFunctionMapping(row)) ||
        new Set(rows.map(row => row.functionKey)).size !== rows.length) {
        throw new Error("vehicles.functions.invalidImport");
      }
      setEdits(current => {
        const next = { ...current };
        for (const row of rows) {
          const existing = current[row.functionKey] || emptyFunctionEdit(row.functionKey);
          next[row.functionKey] = { ...existing, ...row, dirty: true, deleted: false, modeAssigned: true,
            directionDependent: row.flagsVersion === 2 ? row.directionDependent : existing.directionDependent,
            inverted: row.flagsVersion === 2 ? row.inverted : existing.inverted,
            legacyDirectionDependent: existing.legacyDirectionDependent ?? row.legacyDirectionDependent,
            flagsVersion: 2 };
        }
        return next;
      });
      return t("vehicles.functions.importDraft", { count: rows.length });
    }).catch(() => {
      onMessage(t("vehicles.functions.invalidImport"));
      return undefined;
    })
      .finally(() => { if (importInputRef.current) importInputRef.current.value = ""; });
  };

  return {
    state: { edits, showConfiguredOnly, configuredKeys, visibleKeys, summary },
    refs: { importInputRef },
    setters: { setEdits, setShowConfiguredOnly },
    commands: { functionEdit, updateFunctionEdit, reset, loadDetail, mergeDetail, acknowledgeSaved, changes: () => functionDraftChanges(edits), remove, exportValues, importValues }
  };
}
