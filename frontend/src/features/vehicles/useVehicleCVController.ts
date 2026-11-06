import { useMemo, useRef, useState } from "react";

import {
  api, ApiError,
  type Vehicle,
  type VehicleCVValue,
  type VehicleCVValueInput
} from "../../shared/api";
import {
  commonDecoderProfiles,
  isValidCVValueInput
} from "./cvImport";
import { useVehicleCVImportController } from "./useVehicleCVImportController";
import { useI18n } from "../../shared/i18n";
import { emptyCVForm } from "./vehicleViewModel";
import {
  cvDraftCollision, cvDraftConflicts, cvDraftCurrent, cvDraftInput, cvDraftRequest, cvDraftValues, maxManualCVChanges,
  rebaseCVChange, stageCVDeletion, stageCVInput, type CVDraftChange
} from "./cvDraft";

type UseVehicleCVControllerOptions = {
  selected: Vehicle | null;
  decoderNumber: string;
  readonly?: boolean;
  saving?: boolean;
  refreshImportedValues?: (vehicleId: string) => Promise<void>;
  setSaving: (saving: boolean) => void;
  onMessage: (message: string) => void;
  refreshSelectedVehicle: (vehicleId?: string) => Promise<void>;
  reloadSelectedVehicle: (vehicleId?: string) => Promise<void>;
};

