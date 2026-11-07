import { useEffect, useRef, useState } from "react";
import { api, type CreateVehicleRequest, type VehicleSet } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { vehicleSetInputFromForm } from "./VehicleCreateWizard";
import { emptyVehicle } from "./vehicleViewModel";
import { reconcileDraft, retainRefreshConflicts, type DraftRefreshConflict } from "./vehicleDraftRefresh";

function formFromSet(set: VehicleSet): CreateVehicleRequest {
  return { ...emptyVehicle, ...set, inventoryNumber: "" };
}

type Options = {
  setId: string;
  captureScope?: () => () => boolean;
  onUpdated: (set: VehicleSet) => void;
  onCatalogChanged?: (set: VehicleSet) => Promise<void>;
};

export function useVehicleSetEditorController({ setId, captureScope, onUpdated, onCatalogChanged }: Options) {
  const { t } = useI18n();
  const [form, setForm] = useState<CreateVehicleRequest>(emptyVehicle);
  const [vehicleSet, setVehicleSet] = useState<VehicleSet | null>(null);
  const [loading, setLoading] = useState(true), [saving, setSaving] = useState(false);
  const [catalogSaving, setCatalogSaving] = useState(false), [catalogSaved, setCatalogSaved] = useState(false);
  const [error, setError] = useState("");
  const [conflicts, setConflicts] = useState<DraftRefreshConflict[]>([]);
  const mounted = useRef(true), target = useRef(setId), writing = useRef(false), catalogWriting = useRef(false);
  const draft = useRef(form), saved = useRef(vehicleSet), pendingConflicts = useRef(conflicts);
  target.current = setId;
  const capture = () => {
    const id = setId, isCurrent = captureScope?.() ?? (() => true);
    return () => mounted.current && target.current === id && isCurrent();
  };

  useEffect(() => {
    mounted.current = true;
    let active = true;
    setLoading(true); setError(""); setConflicts([]); pendingConflicts.current = [];
    setCatalogSaved(false); setCatalogSaving(false); catalogWriting.current = false;
    saved.current = null;
    api.vehicleSet(setId).then(value => {
      if (!active) return;
      saved.current = value; draft.current = formFromSet(value);
      setVehicleSet(value); setForm(draft.current);
    }).catch((reason: Error) => { if (active) setError(reason.message); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; mounted.current = false; };
  }, [setId]);

  const update = (patch: Partial<CreateVehicleRequest>) => {
    draft.current = { ...draft.current, ...patch }; setForm(draft.current);
  };
  const accept = (value: VehicleSet) => {
    const incoming = formFromSet(value);
    const merged = saved.current ? reconcileDraft(draft.current, formFromSet(saved.current), incoming)
      : { value: incoming, conflicts: [] };
    pendingConflicts.current = retainRefreshConflicts(pendingConflicts.current, merged.value, incoming, merged.conflicts);
    draft.current = merged.value; saved.current = value;
    setForm(merged.value); setVehicleSet(value); setConflicts(pendingConflicts.current);
  };
  const refreshCatalog = async (confirmed: boolean) => {
    const isCurrent = capture();
    if (!isCurrent()) return;
    const value = await api.vehicleSet(setId);
    if (!isCurrent()) return;
    accept(value);
    if (confirmed) setCatalogSaved(true);
    await onCatalogChanged?.(value);
  };
  const resolveConflicts = (choice: "draft" | "saved") => {
    if (choice === "saved") {
      for (const conflict of pendingConflicts.current) draft.current = { ...draft.current, [conflict.field]: conflict.saved };
      setForm(draft.current);
    }
    pendingConflicts.current = []; setConflicts([]);
  };
  const save = async () => {
    if (loading || writing.current || catalogWriting.current || pendingConflicts.current.length) return;
    const isCurrent = capture();
    if (!isCurrent()) return;
    writing.current = true; setSaving(true); setError("");
    try {
      const value = await api.updateVehicleSet(setId, vehicleSetInputFromForm(draft.current));
      if (isCurrent()) { accept(value); onUpdated(value); }
    } catch (reason) {
      if (isCurrent()) setError(reason instanceof Error ? reason.message : t("vehicles.set.updateFailed"));
    } finally {
      if (isCurrent()) { writing.current = false; setSaving(false); }
    }
  };
  const changeCatalogSaving = (value: boolean) => {
    catalogWriting.current = value; setCatalogSaving(value);
  };
  return { form, vehicleSet, loading, saving, catalogSaving, catalogSaved, error, conflicts,
    update, accept, refreshCatalog, resolveConflicts, save, setError, changeCatalogSaving };
}
