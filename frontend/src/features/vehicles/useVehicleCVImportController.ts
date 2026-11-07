import { useCallback, useEffect, useRef, useState, type SetStateAction } from "react";
import { api, ApiError, type Vehicle, type VehicleCVValueInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { buildCVImportPreview, type CVImportPreview } from "./cvImport";
import { readCVImportFile } from "./cvFileImport";
import { CVImportGuardError, guardedCVImportValues, recompareCVImport } from "./cvImportGuard";
import { CVWorkingGroupUnavailable, cvWorkingTarget, refreshCVInputScope, type CVWorkingGroupOptions } from "./cvWorkingGroup";
import { labelWorkingCVPreview, workingCVImportPreview } from "./cvWorkingGroupPreview";

type Options = {
  selected: Vehicle | null; readonly: boolean; saving: boolean;
  setSaving: (saving: boolean) => void; onMessage: (message: string) => void;
  refreshSaved: (vehicleId: string) => Promise<void>;
  workingGroup?: CVWorkingGroupOptions;
};
type State = { preview: CVImportPreview | null; needsReview: boolean; imported: number };

export function useVehicleCVImportController(options: Options) {
  const { t } = useI18n();
  const opts = useRef(options); opts.current = options;
  const [state, setState] = useState<State>({ preview: null, needsReview: false, imported: 0 });
  const current = useRef(state), generation = useRef(0), busy = useRef(false);
  useEffect(() => () => { generation.current++; }, []);
  const importInputRef = useRef<HTMLInputElement | null>(null);
  const set = useCallback((change: (before: State) => State) => {
    current.current = change(current.current); setState(current.current);
  }, []);
  const reset = () => {
    generation.current++;
    if (busy.current) opts.current.setSaving(false);
    busy.current = false;
    set(() => ({ preview: null, needsReview: false, imported: 0 }));
  };
  const setImportPreview = useCallback((action: SetStateAction<CVImportPreview | null>) => {
    generation.current++;
    if (busy.current) opts.current.setSaving(false);
    busy.current = false;
    set(before => {
      let preview = typeof action === "function" ? action(before.preview) : action;
      const { workingGroup, selected } = opts.current;
      if (preview && workingGroup?.groups && preview.workingTarget === undefined) {
        const scoped = workingCVImportPreview(preview.fileName, preview.rows.map(row => row.input), selected?.cvValues || [], workingGroup.groups, workingGroup.target);
        preview = { ...labelWorkingCVPreview(scoped, workingGroup.context, t), vehicleId: preview.vehicleId, rows: scoped.rows.map((row, index) => ({ ...row,
          selected: Boolean(preview?.rows[index]?.selected && row.status !== "invalid" && row.status !== "same") })) };
      }
      return { ...before, preview, needsReview: false };
    });
  }, [set, t]);
  const begin = () => {
    const { selected, readonly, saving, workingGroup } = opts.current;
    if (!selected || readonly || saving || busy.current) return null;
    if (workingGroup && !workingGroup.ready) { opts.current.onMessage(t("cvScope.loadFailed")); return null; }
    const version = ++generation.current, id = selected.id;
    busy.current = true; opts.current.setSaving(true); opts.current.onMessage("");
    const ownerKey = workingGroup?.ownerKey;
    return { id, workingGroup: workingGroup ? { ...workingGroup, groups: workingGroup.groups?.map(group => ({ ...group })), target: workingGroup.target ? { ...workingGroup.target } : null } : undefined,
      active: () => generation.current === version && opts.current.selected?.id === id && opts.current.workingGroup?.ownerKey === ownerKey,
      finish: () => { if (generation.current === version) { busy.current = false; opts.current.setSaving(false); } } };
  };
  const importValues = async (files: FileList | null) => {
    if (!files?.length) return;
    const task = begin(); if (!task) return;
    const [file] = Array.from(files);
    set(() => ({ preview: null, needsReview: false, imported: 0 }));
    try {
      const values = await readCVImportFile(file);
      if (!task.active()) return;
      const preview = task.workingGroup?.groups
        ? labelWorkingCVPreview(workingCVImportPreview(file.name, values, opts.current.selected?.cvValues || [], task.workingGroup.groups, task.workingGroup.target), task.workingGroup.context, t)
        : buildCVImportPreview(file.name, values, opts.current.selected?.cvValues || []);
      if (!preview.rows.some(row => row.status !== "invalid")) throw new Error(t("vehicles.cv.importGuard.noValues"));
      set(before => ({ ...before, preview: { ...preview, vehicleId: task.id } }));
    } catch (error) {
      if (task.active()) opts.current.onMessage(error instanceof Error ? error.message : String(error));
    } finally {
      if (task.active() && importInputRef.current) importInputRef.current.value = "";
      task.finish();
    }
  };
  const applyImportPreview = async () => {
    const snapshot = current.current, vehicle = opts.current.selected, preview = snapshot.preview;
    if (!preview || !vehicle || snapshot.needsReview || (preview.vehicleId && preview.vehicleId !== vehicle.id)) return;
    let values: VehicleCVValueInput[];
    try { values = guardedCVImportValues(preview); }
    catch (error) {
      if (error instanceof CVImportGuardError) opts.current.onMessage(t(`vehicles.cv.importGuard.${error.code}`));
      return;
    }
    const task = begin(); if (!task) return;
    let acknowledged = false;
    try {
      const response = await api.importVehicleCVValues(task.id, values);
      if (!task.active()) return;
      if (response.imported !== values.length) throw new Error(t("vehicles.cv.importGuard.unknown"));
      acknowledged = true;
      set(before => ({ ...before, preview: null, imported: response.imported, needsReview: false }));
      await opts.current.refreshSaved(task.id);
    } catch (error) {
      if (!task.active()) return;
      if (acknowledged) opts.current.onMessage(t("vehicles.cv.importGuard.savedRefreshFailed", { count: values.length }));
      else {
        const stale = error instanceof ApiError && error.status === 409;
        const unknown = !(error instanceof ApiError) || error.status >= 500 || [408, 499].includes(error.status);
        if (stale || unknown) set(before => ({ ...before, needsReview: true }));
        opts.current.onMessage(stale ? t("vehicles.cv.importGuard.stale") : unknown ? t("vehicles.cv.importGuard.unknown")
          : error instanceof Error ? error.message : String(error));
      }
    } finally { task.finish(); }
  };
  const recompareImportPreview = async () => {
    const preview = current.current.preview, vehicle = opts.current.selected;
    if (!preview || !vehicle || (preview.vehicleId && preview.vehicleId !== vehicle.id)) return;
    const task = begin(); if (!task) return;
    try {
      const [values, context] = await Promise.all([api.vehicleCVValues(task.id), task.workingGroup?.reload()]);
      if (!task.active()) return;
      const groups = context?.workspaces;
      if (task.workingGroup && !groups) throw new CVWorkingGroupUnavailable();
      const target = preview.workingTarget && groups?.find(group => group.id === preview.workingTarget?.workspaceId && group.profile === preview.workingTarget.profile);
      if (preview.workingTarget && groups && !target) throw new CVWorkingGroupUnavailable();
      const compared = groups ? { ...preview, ...(target ? { workingTarget: cvWorkingTarget(target) } : {}),
        rows: preview.rows.map(row => ({ ...row, input: refreshCVInputScope(row.input, groups) })) } : preview;
      const refreshed = recompareCVImport(groups ? labelWorkingCVPreview(compared, context, t) : compared, values);
      set(before => ({ ...before, preview: refreshed, needsReview: false }));
    } catch (error) {
      if (task.active()) {
        set(before => ({ ...before, needsReview: true }));
        opts.current.onMessage(error instanceof CVWorkingGroupUnavailable ? t("cvScope.unavailable") : error instanceof Error ? error.message : String(error));
      }
    } finally { task.finish(); }
  };
  const toggleImportRow = (id: string, selected: boolean) => {
    if (busy.current || opts.current.readonly || opts.current.saving) return;
    set(before => ({ ...before, preview: before.preview ? { ...before.preview,
      rows: before.preview.rows.map(row => row.id === id ? { ...row, selected } : row) } : null }));
  };
  const selectImportRows = (mode: "all" | "none" | "empty") => {
    if (busy.current || opts.current.readonly || opts.current.saving) return;
    set(before => ({ ...before, preview: before.preview ? { ...before.preview, rows: before.preview.rows.map(row => ({ ...row,
      selected: row.status !== "invalid" && (mode === "all" || (mode === "empty" && row.status === "new"))
    })) } : null }));
  };
  return { state, refs: { importInputRef }, commands: { reset, importValues, applyImportPreview, recompareImportPreview,
    setImportPreview, toggleImportRow, selectImportRows, discardImportPreview: () => setImportPreview(null) } };
}