export function useVehicleCVController({
  selected,
  decoderNumber,
  readonly = false,
  saving = false,
  refreshImportedValues,
  setSaving,
  onMessage,
  refreshSelectedVehicle,
  reloadSelectedVehicle
}: UseVehicleCVControllerOptions) {
  const { t } = useI18n();
  const [form, setForm] = useState<VehicleCVValueInput>(emptyCVForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const imports = useVehicleCVImportController({ selected, readonly, saving, setSaving, onMessage,
    refreshSaved: refreshImportedValues || refreshSelectedVehicle });
  const { preview: importPreview, needsReview: importNeedsReview, imported: lastImported } = imports.state;
  const { importInputRef } = imports.refs;
  const { importValues, toggleImportRow, selectImportRows, applyImportPreview, setImportPreview,
    recompareImportPreview, discardImportPreview } = imports.commands;
  const [manualChanges, setManualChanges] = useState<CVDraftChange[]>([]);
  const [lastManualSave, setLastManualSave] = useState(0);
  const draftGeneration = useRef(0), nextDraftID = useRef(0);
  const workingCVValues = useMemo(() => cvDraftValues(selected?.cvValues || [], manualChanges, selected?.id || ""),
    [selected, manualChanges]);
  const manualConflicts = useMemo(() => cvDraftConflicts(manualChanges, selected?.cvValues || []),
    [manualChanges, selected]);

  const storedDecoderProfiles = useMemo(() => Array.from(new Set([
    ...(selected?.cvValues || []).map((value) => value.decoderProfile)
      .filter((profile): profile is string => Boolean(profile)),
    ...(selected?.cvFiles || []).map((file) => file.decoderProfile)
      .filter((profile): profile is string => Boolean(profile))
  ])).sort((left, right) => left.localeCompare(right, "de-DE")), [selected]);

  const summary = useMemo(() => ({
    values: selected?.cvValues?.length || 0,
    files: selected?.cvFiles?.length || 0,
    profiles: storedDecoderProfiles.length
  }), [selected, storedDecoderProfiles]);

  const importStats = useMemo(() => ({
    imported: lastImported,
    selected: importPreview?.rows.filter((row) => row.selected && row.status !== "invalid").length || 0,
    new: importPreview?.rows.filter((row) => row.status === "new").length || 0,
    changed: importPreview?.rows.filter((row) => row.status === "changed").length || 0,
    same: importPreview?.rows.filter((row) => row.status === "same").length || 0,
    invalid: importPreview?.rows.filter((row) => row.status === "invalid").length || 0
  }), [importPreview, lastImported]);

  const decoderProfileOptions = useMemo(
    () => Array.from(new Set([...commonDecoderProfiles, ...storedDecoderProfiles])),
    [storedDecoderProfiles]
  );

  const updateForm = (patch: Partial<VehicleCVValueInput>) => {
    setForm((current) => ({ ...current, ...patch }));
  };

  const resetForm = () => {
    setForm(emptyCVForm);
    setEditingId(null);
  };

  const reset = () => {
    draftGeneration.current++;
    setManualChanges([]);
    setLastManualSave(0);
    resetForm();
    imports.commands.reset();
  };

  const edit = (value: VehicleCVValue) => {
    setForm(cvDraftInput(value));
    setEditingId(value.id);
  };

  const save = () => {
    if (!selected) return;
    const payload = { ...form, cvNumber: Number(form.cvNumber), value: Number(form.value) };
    if (!isValidCVValueInput(payload)) {
      onMessage(payload.indexCV31 != null || payload.indexCV32 != null
        ? t("vehicles.cv.indexInvalid") : "CV-Nummer muss 1-1024 und Wert 0-255 sein.");
      return;
    }
    const result = stageCVInput(manualChanges, selected.cvValues || [], selected.id,
      payload, editingId, `cv-draft-${++nextDraftID.current}`);
    if (result.error) { onMessage(t(`vehicles.cv.draft.error.${result.error}`)); return; }
    setManualChanges(result.changes || []);
    setLastManualSave(0);
    resetForm();
    onMessage("");
  };

  const remove = (value: VehicleCVValue) => {
    if (!selected) return;
    const next = stageCVDeletion(manualChanges, value);
    if (!next) {
      onMessage(t(manualChanges.length >= maxManualCVChanges
        ? "vehicles.cv.draft.error.limit" : "vehicles.cv.draft.error.revision"));
      return;
    }
    setManualChanges(next);
    setLastManualSave(0);
    if (editingId === value.id) resetForm();
    onMessage("");
  };

  const applyManualDraft = async () => {
    if (!selected || !manualChanges.length) return;
    if (manualConflicts.length) { onMessage(t("vehicles.cv.draft.conflict")); return; }
    const generation = draftGeneration.current;
    setSaving(true);
    onMessage("");
    try {
      const result = await api.applyVehicleCVChanges(selected.id, cvDraftRequest(manualChanges));
      if (generation !== draftGeneration.current) return;
      setManualChanges([]);
      setLastManualSave(result.changed);
      // The form can contain further input which was not part of the submitted selection.
      // A committed new row no longer has its temporary draft identity; keep its open input.
      if (manualChanges.some(change => change.id === editingId && !change.original)) setEditingId(null);
      onMessage("");
      await refreshSelectedVehicle(selected.id);
    } catch (error) {
      if (generation !== draftGeneration.current) return;
      if (error instanceof ApiError && error.code === "cv_value_conflict") {
        await reloadSelectedVehicle(selected.id);
        if (generation === draftGeneration.current) onMessage(t("vehicles.cv.draft.conflict"));
      } else if (error instanceof ApiError && error.code === "cv_changes_invalid") {
        onMessage(t("vehicles.cv.draft.error.invalid"));
      } else onMessage(error instanceof Error ? error.message : String(error));
    } finally {
      if (generation === draftGeneration.current) setSaving(false);
    }
  };

  const discardManualDraft = () => { setManualChanges([]); resetForm(); onMessage(""); };
  const discardManualChange = (id: string) => {
    setManualChanges(current => current.filter(change => change.id !== id));
    if (editingId === id) resetForm();
  };
  const rebaseManualChange = (id: string) => {
    const change = manualChanges.find(value => value.id === id);
    const current = change && cvDraftCurrent(change, selected?.cvValues || [], manualChanges);
    if (change?.original && cvDraftCollision(change, selected?.cvValues || [], manualChanges)) {
      onMessage(t("vehicles.cv.draft.error.occupied"));
      return;
    }
    if (!current) { onMessage(t("vehicles.cv.draft.missing")); return; }
    const next = rebaseCVChange(manualChanges, id, current);
    if (next === manualChanges) { onMessage(t("vehicles.cv.draft.error.occupied")); return; }
    setManualChanges(next);
    onMessage("");
  };

  const exportValues = () => {
    if (!selected) return;
    const payload = {
      vehicle: { inventoryNumber: selected.inventoryNumber, name: selected.name, decoder: decoderNumber },
      cvValues: selected.cvValues || []
    };
    const url = URL.createObjectURL(new Blob([JSON.stringify(payload, null, 2)], { type: "application/json" }));
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = `${selected.inventoryNumber || "railkeeper"}-cv.json`;
    anchor.click();
    URL.revokeObjectURL(url);
  };


  return {
    state: {
      form,
      editingId,
      manualDraft: { changes: manualChanges, values: workingCVValues, conflicts: manualConflicts.map(change => change.id), lastSavedCount: lastManualSave },
      importPreview,
      importNeedsReview,
      summary,
      importStats,
      storedDecoderProfiles,
      decoderProfileOptions
    },
    refs: { importInputRef },
    commands: {
      updateForm,
      resetForm,
      reset,
      edit,
      save,
      remove,
      applyManualDraft,
      discardManualDraft,
      discardManualChange,
      rebaseManualChange,
      exportValues,
      importValues,
      toggleImportRow,
      selectImportRows,
      applyImportPreview,
      recompareImportPreview,
      setImportPreview,
      discardImportPreview
    }
  };
}
