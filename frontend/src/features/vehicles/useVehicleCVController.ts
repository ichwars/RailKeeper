import { useEffect, useMemo, useRef, useState } from "react";

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
import { CVWorkingGroupUnavailable, cvDeletionScope, refreshCVInputScope, scopeCVInput, type CVWorkingGroupOptions } from "./cvWorkingGroup";

type UseVehicleCVControllerOptions = {
  selected: Vehicle | null;
  decoderNumber: string;
  readonly?: boolean;
  saving?: boolean;
  workingGroup?: CVWorkingGroupOptions;
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
  workingGroup,
  refreshImportedValues,
  setSaving,
  onMessage,
  refreshSelectedVehicle,
  reloadSelectedVehicle
}: UseVehicleCVControllerOptions) {
  const { t } = useI18n();
  const [form, setForm] = useState<VehicleCVValueInput>(emptyCVForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const imports = useVehicleCVImportController({ selected, readonly, saving, workingGroup, setSaving, onMessage,
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
  const bindingConflicts = useMemo(() => workingGroup?.groups ? manualChanges.filter(change => {
    const matches = (scope: { workspaceId?: string; expectedWorkspaceRevision?: string }, profile?: string) => {
      const current = workingGroup.groups?.find(group => group.id === scope.workspaceId);
      return current && current.profile === (profile || "") && current.bindingRevision === scope.expectedWorkspaceRevision;
    };
    return Boolean((change.scope && !matches(change.scope, change.original?.decoderProfile)) ||
      (change.value?.workspaceId && !matches(change.value, change.value.decoderProfile)));
  }) : [], [manualChanges, workingGroup?.groups]);
  const manualConflicts = useMemo(() => [...new Map([...cvDraftConflicts(manualChanges, selected?.cvValues || []), ...bindingConflicts]
    .map(change => [change.id, change])).values()], [manualChanges, selected, bindingConflicts]);
  const writable = () => {
    if (readonly || saving) return false;
    if (workingGroup && !workingGroup.ready) { onMessage(t("cvScope.loadFailed")); return false; }
    return true;
  };
  const withOriginalScopes = (changes: CVDraftChange[]) => workingGroup?.groups ? changes.map(change => change.original && !change.scope
    ? { ...change, scope: cvDeletionScope(change.original.decoderProfile, workingGroup.groups) } : change) : changes;

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
    if (!writable()) return;
    setForm((current) => ({ ...current, ...patch,
      ...(patch.decoderProfile !== undefined && patch.decoderProfile !== current.decoderProfile
        ? { workspaceId: undefined, expectedWorkspaceRevision: undefined } : {}) }));
  };

  const resetForm = () => {
    setForm(scopeCVInput(emptyCVForm, workingGroup?.groups, workingGroup?.target));
    setEditingId(null);
  };

  const reset = () => {
    draftGeneration.current++;
    setManualChanges([]);
    setLastManualSave(0);
    resetForm();
    imports.commands.reset();
  };
  useEffect(() => { if (workingGroup) reset(); }, [workingGroup?.ownerKey]);
  useEffect(() => {
    if (workingGroup && !editingId) setForm(current => ({ ...current, decoderProfile: workingGroup.target?.profile || "",
      workspaceId: workingGroup.target?.workspaceId, expectedWorkspaceRevision: workingGroup.target?.expectedWorkspaceRevision }));
  }, [workingGroup?.target?.workspaceId, workingGroup?.target?.expectedWorkspaceRevision]);

  const edit = (value: VehicleCVValue) => {
    if (!writable()) return;
    setForm(scopeCVInput(cvDraftInput(value), workingGroup?.groups));
    setEditingId(value.id);
  };

  const save = () => {
    if (!selected || !writable()) return;
    const payload = scopeCVInput({ ...form, cvNumber: Number(form.cvNumber), value: Number(form.value) }, workingGroup?.groups);
    if (!isValidCVValueInput(payload)) {
      onMessage(payload.indexCV31 != null || payload.indexCV32 != null
        ? t("vehicles.cv.indexInvalid") : "CV-Nummer muss 1-1024 und Wert 0-255 sein.");
      return;
    }
    const result = stageCVInput(manualChanges, selected.cvValues || [], selected.id,
      payload, editingId, `cv-draft-${++nextDraftID.current}`);
    if (result.error) { onMessage(t(`vehicles.cv.draft.error.${result.error}`)); return; }
    setManualChanges(withOriginalScopes(result.changes || []));
    setLastManualSave(0);
    resetForm();
    onMessage("");
  };

  const remove = (value: VehicleCVValue) => {
    if (!selected || !writable()) return;
    const next = stageCVDeletion(manualChanges, value);
    if (!next) {
      onMessage(t(manualChanges.length >= maxManualCVChanges
        ? "vehicles.cv.draft.error.limit" : "vehicles.cv.draft.error.revision"));
      return;
    }
    setManualChanges(withOriginalScopes(next));
    setLastManualSave(0);
    if (editingId === value.id) resetForm();
    onMessage("");
  };

  const applyManualDraft = async () => {
    if (!selected || !manualChanges.length || !writable()) return;
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
  const rebaseManualChange = async (id: string) => {
    if (!selected || !writable()) return;
    const generation = draftGeneration.current;
    let stored = selected.cvValues || [], groups = workingGroup?.groups;
    if (workingGroup) {
      setSaving(true);
      try {
        const [freshValues, freshContext] = await Promise.all([api.vehicleCVValues(selected.id), workingGroup.reload()]);
        if (generation !== draftGeneration.current) return;
        if (!freshContext) throw new CVWorkingGroupUnavailable();
        stored = freshValues; groups = freshContext.workspaces;
        await reloadSelectedVehicle(selected.id);
      } catch {
        if (generation === draftGeneration.current) onMessage(t("cvScope.unavailable"));
        return;
      } finally { if (generation === draftGeneration.current) setSaving(false); }
    }
    if (generation !== draftGeneration.current) return;
    const change = manualChanges.find(value => value.id === id);
    const current = change && cvDraftCurrent(change, stored, manualChanges);
    if (change?.original && cvDraftCollision(change, stored, manualChanges)) {
      onMessage(t("vehicles.cv.draft.error.occupied"));
      return;
    }
    if (!change || (!current && change.original)) { onMessage(t("vehicles.cv.draft.missing")); return; }
    let next = current ? rebaseCVChange(manualChanges, id, current) : manualChanges.map(entry => entry.id === id ? { ...entry } : entry);
    if (next === manualChanges) { onMessage(t("vehicles.cv.draft.error.occupied")); return; }
    try {
      if (groups) next = next.map(change => change.id === id ? { ...change,
        scope: change.original ? cvDeletionScope(change.original.decoderProfile, groups) : undefined,
        value: change.value ? refreshCVInputScope(change.value, groups) : null } : change);
      setManualChanges(next);
    } catch { onMessage(t("cvScope.unavailable")); return; }
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
      manualDraft: { changes: manualChanges, values: workingCVValues, conflicts: manualConflicts.map(change => change.id),
        bindingChanged: bindingConflicts.length > 0, lastSavedCount: lastManualSave },
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
